package main

import "../../"
import "time"
import "fmt"

func main() {
	engine := v8.NewEngine()
	global := engine.NewObjectTemplate()

	var context *v8.Context

	// echo()
	echo := engine.NewFunctionTemplate(func(info v8.FunctionCallbackInfo) {
		println(info.Get(0).ToString())
	}, nil)
	global.SetAccessor("echo", func(name string, info v8.AccessorCallbackInfo) {
		info.ReturnValue().Set(echo.NewFunction())
	}, nil, nil, v8.PA_None)

	// setTimeout()
	setTimeout := engine.NewFunctionTemplate(func(info v8.FunctionCallbackInfo) {
		d := info.Get(0).ToInteger()
		f := info.Get(1).ToFunction()
		time.AfterFunc(time.Millisecond*time.Duration(d), func() {
			context.Scope(func(cs v8.ContextScope) {
				f.Call()
			})
		})
	}, nil)
	global.SetAccessor("setTimeout", func(name string, info v8.AccessorCallbackInfo) {
		info.ReturnValue().Set(setTimeout.NewFunction())
	}, nil, nil, v8.PA_None)

	// test
	fmt.Println("press any key to exit")

	context = engine.NewContext(global)
	context.Scope(func(cs v8.ContextScope) {
		cs.Eval(`
		echo("begin");
		setTimeout(3000, function(){
			echo("one");
			setTimeout(3000, function(){
				echo("two");
				setTimeout(3000, function(){
					echo("three");
				})
			})
		})`)
	})

	fmt.Scanln()
}
