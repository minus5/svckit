package jse

type Evaluator interface {
	Eval(string) (interface{}, error)
}
