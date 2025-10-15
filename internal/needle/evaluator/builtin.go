package evaluator

import "time"

func loadBuiltins(env *Env) {
	for name, native := range newBuiltins() {
		env.Declare(name, native)
	}
}

func newBuiltins() map[string]*Native {
	return map[string]*Native{
		"clock": {
			Name:  "clock",
			Arity: 0,
			Function: func(e *Evaluator, self Value, args ...Value) Value {
				return &Number{Value: float64(time.Now().Unix())}
			},
		},
		"class_of": {
			Name:  "class_of",
			Arity: 1,
			Function: func(e *Evaluator, self Value, args ...Value) Value {
				switch obj := args[0].(type) {
				case *Boolean:
					return e.globals.Classes[CLASS_BOOLEAN]
				case *Number:
					return e.globals.Classes[CLASS_NUMBER]
				case *String:
					return e.globals.Classes[CLASS_STRING]
				case *Vector:
					return e.globals.Classes[CLASS_VECTOR]
				case *Map:
					return e.globals.Classes[CLASS_MAP]
				case *Instance:
					return obj.Class
				}
				return e.globalNull()
			},
		},
	}
}
