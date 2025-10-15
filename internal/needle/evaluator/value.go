package evaluator

import (
	"errors"
	"fmt"
	"needle/internal/needle/ast"
	"strconv"
	"strings"

	"github.com/fatih/color"
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
	VAL_MODULE   ValueType = "module"

	VAL_VECTOR    ValueType = "vector"
	VAL_MAP       ValueType = "map"
	VAL_EXCEPTION ValueType = "exception"
)

type Value interface {
	Type() ValueType
	Say() string
}

type Null struct{}
type Boolean struct{ Value bool }
type Number struct{ Value float64 }
type String struct{ Value string }

type Function struct {
	Name    string
	Params  []string
	Body    ast.Stmt
	Closure *Env
}

type NativeFunction = func(e *Evaluator, self0 Value, args ...Value) Value
type Native struct {
	Name     string
	Arity    int
	Function NativeFunction
}

type Method struct {
	Function Value
	Self     Value
	IsInit   bool
}

type Module struct {
	Store map[string]Value
}

type Class struct {
	Name  string
	Inits map[string]Value
	Funs  map[string]Value
}

type Instance struct {
	Class  *Class
	Fields map[string]Value
}

type Exception struct {
	Message    string
	StackTrace []Value
}

func (e *Exception) Error() string {
	return fmt.Sprintf(
		"Exception: %s\n%s",
		e.Message,
		sprintTrace(e.StackTrace),
	)
}

type Vector struct{ Elems []Value }
type Map struct{ Pairs *hashTable }

/* == type ================================================================== */

func (n *Null) Type() ValueType      { return VAL_NULL }
func (b *Boolean) Type() ValueType   { return VAL_BOOLEAN }
func (n *Number) Type() ValueType    { return VAL_NUMBER }
func (s *String) Type() ValueType    { return VAL_STRING }
func (f *Function) Type() ValueType  { return VAL_FUNCTION }
func (n *Native) Type() ValueType    { return VAL_NATIVE }
func (m *Method) Type() ValueType    { return VAL_METHOD }
func (m *Module) Type() ValueType    { return VAL_MODULE }
func (c *Class) Type() ValueType     { return VAL_CLASS }
func (i *Instance) Type() ValueType  { return VAL_INSTANCE }
func (e *Exception) Type() ValueType { return VAL_EXCEPTION }
func (v *Vector) Type() ValueType    { return VAL_VECTOR }
func (m *Map) Type() ValueType       { return VAL_MAP }

/* == say =================================================================== */

func (n *Null) Say() string {
	return color.MagentaString("null")
}
func (b *Boolean) Say() string {
	return strconv.FormatBool(b.Value)
}
func (n *Number) Say() string {
	return strconv.FormatFloat(n.Value, 'g', -1, 64)
}
func (s *String) Say() string {
	return fmt.Sprintf("\"%s\"", s.Value)
}
func (f *Function) Say() string {
	return fmt.Sprintf("<function %s %p>", anon(f.Name), f)
}
func (m *Module) Say() string {
	return fmt.Sprintf("<module %p>", m)
}
func (c *Class) Say() string {
	return fmt.Sprintf("<class %s %p>", anon(c.Name), c)
}
func (n *Native) Say() string {
	return fmt.Sprintf("<function %s %p>", anon(n.Name), n)
}
func (m *Method) Say() string {
	return fmt.Sprintf("<method %s of %s>", m.Function.Say(), m.Self.Say())
}
func (i *Instance) Say() string {
	return fmt.Sprintf("<instance %p of %s>", i, i.Class.Say())
}
func (e *Exception) Say() string {
	return fmt.Sprintf("<exception \"%s\" %p>", e.Message, e)
}
func (v *Vector) Say() string {
	return fmt.Sprintf("<vector %p>", v)
}
func (m *Map) Say() string {
	return fmt.Sprintf("<map %p>", m)
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

/* == utils ================================================================= */

func anon(n string) string {
	if n != "" {
		return "'" + n + "'"
	}
	return "(anonymous)"
}

func sprintTrace(trace []Value) string {
	var str strings.Builder
	for _, fun := range trace {
		str.WriteString(fmt.Sprintf("\tin %s\n", fun.Say()))
	}
	return str.String()
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
