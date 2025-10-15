package evaluator

import "strconv"

const (
	CLASS_BOOLEAN   = "Boolean"
	CLASS_NUMBER    = "Number"
	CLASS_STRING    = "String"
	CLASS_VECTOR    = "Vector"
	CLASS_MAP       = "Map"
	CLASS_EXCEPTION = "Exception"
)

func newBooleanClass() *Class {
	inits := map[string]Value{}
	funs := map[string]Value{
		"to_string": &Native{
			Name:  "to_string",
			Arity: 0,
			Function: func(e *Evaluator, self0 Value, args ...Value) Value {
				self := self0.(*Boolean)
				return &String{Value: strconv.FormatBool(self.Value)}
			},
		},
	}
	return &Class{Funs: funs, Inits: inits}
}

func newNumberClass() *Class {
	funs := map[string]Value{
		"to_string": &Native{
			Name:  "to_string",
			Arity: 0,
			Function: func(e *Evaluator, self0 Value, args ...Value) Value {
				self := self0.(*Number)
				return &String{
					Value: strconv.FormatFloat(self.Value, 'g', -1, 64),
				}
			},
		},
	}
	inits := map[string]Value{}
	return &Class{Funs: funs, Inits: inits}
}

func newStringClass() *Class {
	funs := map[string]Value{
		"reverse": &Native{
			Name:  "reverse",
			Arity: 0,
			Function: func(e *Evaluator, self0 Value, args ...Value) Value {
				self := self0.(*String)
				rev := []rune(self.Value)
				for i := 0; i < len(rev)/2; i++ {
					alt := len(rev) - i - 1
					rev[i], rev[alt] = rev[alt], rev[i]
				}
				return &String{Value: string(rev)}
			},
		},
		"to_upper_case": &Native{
			Name:  "to_upper_case",
			Arity: 0,
			Function: func(e *Evaluator, self0 Value, args ...Value) Value {
				self := self0.(*String)
				up := []rune(self.Value)
				for i := range len(up) {
					if 'a' <= up[i] && up[i] <= 'z' {
						up[i] += 'A' - 'a'
					}
				}
				return &String{Value: string(up)}
			},
		},
		"to_lower_case": &Native{
			Name:  "to_lower_case",
			Arity: 0,
			Function: func(e *Evaluator, self0 Value, args ...Value) Value {
				self := self0.(*String)
				low := []rune(self.Value)
				for i := range len(low) {
					if 'A' <= low[i] && low[i] <= 'Z' {
						low[i] -= 'A' - 'a'
					}
				}
				return &String{Value: string(low)}
			},
		},
	}
	inits := map[string]Value{}
	return &Class{Inits: inits, Funs: funs}
}

func newVectorClass() *Class {
	funs := map[string]Value{
		"push": &Native{
			Name:  "push",
			Arity: 1,
			Function: func(e *Evaluator, self0 Value, args ...Value) Value {
				self := self0.(*Vector)
				self.Elems = append(self.Elems, args[0])
				return e.globalNull()
			},
		},
		"pop": &Native{
			Name:  "pop",
			Arity: 0,
			Function: func(e *Evaluator, self0 Value, args ...Value) Value {
				self := self0.(*Vector)
				elem := self.Elems[len(self.Elems)-1]
				self.Elems = self.Elems[:len(self.Elems)-1]
				return elem
			},
		},
		"length": &Native{
			Name:  "length",
			Arity: 0,
			Function: func(e *Evaluator, self0 Value, args ...Value) Value {
				self := self0.(*Vector)
				return &Number{Value: float64(len(self.Elems))}
			},
		},
	}
	inits := map[string]Value{}
	return &Class{Inits: inits, Funs: funs}
}

func newMapClass() *Class {
	funs := map[string]Value{
		"size": &Native{
			Name:  "size",
			Arity: 0,
			Function: func(e *Evaluator, self0 Value, args ...Value) Value {
				self := self0.(*Map)
				return &Number{Value: float64(self.Pairs.Size())}
			},
		},
		"keys": &Native{
			Name:  "keys",
			Arity: 0,
			Function: func(e *Evaluator, self0 Value, args ...Value) Value {
				self := self0.(*Map)
				return &Vector{Elems: self.Pairs.Keys()}
			},
		},
		"values": &Native{
			Name:  "values",
			Arity: 0,
			Function: func(e *Evaluator, self0 Value, args ...Value) Value {
				self := self0.(*Map)
				return &Vector{Elems: self.Pairs.Values()}
			},
		},
	}
	inits := map[string]Value{}
	return &Class{Inits: inits, Funs: funs}
}

func newExceptionClass() *Class {
	funs := map[string]Value{
		"message": &Native{
			Name:  "message",
			Arity: 0,
			Function: func(e *Evaluator, self0 Value, args ...Value) Value {
				self := self0.(*Exception)
				return &String{Value: self.Message}
			},
		},
	}
	inits := map[string]Value{}
	return &Class{Inits: inits, Funs: funs}
}

func newBaseClasses() map[string]*Class {
	cs := map[string]*Class{
		CLASS_BOOLEAN:   newBooleanClass(),
		CLASS_NUMBER:    newNumberClass(),
		CLASS_STRING:    newStringClass(),
		CLASS_VECTOR:    newVectorClass(),
		CLASS_MAP:       newMapClass(),
		CLASS_EXCEPTION: newExceptionClass(),
	}
	for name, cls := range cs {
		cls.Name = name
	}
	return cs
}
