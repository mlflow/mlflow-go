package sql

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/tracking/store/sql/models"
)

// supported expression list.
const (
	LikeExpression           = "LIKE"
	ILikeExpression          = "ILIKE"
	EqualExpression          = "="
	NotEqualExpression       = "!="
	LessExpression           = "<"
	LessOrEqualExpression    = "<="
	GreaterExpression        = ">"
	GreaterOrEqualExpression = ">="
)

//nolint:lll
var (
	filterAnd       = regexp.MustCompile(`(?i)\s+AND\s+`)
	filterCond      = regexp.MustCompile(`^(?:(\w+)\.)?("[^"]+"|` + "`[^`]+`" + `|[\w\.]+)\s+(<|<=|>|>=|=|!=|(?i:I?LIKE)|(?i:(?:NOT )?IN))\s+(\((?:'[^']+'(?:,\s*)?)+\)|"[^"]+"|'[^']+'|[\w\.]+)$`)
	experimentOrder = regexp.MustCompile(`^(?:attr(?:ibutes?)?\.)?(\w+)(?i:\s+(ASC|DESC))?$`)
)

// PageToken.
type PageToken struct {
	Offset int32 `json:"offset"`
}

func applyExperimentsLimitFilter(query *gorm.DB, maxResults int64) (*gorm.DB, int) {
	return query.Limit(int(maxResults) + 1), int(maxResults)
}

func applyExperimentsOffsetFilter(query *gorm.DB, pageToken string) (*gorm.DB, int, *contract.Error) {
	var offset int

	if pageToken != "" {
		var token PageToken
		if err := json.NewDecoder(
			base64.NewDecoder(
				base64.StdEncoding,
				strings.NewReader(pageToken),
			),
		).Decode(&token); err != nil {
			return nil, 0, contract.NewError(
				protos.ErrorCode_INVALID_PARAMETER_VALUE,
				fmt.Sprintf("invalid page_token '%s': %s", pageToken, err),
			)
		}

		offset = int(token.Offset)
	}

	return query.Offset(offset), offset, nil
}

func applyExperimentsLifecycleStagesFilter(query *gorm.DB, runViewType protos.ViewType) *gorm.DB {
	switch runViewType {
	case protos.ViewType_ACTIVE_ONLY:
		query = query.Where(
			"lifecycle_stage IN (?)", []models.LifecycleStage{
				models.LifecycleStageActive,
			},
		)
	case protos.ViewType_DELETED_ONLY:
		query = query.Where(
			"lifecycle_stage IN (?)", []models.LifecycleStage{
				models.LifecycleStageDeleted,
			},
		)
	case protos.ViewType_ALL:
		query = query.Where(
			"lifecycle_stage IN (?)", []models.LifecycleStage{
				models.LifecycleStageActive,
				models.LifecycleStageDeleted,
			},
		)
	}

	return query
}

func applyExperimentsOrderBy(query *gorm.DB, orderBy []string) (*gorm.DB, *contract.Error) {
	expOrder := false

	for _, o := range orderBy {
		parts := experimentOrder.FindStringSubmatch(o)
		if len(parts) == 0 {
			return nil, contract.NewError(
				protos.ErrorCode_INVALID_PARAMETER_VALUE,
				fmt.Sprintf("invalid order_by clause '%s'", o),
			)
		}

		column := parts[1]
		switch column {
		case "experiment_id":
			expOrder = true

			fallthrough
		case "name", "creation_time", "last_update_time":
		default:
			return nil, contract.NewError(
				protos.ErrorCode_INVALID_PARAMETER_VALUE,
				fmt.Sprintf(
					"invalid attribute '%s'. Valid values are ['name', 'experiment_id', 'creation_time', 'last_update_time']",
					column,
				),
			)
		}

		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Name: column},
			Desc:   len(parts) == 3 && strings.ToUpper(parts[2]) == "DESC",
		})
	}

	if len(orderBy) == 0 {
		query = query.Order("experiments.creation_time DESC")
	}

	if !expOrder {
		query = query.Order("experiments.experiment_id ASC")
	}

	return query, nil
}

//nolint:funlen,gocognit,nestif,cyclop,goconst,mnd,forcetypeassert
func applyExperimentsFilter(database, query *gorm.DB, filter string) (*gorm.DB, *contract.Error) {
	if filter != "" {
		for index, f := range filterAnd.Split(filter, -1) {
			parts := filterCond.FindStringSubmatch(f)
			if len(parts) != 5 {
				return nil, contract.NewError(
					protos.ErrorCode_INVALID_PARAMETER_VALUE,
					fmt.Sprintf("malformed filter '%s'", f),
				)
			}

			var value any = parts[4]

			entity, key, comparison := parts[1], strings.Trim(parts[2], "\"`"), parts[3]

			switch entity {
			case "", "attribute", "attributes", "attr":
				switch key {
				case "creation_time", "last_update_time":
					switch comparison {
					case GreaterExpression, GreaterOrEqualExpression, NotEqualExpression,
						EqualExpression, LessExpression, LessOrEqualExpression:
						intValue, err := strconv.Atoi(value.(string))
						if err != nil {
							return nil, contract.NewError(
								protos.ErrorCode_INVALID_PARAMETER_VALUE,
								fmt.Sprintf("invalid numeric value '%s'", value),
							)
						}

						value = intValue
					default:
						return nil, contract.NewError(
							protos.ErrorCode_INVALID_PARAMETER_VALUE,
							fmt.Sprintf(
								"invalid numeric attribute comparison operator '%s'", comparison,
							),
						)
					}
				case "name":
					switch strings.ToUpper(comparison) {
					case NotEqualExpression, EqualExpression, LikeExpression, ILikeExpression:
						if strings.HasPrefix(value.(string), "(") {
							return nil, contract.NewError(
								protos.ErrorCode_INVALID_PARAMETER_VALUE,
								fmt.Sprintf("invalid string value '%s'", value),
							)
						}

						value = strings.Trim(value.(string), `"'`)

						if database.Dialector.Name() == "sqlite" && strings.ToUpper(comparison) == ILikeExpression {
							key = fmt.Sprintf("LOWER(%s)", key)
							comparison = LikeExpression
							value = strings.ToLower(value.(string))
						}
					default:
						return nil, contract.NewError(
							protos.ErrorCode_INVALID_PARAMETER_VALUE,
							fmt.Sprintf(
								"invalid string attribute comparison operator '%s'", comparison,
							),
						)
					}
				default:
					return nil, contract.NewError(
						protos.ErrorCode_INVALID_PARAMETER_VALUE,
						fmt.Sprintf(
							"invalid attribute '%s'. Valid values are ['name', 'creation_time', 'last_update_time']", key,
						),
					)
				}

				query.Where(fmt.Sprintf("%s %s ?", key, comparison), value)
			case "tag", "tags":
				switch strings.ToUpper(comparison) {
				case NotEqualExpression, EqualExpression, LikeExpression, ILikeExpression:
					if strings.HasPrefix(value.(string), "(") {
						return nil, contract.NewError(
							protos.ErrorCode_INVALID_PARAMETER_VALUE,
							fmt.Sprintf("invalid string value '%s'", value),
						)
					}

					value = strings.Trim(value.(string), `"'`)
				default:
					return nil, contract.NewError(
						protos.ErrorCode_INVALID_PARAMETER_VALUE,
						fmt.Sprintf("invalid tag comparison operator '%s'", comparison),
					)
				}

				table := fmt.Sprintf("filter_%d", index)
				where := fmt.Sprintf("value %s ?", comparison)

				if database.Dialector.Name() == "sqlite" && strings.ToUpper(comparison) == ILikeExpression {
					where = "LOWER(value) LIKE ?"
					value = strings.ToLower(value.(string))
				}

				query.Joins(
					fmt.Sprintf("JOIN (?) AS %s ON experiments.experiment_id = %s.experiment_id", table, table),
					database.Select(
						"experiment_id", "value",
					).Where("key = ?", key).Where(where, value).Model(&models.ExperimentTag{}),
				)
			default:
				return nil, contract.NewError(
					protos.ErrorCode_INVALID_PARAMETER_VALUE,
					fmt.Sprintf("invalid entity type '%s'. Valid values are ['tag', 'tags', 'attribute']", entity),
				)
			}
		}
	}

	return query, nil
}

//nolint:gosec // disable G115
func createExperimentsNextPageToken(experiments []models.Experiment, limit, offset int) (string, *contract.Error) {
	var token strings.Builder
	if len(experiments) > limit {
		if err := json.NewEncoder(
			base64.NewEncoder(base64.StdEncoding, &token),
		).Encode(PageToken{
			Offset: int32(offset + limit),
		}); err != nil {
			return "", contract.NewErrorWith(
				protos.ErrorCode_INTERNAL_ERROR,
				"error encoding 'nextPageToken' value",
				err,
			)
		}
	}

	return token.String(), nil
}
