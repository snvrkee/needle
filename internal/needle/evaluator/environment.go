package evaluator

import (
	"errors"
)

var (
	errVarAlreadyExists = errors.New("variable already exists")
	errVarNotExists     = errors.New("variable not exists")
)

type Env struct {
	store map[string]Value
	outer *Env
	self  Value
}

func newEnv(outer *Env) *Env {
	return &Env{
		store: make(map[string]Value, 8),
		outer: outer,
		self:  nil,
	}
}

func (e *Env) Declare(name string, value Value) error {
	if _, exists := e.store[name]; exists {
		return errVarAlreadyExists
	}
	e.store[name] = value
	return nil
}

func (e *Env) Get(name string) (Value, error) {
	v, exists := e.store[name]
	if exists {
		return v, nil
	}
	if e.outer != nil {
		return e.outer.Get(name)
	}
	return nil, errVarNotExists
}

func (e *Env) Set(name string, value Value) error {
	if _, exists := e.store[name]; exists {
		e.store[name] = value
		return nil
	}
	if e.outer != nil {
		return e.outer.Set(name, value)
	}
	return errVarNotExists
}

func (e *Env) GetSelf() Value {
	if e.self != nil {
		return e.self
	}
	if e.outer != nil {
		return e.outer.GetSelf()
	}
	return nil
}

func (e *Env) SetSelf(self Value) {
	e.self = self
}
