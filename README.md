v8.go
=====

V8 JavaScript engine bindings for Go.

Features
=======

* Thread safe
* Thorough and careful testing
* Boolean, Number, String, Object, Array, Regexp, Function
* Compile and run JavaScript
* Create JavaScript context with global object template
* Operate JavaScript object properties and array elements in Go
* Define JavaScript object template in Go with property accessors and interceptors
* Define JavaScript function template in Go
* Catch JavaScript exception in Go
* Throw JavaScript exception by Go
* JSON parse and generate
* Powerful binding API
* C++ plugin

Install
=======

For 'curl' user. please run this shell command:

> curl -OL https://raw.github.com/idada/v8.go/master/get.sh && chmod +x get.sh && ./get.sh v8.go

For 'wget' user. Please run this shell command:

> wget https://raw.github.com/idada/v8.go/master/get.sh && chmod +x get.sh && ./get.sh v8.go

Note: require Go version 1.2 and Git.

Hello World
===========

This 'Hello World' program shows how to use v8.go to compile and run JavaScript code then get the result.

```go
package main

import "github.com/idada/v8.go"

func main() {
	engine := v8.NewEngine()
	script := engine.Compile([]byte("'Hello ' + 'World!'"), nil)
	context := engine.NewContext(nil)

	context.Scope(func(cs v8.ContextScope) {
		result := cs.Run(script)
		println(result.ToString())
	})
}
```

Fast Binding
============

```go
package main

import "fmt"
import "reflect"
import "github.com/idada/v8.go"

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
			a.Data = {
				'x': 1,
				'y': 2
			};
			a.Dump("new");

			a.Callback = function(a, b) {
				print(a, b);
			}

			a.Callback(10, "Hello");

			test(a);
		`)
	})
}

```

C++ plugin
==========

You can implement plugin in C++.

```cpp
#include "v8.h"
#include "v8_plugin.h"

using namespace v8;

extern "C" {

static void LogCallback(const FunctionCallbackInfo<Value>& args) {
	if (args.Length() < 1) return;

	HandleScope scope(args.GetIsolate());
	Handle<Value> arg = args[0];
	String::Utf8Value value(arg);

	printf("%s\n", *value);
}

v8_export_plugin(log, {
	global->Set(isolate, "log",
		FunctionTemplate::New(isolate, LogCallback)
	);
});

}
```

And load the plugin in Go.

```go
package main

/*
#cgo pkg-config: ../../v8.pc

#include "v8_plugin.h"

v8_import_plugin(log);
*/
import "C"
import "github.com/idada/v8.go"

func main() {
	engine := v8.NewEngine()
	global := engine.NewObjectTemplate()

	// C.v8_plugin_log generate by v8_import_plugin(log)
	global.Plugin(C.v8_plugin_log)

	engine.NewContext(global).Scope(func(cs v8.ContextScope) {
		cs.Eval(`
			log("Hello Plugin!")
		`)
	})
}
```

Performance and Stability 
=========================

The benchmark result on my iMac:

```
Benchmark_NewContext   285869 ns/op
Benchmark_NewInteger      707 ns/op
Benchmark_NewString      1869 ns/op
Benchmark_NewObject      3292 ns/op
Benchmark_NewArray0      1004 ns/op
Benchmark_NewArray5      4024 ns/op
Benchmark_NewArray20     8601 ns/op
Benchmark_NewArray100   31963 ns/op
Benchmark_Compile      640988 ns/op
Benchmark_RunScript       888 ns/op
Benchmark_JsFunction     1148 ns/op
Benchmark_GoFunction     1491 ns/op
Benchmark_Getter         2215 ns/op
Benchmark_Setter         3261 ns/op
Benchmark_TryCatch      47366 ns/op
```

Concepts
========

Engine
------

In v8.go, engine type is the wrapper of v8::Isolate.

Because V8 engine use thread-local storage but cgo calls may be execute in different thread. So v8.go use v8::Locker to make sure V8 engine's thread-local data initialized. And the locker make v8.go thread safe.

You can create different engine instance for data isolate or improve efficiency of concurrent purpose.

```go
engine1 := v8.NewEngine()
engine2 := v8.NewEngine()
```

Script
------

When you want to run some JavaScript. You need to compile first.

Scripts can run many times or run in different context.

```go
script := engine.Compile([]byte(`"Hello " + "World!"`), nil)
```

The Engine.Compile() method take 2 arguments. 

The first is the code.

The second is a ScriptOrigin, it stores script's file name or line number offset etc. You can use ScriptOrigin to make error message and stack trace friendly.

```go
name := "my_file.js"
real := ReadFile(name)
code := "function(_export){\n" + real + "\n}"
origin := engine.NewScriptOrigin(name, 1, 0)
script := engine.Compile(code, origin, nil)
```

Context
-------

The description in V8 embedding guide:

> In V8, a context is an execution environment that allows separate, unrelated, JavaScript applications to run in a single instance of V8. You must explicitly specify the context in which you want any JavaScript code to be run.

In v8.go, you can create many contexts from a V8 engine instance. When you want to run some JavaScript in a context. You need to enter the context by calling Scope() and run the JavaScript in the callback.

```go
context.Scope(func(cs v8.ContextScope){
	script.Run()
})
```

Context in V8 is necessary. So in v8.go you can do this:

```go
context.Scope(func(cs v8.ContextScope) {
	context2 := engine.NewContext(nil)
	context2.Scope(func(cs2 v8.ContextScope) {

	})
})
```

More
----

Please read `v8_all_test.go` and the codes in `samples` folder.

中文介绍
========

V8引擎的Go语言绑定。

特性
====

* 线程安全
* 详细的测试
* 数据类型：Boolean, Number, String, Object, Array, Regexp, Function
* 编译并运行JavaScript
* 创建带有全局对象模板的Context
* 在Go语言端操作和访问JavaScript数组的元素
* 在Go语言端操作和访问JavaScript对象的属性
* 用Go语言创建支持属性访问器和拦截器的JavaScript对象模板
* 用Go语言创建JavaScript函数模板
* 在Go语言端捕获JavaScript的异常
* 从Go语言端抛出JavaScript的异常
* JSON解析和生成
* 强大的绑定功能
* C++插件机制

安装
====

'curl'用户请运行以下脚本：

> curl -O https://raw.github.com/idada/v8.go/master/get.sh && chmod +x get.sh && ./get.sh v8.go

'wget'用户请运行以下脚本：

> wget https://raw.github.com/idada/v8.go/master/get.sh && chmod +x get.sh && ./get.sh v8.go

需求本地安装有Go 1.2和git命令。

Hello World
===========

以下是一段Hello World程序，用来展示v8.go如何编译和运行JavaScript并获得结果：

```go
package main

import "github.com/idada/v8.go"

func main() {
	engine := v8.NewEngine()
	script := engine.Compile([]byte("'Hello ' + 'World!'"), nil)
	context := engine.NewContext(nil)

	context.Scope(func(cs v8.ContextScope) {
		result := cs.Run(script)
		println(result.ToString())
	})
}
```

快速绑定
=======

```go
package main

import "fmt"
import "reflect"
import "github.com/idada/v8.go"

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
			a.Data = {
				'x': 1,
				'y': 2
			};
			a.Dump("new");

			a.Callback = function(a, b) {
				print(a, b);
			}

			a.Callback(10, "Hello");

			test(a);
		`)
	})
}

```

C++插件机制
==========

你可以用C++实现插件.

```cpp
#include "v8.h"
#include "v8_plugin.h"

using namespace v8;

extern "C" {

static void LogCallback(const FunctionCallbackInfo<Value>& args) {
	if (args.Length() < 1) return;

	HandleScope scope(args.GetIsolate());
	Handle<Value> arg = args[0];
	String::Utf8Value value(arg);

	printf("%s\n", *value);
}

v8_export_plugin(log, {
	global->Set(isolate, "log",
		FunctionTemplate::New(isolate, LogCallback)
	);
});

}
```

并在Go代码中加载它.

```go
package main

/*
#cgo pkg-config: ../../v8.pc

#include "v8_plugin.h"

v8_import_plugin(log);
*/
import "C"
import "github.com/idada/v8.go"

func main() {
	engine := v8.NewEngine()
	global := engine.NewObjectTemplate()

	// C.v8_plugin_log generate by v8_import_plugin(log)
	global.Plugin(C.v8_plugin_log)

	engine.NewContext(global).Scope(func(cs v8.ContextScope) {
		cs.Eval(`
			log("Hello Plugin!")
		`)
	})
}
```

性能和稳定性 
============

以下是在我的iMac上运行benchmark的输出结果:

```
Benchmark_NewContext   285869 ns/op
Benchmark_NewInteger      707 ns/op
Benchmark_NewString      1869 ns/op
Benchmark_NewObject      3292 ns/op
Benchmark_NewArray0      1004 ns/op
Benchmark_NewArray5      4024 ns/op
Benchmark_NewArray20     8601 ns/op
Benchmark_NewArray100   31963 ns/op
Benchmark_Compile      640988 ns/op
Benchmark_RunScript       888 ns/op
Benchmark_JsFunction     1148 ns/op
Benchmark_GoFunction     1491 ns/op
Benchmark_Getter         2215 ns/op
Benchmark_Setter         3261 ns/op
Benchmark_TryCatch      47366 ns/op
```

概念
====

Engine
------

在v8.go中，Engine类型是对象v8::Isolate的封装。

因为V8引擎使用线程相关的存储机制用来优化性能，但是CGO调用可能会在不同的线程里执行。所以v8.go使用v8::Locker来确定V8引擎的线程数据有初始化，并确保v8.go是线程安全的。

你可以创建多个引擎实例用来隔离数据和优化并发效率。

```go
engine1 := v8.NewEngine()
engine2 := v8.NewEngine()
```

Script
------

当你要运行一段JavaScript代码前，你需要先把它编译成Script对象。

一个Script对象可以在不同的Context中运行多次。

```go
script := engine.Compile([]byte(`"Hello " + "World!"`), nil)
```

Engine.Compile()方法需要三个参数。

第一个参数是所要编译的JavaScript代码。

第二个参数是一个ScriptOrigin对象，其中存储着脚本对应的文件名和行号等。你可以用ScriptOrigin来让错误信息和栈跟踪信息更友好。

```go
name := "my_file.js"
real := ReadFile(name)
code := "function(_export){\n" + real + "\n}"
origin := engine.NewScriptOrigin(name, 1, 0)
script := engine.Compile(code, origin, nil)
```

Context
-------

V8嵌入指南中的解释:

> In V8, a context is an execution environment that allows separate, unrelated, JavaScript applications to run in a single instance of V8. You must explicitly specify the context in which you want any JavaScript code to be run.

在v8.go中，你可以从一个Engine实例中创建多个上下文。当你需要在某个上下文中运行一段JavaScript时，你需要调用Context.Scope()方法进入这个上下文，然后在回调函数中运行JavaScript。

```go
context.Scope(func(cs v8.ContextScope){
	cs.Run(script)
})
```

上下文在V8中是可以嵌套的。所以v8.go中你可以这样做：

```go
context.Scope(func(cs v8.ContextScope) {
	context2 := engine.NewContext(nil)
	context2.Scope(func(cs2 v8.ContextScope) {

	})
})
```

更多
----

请阅读`v8_all_test.go`以及`samples`目录下的示例代码。
