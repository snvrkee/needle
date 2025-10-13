package evaluator

import (
	"errors"
	"fmt"
	"needle/internal/needle/ast"
	"needle/internal/needle/token"
	"needle/internal/pkg"
	"slices"
)

type globals struct {
	Null    *Null
	True    *Boolean
	False   *Boolean
	Classes map[string]*Class
}

type Evaluator struct {
	env       *Env
	callStack *pkg.Stack[Value]
	globals   *globals
}

func New() *Evaluator {
	env := newEnv(nil)
	loadBuiltins(env)
	classes := newBaseClasses()
	for name, class := range classes {
		env.Declare(name, class)
	}
	return &Evaluator{
		env:       env,
		callStack: pkg.NewStack[Value](),
		globals: &globals{
			Null:    &Null{},
			True:    &Boolean{Value: true},
			False:   &Boolean{Value: false},
			Classes: classes,
		},
	}
}

func (e *Evaluator) EvalScript(script *ast.Script) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if exc, ok := r.(*Exception); ok {
				err = exc
				return
			}
			panic(r)
		}
	}()
	defer func() {
		if r := recover(); r != nil {
			if s, ok := r.(*Signal); ok {
				switch s.Type {
				case SIG_RETURN:
					err = errors.New("'return' outside function")
				case SIG_BREAK:
					err = errors.New("'break' outside loop")
				case SIG_CONTINUE:
					err = errors.New("'continue' outside loop")
				default:
					panic(r)
				}
				return
			}
			panic(r)
		}
	}()
	e.Eval(script)
	return nil
}

func (e *Evaluator) Eval(node ast.Node) Value {
	switch node := node.(type) {
	case *ast.Script:
		return e.evalScript(node)
	case *ast.Block:
		return e.evalBlock(node)

	case *ast.VarDecl:
		return e.evalVarDecl(node)
	case *ast.FunDecl:
		return e.evalFunDecl(node)
	case *ast.ClassDecl:
		return e.evalClassDecl(node)
	case *ast.StmtDecl:
		return e.Eval(node.Stmt)

	case *ast.SayStmt:
		return e.evalSayStmt(node)
	case *ast.IfStmt:
		return e.evalIfStmt(node)
	case *ast.ForStmt:
		return e.evalForStmt(node)
	case *ast.WhileStmt:
		return e.evalWhileStmt(node)
	case *ast.DoStmt:
		return e.evalDoStmt(node)
	case *ast.ExprStmt:
		return e.Eval(node.Expr)
	case *ast.AssignStmt:
		return e.evalAssignStmt(node)
	case *ast.TryStmt:
		return e.evalTryStmt(node)
	case *ast.ThrowStmt:
		return e.evalThrowStmt(node)
	case *ast.ReturnStmt:
		panic(&Signal{Type: SIG_RETURN, Value: e.Eval(node.Value)})
	case *ast.BreakStmt:
		panic(&Signal{Type: SIG_BREAK})
	case *ast.ContinueStmt:
		panic(&Signal{Type: SIG_CONTINUE})

	case *ast.InfixExpr:
		return e.evalInfixExpr(node)
	case *ast.PrefixExpr:
		return e.evalPrefixExpr(node)
	case *ast.CallExpr:
		return e.evalCallExpr(node)
	case *ast.PropExpr:
		return e.evalPropExpr(node)
	case *ast.IndexExpr:
		return e.evalIndexExpr(node)
	case *ast.SliceExpr:
		return e.evalSliceExpr(node)

	case *ast.Ident:
		val, err := e.env.Get(node.Name)
		if err != nil {
			e.panicException(err)
		}
		return val
	case *ast.SelfLit:
		if self := e.env.GetSelf(); self != nil {
			return self
		}
		e.panicException("'self' is undefined")
	case *ast.NullLit:
		return e.globalNull()
	case *ast.BooleanLit:
		return e.globalBoolean(node.Value)
	case *ast.NumberLit:
		return &Number{Value: node.Value}
	case *ast.StringLit:
		return &String{Value: node.Value}
	case *ast.FunLit:
		return e.evalFunLit(node)
	case *ast.ClassLit:
		return e.evalClassLit(node)
	case *ast.ArrayLit:
		return e.evalArrayLit(node)
	case *ast.TableLit:
		return e.evalTableLit(node)
	default:
		panic(fmt.Sprintf("unknown node: %s", node.String()))
	}
	return nil
}

func (e *Evaluator) evalScript(node *ast.Script) Value {
	for _, decl := range node.Decls {
		e.Eval(decl)
	}
	return e.globalNull()
}

func (e *Evaluator) evalBlock(node *ast.Block) Value {
	oldEnv := e.env
	defer func() { e.env = oldEnv }()
	e.env = newEnv(oldEnv)
	for _, decl := range node.Decls {
		e.Eval(decl)
	}
	return e.globalNull()
}

/* == eval daclaration ====================================================== */

func (e *Evaluator) evalVarDecl(node *ast.VarDecl) Value {
	if err := e.env.Declare(node.Name.Name, e.Eval(node.Right)); err != nil {
		e.panicException(err)
	}
	return e.globalNull()
}

func (e *Evaluator) evalFunDecl(node *ast.FunDecl) Value {
	fun := e.evalFunLit(node.Func)
	if err := e.env.Declare(node.Name.Name, fun); err != nil {
		e.panicException(err)
	}
	return e.globalNull()
}

func (e *Evaluator) evalClassDecl(node *ast.ClassDecl) Value {
	class := e.evalClassLit(node.Class)
	if err := e.env.Declare(node.Name.Name, class); err != nil {
		e.panicException(err)
	}
	return e.globalNull()
}

/* == eval statement ======================================================== */

func (e *Evaluator) evalSayStmt(node *ast.SayStmt) Value {
	fmt.Println(e.Eval(node.Expr).Say())
	return e.globalNull()
}

func (e *Evaluator) evalIfStmt(node *ast.IfStmt) Value {
	if toBoolean(e.Eval(node.Cond)) {
		return e.Eval(node.Then)
	}
	return e.Eval(node.Else)
}

func (e *Evaluator) evalForStmt(node *ast.ForStmt) (value Value) {
	value = e.globalNull()
	oldEnv := e.env
	e.env = newEnv(oldEnv)
	defer func() { e.env = oldEnv }()
	e.Eval(node.Init)
	cond := e.Eval(node.Cond)
	defer catchBreak()
	for toBoolean(cond) {
		e.runLoop(node.Repeat)
		e.Eval(node.Post)
		cond = e.Eval(node.Cond)
	}
	return
}

func (e *Evaluator) evalWhileStmt(node *ast.WhileStmt) (value Value) {
	cond := e.Eval(node.Cond)
	value = e.globalNull()
	defer catchBreak()
	for toBoolean(cond) {
		e.runLoop(node.Do)
		cond = e.Eval(node.Cond)
	}
	return
}

func (e *Evaluator) evalDoStmt(node *ast.DoStmt) (value Value) {
	var cond Value = e.globalBoolean(true)
	value = e.globalNull()
	defer catchBreak()
	for toBoolean(cond) {
		e.runLoop(node.Do)
		cond = e.Eval(node.While)
	}
	return
}

func (e *Evaluator) evalTryStmt(node *ast.TryStmt) Value {
	value, excTry := pkg.Catch[ast.Node, Value, *Exception](e.Eval, node.Try)
	var excCatch *Exception
	if excTry != nil {
		oldEnv := e.env
		e.env = newEnv(oldEnv)
		defer func() { e.env = oldEnv }()
		e.env.Declare(node.As.Name, excTry)
		_, excCatch = pkg.Catch[ast.Node, Value, *Exception](e.Eval, node.Catch)
	}
	_, excFin := pkg.Catch[ast.Node, Value, *Exception](e.Eval, node.Finally)

	if excFin != nil {
		panic(excFin)
	} else if excCatch != nil {
		panic(excCatch)
	}
	return value
}

func (e *Evaluator) evalThrowStmt(node *ast.ThrowStmt) Value {
	e.panicException(e.Eval(node.Error).Say())
	return nil
}

func (e *Evaluator) evalAssignStmt(node *ast.AssignStmt) Value {
	right := e.Eval(node.Right)

	switch left := node.Left.(type) {
	case *ast.Ident: // name = value;
		if err := e.env.Set(left.Name, right); err != nil {
			e.panicException(err)
		}
	case *ast.PropExpr: // obj.prop = value;
		e.propAssign(left, right)
	case *ast.IndexExpr: // obj[index] = value;
		e.indexAssign(left, right)
	default:
		e.panicException("can't assign to")
	}
	return e.globalNull()
}

func (e *Evaluator) propAssign(left *ast.PropExpr, right Value) {
	prop := left.Prop.Name

	if _, isSelf := left.Left.(*ast.SelfLit); isSelf {
		self := e.env.GetSelf()
		if self == nil {
			e.panicException("'self' is undefined")
		}
		self.(*Instance).Fields[prop] = right
		return
	}

	obj := e.Eval(left.Left)
	switch obj.(type) {
	case *Instance:
		panic("not done yet")
	default:
		panic("not done yet")
	}
}

func (e *Evaluator) indexAssign(left *ast.IndexExpr, right Value) {
	index := e.Eval(left.Index)
	obj := e.Eval(left.Left)

	switch obj := obj.(type) {
	case *Array:
		idx, err := checkIndex(index, len(obj.Elems))
		if err != nil {
			e.panicException(err)
		}
		obj.Elems[idx] = right
	case *Table:
		_, err := obj.Pairs.Set(index, right)
		if err != nil {
			e.panicException(err)
		}
	case *Instance:
		panic("not done yet")
	default:
		e.panicException("index assign is not supported")
	}
}

/* == eval expression ======================================================= */

func (e Evaluator) evalPrefixExpr(node *ast.PrefixExpr) Value {
	right := e.Eval(node.Right)

	if node.Op.Type == token.WOW {
		return &Boolean{
			Value: !toBoolean(right),
		}
	}

	if node.Op.Type == token.PLUS ||
		node.Op.Type == token.MINUS {
		if right.Type() != VAL_NUMBER {
			e.panicException("expected 'number', got '%s'", right.Type())
		}
		if node.Op.Type == token.MINUS {
			return &Number{Value: -right.(*Number).Value}
		}
		return &Number{Value: +right.(*Number).Value}
	}

	panic("unknown prefix operator")
}

func (e *Evaluator) evalInfixExpr(node *ast.InfixExpr) Value {
	left := e.Eval(node.Left)
	right := e.Eval(node.Right)

	switch node.Op.Type {
	case token.IS:
		return &Boolean{Value: right == left}
	case token.ISNT:
		return &Boolean{Value: right != left}
	case token.OR:
		if toBoolean(left) {
			return left
		}
		return right
	case token.AND:
		if toBoolean(left) {
			return right
		}
		return left
	}

	var f binOp
	var ok bool
	switch left.(type) {
	case *Boolean:
		f, ok = boolBinOps[node.Op.Type]
	case *Number:
		f, ok = numBinOps[node.Op.Type]
	case *String:
		f, ok = strBinOps[node.Op.Type]
	default:
		e.panicException("unsupported type")
	}
	if !ok {
		e.panicException("unsupported operator for type")
	}
	res, err := f(left, right)
	if err != nil {
		e.panicException(err)
	}
	return res
}

func (e *Evaluator) evalCallExpr(node *ast.CallExpr) Value {
	left := e.Eval(node.Left)
	args := e.evalExprs(node.Arguments)
	var fun, self Value
	var isInit bool
	switch left := left.(type) {
	case *Method:
		self = left.Self
		fun = left.Function
		isInit = left.IsInit
	case *Function, *Native:
		self = nil
		fun = left
		isInit = false
	default:
		e.panicException("call not callable")
	}
	value := e.runCall(fun, self, args)
	if isInit {
		return self
	}
	return value
}

func (e *Evaluator) evalPropExpr(node *ast.PropExpr) Value {
	left := e.Eval(node.Left)
	prop := node.Prop.Name
	var className string

	switch left := left.(type) {
	case *Class:
		init, ok := left.Inits[prop]
		if !ok {
			e.panicException("missing initializer")
		}
		self := &Instance{
			Class:  left,
			Fields: map[string]Value{},
		}
		return &Method{
			Function: init,
			Self:     self,
			IsInit:   true,
		}
	case *Instance:
		if _, isSelf := node.Left.(*ast.SelfLit); isSelf {
			value, ok := left.Fields[prop]
			if ok {
				return value
			}
			pub, ok := left.Class.Funs[prop]
			if ok {
				return &Method{
					Function: pub,
					Self:     left,
					IsInit:   false,
				}
			}
			e.panicException("missing field or method")
		}
		if fun, ok := left.Class.Funs[prop]; ok {
			return &Method{
				Function: fun,
				Self:     left,
				IsInit:   false,
			}
		}
		e.panicException("missing property")
	case *String:
		className = CLASS_STRING
	case *Number:
		className = CLASS_NUMBER
	case *Array:
		className = CLASS_ARRAY
	case *Table:
		className = CLASS_TABLE
	case *Exception:
		className = CLASS_EXCEPTION
	default:
		panic("getting property from unsupported type")
	}
	class := e.globals.Classes[className]
	f, ok := class.Funs[prop]
	if !ok {
		e.panicException("missing property")
	}
	return &Method{Function: f, Self: left, IsInit: false}
}

func (e *Evaluator) evalIndexExpr(node *ast.IndexExpr) Value {
	left := e.Eval(node.Left)
	index := e.Eval(node.Index)

	switch left := left.(type) {
	case *Array:
		intIndex, err := checkIndex(index, len(left.Elems))
		if err != nil {
			e.panicException(err)
		}
		return left.Elems[intIndex]
	case *Table:
		val, err := left.Pairs.Get(index)
		if err != nil {
			e.panicException(err)
		}
		return val
	case *String:
		chars := []rune(left.Value)
		intIndex, err := checkIndex(index, len(chars))
		if err != nil {
			e.panicException(err)
		}
		return &String{Value: string(chars[intIndex])}
	}
	e.panicException("type not supports index access")
	return nil
}

func (e *Evaluator) evalSliceExpr(node *ast.SliceExpr) Value {
	left := e.Eval(node.Left)
	start := e.Eval(node.Start)
	end := e.Eval(node.End)

	switch left := left.(type) {
	case *Array:
		intStart, intEnd, err := checkSlice(start, end, len(left.Elems))
		if err != nil {
			e.panicException(err)
		}
		return &Array{Elems: slices.Clone(left.Elems[intStart:intEnd])}
	case *String:
		chars := []rune(left.Value)
		intStart, intEnd, err := checkSlice(start, end, len(chars))
		if err != nil {
			e.panicException(err)
		}
		return &String{Value: string(chars[intStart:intEnd])}
	default:
		e.panicException("type not supports slice")
	}
	return nil
}

/* == eval literal ========================================================== */

func (e *Evaluator) evalFunLit(node *ast.FunLit) Value {
	params, _ := pkg.SliceMap(
		node.Params,
		func(e *ast.Ident) (string, error) {
			return e.Name, nil
		},
	)
	return &Function{
		Closure: e.env,
		Body:    node.Body,
		Params:  params,
	}
}

func (e *Evaluator) evalArrayLit(node *ast.ArrayLit) Value {
	arr := &Array{Elems: []Value{}}
	for _, expr := range node.Elems {
		arr.Elems = append(arr.Elems, e.Eval(expr))
	}
	return arr
}

func (e *Evaluator) evalTableLit(node *ast.TableLit) Value {
	table := &Table{Pairs: newHashTable()}
	for kExpr, vExpr := range node.Pairs {
		table.Pairs.Set(e.Eval(kExpr), e.Eval(vExpr))
	}
	return table
}

func (e *Evaluator) evalClassLit(node *ast.ClassLit) Value {
	class := &Class{}
	fmm := func(
		ident *ast.Ident, lit *ast.FunLit,
	) (
		string, Value, error,
	) {
		return ident.Name, e.Eval(lit), nil
	}
	class.Inits, _ = pkg.MapMap(node.Inits, fmm)
	class.Funs, _ = pkg.MapMap(node.Funs, fmm)
	return class
}

/* == utils ================================================================= */

func (e *Evaluator) assertArgsLength(arity, args int) {
	if arity != args {
		e.panicException(
			"expected %d arguments, got %d",
			arity,
			args,
		)
	}
}

func (e *Evaluator) evalExprs(exprs []ast.Expr) []Value {
	vals := []Value{}
	for _, expr := range exprs {
		vals = append(vals, e.Eval(expr))
	}
	return vals
}

func checkIndex(index0 Value, length int) (int, error) {
	index, ok := index0.(*Number)
	if !ok {
		return 0, errors.New("non number index")
	}
	intIndex := int(index.Value)
	if intIndex < 0 || intIndex >= length {
		return 0, errors.New("index out of range")
	}
	return intIndex, nil
}

func checkSlice(start0, end0 Value, length int) (int, int, error) {
	start, ok1 := start0.(*Number)
	end, ok2 := end0.(*Number)
	if !ok1 || !ok2 {
		return 0, 0, errors.New("non number index")
	}
	intStart := int(start.Value)
	intEnd := int(end.Value)
	if (intStart < 0 || intStart >= length) ||
		(intEnd < 0 || intEnd > length) ||
		intStart > intEnd {
		return 0, 0, errors.New("index out of range")
	}
	return intStart, intEnd, nil
}

func (e *Evaluator) runLoop(loop ast.Stmt) {
	defer catchContinue()
	e.Eval(loop)
}

func (e *Evaluator) runCall(
	fun Value, self Value, args []Value,
) (value Value) {
	e.callStack.Push(fun)
	defer e.callStack.Pop()
	switch fun := fun.(type) {
	case *Function:
		oldEnv := e.env
		e.env = newEnv(fun.Closure)
		defer func() { e.env = oldEnv }()
		e.env.SetSelf(self)
		e.assertArgsLength(len(fun.Params), len(args))
		for i, arg := range args {
			e.env.Declare(fun.Params[i], arg)
		}
		defer catchReturn(&value)
		return e.Eval(fun.Body)
	case *Native:
		e.assertArgsLength(fun.Arity, len(args))
		return fun.Function(e, self, args...)
	default:
		panic("unknown function type")
	}
}

func catchReturn(value *Value) {
	if r := recover(); r != nil {
		if s, ok := r.(*Signal); ok {
			if s.Type == SIG_RETURN {
				*value = s.Value
				return
			}
		}
		panic(r)
	}
}

func catchBreak() {
	if r := recover(); r != nil {
		if s, ok := r.(*Signal); ok {
			if s.Type == SIG_BREAK {
				return
			}
		}
		panic(r)
	}
}

func catchContinue() {
	if r := recover(); r != nil {
		if s, ok := r.(*Signal); ok {
			if s.Type == SIG_CONTINUE {
				return
			}
		}
		panic(r)
	}
}

func toBoolean(value Value) bool {
	if value.Type() == VAL_NULL {
		return false
	}
	if value.Type() == VAL_BOOLEAN {
		return value.(*Boolean).Value
	}
	return true
}

func (e *Evaluator) panicException(message any, a ...any) {
	st := e.callStack.Shot()
	nst := []Value{}
	for f, err := st.Pop(); err == nil; f, err = st.Pop() {
		nst = append(nst, f)
	}
	msg0 := fmt.Sprintf("%s", message)
	msg := fmt.Sprintf(msg0, a...)
	panic(&Exception{
		Message:    msg,
		StackTrace: nst,
	})
}

/* == new value ============================================================= */

func (e *Evaluator) globalNull() *Null {
	return e.globals.Null
}

func (e *Evaluator) globalBoolean(value bool) *Boolean {
	if value {
		return e.globals.True
	}
	return e.globals.False
}

/* == bin ops =============================================================== */

type binOp func(Value, Value) (Value, error)

var boolBinOps = map[token.TokenType]binOp{
	token.EQ: func(v1, v2 Value) (Value, error) {
		if v2.Type() != VAL_BOOLEAN {
			return &Boolean{Value: false}, nil
		}
		return &Boolean{Value: v1.(*Boolean).Value == v2.(*Boolean).Value}, nil
	},
	token.NE: func(v1, v2 Value) (Value, error) {
		if v2.Type() != VAL_BOOLEAN {
			return &Boolean{Value: true}, nil
		}
		return &Boolean{Value: v1.(*Boolean).Value != v2.(*Boolean).Value}, nil
	},
}

var strBinOps = map[token.TokenType]binOp{
	token.PLUS: func(v1, v2 Value) (Value, error) {
		if v2.Type() != VAL_STRING {
			return nil, errors.New("expected string")
		}
		return &String{Value: v1.(*String).Value + v2.(*String).Value}, nil
	},
	token.EQ: func(v1, v2 Value) (Value, error) {
		if v2.Type() != VAL_STRING {
			return &Boolean{Value: false}, nil
		}
		return &Boolean{Value: v1.(*String).Value == v2.(*String).Value}, nil
	},
	token.NE: func(v1, v2 Value) (Value, error) {
		if v2.Type() != VAL_STRING {
			return &Boolean{Value: true}, nil
		}
		return &Boolean{Value: v1.(*String).Value != v2.(*String).Value}, nil
	},
}

var numBinOps = map[token.TokenType]binOp{
	token.PLUS: func(v1, v2 Value) (Value, error) {
		if v2.Type() != VAL_NUMBER {
			return nil, errors.New("expected number")
		}
		return &Number{Value: v1.(*Number).Value + v2.(*Number).Value}, nil
	},
	token.MINUS: func(v1, v2 Value) (Value, error) {
		if v2.Type() != VAL_NUMBER {
			return nil, errors.New("expected number")
		}
		return &Number{Value: v1.(*Number).Value - v2.(*Number).Value}, nil
	},
	token.STAR: func(v1, v2 Value) (Value, error) {
		if v2.Type() != VAL_NUMBER {
			return nil, errors.New("expected number")
		}
		return &Number{Value: v1.(*Number).Value * v2.(*Number).Value}, nil
	},
	token.SLASH: func(v1, v2 Value) (Value, error) {
		if v2.Type() != VAL_NUMBER {
			return nil, errors.New("expected number")
		}
		return &Number{Value: v1.(*Number).Value / v2.(*Number).Value}, nil
	},
	token.EQ: func(v1, v2 Value) (Value, error) {
		if v2.Type() != VAL_NUMBER {
			return &Boolean{Value: false}, nil
		}
		return &Boolean{Value: v1.(*Number).Value == v2.(*Number).Value}, nil
	},
	token.LT: func(v1, v2 Value) (Value, error) {
		if v2.Type() != VAL_NUMBER {
			return nil, errors.New("expected number")
		}
		return &Boolean{Value: v1.(*Number).Value < v2.(*Number).Value}, nil
	},
	token.LE: func(v1, v2 Value) (Value, error) {
		if v2.Type() != VAL_NUMBER {
			return nil, errors.New("expected number")
		}
		return &Boolean{Value: v1.(*Number).Value <= v2.(*Number).Value}, nil
	},
	token.NE: func(v1, v2 Value) (Value, error) {
		if v2.Type() != VAL_NUMBER {
			return &Boolean{Value: true}, nil
		}
		return &Boolean{Value: v1.(*Number).Value != v2.(*Number).Value}, nil
	},
	token.GT: func(v1, v2 Value) (Value, error) {
		if v2.Type() != VAL_NUMBER {
			return nil, errors.New("expected number")
		}
		return &Boolean{Value: v1.(*Number).Value > v2.(*Number).Value}, nil
	},
	token.GE: func(v1, v2 Value) (Value, error) {
		if v2.Type() != VAL_NUMBER {
			return nil, errors.New("expected number")
		}
		return &Boolean{Value: v1.(*Number).Value >= v2.(*Number).Value}, nil
	},
}
