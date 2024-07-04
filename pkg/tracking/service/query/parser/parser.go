package parser

import (
	"fmt"
	"strconv"

	"github.com/mlflow/mlflow-go/pkg/tracking/service/query/lexer"
)

type parser struct {
	tokens []lexer.Token
	pos    int
}

func newParser(tokens []lexer.Token) *parser {
	return &parser{
		tokens: tokens,
		pos:    0,
	}
}

func (p *parser) currentTokenKind() lexer.TokenKind {
	return p.tokens[p.pos].Kind
}

func (p *parser) hasTokens() bool {
	return p.pos < len(p.tokens) && p.currentTokenKind() != lexer.EOF
}

func (p *parser) printCurrentToken() string {
	return p.tokens[p.pos].Debug()
}

func (p *parser) currentToken() lexer.Token {
	return p.tokens[p.pos]
}

func (p *parser) advance() lexer.Token {
	tk := p.currentToken()
	p.pos++

	return tk
}

type Error struct {
	message string
}

func NewParserError(format string, a ...any) *Error {
	return &Error{message: fmt.Sprintf(format, a...)}
}

func (e *Error) Error() string {
	return e.message
}

func (p *parser) parseIdentifier() (Identifier, error) {
	emptyIdentifier := Identifier{Identifier: "", Key: ""}
	if p.hasTokens() && p.currentTokenKind() != lexer.Identifier {
		return emptyIdentifier, NewParserError(
			"expected identifier, got %s",
			p.printCurrentToken(),
		)
	}

	identToken := p.advance()

	if p.currentTokenKind() == lexer.Dot {
		p.advance() // Consume the DOT
		//nolint:exhaustive
		switch p.currentTokenKind() {
		case lexer.Identifier:
			column := p.advance().Value

			return Identifier{Identifier: identToken.Value, Key: column}, nil
		case lexer.String:
			column := p.advance().Value
			column = column[1 : len(column)-1] // Remove quotes

			return Identifier{Identifier: identToken.Value, Key: column}, nil
		default:
			return emptyIdentifier, NewParserError(
				"expected IDENTIFIER or STRING, got %s",
				p.printCurrentToken(),
			)
		}
	} else {
		return Identifier{Identifier: "", Key: identToken.Value}, nil
	}
}

func (p *parser) parseOperator() (OperatorKind, error) {
	//nolint:exhaustive
	switch p.advance().Kind {
	case lexer.Equals:
		return Equals, nil
	case lexer.NotEquals:
		return NotEquals, nil
	case lexer.Less:
		return Less, nil
	case lexer.LessEquals:
		return LessEquals, nil
	case lexer.Greater:
		return Greater, nil
	case lexer.GreaterEquals:
		return GreaterEquals, nil
	case lexer.Like:
		return Like, nil
	case lexer.ILike:
		return ILike, nil
	default:
		return -1, NewParserError("expected operator, got %s", p.printCurrentToken())
	}
}

//nolint:ireturn
func (p *parser) parseValue() (Value, error) {
	//nolint:exhaustive
	switch p.currentTokenKind() {
	case lexer.Number:
		n, err := strconv.ParseFloat(p.advance().Value, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse number token to float: %w", err)
		}

		return NumberExpr{Value: n}, nil
	case lexer.String:
		value := p.advance().Value
		value = value[1 : len(value)-1] // Remove quotes

		return StringExpr{Value: value}, nil
	default:
		return nil, NewParserError(
			"Expected NUMBER or STRING, got %s",
			p.printCurrentToken(),
		)
	}
}

func (p *parser) parseInSetExpr(ident Identifier) (*CompareExpr, error) {
	if p.currentTokenKind() != lexer.OpenParen {
		return nil, NewParserError(
			"expected '(', got %s",
			p.printCurrentToken(),
		)
	}

	p.advance() // Consume the OPEN_PAREN

	set := make([]string, 0)

	for p.hasTokens() && p.currentTokenKind() != lexer.CloseParen {
		if p.currentTokenKind() != lexer.String {
			return nil, NewParserError(
				"expected STRING, got %s",
				p.printCurrentToken(),
			)
		}

		value := p.advance().Value
		value = value[1 : len(value)-1] // Remove quotes

		set = append(set, value)

		if p.currentTokenKind() == lexer.Comma {
			p.advance() // Consume the COMMA
		}
	}

	if p.currentTokenKind() != lexer.CloseParen {
		return nil, NewParserError(
			"expected ')', got %s",
			p.printCurrentToken(),
		)
	}

	p.advance() // Consume the CLOSE_PAREN

	return &CompareExpr{Left: ident, Operator: In, Right: StringListExpr{Values: set}}, nil
}

func (p *parser) parseExpression() (*CompareExpr, error) {
	ident, err := p.parseIdentifier()
	if err != nil {
		return nil, err
	}

	//nolint:exhaustive
	switch p.currentTokenKind() {
	case lexer.In:
		p.advance() // Consume the IN

		return p.parseInSetExpr(ident)
	case lexer.Not:
		p.advance() // Consume the NOT

		if p.currentTokenKind() != lexer.In {
			return nil, NewParserError(
				"expected IN after NOT, got %s",
				p.printCurrentToken(),
			)
		}

		p.advance() // Consume the IN

		expr, err := p.parseInSetExpr(ident)
		if err != nil {
			return nil, err
		}

		expr.Operator = NotIn

		return expr, nil
	default:
		operator, err := p.parseOperator()
		if err != nil {
			return nil, err
		}

		value, err := p.parseValue()
		if err != nil {
			return nil, err
		}

		return &CompareExpr{Left: ident, Operator: operator, Right: value}, nil
	}
}

func (p *parser) parse() (*AndExpr, error) {
	exprs := make([]*CompareExpr, 0)

	leftExpr, err := p.parseExpression()
	if err != nil {
		return nil, fmt.Errorf("error while parsing initial expression: %w", err)
	}

	exprs = append(exprs, leftExpr)

	// While there are tokens and the next token is AND
	for p.currentTokenKind() == lexer.And {
		p.advance() // Consume the AND

		rightExpr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		exprs = append(exprs, rightExpr)
	}

	if p.hasTokens() {
		return nil, NewParserError(
			"unexpected leftover token(s) after parsing: %s",
			p.printCurrentToken(),
		)
	}

	return &AndExpr{Exprs: exprs}, nil
}

func Parse(tokens []lexer.Token) (*AndExpr, error) {
	parser := newParser(tokens)

	return parser.parse()
}
