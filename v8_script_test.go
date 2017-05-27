package v8

import (
	"fmt"
	"io/ioutil"
	"runtime"
	"testing"
)

func TestReturnValue(t *testing.T) {
	template := engine.NewObjectTemplate()

	template.Bind("Call", func() *Value {
		val := engine.NewObject()
		obj := val.ToObject()
		obj.SetProperty("name", engine.NewString("test object"))
		obj.SetProperty("id", engine.NewInteger(1234))
		return val
	})

	engine.NewContext(template).Scope(func(cs ContextScope) {
		script := engine.Compile([]byte(`
		a = Call();
		if (a.name != "test object" || a.id != 1234) {
			{};
		} else {
			a;
		}
		`), nil)

		retVal := cs.Run(script)
		if !retVal.IsObject() {
			t.Fatalf("expected object")
		}
		retObj := retVal.ToObject()
		if !retObj.HasProperty("name") || retObj.GetProperty("name").ToString() != "test object" ||
			!retObj.HasProperty("id") || retObj.GetProperty("id").ToNumber() != 1234 {
			t.Fatalf("value should be %q not %q", "{\"name\":\"test object\",\"id\":1234}", string(ToJSON(retVal)))
		}
	})

	runtime.GC()
}

func TestTypescript(t *testing.T) {
	code, err := ioutil.ReadFile("./typescript.js")
	if err != nil {
		t.Errorf("Read typescript file failed: %s\n", err)
	}
	template := engine.NewObjectTemplate()

	engine.NewContext(template).Scope(func(cs ContextScope) {
		script1 := engine.Compile(code, nil)

		cs.Run(script1)

		script2 := engine.Compile([]byte(`"use strict";
		function _go_transpile(source) {
    		var result = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS } });
    		return result.outputText;
		}`), nil)

		cs.Run(script2)

		script3 := engine.Compile([]byte(`
			f = _go_transpile;
		`), nil)

		value := cs.Run(script3)

		if value.IsFunction() == false {
			t.Fatal("value not a function")
		}

		result := value.ToFunction().Call(
			engine.NewString(
				`class Greeter {
    greeting: string;
    constructor(message: string) {
        this.greeting = message;
    }
    greet() {
        return "Hello, " + this.greeting;
    }
}
let greeter = new Greeter("world");`))

		fmt.Println(result)

		result = value.ToFunction().Call(engine.NewString(
			`class Point {
  constructor(x, y) {
    this.x = x;
    this.y = y;
  }
  toString() {
    return '(' + this.x + ', ' + this.y + ')';
  }
}`))

		fmt.Println(result)

	})

	runtime.GC()
}
