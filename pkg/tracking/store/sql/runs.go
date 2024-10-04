package sql

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/entities"
	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/tracking/service/query"
	"github.com/mlflow/mlflow-go/pkg/tracking/service/query/parser"
	"github.com/mlflow/mlflow-go/pkg/tracking/store/sql/models"
	"github.com/mlflow/mlflow-go/pkg/utils"
)

type PageToken struct {
	Offset int32 `json:"offset"`
}

func checkRunIsActive(transaction *gorm.DB, runID string) *contract.Error {
	var run models.Run

	err := transaction.
		Model(&models.Run{}).
		Where("run_uuid = ?", runID).
		Select("lifecycle_stage").
		First(&run).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return contract.NewError(
				protos.ErrorCode_RESOURCE_DOES_NOT_EXIST,
				fmt.Sprintf("Run with id=%s not found", runID),
			)
		}

		return contract.NewErrorWith(
			protos.ErrorCode_INTERNAL_ERROR,
			fmt.Sprintf(
				"failed to get lifecycle stage for run %q",
				runID,
			),
			err,
		)
	}

	if run.LifecycleStage != models.LifecycleStageActive {
		return contract.NewError(
			protos.ErrorCode_INVALID_PARAMETER_VALUE,
			fmt.Sprintf(
				"The run %s must be in the 'active' state.\n"+
					"Current state is %v.",
				runID,
				run.LifecycleStage,
			),
		)
	}

	return nil
}

func getLifecyleStages(runViewType protos.ViewType) []models.LifecycleStage {
	switch runViewType {
	case protos.ViewType_ACTIVE_ONLY:
		return []models.LifecycleStage{
			models.LifecycleStageActive,
		}
	case protos.ViewType_DELETED_ONLY:
		return []models.LifecycleStage{
			models.LifecycleStageDeleted,
		}
	case protos.ViewType_ALL:
		return []models.LifecycleStage{
			models.LifecycleStageActive,
			models.LifecycleStageDeleted,
		}
	}

	return []models.LifecycleStage{
		models.LifecycleStageActive,
		models.LifecycleStageDeleted,
	}
}

func getOffset(pageToken string) (int, *contract.Error) {
	if pageToken != "" {
		var token PageToken
		if err := json.NewDecoder(
			base64.NewDecoder(
				base64.StdEncoding,
				strings.NewReader(pageToken),
			),
		).Decode(&token); err != nil {
			return 0, contract.NewErrorWith(
				protos.ErrorCode_INVALID_PARAMETER_VALUE,
				fmt.Sprintf("invalid page_token: %q", pageToken),
				err,
			)
		}

		return int(token.Offset), nil
	}

	return 0, nil
}

//nolint:funlen,cyclop,gocognit
func applyFilter(ctx context.Context, database, transaction *gorm.DB, filter string) *contract.Error {
	filterConditions, err := query.ParseFilter(filter)
	if err != nil {
		return contract.NewErrorWith(
			protos.ErrorCode_INVALID_PARAMETER_VALUE,
			"error parsing search filter",
			err,
		)
	}

	utils.GetLoggerFromContext(ctx).Debugf("Filter conditions: %v", filterConditions)

	for index, clause := range filterConditions {
		var kind any

		key := clause.Key
		comparison := strings.ToUpper(clause.Operator.String())
		value := clause.Value

		switch clause.Identifier {
		case parser.Metric:
			kind = &models.LatestMetric{}
		case parser.Parameter:
			kind = &models.Param{}
		case parser.Tag:
			kind = &models.Tag{}
		case parser.Dataset:
			kind = &models.Dataset{}
		case parser.Attribute:
			kind = nil
		}

		// Treat "attributes.run_name == <value>" as "tags.`mlflow.runName` == <value>".
		// The name column in the runs table is empty for runs logged in MLflow <= 1.29.0.
		if key == "run_name" {
			kind = &models.Tag{}
			key = utils.TagRunName
		}

		isSqliteAndILike := database.Dialector.Name() == "sqlite" && comparison == "ILIKE"
		table := fmt.Sprintf("filter_%d", index)

		switch {
		case kind == nil:
			if isSqliteAndILike {
				key = fmt.Sprintf("LOWER(runs.%s)", key)
				comparison = "LIKE"

				if str, ok := value.(string); ok {
					value = strings.ToLower(str)
				}

				transaction.Where(fmt.Sprintf("%s %s ?", key, comparison), value)
			} else {
				transaction.Where(fmt.Sprintf("runs.%s %s ?", key, comparison), value)
			}
		case clause.Identifier == parser.Dataset && key == "context":
			// SELECT *
			// FROM runs
			// JOIN (
			//   SELECT inputs.destination_id AS run_uuid
			//   FROM inputs
			//   JOIN input_tags
			//   ON inputs.input_uuid = input_tags.input_uuid
			//   AND input_tags.name = 'mlflow.data.context'
			//   AND input_tags.value %s ?
			//   WHERE inputs.destination_type = 'RUN'
			// ) AS filter_0
			// ON runs.run_uuid = filter_0.run_uuid
			valueColumn := "input_tags.value "
			if isSqliteAndILike {
				valueColumn = "LOWER(input_tags.value) "
				comparison = "LIKE"

				if str, ok := value.(string); ok {
					value = strings.ToLower(str)
				}
			}

			transaction.Joins(
				fmt.Sprintf("JOIN (?) AS %s ON runs.run_uuid = %s.run_uuid", table, table),
				database.Select("inputs.destination_id AS run_uuid").
					Joins(
						"JOIN input_tags ON inputs.input_uuid = input_tags.input_uuid"+
							" AND input_tags.name = 'mlflow.data.context'"+
							" AND "+valueColumn+comparison+" ?",
						value,
					).
					Where("inputs.destination_type = 'RUN'").
					Model(&models.Input{}),
			)
		case clause.Identifier == parser.Dataset:
			// add join with datasets
			// JOIN (
			// 		SELECT "experiment_id", key
			//		FROM datasests d
			// 		JOIN inputs ON inputs.source_id = datasets.dataset_uuid
			//		WHERE key comparison value
			// ) AS filter_0 ON runs.experiment_id = dataset.experiment_id
			//
			// columns: name, digest, context
			where := key + " " + comparison + " ?"
			if isSqliteAndILike {
				where = "LOWER(" + key + ") LIKE ?"

				if str, ok := value.(string); ok {
					value = strings.ToLower(str)
				}
			}

			transaction.Joins(
				fmt.Sprintf("JOIN (?) AS %s ON runs.run_uuid = %s.destination_id", table, table),
				database.Model(kind).
					Joins("JOIN inputs ON inputs.source_id = datasets.dataset_uuid").
					Where(where, value).
					Select("destination_id", key),
			)
		default:
			where := fmt.Sprintf("value %s ?", comparison)
			if isSqliteAndILike {
				where = "LOWER(value) LIKE ?"

				if str, ok := value.(string); ok {
					value = strings.ToLower(str)
				}
			}

			transaction.Joins(
				fmt.Sprintf("JOIN (?) AS %s ON runs.run_uuid = %s.run_uuid", table, table),
				database.Select("run_uuid", "value").Where("key = ?", key).Where(where, value).Model(kind),
			)
		}
	}

	return nil
}

type orderByExpr struct {
	identifier *string
	key        string
	order      *string
}

var ErrInvalidOrderClauseInput = errors.New("input string is empty or only contains quote characters")

const (
	identifierAndKeyLength = 2
	startTime              = "start_time"
	name                   = "name"
	attribute              = "attribute"
	metric                 = "metric"
)

func orderByKeyAlias(input string) string {
	switch input {
	case "created", "Created":
		return startTime
	case "run_name", "run name", "Run name", "Run Name":
		return name
	case "run_id":
		return "run_uuid"
	default:
		return input
	}
}

func handleInsideQuote(
	char, quoteChar rune, insideQuote bool, current strings.Builder, result []string,
) (bool, strings.Builder, []string) {
	if char == quoteChar {
		insideQuote = false

		result = append(result, current.String())
		current.Reset()
	} else {
		current.WriteRune(char)
	}

	return insideQuote, current, result
}

func handleOutsideQuote(
	char rune, insideQuote bool, quoteChar rune, current strings.Builder, result []string,
) (bool, rune, strings.Builder, []string) {
	switch char {
	case ' ':
		if current.Len() > 0 {
			result = append(result, current.String())
			current.Reset()
		}
	case '"', '\'', '`':
		insideQuote = true
		quoteChar = char
	default:
		current.WriteRune(char)
	}

	return insideQuote, quoteChar, current, result
}

// Process an order by input string to split the string into the separate parts.
// We can't simply split by space, because the column name could be wrapped in quotes, e.g. "Run name" ASC.
func splitOrderByClauseWithQuotes(input string) []string {
	input = strings.ToLower(strings.Trim(input, " "))

	var result []string

	var current strings.Builder

	var insideQuote bool

	var quoteChar rune

	// Process char per char, split items on spaces unless inside a quoted entry.
	for _, char := range input {
		if insideQuote {
			insideQuote, current, result = handleInsideQuote(char, quoteChar, insideQuote, current, result)
		} else {
			insideQuote, quoteChar, current, result = handleOutsideQuote(char, insideQuote, quoteChar, current, result)
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

func translateIdentifierAlias(identifier string) string {
	switch strings.ToLower(identifier) {
	case "metrics":
		return metric
	case "parameters", "param", "params":
		return "parameter"
	case "tags":
		return "tag"
	case "attr", "attributes", "run":
		return attribute
	case "datasets":
		return "dataset"
	default:
		return identifier
	}
}

func processOrderByClause(input string) (orderByExpr, error) {
	parts := splitOrderByClauseWithQuotes(input)

	if len(parts) == 0 {
		return orderByExpr{}, ErrInvalidOrderClauseInput
	}

	var expr orderByExpr

	identifierKey := strings.Split(parts[0], ".")

	if len(identifierKey) == identifierAndKeyLength {
		expr.identifier = utils.PtrTo(translateIdentifierAlias(identifierKey[0]))
		expr.key = orderByKeyAlias(identifierKey[1])
	} else if len(identifierKey) == 1 {
		expr.key = orderByKeyAlias(identifierKey[0])
	}

	if len(parts) > 1 {
		expr.order = utils.PtrTo(strings.ToUpper(parts[1]))
	}

	return expr, nil
}

//nolint:funlen, cyclop, gocognit
func applyOrderBy(ctx context.Context, database, transaction *gorm.DB, orderBy []string) *contract.Error {
	startTimeOrder := false
	columnSelection := "runs.*"

	for index, orderByClause := range orderBy {
		orderByExpr, err := processOrderByClause(orderByClause)
		if err != nil {
			return contract.NewError(
				protos.ErrorCode_INVALID_PARAMETER_VALUE,
				fmt.Sprintf(
					"invalid order_by clause %q.",
					orderByClause,
				),
			)
		}

		logger := utils.GetLoggerFromContext(ctx)
		logger.
			Debugf(
				"OrderByExpr: identifier: %v, key: %v, order: %v",
				utils.DumpStringPointer(orderByExpr.identifier),
				orderByExpr.key,
				utils.DumpStringPointer(orderByExpr.order),
			)

		var kind any

		if orderByExpr.identifier == nil && orderByExpr.key == "start_time" {
			startTimeOrder = true
		} else if orderByExpr.identifier != nil {
			switch {
			case *orderByExpr.identifier == attribute && orderByExpr.key == "start_time":
				startTimeOrder = true
			case *orderByExpr.identifier == metric:
				kind = &models.LatestMetric{}
			case *orderByExpr.identifier == "parameter":
				kind = &models.Param{}
			case *orderByExpr.identifier == "tag":
				kind = &models.Tag{}
			}
		}

		table := fmt.Sprintf("order_%d", index)

		if kind != nil {
			columnsInJoin := []string{"run_uuid", "value"}
			if *orderByExpr.identifier == metric {
				columnsInJoin = append(columnsInJoin, "is_nan")
			}

			transaction.Joins(
				fmt.Sprintf("LEFT OUTER JOIN (?) AS %s ON runs.run_uuid = %s.run_uuid", table, table),
				database.Select(columnsInJoin).Where("key = ?", orderByExpr.key).Model(kind),
			)

			orderByExpr.key = table + ".value"
		}

		desc := false
		if orderByExpr.order != nil {
			desc = *orderByExpr.order == "DESC"
		}

		nullableColumnAlias := fmt.Sprintf("order_null_%d", index)

		if orderByExpr.identifier == nil || *orderByExpr.identifier != metric {
			var originalColumn string

			switch {
			case orderByExpr.identifier != nil && *orderByExpr.identifier == "attribute":
				originalColumn = "runs." + orderByExpr.key
			case orderByExpr.identifier != nil:
				originalColumn = table + ".value"
			default:
				originalColumn = orderByExpr.key
			}

			columnSelection = fmt.Sprintf(
				"%s, (CASE WHEN (%s IS NULL) THEN 1 ELSE 0 END) AS %s",
				columnSelection,
				originalColumn,
				nullableColumnAlias,
			)

			transaction.Order(nullableColumnAlias)
		}

		// the metric table has the is_nan column
		if orderByExpr.identifier != nil && *orderByExpr.identifier == metric {
			trueColumnValue := "true"
			if database.Dialector.Name() == "sqlite" {
				trueColumnValue = "1"
			}

			columnSelection = fmt.Sprintf(
				"%s, (CASE WHEN (%s.is_nan = %s) THEN 1 WHEN (%s.value IS NULL) THEN 2 ELSE 0 END) AS %s",
				columnSelection,
				table,
				trueColumnValue,
				table,
				nullableColumnAlias,
			)

			transaction.Order(nullableColumnAlias)
		}

		transaction.Order(clause.OrderByColumn{
			Column: clause.Column{
				Name: orderByExpr.key,
			},
			Desc: desc,
		})
	}

	if !startTimeOrder {
		transaction.Order("runs.start_time DESC")
	}

	transaction.Order("runs.run_uuid")

	// mlflow orders all nullable columns to have null last.
	// For each order by clause, an additional dynamic order clause was added.
	// We need to include these columns in the select clause.
	transaction.Select(columnSelection)

	return nil
}

func mkNextPageToken(runLength, maxResults, offset int) (string, *contract.Error) {
	var nextPageToken string

	if runLength == maxResults {
		var token strings.Builder
		if err := json.NewEncoder(
			base64.NewEncoder(base64.StdEncoding, &token),
		).Encode(PageToken{
			Offset: int32(offset + maxResults),
		}); err != nil {
			return "", contract.NewErrorWith(
				protos.ErrorCode_INTERNAL_ERROR,
				"error encoding 'nextPageToken' value",
				err,
			)
		}

		nextPageToken = token.String()
	}

	return nextPageToken, nil
}

//nolint:funlen
func (s TrackingSQLStore) SearchRuns(
	ctx context.Context,
	experimentIDs []string, filter string,
	runViewType protos.ViewType, maxResults int, orderBy []string, pageToken string,
) ([]*entities.Run, string, *contract.Error) {
	// ViewType
	lifecyleStages := getLifecyleStages(runViewType)
	transaction := s.db.WithContext(ctx).Where(
		"runs.experiment_id IN ?", experimentIDs,
	).Where(
		"runs.lifecycle_stage IN ?", lifecyleStages,
	)

	// MaxResults
	transaction.Limit(maxResults)

	// PageToken
	offset, contractError := getOffset(pageToken)
	if contractError != nil {
		return nil, "", contractError
	}

	transaction.Offset(offset)

	// Filter
	contractError = applyFilter(ctx, s.db, transaction, filter)
	if contractError != nil {
		return nil, "", contractError
	}

	// OrderBy
	contractError = applyOrderBy(ctx, s.db, transaction, orderBy)
	if contractError != nil {
		return nil, "", contractError
	}

	// Actual query
	var runs []models.Run

	transaction.Preload("LatestMetrics").Preload("Params").Preload("Tags").
		Preload("Inputs", "inputs.destination_type = 'RUN'").
		Preload("Inputs.Dataset").Preload("Inputs.Tags").Find(&runs)

	if transaction.Error != nil {
		return nil, "", contract.NewErrorWith(
			protos.ErrorCode_INTERNAL_ERROR,
			"Failed to query search runs",
			transaction.Error,
		)
	}

	entityRuns := make([]*entities.Run, len(runs))
	for i, run := range runs {
		entityRuns[i] = run.ToEntity()
	}

	nextPageToken, contractError := mkNextPageToken(len(runs), maxResults, offset)
	if contractError != nil {
		return nil, "", contractError
	}

	return entityRuns, nextPageToken, nil
}

const RunIDMaxLength = 32

const (
	ArtifactFolderName  = "artifacts"
	RunNameIntegerScale = 3
	RunNameMaxLength    = 20
)

func getRunNameFromTags(tags []models.Tag) string {
	for _, tag := range tags {
		if tag.Key == utils.TagRunName {
			return tag.Value
		}
	}

	return ""
}

func ensureRunName(runModel *models.Run) *contract.Error {
	runNameFromTags := getRunNameFromTags(runModel.Tags)

	switch {
	// run_name and name in tags differ
	case utils.IsNotNilOrEmptyString(&runModel.Name) && runNameFromTags != "" && runModel.Name != runNameFromTags:
		return contract.NewError(
			protos.ErrorCode_INVALID_PARAMETER_VALUE,
			fmt.Sprintf(
				"Both 'run_name' argument and 'mlflow.runName' tag are specified, but with "+
					"different values (run_name='%s', mlflow.runName='%s').",
				runModel.Name,
				runNameFromTags,
			),
		)
	// no name was provided, generate a random name
	case utils.IsNilOrEmptyString(&runModel.Name) && runNameFromTags == "":
		randomName, err := utils.GenerateRandomName()
		if err != nil {
			return contract.NewErrorWith(
				protos.ErrorCode_INTERNAL_ERROR,
				"failed to generate random run name",
				err,
			)
		}

		runModel.Name = randomName
	// use name from tags
	case utils.IsNilOrEmptyString(&runModel.Name) && runNameFromTags != "":
		runModel.Name = runNameFromTags
	}

	if runNameFromTags == "" {
		runModel.Tags = append(runModel.Tags, models.Tag{
			Key:   utils.TagRunName,
			Value: runModel.Name,
		})
	}

	return nil
}

func (s TrackingSQLStore) GetRun(ctx context.Context, runID string) (*entities.Run, *contract.Error) {
	var run models.Run
	if err := s.db.WithContext(ctx).Where(
		"run_uuid = ?", runID,
	).Preload(
		"Tags",
	).Preload(
		"Params",
	).Preload(
		"Inputs.Tags",
	).Preload(
		"LatestMetrics",
	).Preload(
		"Inputs.Dataset",
	).First(&run).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, contract.NewError(
				protos.ErrorCode_RESOURCE_DOES_NOT_EXIST,
				fmt.Sprintf("Run with id=%s not found", runID),
			)
		}

		return nil, contract.NewErrorWith(protos.ErrorCode_INTERNAL_ERROR, "failed to get run", err)
	}

	return run.ToEntity(), nil
}

//nolint:funlen
func (s TrackingSQLStore) CreateRun(
	ctx context.Context,
	experimentID, userID string,
	startTime int64,
	tags []*entities.RunTag,
	runName string,
) (*entities.Run, *contract.Error) {
	experiment, err := s.GetExperiment(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	if models.LifecycleStage(experiment.LifecycleStage) != models.LifecycleStageActive {
		return nil, contract.NewError(
			protos.ErrorCode_INVALID_PARAMETER_VALUE,
			fmt.Sprintf(
				"The experiment %q must be in the 'active' state.\n"+
					"Current state is %q.",
				experiment.ExperimentID,
				experiment.LifecycleStage,
			),
		)
	}

	runModel := &models.Run{
		ID:             utils.NewUUID(),
		Name:           runName,
		ExperimentID:   utils.ConvertStringPointerToInt32Pointer(&experimentID),
		StartTime:      startTime,
		UserID:         userID,
		Tags:           make([]models.Tag, 0, len(tags)),
		LifecycleStage: models.LifecycleStageActive,
		Status:         models.RunStatusRunning,
		SourceType:     models.SourceTypeUnknown,
	}

	for _, tag := range tags {
		runModel.Tags = append(runModel.Tags, models.NewTagFromEntity(runModel.ID, tag))
	}

	artifactLocation, appendErr := utils.AppendToURIPath(
		experiment.ArtifactLocation,
		runModel.ID,
		ArtifactFolderName,
	)
	if appendErr != nil {
		return nil, contract.NewError(
			protos.ErrorCode_INTERNAL_ERROR,
			"failed to append run ID to experiment artifact location",
		)
	}

	runModel.ArtifactURI = artifactLocation

	errRunName := ensureRunName(runModel)
	if errRunName != nil {
		return nil, errRunName
	}

	if err := s.db.Create(&runModel).Error; err != nil {
		return nil, contract.NewErrorWith(
			protos.ErrorCode_INTERNAL_ERROR,
			fmt.Sprintf(
				"failed to create run for experiment_id %q",
				experiment.ExperimentID,
			),
			err,
		)
	}

	return runModel.ToEntity(), nil
}

func (s TrackingSQLStore) UpdateRun(
	ctx context.Context,
	runID string,
	runStatus string,
	endTime *int64,
	runName string,
) *contract.Error {
	runTag, err := s.GetRunTag(ctx, runID, utils.TagRunName)
	if err != nil {
		return err
	}

	tags := make([]models.Tag, 0, 1)
	if runTag == nil {
		tags = append(tags, models.Tag{
			RunID: runID,
			Key:   utils.TagRunName,
			Value: runName,
		})
	}

	var endTimeValue sql.NullInt64
	if endTime == nil {
		endTimeValue = sql.NullInt64{}
	} else {
		endTimeValue = sql.NullInt64{Int64: *endTime, Valid: true}
	}

	if err := s.db.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		if err := transaction.Model(&models.Run{}).
			Where("run_uuid = ?", runID).
			Updates(&models.Run{
				Name:    runName,
				Status:  models.RunStatus(runStatus),
				EndTime: endTimeValue,
			}).Error; err != nil {
			return err
		}

		if len(tags) > 0 {
			if err := transaction.Clauses(clause.OnConflict{
				UpdateAll: true,
			}).CreateInBatches(tags, tagsBatchSize).Error; err != nil {
				return fmt.Errorf("failed to create tags for run %q: %w", runID, err)
			}
		}

		return nil
	}); err != nil {
		return contract.NewErrorWith(protos.ErrorCode_INTERNAL_ERROR, "failed to update run", err)
	}

	return nil
}

func (s TrackingSQLStore) DeleteRun(ctx context.Context, runID string) *contract.Error {
	run, err := s.GetRun(ctx, runID)
	if err != nil {
		return err
	}

	if err := s.db.WithContext(ctx).Model(&models.Run{}).
		Where("run_uuid = ?", run.Info.RunID).
		Updates(&models.Run{
			DeletedTime:    sql.NullInt64{Valid: true, Int64: time.Now().UnixMilli()},
			LifecycleStage: models.LifecycleStageDeleted,
		}).Error; err != nil {
		return contract.NewErrorWith(protos.ErrorCode_INTERNAL_ERROR, "failed to delete run", err)
	}

	return nil
}

func (s TrackingSQLStore) RestoreRun(ctx context.Context, runID string) *contract.Error {
	run, err := s.GetRun(ctx, runID)
	if err != nil {
		return err
	}

	if err := s.db.WithContext(ctx).Model(&models.Run{}).
		Where("run_uuid = ?", run.Info.RunID).
		// Force GORM to update fields with zero values by selecting them.
		Select("DeletedTime", "LifecycleStage").
		Updates(&models.Run{
			DeletedTime:    sql.NullInt64{},
			LifecycleStage: models.LifecycleStageActive,
		}).Error; err != nil {
		return contract.NewErrorWith(protos.ErrorCode_INTERNAL_ERROR, "failed to restore run", err)
	}

	return nil
}

func (s TrackingSQLStore) LogBatch(
	ctx context.Context, runID string, metrics []*entities.Metric, params []*entities.Param, tags []*entities.RunTag,
) *contract.Error {
	err := s.db.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		contractError := checkRunIsActive(transaction, runID)
		if contractError != nil {
			return contractError
		}

		err := s.setTagsWithTransaction(transaction, runID, tags)
		if err != nil {
			return fmt.Errorf("error setting tags for run_id %q: %w", runID, err)
		}

		contractError = s.logParamsWithTransaction(transaction, runID, params)
		if contractError != nil {
			return contractError
		}

		contractError = s.logMetricsWithTransaction(transaction, runID, metrics)
		if contractError != nil {
			return contractError
		}

		return nil
	})
	if err != nil {
		var contractError *contract.Error
		if errors.As(err, &contractError) {
			return contractError
		}

		return contract.NewErrorWith(
			protos.ErrorCode_INTERNAL_ERROR,
			fmt.Sprintf("log batch transaction failed for %q", runID),
			err,
		)
	}

	return nil
}
