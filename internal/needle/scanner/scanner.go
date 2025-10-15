package scanner

import (
	"needle/internal/needle/token"
	"strings"
)

const eof = -1

type Scanner struct {
	source []rune
	arrow  int
	line   int
	column int
}

func New(source []rune) *Scanner {
	return &Scanner{
		source: source,
		arrow:  0,
		line:   1,
		column: 1,
	}
}

func (s *Scanner) Reset() {
	s.arrow = 0
	s.column = 1
	s.line = 1
}

func (s *Scanner) NextToken() *token.Token {
	s.skipWhite()
	r := s.read()

	if (r == '=' || r == '!') && s.peek() == '=' {
		first := r
		s.read()
		if s.peek() == '=' {
			s.read()
			if first == '=' {
				return token.NewToken(token.IS, "===", s.line, s.column-3)
			} else {
				return token.NewToken(token.ISNT, "!==", s.line, s.column-3)
			}
		} else {
			if first == '=' {
				return token.NewToken(token.EQ, "==", s.line, s.column-2)
			} else {
				return token.NewToken(token.NE, "!=", s.line, s.column-2)
			}
		}
	} else if r == '/' && (s.peek() == '/' || s.peek() == '*') {
		if errToken := s.skipComment(); errToken != nil {
			return errToken
		}
		return s.NextToken()
	} else if t, ok := dual[string([]rune{r, s.peek()})]; ok {
		literal := string([]rune{r, s.peek()})
		s.read()
		return token.NewToken(t, literal, s.line, s.column-2)
	} else if t, ok := mono[r]; ok {
		return token.NewToken(t, string(r), s.line, s.column-1)
	} else if isAlpha(r) {
		return s.readIdentifier(r)
	} else if isDigit(r) {
		return s.readNumber(r)
	} else if r == '"' {
		return s.readString()
	} else if r == '`' {
		return s.readUniversalIdentifier()
	} else if r == eof {
		return token.NewToken(token.EOF, "", s.line, s.column-1)
	} else {
		return token.NewToken(token.ERROR, string(r), s.line, s.column-1)
	}
}

func (s *Scanner) read() rune {
	r := s.peek()
	if r == '\n' {
		s.line++
		s.column = 1
	} else {
		s.column++
	}
	s.arrow++
	return r
}

func (s *Scanner) peek() rune {
	if s.arrow >= len(s.source) {
		return eof
	}
	return s.source[s.arrow]
}

func (s *Scanner) skipWhite() {
	for {
		next := s.peek()
		if next != ' ' &&
			next != '\n' &&
			next != '\t' &&
			next != '\r' {
			return
		}
		s.read()
	}
}

func (s *Scanner) skipComment() *token.Token {
	mode := s.read() // '/' or '*'
	if mode == '/' {
		for {
			r := s.read()
			if r == '\n' || r == eof {
				return nil
			}
		}
	} else {
		ln, col := s.line, s.column-2
		for {
			r := s.read()
			if r == '*' && s.peek() == '/' {
				s.read()
				return nil
			}
			if r == eof {
				return token.NewToken(token.ERROR, "/*...", ln, col)
			}
		}
	}
}

func (s *Scanner) readIdentifier(firstChar rune) *token.Token {
	column := s.column - 1
	var str strings.Builder
	str.WriteRune(firstChar)
	for {
		next := s.peek()
		if !isAlpha(next) && !isDigit(next) {
			break
		}
		str.WriteRune(s.read())
	}
	literal := str.String()
	var type_ token.TokenType
	if t, ok := indentifiers[literal]; ok {
		type_ = t
	} else {
		type_ = token.IDENT
	}
	return token.NewToken(type_, literal, s.line, column)
}

func (s *Scanner) readUniversalIdentifier() *token.Token {
	if s.peek() == '`' {
		s.read()
		return token.NewToken(token.ERROR, "``", s.line, s.column-2)
	}
	column := s.column - 1
	var str strings.Builder
	for {
		r := s.read()
		if r == '`' {
			break
		}
		if r == eof || r == '\r' || r == '\n' || r == '\t' {
			return token.NewToken(token.ERROR, "`"+str.String(), s.line, column)
		}
		str.WriteRune(r)
	}
	return token.NewToken(token.IDENT, str.String(), s.line, column)
}

func (s *Scanner) readNumber(firstChar rune) *token.Token {
	column := s.column - 1
	var str strings.Builder
	str.WriteRune(firstChar)

	for {
		next := s.peek()
		if !isDigit(next) && next != '.' {
			break
		}
		if next == '.' {
			str.WriteRune(s.read())
			for {
				next := s.peek()
				if !isDigit(next) {
					break
				}
				str.WriteRune(s.read())
			}
			break
		}
		str.WriteRune(s.read())
	}
	return token.NewToken(token.NUMBER, str.String(), s.line, column)
}

func (s *Scanner) readString() *token.Token {
	column := s.column - 1
	var str strings.Builder
	for {
		r := s.read()
		if esc, ok := escapes[string([]rune{r, s.peek()})]; ok {
			str.WriteRune(esc)
			s.read()
			continue
		}
		if r == '"' {
			break
		}
		if r == '\n' || r == '\r' || r == eof {
			return token.NewToken(token.ERROR, "\""+str.String(), s.line, column)
		}
		str.WriteRune(r)
	}
	return token.NewToken(token.STRING, str.String(), s.line, column)
}

// Include underscore
func isAlpha(char rune) bool {
	return 'a' <= char && char <= 'z' ||
		'A' <= char && char <= 'Z' ||
		char == '_'
}

func isDigit(char rune) bool {
	return '0' <= char && char <= '9'
}

var mono = map[rune]token.TokenType{
	'(': token.L_PAREN,
	')': token.R_PAREN,
	'{': token.L_BRACE,
	'}': token.R_BRACE,
	'[': token.L_BRACK,
	']': token.R_BRACK,

	';': token.SEMI,
	':': token.COLON,
	'?': token.QUEST,
	',': token.COMMA,
	'=': token.ASSIGN,
	'.': token.DOT,
	'!': token.WOW,

	'<': token.LT,
	'>': token.GT,

	'+': token.PLUS,
	'-': token.MINUS,
	'*': token.STAR,
	'/': token.SLASH,
}

var dual = map[string]token.TokenType{
	"<=": token.LE,
	">=": token.GE,

	"->": token.ARROW,
	":=": token.DEF,

	"+=": token.PLUS_ASSIGN,
	"-=": token.MINUS_ASSIGN,
	"*=": token.STAR_ASSIGN,
	"/=": token.SLASH_ASSIGN,
}

var indentifiers = map[string]token.TokenType{
	"or":  token.OR,
	"and": token.AND,

	"var": token.VAR,

	"fun":   token.FUN,
	"class": token.CLASS,
	"vec":   token.VEC,
	"map":   token.MAP,

	"null":  token.NULL,
	"true":  token.BOOLEAN,
	"false": token.BOOLEAN,

	"for":     token.FOR,
	"while":   token.WHILE,
	"do":      token.DO,
	"if":      token.IF,
	"else":    token.ELSE,
	"throw":   token.THROW,
	"try":     token.TRY,
	"catch":   token.CATCH,
	"finally": token.FINALLY,

	"self": token.SELF,

	"return":   token.RETURN,
	"break":    token.BREAK,
	"continue": token.CONTINUE,

	"import": token.IMPORT,

	"say": token.SAY,
}

var escapes = map[string]rune{
	`\n`: '\n',
	`\r`: '\r',
	`\t`: '\t',
	`\"`: '"',
	`\\`: '\\',
}
