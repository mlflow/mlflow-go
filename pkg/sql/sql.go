package sql

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"

	"github.com/mlflow/mlflow-go/pkg/utils"
)

func NewDatabase(ctx context.Context, storeURL string) (*gorm.DB, error) {
	logger := utils.GetLoggerFromContext(ctx)

	uri, err := url.Parse(storeURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse store URL %q: %w", storeURL, err)
	}

	var dialector gorm.Dialector

	uri.Scheme, _, _ = strings.Cut(uri.Scheme, "+")

	switch uri.Scheme {
	case "mssql":
		uri.Scheme = "sqlserver"
		dialector = sqlserver.Open(uri.String())
	case "mysql":
		dialector = mysql.Open(fmt.Sprintf("%s@tcp(%s)%s?%s", uri.User, uri.Host, uri.Path, uri.RawQuery))
	case "postgres", "postgresql":
		dialector = postgres.Open(uri.String())
	case "sqlite":
		uri.Scheme = ""
		uri.Path = uri.Path[1:]
		dialector = sqlite.Open(uri.String())
	default:
		return nil, fmt.Errorf("unsupported store URL scheme %q", uri.Scheme) //nolint:err113
	}

	database, err := gorm.Open(dialector, &gorm.Config{
		TranslateError: true,
		Logger:         NewLoggerAdaptor(logger, LoggerAdaptorConfig{}),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database %q: %w", uri.String(), err)
	}

	if dialector.Name() == "sqlite" {
		database.Exec("PRAGMA case_sensitive_like = true;")

		sqlDB, err := database.DB()
		if err != nil {
			return nil, fmt.Errorf("failed to get database instance: %w", err)
		}
		// set SetMaxOpenConns to be 1 only in case of SQLite to avoid `database is locked`
		// in case of parallel calls to some endpoints that use `transactions`.
		sqlDB.SetMaxOpenConns(1)
	}

	return database, nil
}
