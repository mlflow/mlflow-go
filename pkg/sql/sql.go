package sql

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"runtime"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"

	"github.com/mlflow/mlflow-go/pkg/utils"
)

var (
	errSqliteMemory             = errors.New("go implementation does not support :memory: for sqlite")
	errSqliteQueryParamsWindows = errors.New("query parameters are not supported on Windows")
)

//nolint:ireturn
func getDialector(uri *url.URL) (gorm.Dialector, error) {
	uri.Scheme, _, _ = strings.Cut(uri.Scheme, "+")

	switch uri.Scheme {
	case "mssql":
		uri.Scheme = "sqlserver"

		return sqlserver.Open(uri.String()), nil
	case "mysql":
		return mysql.Open(fmt.Sprintf("%s@tcp(%s)%s?%s", uri.User, uri.Host, uri.Path, uri.RawQuery)), nil
	case "postgres", "postgresql":
		return postgres.Open(uri.String()), nil
	case "sqlite":
		uri.Scheme = ""
		uri.Path = uri.Path[1:]
		dsn := uri.String()

		if uri.Path == ":memory:" {
			return nil, errSqliteMemory
		}

		if runtime.GOOS == "windows" {
			if uri.RawQuery != "" {
				return nil, errSqliteQueryParamsWindows
			}

			dsn = strings.ReplaceAll(uri.Path, "/", "\\")
		}

		return sqlite.Open(dsn), nil
	default:
		return nil, fmt.Errorf("unsupported store URL scheme %q", uri.Scheme) //nolint:err113
	}
}

func initSqlite(database *gorm.DB) error {
	database.Exec("PRAGMA case_sensitive_like = true;")

	sqlDB, err := database.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}
	// set SetMaxOpenConns to be 1 only in case of SQLite to avoid `database is locked`
	// in case of parallel calls to some endpoints that use `transactions`.
	sqlDB.SetMaxOpenConns(1)

	return nil
}

func NewDatabase(ctx context.Context, storeURL string) (*gorm.DB, error) {
	logger := utils.GetLoggerFromContext(ctx)

	uri, err := url.Parse(storeURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse store URL %q: %w", storeURL, err)
	}

	dialector, err := getDialector(uri)
	if err != nil {
		return nil, err
	}

	database, err := gorm.Open(dialector, &gorm.Config{
		TranslateError: true,
		Logger:         NewLoggerAdaptor(logger, LoggerAdaptorConfig{}),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database %q: %w", uri.String(), err)
	}

	if dialector.Name() == "sqlite" {
		if err := initSqlite(database); err != nil {
			return nil, err
		}
	}

	return database, nil
}

func CloseDatabase(database *gorm.DB) error {
	db, err := database.DB()
	if err != nil {
		return fmt.Errorf("error while getting database connection: %w", err)
	}

	database.Logger.Info(database.Statement.Context, "closing sql connection")

	err = db.Close()
	if err != nil {
		return fmt.Errorf("error while closing store: %w", err)
	}

	inUse := db.Stats().InUse
	if inUse > 0 {
		return fmt.Errorf("there are still %d in use connections", inUse)
	}

	return nil
}
