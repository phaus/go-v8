package main

import (
	"fmt"
	"github.com/idada/v8.go"
)

const (
	code = `
        var obj = new Obj()
        obj.PrintMe()

        var o2 = getObj()
        o2.PrintMe()
    `
)

type Obj struct {
	Id int
}

func (this *Obj) PrintMe() {
	fmt.Println("print", this.Id)
}

func GetObj() *Obj {
	r := new(Obj)
	r.Id = 123
	return r
}

func main() {
	engine := v8.NewEngine()
	script := engine.Compile([]byte(code), nil)
	if nil == script {
		panic("cannot compile")
	}
	global := engine.NewObjectTemplate()

	global.Bind("Obj", Obj{})
	global.Bind("getObj", GetObj)

	context := engine.NewContext(global)

	context.Scope(func(cs v8.ContextScope) {
		cs.Run(script)
		//fmt.Println(r.ToString())
	})
}
