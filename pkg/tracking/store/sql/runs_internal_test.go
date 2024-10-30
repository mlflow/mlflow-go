//nolint:ireturn
package sql

import (
	"context"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/iancoleman/strcase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"

	"github.com/mlflow/mlflow-go/pkg/tracking/store/sql/models"
	"github.com/mlflow/mlflow-go/pkg/utils"
)

type testData struct {
	name         string
	query        string
	orderBy      []string
	expectedSQL  map[string]string
	expectedVars []any
}

var whitespaceRegex = regexp.MustCompile(`\s` + "|`")

func removeWhitespace(s string) string {
	return whitespaceRegex.ReplaceAllString(s, "")
}

var tests = []testData{
	{
		name:  "SimpleMetricQuery",
		query: "metrics.accuracy > 0.72",
		expectedSQL: map[string]string{
			"postgres": `
	SELECT "run_uuid" FROM "runs"
	JOIN (SELECT "run_uuid","value" FROM "latest_metrics" WHERE key = $1 AND value > $2)
	AS filter_0
	ON runs.run_uuid = filter_0.run_uuid
	ORDER BY runs.start_time DESC,runs.run_uuid`,
			"sqlite": `
	SELECT run_uuid FROM runs
	JOIN (SELECT run_uuid,value FROM latest_metrics WHERE key = ? AND value > ?)
	AS filter_0 ON runs.run_uuid = filter_0.run_uuid
	ORDER BY runs.start_time DESC,runs.run_uuid`,
			"sqlserver": `
	SELECT "run_uuid" FROM "runs"
	JOIN (SELECT "run_uuid","value" FROM "latest_metrics" WHERE key = @p1 AND value > @p2)
	AS filter_0 ON runs.run_uuid = filter_0.run_uuid
	ORDER BY runs.start_time DESC,runs.run_uuid`,
			"mysql": `
	SELECT run_uuid FROM runs
	JOIN (SELECT run_uuid,value FROM latest_metrics WHERE key = ? AND value > ?)
	AS filter_0 ON runs.run_uuid = filter_0.run_uuid
	ORDER BY runs.start_time DESC,runs.run_uuid`,
		},
		expectedVars: []any{"accuracy", 0.72},
	},
	{
		name:  "SimpleMetricAndParamQuery",
		query: "metrics.accuracy > 0.72 AND params.batch_size = '2'",
		expectedSQL: map[string]string{
			"postgres": `
	SELECT "run_uuid" FROM "runs"
	JOIN (SELECT "run_uuid","value" FROM "latest_metrics" WHERE key = $1 AND value > $2)
	AS filter_0 ON runs.run_uuid = filter_0.run_uuid
	JOIN (SELECT "run_uuid","value" FROM "params" WHERE key = $3 AND value = $4)
	AS filter_1 ON runs.run_uuid = filter_1.run_uuid
	ORDER BY runs.start_time DESC,runs.run_uuid`,
			"sqlite": `
	SELECT run_uuid FROM runs
	JOIN (SELECT run_uuid,value FROM latest_metrics WHERE key = ? AND value > ?)
	AS filter_0 ON runs.run_uuid = filter_0.run_uuid
	JOIN (SELECT run_uuid,value FROM params WHERE key = ? AND value = ?)
	AS filter_1 ON runs.run_uuid = filter_1.run_uuid
	ORDER BY runs.start_time DESC,runs.run_uuid`,
		},
		expectedVars: []any{"accuracy", 0.72, "batch_size", "2"},
	},
	{
		name:  "TagQuery",
		query: "tags.environment = 'notebook' AND tags.task ILIKE 'classif%'",
		expectedSQL: map[string]string{
			"postgres": `
	SELECT "run_uuid" FROM "runs"
	JOIN (SELECT "run_uuid","value" FROM "tags" WHERE key = $1 AND value = $2)
	AS filter_0 ON runs.run_uuid = filter_0.run_uuid
	JOIN (SELECT "run_uuid","value" FROM "tags" WHERE key = $3 AND value ILIKE $4)
	AS filter_1 ON runs.run_uuid = filter_1.run_uuid
	ORDER BY runs.start_time DESC,runs.run_uuid`,
			"sqlite": `
	SELECT run_uuid FROM runs
	JOIN (SELECT run_uuid,value FROM tags WHERE key = ? AND value = ?)
	AS filter_0 ON runs.run_uuid = filter_0.run_uuid
	JOIN (SELECT run_uuid,value FROM tags WHERE key = ? AND LOWER(value) LIKE ?)
	AS filter_1 ON runs.run_uuid = filter_1.run_uuid
	ORDER BY runs.start_time DESC,runs.run_uuid`,
		},
		expectedVars: []any{"environment", "notebook", "task", "classif%"},
	},
	{
		name:  "DatasestsInQuery",
		query: "datasets.digest IN ('s8ds293b', 'jks834s2')",
		expectedSQL: map[string]string{
			"postgres": `
	SELECT "run_uuid" FROM "runs"
	JOIN (
		SELECT destination_id,"digest"
		FROM "datasets" JOIN inputs ON inputs.source_id = datasets.dataset_uuid
		WHERE digest IN ($1,$2)
	)
	AS filter_0 ON runs.run_uuid = filter_0.destination_id
	ORDER BY runs.start_time DESC,runs.run_uuid`,
			"sqlite": `
	SELECT run_uuid FROM runs
	JOIN (
		SELECT destination_id,digest
		FROM datasets JOIN inputs
		ON inputs.source_id = datasets.dataset_uuid
		WHERE digest IN (?,?)
	)
	AS filter_0
	ON runs.run_uuid = filter_0.destination_id
	ORDER BY runs.start_time DESC,runs.run_uuid`,
		},
		expectedVars: []any{"s8ds293b", "jks834s2"},
	},
	{
		name:  "AttributesQuery",
		query: "attributes.run_id = 'a1b2c3d4'",
		expectedSQL: map[string]string{
			"postgres": `
	SELECT "run_uuid" FROM "runs"
	WHERE runs.run_uuid = $1
	ORDER BY runs.start_time DESC,runs.run_uuid
		`,
			"sqlite": `SELECT run_uuid FROM runs WHERE runs.run_uuid = ? ORDER BY runs.start_time DESC,runs.run_uuid`,
		},
		expectedVars: []any{"a1b2c3d4"},
	},
	{
		name:  "Run_nameQuery",
		query: "attributes.run_name = 'my-run'",
		expectedSQL: map[string]string{
			"postgres": `
	SELECT "run_uuid" FROM "runs"
	JOIN (SELECT "run_uuid","value" FROM "tags" WHERE key = $1 AND value = $2)
	AS filter_0 ON runs.run_uuid = filter_0.run_uuid
	ORDER BY runs.start_time DESC,runs.run_uuid`,
			"sqlite": `
	SELECT run_uuid FROM runs
	JOIN (SELECT run_uuid,value FROM tags WHERE key = ? AND value = ?)
	AS filter_0 ON runs.run_uuid = filter_0.run_uuid
	ORDER BY runs.start_time DESC,runs.run_uuid`,
		},
		expectedVars: []any{"mlflow.runName", "my-run"},
	},
	{
		name:  "DatasetsContextQuery",
		query: "datasets.context = 'train'",
		expectedSQL: map[string]string{
			"postgres": `
	SELECT "run_uuid" FROM "runs"
	JOIN (
		SELECT inputs.destination_id AS run_uuid
		FROM "inputs"
		JOIN input_tags
		ON inputs.input_uuid = input_tags.input_uuid
		AND input_tags.name = 'mlflow.data.context'
		AND input_tags.value = $1
		WHERE inputs.destination_type = 'RUN'
	) AS filter_0 ON runs.run_uuid = filter_0.run_uuid
	ORDER BY runs.start_time DESC,runs.run_uuid`,
			"sqlite": `
	SELECT run_uuid FROM runs
	JOIN (
		SELECT inputs.destination_id AS run_uuid
		FROM inputs
		JOIN input_tags ON inputs.input_uuid = input_tags.input_uuid
		AND input_tags.name = 'mlflow.data.context'
		AND input_tags.value = ? WHERE inputs.destination_type = 'RUN'
	) AS filter_0 ON runs.run_uuid = filter_0.run_uuid
	ORDER BY runs.start_time DESC,runs.run_uuid`,
		},
		expectedVars: []any{"train"},
	},
	{
		name:  "Run_nameQuery",
		query: "attributes.run_name ILIKE 'my-run%'",
		expectedSQL: map[string]string{
			"postgres": `
	SELECT "run_uuid" FROM "runs"
	JOIN (SELECT "run_uuid","value" FROM "tags" WHERE key = $1 AND value ILIKE $2)
	AS filter_0 ON runs.run_uuid = filter_0.run_uuid
	ORDER BY runs.start_time DESC,runs.run_uuid`,
			"sqlite": `
	SELECT run_uuid FROM runs
	JOIN (SELECT run_uuid, value FROM tags WHERE key = ? AND LOWER(value) LIKE ?)
	AS filter_0 ON runs.run_uuid = filter_0.run_uuid
	ORDER BY runs.start_time DESC,runs.run_uuid`,
		},
		expectedVars: []any{"mlflow.runName", "my-run%"},
	},
	{
		name:  "DatasetsContextQuery",
		query: "datasets.context ILIKE '%train'",
		expectedSQL: map[string]string{
			"postgres": `
	SELECT "run_uuid" FROM "runs"
	JOIN (
		SELECT inputs.destination_id AS run_uuid FROM "inputs"
		JOIN input_tags ON inputs.input_uuid = input_tags.input_uuid
		AND input_tags.name = 'mlflow.data.context'
		AND input_tags.value ILIKE $1 WHERE inputs.destination_type = 'RUN'
	) AS filter_0 ON runs.run_uuid = filter_0.run_uuid
	ORDER BY runs.start_time DESC,runs.run_uuid`,
			"sqlite": `
	SELECT run_uuid FROM runs
	JOIN (
		SELECT inputs.destination_id AS run_uuid FROM inputs
		JOIN input_tags ON inputs.input_uuid = input_tags.input_uuid
		AND input_tags.name = 'mlflow.data.context'
		AND LOWER(input_tags.value) LIKE ? WHERE inputs.destination_type = 'RUN')
	AS filter_0 ON runs.run_uuid = filter_0.run_uuid
	ORDER BY runs.start_time DESC,runs.run_uuid
		`,
		},
		expectedVars: []any{"%train"},
	},
	{
		name:  "DatasestsDigest",
		query: "datasets.digest ILIKE '%s'",
		expectedSQL: map[string]string{
			"postgres": `
	SELECT "run_uuid" FROM "runs"
	JOIN (
		SELECT destination_id,"digest"
		FROM "datasets"
		JOIN inputs ON inputs.source_id = datasets.dataset_uuid
		WHERE digest ILIKE $1
	)
	AS filter_0 ON runs.run_uuid = filter_0.destination_id
	ORDER BY runs.start_time DESC,runs.run_uuid`,
			"sqlite": `
	SELECT run_uuid FROM runs
	JOIN (
		SELECT destination_id,digest
		FROM datasets
		JOIN inputs ON inputs.source_id = datasets.dataset_uuid
		WHERE LOWER(digest) LIKE ?)
	AS filter_0 ON runs.run_uuid = filter_0.destination_id
	ORDER BY runs.start_time DESC,runs.run_uuid`,
		},
		expectedVars: []any{"%s"},
	},
	{
		name:  "ParamQuery",
		query: "metrics.accuracy > 0.72 AND params.batch_size ILIKE '%a'",
		expectedSQL: map[string]string{
			"postgres": `
	SELECT "run_uuid" FROM "runs"
	JOIN (SELECT "run_uuid","value" FROM "latest_metrics" WHERE key = $1 AND value > $2)
	AS filter_0 ON runs.run_uuid = filter_0.run_uuid
	JOIN (SELECT "run_uuid","value" FROM "params" WHERE key = $3 AND value ILIKE $4)
	AS filter_1 ON runs.run_uuid = filter_1.run_uuid
	ORDER BY runs.start_time DESC,runs.run_uuid`,
			"sqlite": `
	SELECT run_uuid FROM runs
	JOIN (SELECT run_uuid, value FROM latest_metrics WHERE key = ? AND value > ?)
	AS filter_0 ON runs.run_uuid = filter_0.run_uuid
	JOIN (SELECT run_uuid,value FROM params WHERE key = ? AND LOWER(value) LIKE ?)
	AS filter_1 ON runs.run_uuid = filter_1.run_uuid
	ORDER BY runs.start_time DESC,runs.run_uuid
		`,
		},
		expectedVars: []any{"accuracy", 0.72, "batch_size", "%a"},
	},
	{
		name:    "OrderByStartTimeASC",
		query:   "",
		orderBy: []string{"start_time ASC"},
		expectedSQL: map[string]string{
			"postgres": `SELECT "run_uuid" FROM "runs" ORDER BY order_null_0, "start_time",runs.run_uuid`,
		},
		expectedVars: []any{},
	},
	{
		name:  "OrderByStatusDesc",
		query: "",
		expectedSQL: map[string]string{
			"postgres": `SELECT "run_uuid" FROM "runs" ORDER BY order_null_0, "status" DESC,runs.start_time DESC,runs.run_uuid`,
		},
		orderBy:      []string{"status DESC"},
		expectedVars: []any{},
	},
	{
		name:  "OrderByRunNameSnakeCase",
		query: "",
		expectedSQL: map[string]string{
			"postgres": `SELECT "run_uuid" FROM "runs" ORDER BY order_null_0, "name",runs.start_time DESC,runs.run_uuid`,
		},
		orderBy:      []string{"run_name"},
		expectedVars: []any{},
	},
	{
		name:  "OrderByRunNameLowerName",
		query: "",
		expectedSQL: map[string]string{
			"postgres": `SELECT "run_uuid" FROM "runs" ORDER BY order_null_0, "name",runs.start_time DESC,runs.run_uuid`,
		},
		orderBy:      []string{"`Run name`"},
		expectedVars: []any{},
	},
	{
		name:  "OrderByRunNamePascal",
		query: "",
		expectedSQL: map[string]string{
			"postgres": `SELECT "run_uuid" FROM "runs" ORDER BY order_null_0, "name",runs.start_time DESC,runs.run_uuid`,
		},
		orderBy:      []string{"`Run Name`"},
		expectedVars: []any{},
	},
}

func newPostgresDialector() gorm.Dialector {
	mockedDB, _, _ := sqlmock.New()

	return postgres.New(postgres.Config{
		Conn:       mockedDB,
		DriverName: "postgres",
	})
}

func newSqliteDialector() gorm.Dialector {
	mockedDB, mock, _ := sqlmock.New()
	mock.ExpectQuery("select sqlite_version()").WillReturnRows(
		sqlmock.NewRows([]string{"sqlite_version()"}).AddRow("3.41.1"))

	return sqlite.New(sqlite.Config{
		DriverName: "sqlite3",
		Conn:       mockedDB,
	})
}

func newSQLServerDialector() gorm.Dialector {
	mockedDB, _, _ := sqlmock.New()

	return sqlserver.New(sqlserver.Config{
		DriverName: "sqlserver",
		Conn:       mockedDB,
	})
}

func newMySQLDialector() gorm.Dialector {
	mockedDB, _, _ := sqlmock.New()

	return mysql.New(mysql.Config{
		DriverName:                "mysql",
		Conn:                      mockedDB,
		SkipInitializeWithVersion: true,
	})
}

var dialectors = []gorm.Dialector{
	newPostgresDialector(),
	newSqliteDialector(),
	newSQLServerDialector(),
	newMySQLDialector(),
}

func assertTestData(
	t *testing.T, database *gorm.DB, expectedSQL string, testData testData,
) {
	t.Helper()

	transaction := database.Model(&models.Run{})

	contractErr := applyFilter(context.Background(), database, transaction, testData.query)
	if contractErr != nil {
		t.Fatal("contractErr: ", contractErr)
	}

	contractErr = applyOrderBy(context.Background(), database, transaction, testData.orderBy)
	if contractErr != nil {
		t.Fatal("contractErr: ", contractErr)
	}

	sqlErr := transaction.Select("ID").Find(&models.Run{}).Error
	require.NoError(t, sqlErr)

	actualSQL := transaction.Statement.SQL.String()

	// if removeWhitespace(expectedSQL) != removeWhitespace(actualSQL) {
	// 	fmt.Println(strings.ReplaceAll(actualSQL, "`", ""))
	// }

	assert.Equal(t, removeWhitespace(expectedSQL), removeWhitespace(actualSQL))
	assert.Equal(t, testData.expectedVars, transaction.Statement.Vars)
}

func TestSearchRuns(t *testing.T) {
	t.Parallel()

	for _, dialector := range dialectors {
		database, err := gorm.Open(dialector, &gorm.Config{DryRun: true})
		require.NoError(t, err)

		dialectorName := database.Dialector.Name()

		for _, testData := range tests {
			currentTestData := testData
			if expectedSQL, ok := currentTestData.expectedSQL[dialectorName]; ok {
				t.Run(currentTestData.name+"_"+dialectorName, func(t *testing.T) {
					t.Parallel()
					assertTestData(t, database, expectedSQL, currentTestData)
				})
			}
		}
	}
}

func TestInvalidSearchRunsQuery(t *testing.T) {
	t.Parallel()

	database, err := gorm.Open(newSqliteDialector(), &gorm.Config{DryRun: true})
	require.NoError(t, err)

	transaction := database.Model(&models.Run{})

	contractErr := applyFilter(context.Background(), database, transaction, "⚡✱*@❖$#&")
	if contractErr == nil {
		t.Fatal("expected contract error")
	}
}

//nolint:funlen
func TestOrderByClauseParsing(t *testing.T) {
	t.Parallel()

	testData := []struct {
		input    string
		expected orderByExpr
	}{
		{
			input: "status DESC",
			expected: orderByExpr{
				key:   "status",
				order: utils.PtrTo("DESC"),
			},
		},
		{
			input: "run_name",
			expected: orderByExpr{
				key: "name",
			},
		},
		{
			input: "params.input DESC",
			expected: orderByExpr{
				identifier: utils.PtrTo("parameter"),
				key:        "input",
				order:      utils.PtrTo("DESC"),
			},
		},
		{
			input: "metrics.alpha ASC",
			expected: orderByExpr{
				identifier: utils.PtrTo("metric"),
				key:        "alpha",
				order:      utils.PtrTo("ASC"),
			},
		},
		{
			input: "`Run name`",
			expected: orderByExpr{
				key: "name",
			},
		},
		{
			input: "tags.`foo bar` ASC",
			expected: orderByExpr{
				identifier: utils.PtrTo("tag"),
				key:        "foo bar",
				order:      utils.PtrTo("ASC"),
			},
		},
	}

	for _, testData := range testData {
		t.Run(strcase.ToKebab(testData.input), func(t *testing.T) {
			t.Parallel()

			result, err := processOrderByClause(testData.input)
			if err != nil {
				t.Fatalf("unexpected error: %A", err)
			}

			if !reflect.DeepEqual(testData.expected, result) {
				t.Fatalf("expected (%s, %s, %s), got (%s, %s, %s)",
					*testData.expected.identifier,
					testData.expected.key,
					*testData.expected.order,
					*result.identifier,
					result.key,
					*result.order,
				)
			}
		})
	}
}
