package sql

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/mlflow/mlflow-go/pkg/entities"
)

const (
	ModelsURISuffixLatest = "latest"
)

//nolint
var ErrImproperModelURI = func(uri string) error {
	return fmt.Errorf(`
		Not a proper models:/ URI: %s. "Models URIs must be of the form 'models:/model_name/suffix' or 
		'models:/model_name@alias' where suffix is a model version, stage, or the string latest 
		and where alias is a registered model alias. Only one of suffix or alias can be defined at a time."`,
		uri,
	)
}

type ParsedModelURI struct {
	Name    string
	Stage   string
	Alias   string
	Version string
}

func GetModelNextVersion(registeredModel *entities.RegisteredModel) int32 {
	if len(registeredModel.Versions) == 0 {
		return 1
	}

	maxVersion := int32(0)
	for _, version := range registeredModel.Versions {
		if version.Version > maxVersion {
			maxVersion = version.Version
		}
	}

	return maxVersion + 1
}

//nolint
func ParseModelURI(uri string) (*ParsedModelURI, error) {
	parsedURI, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	if parsedURI.Scheme != "models" {
		return nil, ErrImproperModelURI(uri)
	}

	if !strings.HasSuffix(parsedURI.Path, "/") || len(parsedURI.Path) <= 1 {
		return nil, ErrImproperModelURI(uri)
	}

	parts := strings.Split(strings.TrimLeft(parsedURI.Path, "/"), "/")
	if len(parts) > 2 || strings.Trim(parts[0], " ") == "" {
		return nil, ErrImproperModelURI(uri)
	}

	if len(parts) == 2 {
		name, suffix := parts[0], parts[1]
		if strings.Trim(suffix, " ") == "" {
			return nil, ErrImproperModelURI(uri)
		}
		// The suffix is a specific version, e.g. "models:/AdsModel1/123"
		if _, err := strconv.Atoi(suffix); err == nil {
			return &ParsedModelURI{
				Name:    name,
				Version: suffix,
			}, nil
		}
		// The suffix is the 'latest' string (case insensitive), e.g. "models:/AdsModel1/latest"
		if (strings.ToLower(suffix)) == ModelsURISuffixLatest {
			return &ParsedModelURI{
				Name: name,
			}, nil
		}
		// The suffix is a specific stage (case insensitive), e.g. "models:/AdsModel1/Production"
		return &ParsedModelURI{
			Name:  name,
			Stage: suffix,
		}, nil
	}

	aliasParts := strings.SplitN(parts[0], "@", 1)
	if len(aliasParts) != 2 || strings.Trim(aliasParts[1], " ") == "" {
		return nil, ErrImproperModelURI(uri)
	}

	return &ParsedModelURI{
		Name:  aliasParts[0],
		Alias: aliasParts[1],
	}, nil
}
