package main

import "fmt"
import "github.com/saibing/go-v8"

type MyType struct {
	Id       int
	Name     string
	Data     map[string]int
	Callback func(a int, b string)
}

func (mt *MyType) Dump(info string) {
	fmt.Printf(
		"Info: \"%s\", Id: %d, Name: \"%s\", Data: %v\n",
		info, mt.Id, mt.Name, mt.Data,
	)
}

func main() {
	engine := v8.NewEngine()

	global := engine.NewObjectTemplate()

	global.Bind("MyType", MyType{})

	global.Bind("print", func(v ...interface{}) {
		fmt.Println(v...)
	})

	//global.Bind("test", func(obj *v8.Object) {
	//	raw := obj.GetInternalField(0).(*v8.BindObject).Target
	//	raw.Interface().(*MyType).Callback(123, "dada")
	//})

	engine.NewContext(global).Scope(func(cs v8.ContextScope) {
		cs.Eval(`
			var a = new MyType();

			a.Dump("old");

			a.Id = 10;
			a.Name = "Hello";
			a.Data = {
				'x': 1,
				'y': 2
			};
			a.Dump("new");

			a.Callback = function(a, b) {
				print(a, b);
			}

			a.Callback(10, "Hello");

			//test(a);
		`)
	})
}
