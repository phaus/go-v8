package main

import (
	"fmt"
	"github.com/idada/v8.go"
)

const (
	code = `
        var obj1 = new Obj()
        obj1.Id = 100
        obj1.PrintMe()
        obj1.Test = function() {
        	check(this.Id)
        }
        obj1.Test()

        var obj2 = getObj()
        obj2.PrintMe()
        obj2.Test = function() {
        	check(this.Id)
        }
        obj2.Test()
    `
)

func check(a int) {
	fmt.Println("check:", a)
}

type Obj struct {
	Id int
}

func (this *Obj) PrintMe() {
	fmt.Println("print:", this.Id)
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

	global.Bind("check", check)
	global.Bind("Obj", Obj{})
	global.Bind("getObj", GetObj)

	context := engine.NewContext(global)

	context.Scope(func(cs v8.ContextScope) {
		cs.Run(script)
		//fmt.Println(r.ToString())
	})
}
