package parser

import (
	"errors"
	"fmt"
	"needle/internal/needle/ast"
	"needle/internal/needle/token"
	"strconv"
)

type Tokenizer interface {
	NextToken() *token.Token
}

type Parser struct {
	tokenizer Tokenizer
	current   *token.Token
	backpack  *token.Token
	errors    []error
}

func New(tokenizer Tokenizer) *Parser {
	p := &Parser{
		tokenizer: tokenizer,
		backpack:  nil,
		errors:    nil,
	}
	p.advance()
	return p
}

func (p *Parser) Parse() (*ast.Script, []error) {
	script := &ast.Script{
		Decls: []ast.Decl{},
	}

	for !p.check(token.EOF) {
		decl := p.catch(p.declaration)
		if decl == nil {
			p.synchronize()
			decl = newBadDecl()
		}
		script.Decls = append(script.Decls, decl)
		p.advance()
	}

	return script, p.errors
}

func (p *Parser) declaration() ast.Decl {
	switch p.current.Type {
	case token.VAR:
		return p.varDecl()
	case token.FUN:
		if p.peek().Type == token.IDENT {
			return nil // TODO
		}
		fallthrough
	case token.CLASS:
		if p.peek().Type == token.IDENT {
			return nil // TODO
		}
		fallthrough
	default:
		return &ast.StmtDecl{
			Stmt: p.statement(),
		}
	}
}

func (p *Parser) statement() ast.Stmt {
	switch p.current.Type {
	case token.SEMI:
		return newNullStmt()
	case token.L_BRACE:
		return p.block()
	case token.FOR:
		return p.forStmt()
	case token.WHILE:
		return p.whileStmt()
	case token.DO:
		return p.doStmt()
	case token.IF:
		return p.ifStmt()
	case token.SAY:
		return p.sayStmt()
	case token.TRY:
		return p.tryStmt()
	case token.THROW:
		return p.throwStmt()
	case token.RETURN:
		return p.returnStmt()
	case token.BREAK:
		p.expect(token.SEMI)
		return &ast.BreakStmt{}
	case token.CONTINUE:
		p.expect(token.SEMI)
		return &ast.ContinueStmt{}
	}

	expr := p.expression(LOWEST)
	if p.peek().Type == token.ASSIGN {
		p.advance()
		return p.assignStmt(expr)
	}

	p.expect(token.SEMI)
	return &ast.ExprStmt{Expr: expr}
}

func (p *Parser) expression(prec precedence) ast.Expr {
	var expr ast.Expr
	switch p.current.Type {
	case token.L_PAREN:
		p.advance()
		if p.check(token.R_PAREN) {
			panicParseError(
				p.current,
				"unexpected ')'",
			)
		}
		expr = p.expression(LOWEST)
		p.expect(token.R_PAREN)

	case token.CLASS:
		expr = p.classLit()
	case token.FUN:
		expr = p.funLit()
	case token.ARRAY:
		expr = p.arrayLit()
	case token.TABLE:
		expr = p.tableLit()

	case token.NULL:
		expr = &ast.NullLit{}
	case token.BOOLEAN:
		if val, err := strconv.ParseBool(p.current.Literal); err != nil {
			panic(err)
		} else {
			expr = &ast.BooleanLit{Value: val}
		}
	case token.NUMBER:
		if val, err := strconv.ParseFloat(p.current.Literal, 64); err != nil {
			panic(err)
		} else {
			expr = &ast.NumberLit{Value: val}
		}
	case token.STRING:
		expr = &ast.StringLit{Value: p.current.Literal}

	case token.IDENT:
		expr = p.ident()
	case token.SELF:
		expr = &ast.SelfLit{}

	case token.MINUS, token.PLUS, token.WOW:
		op := p.current
		p.advance()
		e := p.expression(UN)
		expr = &ast.PrefixExpr{Right: e, Op: op}
	default:
		panicParseError(
			p.current,
			"unexpected '%s'",
			p.current.Literal,
		)
	}

	for prec < p.peekPrecedence() {
		p.advance()
		switch p.current.Type {
		case token.PLUS, token.MINUS, token.STAR, token.SLASH,
			token.LT, token.LE, token.GT, token.GE, token.EQ, token.NE,
			token.AND, token.OR, token.IS, token.ISNT:
			expr = p.infixExpr(expr)
		case token.L_PAREN:
			expr = p.callExpr(expr)
		case token.DOT:
			expr = p.propExpr(expr)
		case token.L_BRACK:
			expr = p.indexOrSliceExpr(expr)
		default:
			panicParseError(
				p.current,
				"unexpected '%s'",
				p.current.Literal,
			)
		}
	}

	return expr
}

/* == declarations ========================================================== */

func (p *Parser) varDecl() *ast.VarDecl {
	decl := &ast.VarDecl{}

	p.expect(token.IDENT)
	decl.Name = p.ident()

	p.advance()
	if p.check(token.SEMI) {
		decl.Right = newNullExpr()
		return decl
	} else if p.check(token.ASSIGN) {
		p.advance()
		decl.Right = p.expression(LOWEST)
		p.expect(token.SEMI)
		return decl
	}
	panicParseError(
		p.current,
		"expected ';' or '='",
	)
	return nil
}

/* == statements ============================================================ */

func (p *Parser) block() *ast.Block {
	block := &ast.Block{
		Decls: []ast.Decl{},
	}

	p.advance()
	for !p.check(token.R_BRACE) {
		decl := p.catch(p.declaration)
		if decl == nil {
			p.synchronize()
			decl = newBadDecl()
		}
		block.Decls = append(block.Decls, decl)
		p.advance()
		if p.check(token.EOF) {
			panicParseError(
				p.current,
				"expected '}'",
			)
		}
	}

	return block
}

func (p *Parser) forStmt() *ast.ForStmt {
	stmt := &ast.ForStmt{}
	p.advance()
	stmt.Init = p.declaration()
	p.advance()
	stmt.Cond = p.expression(LOWEST)
	p.expect(token.SEMI)
	p.advance()
	post := p.expression(LOWEST)
	if p.peek().Type == token.ASSIGN {
		p.advance()
		p.advance()
		stmt.Post = &ast.AssignStmt{Left: post, Right: p.expression(LOWEST)}
	} else {
		stmt.Post = &ast.ExprStmt{Expr: post}
	}
	if p.peek().Type != token.L_BRACE {
		p.expect(token.ARROW)
	}
	p.advance()
	stmt.Repeat = p.statement()
	return stmt
}

func (p *Parser) whileStmt() *ast.WhileStmt {
	stmt := &ast.WhileStmt{}
	p.advance()
	stmt.Cond = p.expression(LOWEST)
	if p.peek().Type != token.L_BRACE {
		p.expect(token.ARROW)
	}
	p.advance()
	stmt.Do = p.statement()
	return stmt
}

func (p *Parser) doStmt() *ast.DoStmt {
	stmt := &ast.DoStmt{}
	p.advance()
	stmt.Do = p.statement()
	p.expect(token.WHILE)
	p.advance()
	stmt.While = p.expression(LOWEST)
	p.expect(token.SEMI)
	return stmt
}

func (p *Parser) ifStmt() *ast.IfStmt {
	stmt := &ast.IfStmt{}
	p.advance()
	stmt.Cond = p.expression(LOWEST)
	if p.peek().Type != token.L_BRACE {
		p.expect(token.ARROW)
	}
	p.advance()
	stmt.Then = p.statement()
	if p.peek().Type == token.ELSE {
		p.advance()
		p.advance()
		stmt.Else = p.statement()
	} else {
		stmt.Else = newNullStmt()
	}
	return stmt
}

func (p *Parser) sayStmt() *ast.SayStmt {
	stmt := &ast.SayStmt{}
	p.advance()
	stmt.Expr = p.expression(LOWEST)
	p.expect(token.SEMI)
	return stmt
}

func (p *Parser) tryStmt() *ast.TryStmt {
	stmt := &ast.TryStmt{}
	ended := false
	p.advance()
	stmt.Try = p.statement()
	if p.peek().Type == token.CATCH {
		p.advance()
		p.expect(token.IDENT)
		stmt.As = p.ident()
		if p.peek().Type != token.L_BRACE {
			p.expect(token.ARROW)
		}
		p.advance()
		stmt.Catch = p.statement()
		ended = true
	} else {
		stmt.As = &ast.Ident{Name: "_"}
		stmt.Catch = newNullStmt()
	}
	if p.peek().Type == token.FINALLY {
		p.advance()
		p.advance()
		stmt.Finally = p.statement()
		ended = true
	} else {
		stmt.Finally = newNullStmt()
	}
	if !ended {
		panicParseError(
			p.current,
			"expected 'catch' or 'finally'",
		)
	}
	return stmt
}

func (p *Parser) throwStmt() *ast.ThrowStmt {
	stmt := &ast.ThrowStmt{}
	p.advance()
	stmt.Error = p.expression(LOWEST)
	p.expect(token.SEMI)
	return stmt
}

func (p *Parser) returnStmt() *ast.ReturnStmt {
	stmt := &ast.ReturnStmt{}
	p.advance()
	if p.check(token.SEMI) {
		stmt.Value = newNullExpr()
		return stmt
	}
	stmt.Value = p.expression(LOWEST)
	p.expect(token.SEMI)
	return stmt
}

func (p *Parser) assignStmt(left ast.Expr) *ast.AssignStmt {
	stmt := &ast.AssignStmt{Left: left}
	p.advance()
	stmt.Right = p.expression(LOWEST)
	p.expect(token.SEMI)
	return stmt
}

/* == expressions =========================================================== */

func (p *Parser) ident() *ast.Ident {
	return &ast.Ident{
		Name: p.current.Literal,
	}
}

func (p *Parser) classLit() *ast.ClassLit {
	lit := &ast.ClassLit{
		Inits: map[*ast.Ident]*ast.FunLit{},
		Funs:  map[*ast.Ident]*ast.FunLit{},
	}
	p.expect(token.L_BRACE)
	p.advance()
	for !p.check(token.R_BRACE) {
		if p.current.Literal == LIT_INIT {
			p.expect(token.IDENT)
			name := p.ident()
			lit.Inits[name] = p.funLit()
		} else if p.check(token.FUN) {
			p.expect(token.IDENT)
			name := p.ident()
			lit.Funs[name] = p.funLit()
		} else {
			panicParseError(
				p.current,
				"expected method declaration",
			)
		}
		p.advance()
		if p.check(token.EOF) {
			panicParseError(
				p.current,
				"expected '}'",
			)
		}
	}
	return lit
}

func (p *Parser) funLit() *ast.FunLit {
	lit := &ast.FunLit{}
	p.expect(token.L_PAREN)
	lit.Params = p.parameters()
	if p.peek().Type != token.L_BRACE {
		p.expect(token.ARROW)
	}
	p.advance()
	lit.Body = p.statement()
	return lit
}

func (p *Parser) arrayLit() *ast.ArrayLit {
	lit := &ast.ArrayLit{}
	p.expect(token.L_BRACE)
	lit.Elems = p.arrayElements()
	return lit
}

func (p *Parser) tableLit() *ast.TableLit {
	lit := &ast.TableLit{}
	p.expect(token.L_BRACE)
	lit.Pairs = p.tablePairs()
	return lit
}

func (p *Parser) infixExpr(left ast.Expr) *ast.InfixExpr {
	expr := &ast.InfixExpr{
		Left: left,
		Op:   p.current,
	}
	prec := p.currentPrecedence()
	p.advance()
	expr.Right = p.expression(prec)
	return expr
}

func (p *Parser) callExpr(left ast.Expr) *ast.CallExpr {
	expr := &ast.CallExpr{Left: left}
	expr.Arguments = p.arguments()
	return expr
}

func (p *Parser) propExpr(left ast.Expr) *ast.PropExpr {
	expr := &ast.PropExpr{Left: left}
	p.expect(token.IDENT)
	expr.Prop = p.ident()
	return expr
}

func (p *Parser) indexOrSliceExpr(left ast.Expr) ast.Expr {
	p.advance()
	index := p.expression(LOWEST)
	p.advance()
	if p.check(token.R_BRACK) {
		return &ast.IndexExpr{Left: left, Index: index}
	}
	if !p.check(token.COLON) {
		panicParseError(
			p.current,
			"expected ']' or ':'",
		)
	}
	p.advance()
	end := p.expression(LOWEST)
	p.expect(token.R_BRACK)
	return &ast.SliceExpr{Left: left, Start: index, End: end}
}

/* == parse utility ========================================================= */

func (p *Parser) tablePairs() map[ast.Expr]ast.Expr {
	pairs := map[ast.Expr]ast.Expr{}
	if p.peek().Type == token.R_BRACE {
		p.advance()
		return pairs
	}
	for {
		p.advance()
		k := p.expression(LOWEST)
		p.expect(token.COLON)
		p.advance()
		v := p.expression(LOWEST)
		pairs[k] = v
		p.advance()
		if p.check(token.R_BRACE) {
			break
		}
		if !p.check(token.COMMA) {
			panicParseError(
				p.current,
				"expected ',' or '}'",
			)
		}
		if p.peek().Type == token.R_BRACE {
			p.advance()
			break
		}
	}
	return pairs
}

func (p *Parser) arrayElements() []ast.Expr {
	elems := []ast.Expr{}
	p.advance()
	if p.check(token.R_BRACE) {
		return elems
	}
	for {
		expr := p.expression(LOWEST)
		elems = append(elems, expr)
		p.advance()
		if p.check(token.R_BRACE) {
			break
		}
		if !p.check(token.COMMA) {
			panicParseError(
				p.current,
				"expected ',' or '}'",
			)
		}
		p.advance()
		if p.check(token.R_BRACE) {
			break
		}
	}
	return elems
}

func (p *Parser) arguments() []ast.Expr {
	args := []ast.Expr{}
	p.advance()
	if p.check(token.R_PAREN) {
		return args
	}
	for {
		expr := p.expression(LOWEST)
		args = append(args, expr)
		p.advance()
		if p.check(token.R_PAREN) {
			break
		}
		if !p.check(token.COMMA) {
			panicParseError(
				p.current,
				"expected ',' or ')'",
			)
		}
		p.advance()
		if p.check(token.R_PAREN) {
			break
		}
	}
	return args
}

func (p *Parser) parameters() []*ast.Ident {
	params := []*ast.Ident{}
	p.advance()
	if p.check(token.R_PAREN) {
		return params
	}
	for {
		if !p.check(token.IDENT) {
			panicParseError(
				p.current,
				"expected 'identifier'",
			)
		}
		params = append(
			params,
			p.ident(),
		)
		p.advance()
		if p.check(token.R_PAREN) {
			break
		}
		if !p.check(token.COMMA) {
			panicParseError(
				p.current,
				"expected ',' or ')'",
			)
		}
		p.advance()
		if p.check(token.R_PAREN) {
			break
		}
	}
	return params
}

/* == utility =============================================================== */

func (p *Parser) currentPrecedence() precedence {
	return precedences[p.current.Type]
}

func (p *Parser) peekPrecedence() precedence {
	return precedences[p.peek().Type]
}

func (p *Parser) catch(f func() ast.Decl) (result ast.Decl) {
	defer func() {
		if pa := recover(); pa != nil {
			if pErr, ok := pa.(*parseError); ok {
				p.errors = append(p.errors, pErr.Error)
				result = nil
				return
			}
			panic(pa)
		}
	}()
	return f()
}

func (p *Parser) synchronize() {
	for !p.check(token.EOF) {
		if p.check(token.SEMI) || p.check(token.R_BRACE) {
			return
		}
		switch p.peek().Type {
		case token.L_BRACE, token.VAR, token.WHILE, token.DO,
			token.SAY, token.IF, token.RETURN,
			token.BREAK, token.CONTINUE, token.TRY:
			return
		}
		p.advance()
	}
}

func (p *Parser) expect(t token.TokenType) {
	p.advance()
	if t == p.current.Type {
		return
	}
	panicParseError(p.current, "expected '%s'", t)
}

func (p *Parser) consume(t token.TokenType) {
	if t == p.current.Type {
		p.advance()
		return
	}
	panicParseError(p.current, "expected '%s'", t)
}

func (p *Parser) check(t token.TokenType) bool {
	return p.current.Type == t
}

func (p *Parser) peek() *token.Token {
	temp := p.current
	p.advance()
	tkn := p.current
	p.current = temp
	p.backpack = tkn
	return tkn
}

func (p *Parser) advance() {
	if p.backpack != nil {
		p.current = p.backpack
		p.backpack = nil
		return
	}
	next := p.tokenizer.NextToken()
	if next.Type == token.ERROR {
		panicParseError(
			next,
			"error token",
		)
	}
	p.current = next
}

const (
	LIT_INIT  = "init"
	LIT_GET   = "get"
	LIT_SET   = "set"
	LIT_INFIX = "infix"
)

type precedence int

const (
	LOWEST precedence = iota
	OR                // or
	AND               // and
	EQ                // == != === !==
	COMP              // < <= > >=
	TERM              // + -
	FACTOR            // * /
	UN                // - + !
	CALL              // . () []
	HIGHEST
)

var precedences = map[token.TokenType]precedence{
	token.OR: OR,

	token.AND: AND,

	token.EQ:   EQ,
	token.NE:   EQ,
	token.IS:   EQ,
	token.ISNT: EQ,

	token.LT: COMP,
	token.LE: COMP,
	token.GT: COMP,
	token.GE: COMP,

	token.PLUS:  TERM,
	token.MINUS: TERM,

	token.STAR:  FACTOR,
	token.SLASH: FACTOR,

	token.L_PAREN: CALL,
	token.L_BRACK: CALL,
	token.DOT:     CALL,
}

func newNullStmt() *ast.ExprStmt {
	return &ast.ExprStmt{
		Expr: newNullExpr(),
	}
}

func newNullExpr() ast.Expr {
	return &ast.NullLit{}
}

func newBadDecl() *ast.BadDecl {
	return &ast.BadDecl{}
}

func newBadStmt() *ast.BadStmt {
	return &ast.BadStmt{}
}

/* == error ================================================================= */

type parseError struct {
	Error error
}

func panicParseError(token *token.Token, message string, a ...any) {
	finalMessage := fmt.Sprintf(
		"%s at line %d, column %d",
		fmt.Sprintf(message, a...),
		token.Position.Line,
		token.Position.Column,
	)
	panic(&parseError{Error: errors.New(finalMessage)})
}
