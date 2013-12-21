v8.go
=====

V8 JavaScript engine bindings for Go.

Features
=======

* Thread safe
* Thorough and careful testing
* Boolean, Number, String, Object, Array, Regexp, Function
* Compile and run JavaScript
* Save and load pre-compiled script data
* Create JavaScript context with global object template
* Operate JavaScript object properties and array elements in Go
* Define JavaScript object template in Go with property accessors and interceptors
* Define JavaScript function template in Go
* Catch JavaScript exception in Go
* Throw JavaScript exception by Go
* JSON parse and generate

Install
=======

For 'curl' user. please run this shell command:

> curl -O https://raw.github.com/idada/v8.go/master/get.sh && chmod +x get.sh && ./get.sh v8.go

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
	script := engine.Compile([]byte("'Hello ' + 'World!'"), nil, nil)
	context := engine.NewContext(nil)

	context.Scope(func(cs v8.ContextScope) {
		result := script.Run()
		println(result.ToString())
	})
}
```

Performance and Stability 
=========================

The benchmark result on my iMac:

```
NewContext     249474 ns/op
NewInteger        984 ns/op
NewString         983 ns/op
NewObject        1036 ns/op
NewArray0        1130 ns/op
NewArray5        1314 ns/op
NewArray20       1666 ns/op
NewArray100      3124 ns/op
Compile         11059 ns/op
PreCompile      11697 ns/op
RunScript        1085 ns/op
JsFunction       1114 ns/op
GoFunction       1630 ns/op
Getter           2060 ns/op
Setter           2934 ns/op
TryCatch        43097 ns/op
```

I write many test and benchmark to make sure v8.go stable and efficient.

There is a shell script named 'test.sh' in the project. 

It can auto configure cgo environment variables and run test.

For example:

```
./test.sh . .
```

The above command will run all of test and benchmark.

The first argument of test.sh is test name pattern, second argument is benchmark name pattern.

For example:

```
./test.sh ThreadSafe Array
```

The above command will run all of thread safe test and all of benchmark about Array type.

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
script := engine.Compile([]byte(`"Hello " + "World!"`), nil, nil)
```

The Engine.Compile() method take 3 arguments. 

The first is the code.

The second is a ScriptOrigin, it stores script's file name or line number offset etc. You can use ScriptOrigin to make error message and stack trace friendly.

```go
name := "my_file.js"
real := ReadFile(name)
code := "function(_export){\n" + real + "\n}"
origin := engine.NewScriptOrigin(name, 1, 0)
script := engine.Compile(code, origin, nil)
```

The third is a ScriptData, it's pre-parsing data, as obtained by Engine.PreCompile(). If you want to compile a script many time, you can use ScriptData to speeds compilation. 

```go
code := []byte(`"Hello " + "World!"`)
data := engine.PreCompile(code)
script1 := engine.Compile(code, nil, data)
script2 := engine.Compile(code, nil, data)
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

Please note. Don't take any JavaScript value out scope.

When the outermost Scope() return, all of the JavaScript value created in this scope or nested scope will be destroyed.

More
----

Please read 'v8_all_test.go' for more information.

中文介绍
========

V8引擎的Go语言绑定。

特性
====

* 线程安全
* 详细的测试
* 数据类型：Boolean, Number, String, Object, Array, Regexp, Function
* 编译并运行JavaScript
* 保存和加载预编译的JavaScript数据
* 创建带有全局对象模板的Context
* 在Go语言端操作和访问JavaScript数组的元素
* 在Go语言端操作和访问JavaScript对象的属性
* 用Go语言创建支持属性访问器和拦截器的JavaScript对象模板
* 用Go语言创建JavaScript函数模板
* 在Go语言端捕获JavaScript的异常
* 从Go语言端抛出JavaScript的异常
* JSON解析和生成

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
	script := engine.Compile([]byte("'Hello ' + 'World!'"), nil, nil)
	context := engine.NewContext(nil)

	context.Scope(func(cs v8.ContextScope) {
		result := script.Run()
		println(result.ToString())
	})
}
```

性能和稳定性 
============

以下是在我的iMac上运行benchmark的输出结果:

```
NewContext     249474 ns/op
NewInteger        984 ns/op
NewString         983 ns/op
NewObject        1036 ns/op
NewArray0        1130 ns/op
NewArray5        1314 ns/op
NewArray20       1666 ns/op
NewArray100      3124 ns/op
Compile         11059 ns/op
PreCompile      11697 ns/op
RunScript        1085 ns/op
JsFunction       1114 ns/op
GoFunction       1630 ns/op
Getter           2060 ns/op
Setter           2934 ns/op
TryCatch        43097 ns/op
```

我写了很多的单元测试和基准测试用来确定v8.go是否稳定和高效。

项目根目录下有一个叫'text.sh'的shell脚本。这个脚本可以自动配置CGO的环境变量并运行v8.go的测试。

举个例子:

```
./test.sh . .
```

以下命令将执行所以单元测试和基准测试。

test.sh的第一个参数是单元测试的名称匹配模式，第二个参数是基准测试的名称匹配模式。

再举个例子:

```
./test.sh ThreadSafe Array
```

以上命令将运行所有线程安全相关的单元测试和所有Array相关的基准测试。

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
script := engine.Compile([]byte(`"Hello " + "World!"`), nil, nil)
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

第三个参数是一个ScriptData对象，它是Engine.PreCompile()方法预解析脚本后得到的数据。如果有一段代码你需要反复编译多次，那么你可以先预解析后，用ScriptData来加速编译。

```go
code := []byte(`"Hello " + "World!"`)
data := engine.PreCompile(code)
script1 := engine.Compile(code, nil, data)
script2 := engine.Compile(code, nil, data)
```

Context
-------

V8嵌入指南中的解释:

> In V8, a context is an execution environment that allows separate, unrelated, JavaScript applications to run in a single instance of V8. You must explicitly specify the context in which you want any JavaScript code to be run.

在v8.go中，你可以从一个Engine实例中创建多个上下文。当你需要在某个上下文中运行一段JavaScript时，你需要调用Context.Scope()方法进入这个上下文，然后在回调函数中运行JavaScript。

```go
context.Scope(func(cs v8.ContextScope){
	script.Run()
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

请注意！不要把JavaScript的Value对象带到最外层的Context.Scope()的回调函数外。

当最外层的Scope()返回，所有在这个Scope中或嵌套的Scope中创建的JavaScript值都将被销毁。

更多
----

请阅读'v8\_all\_test.go'获得更多信息。
