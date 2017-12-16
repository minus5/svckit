package ers

import "errors"

type Errors struct {
	// map text to error constant
	m map[string]error
}

func New() *Errors {
	return &Errors{
		m: make(map[string]error),
	}
}

// New creates new error
func (e *Errors) New(text string) error {
	err := errors.New(text)
	e.m[text] = err
	return err
}

// Parse pretvara text u typed error
func (e *Errors) Parse(text string) error {
	if text == "" {
		return nil
	}
	if err, ok := e.m[text]; ok {
		return err
	}
	return errors.New(text)
}
