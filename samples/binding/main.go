package main

import "fmt"
import "reflect"
import "../../"

type MyType struct {
	Id       int
	Name     string
	Callback func(a int, b string)
}

func (mt *MyType) Dump(add string) {
	println("Id =", mt.Id, "| Name = '"+mt.Name+"'", "| Add = '"+add+"'")
}

func main() {
	engine := v8.NewEngine()

	global := engine.NewObjectTemplate()

	global.Bind("MyType", MyType{})

	global.Bind("print", func(v ...interface{}) {
		fmt.Println(v...)
	})

	global.Bind("test", func(obj *v8.Object) {
		raw := obj.GetInternalField(0).(*reflect.Value)
		raw.Interface().(*MyType).Callback(123, "dada")
	})

	engine.NewContext(global).Scope(func(cs v8.ContextScope) {
		cs.Eval(`
			var a = new MyType();

			a.Dump("old");

			a.Id = 10;
			a.Name = "Hello";
			a.Dump("new");

			a.Callback = function(a, b) {
				print(a, b);
			}

			a.Callback(10, "Hello");

			test(a);
		`)
	})
}
