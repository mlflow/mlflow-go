package sql

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/tracking/service/query"
	"github.com/mlflow/mlflow-go/pkg/tracking/service/query/parser"
	"github.com/mlflow/mlflow-go/pkg/tracking/store/sql/models"
	"github.com/mlflow/mlflow-go/pkg/utils"
)

func applyLifecycleStagesFilter(runViewType protos.ViewType) []models.LifecycleStage {
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

//nolint:gosec // disable G115
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
