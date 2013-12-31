package main

import "../../"

type MyType struct {
	Id   int
	Name string
}

func (mt *MyType) Dump(add string) {
	println("Id =", mt.Id, "| Name = '"+mt.Name+"'", "| Add = '"+add+"'")
}

func main() {
	engine := v8.NewEngine()

	global := engine.NewObjectTemplate()
	global.Bind("MyType", new(MyType))

	engine.NewContext(global).Scope(func(cs v8.ContextScope) {
		cs.Eval(`
			var a = new MyType();
			a.Dump("old");
			a.Id = 10;
			a.Name = "Hello";
			a.Dump("new");
		`)
	})
}
