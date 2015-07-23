package v8

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/idada/v8.go"
)

type V8 struct {
	ctx    *v8.Context
	script *v8.Script
}

func NewFromFile(jsFile string) (*V8, error) {
	buf, err := ioutil.ReadFile(jsFile)
	if err != nil {
		return nil, err
	}
	return New(buf)
}

func New(buf []byte) (*V8, error) {
	engine := v8.NewEngine()
	global := engine.NewObjectTemplate()
	global.Bind("_console_log", func(args ...interface{}) {
		msg := fmt.Sprintf("js console: ")
		for i := 0; i < len(args); i++ {
			msg += fmt.Sprintf("%v ", args[i])
		}
		log.Printf(msg)
	})
	return &V8{
		ctx:    engine.NewContext(global),
		script: engine.Compile(buf, nil, nil),
	}, nil
}

func (t *V8) Eval(src string) (interface{}, error) {
	ret := ""
	exception := ""
	t.ctx.Scope(func(cs v8.ContextScope) {
		exception = cs.TryCatch(true, func() {
			cs.Run(t.script)
			cs.Eval(`this.console = { "log": function(args) { _console_log.apply(null, arguments); }};`)
			e := cs.Eval(src)
			if e != nil {
				ret = e.ToString()
			}
		})
	})
	if exception != "" {
		return nil, errors.New(exception)
	}
	return ret, nil
}
