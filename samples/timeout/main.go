package main

import "github.com/saibing/go-v8"
import "time"
import "fmt"

func main() {
	engine := v8.NewEngine()
	global := engine.NewObjectTemplate()

	var context *v8.Context

	global.Bind("print", func(v ...interface{}) {
		fmt.Println(v...)
	})

	global.Bind("setTimeout", func(d int64, callback *v8.Function) {
		time.AfterFunc(time.Millisecond*time.Duration(d), func() {
			context.Scope(func(cs v8.ContextScope) {
				callback.Call()
			})
		})
	})

	// test
	context = engine.NewContext(global)
	context.Scope(func(cs v8.ContextScope) {
		cs.Eval(`
		print("begin");
		setTimeout(1500, function(){
			print("one");
			setTimeout(1500, function(){
				print("two");
				setTimeout(1500, function(){
					print("three");
					print("press any key to exit");
				})
			})
		})`)
	})

	fmt.Scanln()
}
