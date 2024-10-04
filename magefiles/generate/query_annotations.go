package generate

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Inspect the AST of the incoming file and add a query annotation to the struct tags which have a json tag.
//
//nolint:funlen,cyclop
func addQueryAnnotation(generatedGoFile string) error {
	// Parse the file into an AST
	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, generatedGoFile, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("add query annotation failed: %w", err)
	}

	// Create an AST inspector to modify specific struct tags
	ast.Inspect(node, func(n ast.Node) bool {
		// Look for struct type declarations
		typeSpec, isTypeSpec := n.(*ast.TypeSpec)
		if !isTypeSpec {
			return true
		}

		structType, isStructType := typeSpec.Type.(*ast.StructType)

		if !isStructType {
			return true
		}

		// Iterate over fields in the struct
		for _, field := range structType.Fields.List {
			if field.Tag == nil {
				continue
			}

			tagValue := field.Tag.Value

			hasQuery := strings.Contains(tagValue, "query:")
			hasValidate := strings.Contains(tagValue, "validate:")
			validationKey := fmt.Sprintf("%s_%s", typeSpec.Name, field.Names[0])
			validationRule, needsValidation := validations[validationKey]

			if hasQuery && (!needsValidation || needsValidation && hasValidate) {
				continue
			}

			// With opening ` tick
			newTag := tagValue[0 : len(tagValue)-1]

			matches := jsonFieldTagRegexp.FindStringSubmatch(tagValue)
			if len(matches) > 0 && !hasQuery {
				// Modify the tag here
				// The json annotation could be something like `json:"key,omitempty"`
				// We only want the key part, the `omitempty` is not relevant for the query annotation
				key := matches[1]
				if strings.Contains(key, ",") {
					key = strings.Split(key, ",")[0]
				}
				// Add query annotation
				newTag += fmt.Sprintf(" query:\"%s\"", key)
			}

			if needsValidation {
				// Add validation annotation
				newTag += fmt.Sprintf(" validate:\"%s\"", validationRule)
			}

			// Closing ` tick
			newTag += "`"
			field.Tag.Value = newTag
		}

		return false
	})

	return saveASTToFile(fset, node, false, generatedGoFile)
}

var jsonFieldTagRegexp = regexp.MustCompile(`json:"([^"]+)"`)

//nolint:err113
func AddQueryAnnotations(pkgFolder string) error {
	protoFolder := filepath.Join(pkgFolder, "protos")

	if _, pathError := os.Stat(protoFolder); os.IsNotExist(pathError) {
		return fmt.Errorf("the %s folder does not exist. Are the Go protobuf files generated?", protoFolder)
	}

	err := filepath.WalkDir(protoFolder, func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == ".go" {
			err = addQueryAnnotation(path)
		}

		return err
	})
	if err != nil {
		return fmt.Errorf("failed to add query annotation: %w", err)
	}

	return nil
}
