package token

import (
	"fmt"
	"needle/internal/pkg"
)

type TokenType string

type Position struct {
	Line   int
	Column int
}

type Token struct {
	Type     TokenType
	Literal  string
	Position Position
}

func NewToken(t TokenType, lit string, ln, col int) *Token {
	return &Token{
		Type:    t,
		Literal: lit,
		Position: Position{
			Line:   ln,
			Column: col,
		},
	}
}

func PrintTokens(tkns []*Token) {
	fmt.Println("| type         | literal      | line | column |")
	fmt.Println("|--------------|--------------|------|--------|")
	for _, tkn := range tkns {
		fmt.Printf(
			"| %-12s | %-12s | %-4d | %-6d |\n",
			tkn.Type,
			pkg.ShortString(tkn.Literal, 12),
			tkn.Position.Line,
			tkn.Position.Column,
		)
	}
}

const (
	ERROR TokenType = "__error"
	EOF   TokenType = "__eof"

	PLUS  TokenType = "+"
	MINUS TokenType = "-"
	STAR  TokenType = "*"
	SLASH TokenType = "/"

	LT   TokenType = "<"
	LE   TokenType = "<="
	GT   TokenType = ">"
	GE   TokenType = ">="
	EQ   TokenType = "=="
	NE   TokenType = "!="
	IS   TokenType = "==="
	ISNT TokenType = "!=="

	L_PAREN TokenType = "("
	R_PAREN TokenType = ")"
	L_BRACE TokenType = "{"
	R_BRACE TokenType = "}"
	L_BRACK TokenType = "["
	R_BRACK TokenType = "]"

	SEMI   TokenType = ";"
	COLON  TokenType = ":"
	QUEST  TokenType = "?"
	COMMA  TokenType = ","
	ASSIGN TokenType = "="
	DOT    TokenType = "."
	WOW    TokenType = "!"

	ARROW  TokenType = "->"

	OR  TokenType = "or"
	AND TokenType = "and"

	VAR TokenType = "var"

	IDENT   TokenType = "identifier"
	NULL    TokenType = "null"
	BOOLEAN TokenType = "boolean"
	NUMBER  TokenType = "number"
	STRING  TokenType = "string"

	FUN   TokenType = "fun"
	CLASS TokenType = "class"
	ARRAY TokenType = "array"
	TABLE TokenType = "table"

	FOR     TokenType = "for"
	WHILE   TokenType = "while"
	DO      TokenType = "do"
	IF      TokenType = "if"
	ELSE    TokenType = "else"
	THROW   TokenType = "throw"
	TRY     TokenType = "try"
	CATCH   TokenType = "catch"
	FINALLY TokenType = "finally"

	SELF TokenType = "self"

	RETURN   TokenType = "return"
	BREAK    TokenType = "break"
	CONTINUE TokenType = "continue"

	IMPORT TokenType = "import"

	SAY TokenType = "say"
)
