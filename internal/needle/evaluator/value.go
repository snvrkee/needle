package evaluator

import (
	"errors"
	"fmt"
	"needle/internal/needle/ast"
	"strconv"
	"strings"
)

type ValueType string

const (
	VAL_NULL    ValueType = "null"
	VAL_BOOLEAN ValueType = "boolean"
	VAL_NUMBER  ValueType = "number"
	VAL_STRING  ValueType = "string"

	VAL_FUNCTION ValueType = "function"
	VAL_NATIVE   ValueType = "native"
	VAL_METHOD   ValueType = "method"
	VAL_CLASS    ValueType = "class"
	VAL_INSTANCE ValueType = "instance"

	VAL_ARRAY     ValueType = "array"
	VAL_TABLE     ValueType = "table"
	VAL_EXCEPTION ValueType = "exception"
)

type Value interface {
	Type() ValueType
	Say() string
}

type Null struct{}

func (n *Null) Type() ValueType { return VAL_NULL }
func (n *Null) Say() string     { return "null" }

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ValueType { return VAL_BOOLEAN }
func (b *Boolean) Say() string {
	return strconv.FormatBool(b.Value)
}

type Number struct {
	Value float64
}

func (n *Number) Type() ValueType {
	return VAL_NUMBER
}
func (n *Number) Say() string {
	return strconv.FormatFloat(n.Value, 'g', -1, 64)
}

type String struct {
	Value string
}

func (s *String) Type() ValueType { return VAL_STRING }
func (s *String) Say() string {
	return fmt.Sprintf("\"%s\"", s.Value)
}

type Function struct {
	Params  []string
	Body    ast.Stmt
	Closure *Env
}

func (f *Function) Type() ValueType {
	return VAL_FUNCTION
}
func (f *Function) Say() string {
	return fmt.Sprintf("<function %p>", f)
}

type NativeFunction = func(e *Evaluator, self0 Value, args ...Value) Value

type Native struct {
	Arity    int
	Function NativeFunction
}

func (n *Native) Type() ValueType {
	return VAL_NATIVE
}
func (n *Native) Say() string {
	return fmt.Sprintf("<function %p>", n)
}

type Method struct {
	Function Value
	Self     Value
	IsInit   bool
}

func (m *Method) Type() ValueType {
	return VAL_METHOD
}
func (m *Method) Say() string {
	return fmt.Sprintf("<function %p>", m)
}

type Class struct {
	Inits map[string]Value
	Funs  map[string]Value
}

func (c *Class) Type() ValueType { return VAL_CLASS }
func (c *Class) Say() string {
	return fmt.Sprintf("<class %p>", c)
}

type Instance struct {
	Class  *Class
	Fields map[string]Value
}

func (i *Instance) Type() ValueType { return VAL_INSTANCE }
func (i *Instance) Say() string {
	return fmt.Sprintf("<instance %p of class %p>", i, i.Class)
}

type Exception struct {
	Message    string
	StackTrace []Value
}

func (e *Exception) Type() ValueType { return VAL_EXCEPTION }
func (e *Exception) Say() string {
	return fmt.Sprintf("<exception %p>", e)
}
func (e *Exception) Error() string {
	var trace strings.Builder
	for _, fun := range e.StackTrace {
		trace.WriteString("\t")
		trace.WriteString(fun.Say())
		trace.WriteString("\n")
	}
	return fmt.Sprintf(
		"%s\n%s",
		e.Message,
		trace.String(),
	)
}

type Array struct {
	Elems []Value
}

func (a *Array) Type() ValueType { return VAL_ARRAY }
func (a *Array) Say() string {
	return fmt.Sprintf("<array %p>", a)
}

type Table struct {
	Pairs *hashTable
}

func (t *Table) Type() ValueType { return VAL_TABLE }
func (t *Table) Say() string {
	return fmt.Sprintf("<table %p>", t)
}

/* == signal ================================================================ */

type SignalType int

const (
	SIG_RETURN SignalType = iota
	SIG_BREAK
	SIG_CONTINUE
)

type Signal struct {
	Type  SignalType
	Value Value
}

/* == hash table ============================================================ */

type hashTable struct {
	numMap map[float64]Value
	strMap map[string]Value
}

func newHashTable() *hashTable {
	return &hashTable{
		numMap: map[float64]Value{},
		strMap: map[string]Value{},
	}
}

func (ht *hashTable) Get(key Value) (Value, error) {
	switch key := key.(type) {
	case *Number:
		if v, ok := ht.numMap[key.Value]; ok {
			return v, nil
		}
		return nil, errors.New("missing key")
	case *String:
		if v, ok := ht.strMap[key.Value]; ok {
			return v, nil
		}
		return nil, errors.New("missing key")
	default:
		return nil, errors.New("unhashable type")
	}
}

func (ht *hashTable) Delete(key Value) (bool, error) {
	switch key := key.(type) {
	case *Number:
		_, ok := ht.numMap[key.Value]
		delete(ht.numMap, key.Value)
		return ok, nil
	case *String:
		_, ok := ht.strMap[key.Value]
		delete(ht.strMap, key.Value)
		return ok, nil
	default:
		return false, errors.New("unhashable type")
	}
}

func (ht *hashTable) Set(key Value, value Value) (bool, error) {
	switch key := key.(type) {
	case *Number:
		_, ok := ht.numMap[key.Value]
		ht.numMap[key.Value] = value
		return ok, nil
	case *String:
		_, ok := ht.strMap[key.Value]
		ht.strMap[key.Value] = value
		return ok, nil
	default:
		return false, errors.New("unhashable type")
	}
}

func (ht *hashTable) Size() int {
	return len(ht.strMap) + len(ht.numMap)
}

func (ht *hashTable) Keys() []Value {
	keys := []Value{}
	for k := range ht.numMap {
		keys = append(keys, &Number{Value: k})
	}
	for k := range ht.strMap {
		keys = append(keys, &String{Value: k})
	}
	return keys
}

func (ht *hashTable) Values() []Value {
	vals := []Value{}
	for _, v := range ht.numMap {
		vals = append(vals, v)
	}
	for _, v := range ht.strMap {
		vals = append(vals, v)
	}
	return vals
}
