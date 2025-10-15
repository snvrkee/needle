package evaluator

import "math"

func newBaseModules() map[string]*Module {
	mods := map[string]*Module{
		"math": newMathModule(),
	}
	return mods
}

func newMathModule() *Module {
	store := map[string]Value{
		"PI": &Number{Value: math.Pi},
		"pow": &Native{
			Name:  "pow",
			Arity: 2,
			Function: func(e *Evaluator, self Value, args ...Value) Value {
				if args[0].Type() != VAL_NUMBER ||
					args[1].Type() != VAL_NUMBER {
					e.panicException("non number agrument")
				}
				return &Number{
					Value: math.Pow(
						args[0].(*Number).Value, args[1].(*Number).Value,
					),
				}
			},
		},
		"sqrt": &Native{
			Name:  "sqrt",
			Arity: 1,
			Function: func(e *Evaluator, self Value, args ...Value) Value {
				if args[0].Type() != VAL_NUMBER {
					e.panicException("non number agrument")
				}
				return &Number{Value: math.Sqrt(args[0].(*Number).Value)}
			},
		},
	}
	return &Module{Store: store}
}
