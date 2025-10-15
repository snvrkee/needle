package ast

import (
	"fmt"
	"needle/internal/needle/token"
	"strconv"
	"strings"
)

type Node interface {
	fmt.Stringer
	Node()
}

type Decl interface {
	Node
	Decl()
}

type Stmt interface {
	Node
	Stmt()
}

type Expr interface {
	Node
	Expr()
}

type Script struct {
	Decls []Decl
}

func (s *Script) Node() {}
func (s *Script) String() string {
	var str strings.Builder
	for i, decl := range s.Decls {
		str.WriteString(decl.String())
		if i != len(s.Decls)-1 {
			str.WriteByte('\n')
		}
	}
	return str.String()
}

/* == declarations ========================================================== */

type BadDecl struct{}

func (bd *BadDecl) Node()          {}
func (bd *BadDecl) Decl()          {}
func (bd *BadDecl) String() string { return "__bad_decl" }

type StmtDecl struct {
	Stmt Stmt
}

func (sd *StmtDecl) Node() {}
func (sd *StmtDecl) Decl() {}
func (sd *StmtDecl) String() string {
	return sd.Stmt.String()
}

type VarDecl struct {
	Name  *Ident
	Right Expr
}

func (vd *VarDecl) Node() {}
func (vd *VarDecl) Decl() {}
func (vd *VarDecl) String() string {
	return fmt.Sprintf(
		"var %s = %s;",
		vd.Name,
		vd.Right,
	)
}

type FunDecl struct {
	Name *Ident
	Fun  *FunLit
}

func (fd *FunDecl) Node() {}
func (fd *FunDecl) Decl() {}
func (fd *FunDecl) String() string {
	return "<fun decl>"
}

type ClassDecl struct {
	Name  *Ident
	Class *ClassLit
}

func (cd *ClassDecl) Node() {}
func (cd *ClassDecl) Decl() {}
func (cd *ClassDecl) String() string {
	return "<class decl>"
}

type ImportDecl struct {
	Path   *StringLit
	Unwrap bool
	Alias  *Ident
}

func (id *ImportDecl) Node() {}
func (id *ImportDecl) Decl() {}
func (id *ImportDecl) String() string {
	return fmt.Sprintf("import %s %s;", id.Alias, id.Path)
}

/* == statements ============================================================ */

type BadStmt struct{}

func (bs *BadStmt) Node()          {}
func (bs *BadStmt) Stmt()          {}
func (bs *BadStmt) String() string { return "__bad_stmt" }

type Block struct {
	Decls []Decl
}

func (b *Block) Node() {}
func (b *Block) Stmt() {}
func (b *Block) String() string {
	var str strings.Builder
	str.WriteString("{")
	for i, decl := range b.Decls {
		str.WriteString(decl.String())
		if i != len(b.Decls)-1 {
			str.WriteByte(' ')
		}
	}
	str.WriteString("}")
	return str.String()
}

type ExprStmt struct {
	Expr Expr
}

func (es *ExprStmt) Node() {}
func (es *ExprStmt) Stmt() {}
func (es *ExprStmt) String() string {
	return fmt.Sprintf(
		"%s;",
		es.Expr,
	)
}

type IfStmt struct {
	Cond Expr
	Then Stmt
	Else Stmt
}

func (is *IfStmt) Node() {}
func (is *IfStmt) Stmt() {}
func (is *IfStmt) String() string {
	return fmt.Sprintf(
		"if (%s) %s else %s",
		is.Cond,
		is.Then,
		is.Else,
	)
}

type WhileStmt struct {
	Cond Expr
	Do   Stmt
}

func (ws *WhileStmt) Node() {}
func (ws *WhileStmt) Stmt() {}
func (ws *WhileStmt) String() string {
	return fmt.Sprintf(
		"while (%s) %s",
		ws.Cond,
		ws.Do,
	)
}

type DoStmt struct {
	Do    Stmt
	While Expr
}

func (ds *DoStmt) Node() {}
func (ds *DoStmt) Stmt() {}
func (ds *DoStmt) String() string {
	return fmt.Sprintf(
		"do %s while (%s);",
		ds.Do,
		ds.While,
	)
}

type ForStmt struct {
	Repeat Stmt
	Init   Decl
	Cond   Expr
	Post   Stmt
}

func (fs *ForStmt) Node() {}
func (fs *ForStmt) Stmt() {}
func (fs *ForStmt) String() string {
	return fmt.Sprintf(
		"for (%s %s; %s) %s",
		fs.Init,
		fs.Cond,
		fs.Post,
		fs.Repeat,
	)
}

type AssignStmt struct {
	Left  Expr
	Right Expr
}

func (as *AssignStmt) Node() {}
func (as *AssignStmt) Stmt() {}
func (as *AssignStmt) String() string {
	return fmt.Sprintf(
		"%s = %s;",
		as.Left,
		as.Right,
	)
}

type SayStmt struct {
	Expr Expr
}

func (ss *SayStmt) Node() {}
func (ss *SayStmt) Stmt() {}
func (ss *SayStmt) String() string {
	return fmt.Sprintf(
		"say %s;",
		ss.Expr,
	)
}

type ReturnStmt struct {
	Value Expr
}

func (rs *ReturnStmt) Node() {}
func (rs *ReturnStmt) Stmt() {}
func (rs *ReturnStmt) String() string {
	return fmt.Sprintf(
		"return %s;",
		rs.Value,
	)
}

type BreakStmt struct{}

func (bs *BreakStmt) Node() {}
func (bs *BreakStmt) Stmt() {}
func (bs *BreakStmt) String() string {
	return "break;"
}

type ContinueStmt struct{}

func (cs *ContinueStmt) Node() {}
func (cs *ContinueStmt) Stmt() {}
func (cs *ContinueStmt) String() string {
	return "continue;"
}

type TryStmt struct {
	Try     Stmt
	Catch   Stmt
	As      *Ident
	Finally Stmt
}

func (ts *TryStmt) Node() {}
func (ts *TryStmt) Stmt() {}
func (ts *TryStmt) String() string {
	return fmt.Sprintf(
		"try %s catch (%s) %s finally %s",
		ts.Try,
		ts.As,
		ts.Catch,
		ts.Finally,
	)
}

type ThrowStmt struct {
	Error Expr
}

func (ts *ThrowStmt) Node() {}
func (ts *ThrowStmt) Stmt() {}
func (ts *ThrowStmt) String() string {
	return fmt.Sprintf(
		"throw %s;",
		ts.Error,
	)
}

/* == expressions =========================================================== */

type Ident struct {
	Name string
}

func (l *Ident) Node()          {}
func (i *Ident) Expr()          {}
func (i *Ident) String() string { return i.Name }

type InfixExpr struct {
	Left  Expr
	Right Expr
	Op    *token.Token
}

func (ie *InfixExpr) Node() {}
func (ie *InfixExpr) Expr() {}
func (ie *InfixExpr) String() string {
	return fmt.Sprintf(
		"(%s %s %s)",
		ie.Left,
		ie.Op.Literal,
		ie.Right,
	)
}

type PrefixExpr struct {
	Right Expr
	Op    *token.Token
}

func (pe *PrefixExpr) Node() {}
func (pe *PrefixExpr) Expr() {}
func (pe *PrefixExpr) String() string {
	return fmt.Sprintf(
		"(%s %s)",
		pe.Op.Literal,
		pe.Right,
	)
}

type CallExpr struct {
	Left      Expr
	Arguments []Expr
}

func (ce *CallExpr) Node() {}
func (ce *CallExpr) Expr() {}
func (ce *CallExpr) String() string {
	var args strings.Builder
	for i, arg := range ce.Arguments {
		args.WriteString(arg.String())
		if i != len(ce.Arguments)-1 {
			args.WriteString(", ")
		}
	}
	return fmt.Sprintf(
		"%s(%s)",
		ce.Left,
		args.String(),
	)
}

type PropExpr struct {
	Left Expr
	Prop *Ident
}

func (pe *PropExpr) Node() {}
func (pe *PropExpr) Expr() {}
func (pe *PropExpr) String() string {
	return fmt.Sprintf(
		"%s.%s",
		pe.Left,
		pe.Prop,
	)
}

type IndexExpr struct {
	Left  Expr
	Index Expr
}

func (ie *IndexExpr) Node() {}
func (ie *IndexExpr) Expr() {}
func (ie *IndexExpr) String() string {
	return fmt.Sprintf(
		"%s[%s]",
		ie.Left,
		ie.Index,
	)
}

type SliceExpr struct {
	Left  Expr
	Start Expr
	End   Expr
}

func (se *SliceExpr) Node() {}
func (se *SliceExpr) Expr() {}
func (se *SliceExpr) String() string {
	return fmt.Sprintf(
		"%s[%s:%s]",
		se.Left,
		se.Start,
		se.End,
	)
}

/* == literals ============================================================== */

type NullLit struct{}

func (nl *NullLit) Node() {}
func (nl *NullLit) Expr() {}
func (nl *NullLit) String() string {
	return "null"
}

type BooleanLit struct {
	Value bool
}

func (bl *BooleanLit) Node() {}
func (bl *BooleanLit) Expr() {}
func (bl *BooleanLit) String() string {
	if bl.Value {
		return "true"
	}
	return "false"
}

type NumberLit struct {
	Value float64
}

func (nl *NumberLit) Node() {}
func (nl *NumberLit) Expr() {}
func (nl *NumberLit) String() string {
	return strconv.FormatFloat(nl.Value, 'g', -1, 64)
}

type StringLit struct {
	Value string
}

func (sl *StringLit) Node() {}
func (sl *StringLit) Expr() {}
func (sl *StringLit) String() string {
	return fmt.Sprintf(
		"\"%s\"",
		sl.Value,
	)
}

type ClassLit struct {
	Inits map[*Ident]*FunLit
	Funs  map[*Ident]*FunLit
}

func (cl *ClassLit) Node() {}
func (cl *ClassLit) Expr() {}
func (cl *ClassLit) String() string {
	var str strings.Builder
	str.WriteString("class{")
	for ident, fun := range cl.Inits {
		lit := fmt.Sprintf(
			"init %s %s",
			ident,
			fun,
		)
		str.WriteString(lit + " ")
	}
	for ident, fun := range cl.Funs {
		lit := fmt.Sprintf(
			"fun %s %s",
			ident,
			fun,
		)
		str.WriteString(lit + " ")
	}
	str.WriteString("}")
	return str.String()
}

type FunLit struct {
	Body   Stmt
	Params []*Ident
}

func (fl *FunLit) Node() {}
func (fl *FunLit) Expr() {}
func (fl *FunLit) String() string {
	var str strings.Builder
	for i, param := range fl.Params {
		str.WriteString(param.String())
		if i != len(fl.Params)-1 {
			str.WriteString(", ")
		}
	}
	params := str.String()
	return fmt.Sprintf(
		"fun(%s) %s",
		params,
		fl.Body,
	)
}

type VectorLit struct {
	Elems []Expr
}

func (vl *VectorLit) Node() {}
func (vl *VectorLit) Expr() {}
func (vl *VectorLit) String() string {
	var str strings.Builder
	str.WriteString("vec{")
	for i, elem := range vl.Elems {
		str.WriteString(elem.String())
		if i != len(vl.Elems)-1 {
			str.WriteString(", ")
		}
	}
	str.WriteString("}")
	return str.String()
}

type MapLit struct {
	Pairs map[Expr]Expr
}

func (ml *MapLit) Node() {}
func (ml *MapLit) Expr() {}
func (ml *MapLit) String() string {
	var str strings.Builder
	str.WriteString("map{")
	i := 0
	for k, v := range ml.Pairs {
		str.WriteString(
			fmt.Sprintf("%s: %s", k, v),
		)
		if i != len(ml.Pairs)-1 {
			str.WriteString(", ")
		}
		i++
	}
	str.WriteString("}")
	return str.String()
}

type SelfLit struct{}

func (sl *SelfLit) Node()          {}
func (sl *SelfLit) Expr()          {}
func (sl *SelfLit) String() string { return "self" }
