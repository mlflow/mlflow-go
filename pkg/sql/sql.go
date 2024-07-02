package sql

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/ncruces/go-sqlite3/gormlite"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

func NewDatabase(logger *logrus.Logger, storeURL string) (*gorm.DB, error) {
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
		dialector = gormlite.Open(uri.String())
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

	return database, nil
}
