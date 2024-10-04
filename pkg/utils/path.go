package utils

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"
)

var (
	errFailedToDecodeURL  = errors.New("failed to decode url")
	errInvalidQueryString = errors.New("invalid query string")
)

func decode(input string) (string, error) {
	current := input

	for range 10 {
		decoded, err := url.QueryUnescape(current)
		if err != nil {
			return "", fmt.Errorf("could not unescape %s: %w", current, err)
		}

		parsed, err := url.Parse(decoded)
		if err != nil {
			return "", fmt.Errorf("could not parsed %s: %w", decoded, err)
		}

		if current == parsed.String() {
			return current, nil
		}
	}

	return "", errFailedToDecodeURL
}

func validateQueryString(query string) error {
	query, err := decode(query)
	if err != nil {
		return err
	}

	if strings.Contains(query, "..") {
		return errInvalidQueryString
	}

	return nil
}

func joinPosixPathsAndAppendAbsoluteSuffixes(prefixPath, suffixPath string) string {
	if len(prefixPath) == 0 {
		return suffixPath
	}

	suffixPath = strings.TrimPrefix(suffixPath, "/")

	return path.Join(prefixPath, suffixPath)
}

func AppendToURIPath(uri string, paths ...string) (string, error) {
	path := ""
	for _, subpath := range paths {
		path = joinPosixPathsAndAppendAbsoluteSuffixes(path, subpath)
	}

	parsedURI, err := url.Parse(uri)
	if err != nil {
		return "", fmt.Errorf("failed to parse uri %s: %w", uri, err)
	}

	if err := validateQueryString(parsedURI.RawQuery); err != nil {
		return "", err
	}

	if len(parsedURI.Scheme) == 0 {
		return joinPosixPathsAndAppendAbsoluteSuffixes(uri, path), nil
	}

	prefix := ""
	if !strings.HasPrefix(parsedURI.Path, "/") {
		prefix = parsedURI.Scheme + ":"
		parsedURI.Scheme = ""
	}

	newURIPath := joinPosixPathsAndAppendAbsoluteSuffixes(parsedURI.Path, path)
	parsedURI.Path = newURIPath

	return prefix + parsedURI.String(), nil
}
