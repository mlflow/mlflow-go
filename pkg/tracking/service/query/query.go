package query

import (
	"fmt"

	"github.com/mlflow/mlflow-go/pkg/tracking/service/query/lexer"
	"github.com/mlflow/mlflow-go/pkg/tracking/service/query/parser"
)

func ParseFilter(input string) ([]*parser.ValidCompareExpr, error) {
	if input == "" {
		return make([]*parser.ValidCompareExpr, 0), nil
	}

	tokens, err := lexer.Tokenize(&input)
	if err != nil {
		return nil, fmt.Errorf("error while lexing %s: %w", input, err)
	}

	ast, err := parser.Parse(tokens)
	if err != nil {
		return nil, fmt.Errorf("error while parsing %s: %w", input, err)
	}

	validExpressions := make([]*parser.ValidCompareExpr, 0, len(ast.Exprs))

	for _, expr := range ast.Exprs {
		ve, err := parser.ValidateExpression(expr)
		if err != nil {
			return nil, fmt.Errorf("error while validating %s: %w", input, err)
		}

		validExpressions = append(validExpressions, ve)
	}

	return validExpressions, nil
}
