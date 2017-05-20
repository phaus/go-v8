package v8

import "testing"
import "runtime"
import "io/ioutil"

func TestHelloWorld(t *testing.T) {
	engine.NewContext(nil).Scope(func(cs ContextScope) {
		if cs.Eval("'Hello ' + 'World!'").ToString() != "Hello World!" {
			t.Fatal("result not match")
		}
	})

	runtime.GC()
}

func TestContext(t *testing.T) {
	script1 := engine.Compile([]byte("typeof(MyTestContext) === 'undefined';"), nil)
	script2 := engine.Compile([]byte("MyTestContext = 1;"), nil)
	script3 := engine.Compile([]byte("MyTestContext = MyTestContext + 7;"), nil)

	test_func := func(cs ContextScope) {
		if !cs.Run(script1).IsTrue() {
			t.Fatal(`!cs.Run(script1).IsTrue()`)
		}

		if cs.Run(script1).IsFalse() {
			t.Fatal(`cs.Run(script1).IsFalse()`)
		}

		if cs.Run(script2).ToInteger() != 1 {
			t.Fatal(`cs.Run(script2).ToInteger() != 1`)
		}

		if cs.Run(script3).ToInteger() != 8 {
			t.Fatal(`cs.Run(script3).ToInteger() != 8`)
		}
	}

	engine.NewContext(nil).Scope(func(cs ContextScope) {
		t.Log("context1")
		engine.NewContext(nil).Scope(test_func)
		t.Log("context2")
		engine.NewContext(nil).Scope(test_func)
		test_func(cs)
	})

	functionTemplate := engine.NewFunctionTemplate(func(info FunctionCallbackInfo) {
		for i := 0; i < info.Length(); i++ {
			t.Log(info.Get(i).ToString())
		}
	}, nil)

	// Test Global Template
	globalTemplate := engine.NewObjectTemplate()

	globalTemplate.SetAccessor("log", func(name string, info AccessorCallbackInfo) {
		info.ReturnValue().Set(functionTemplate.NewFunction())
	}, nil, nil, PA_None)

	engine.NewContext(globalTemplate).Scope(func(cs ContextScope) {
		cs.Eval(`log("Hello World!")`)
	})

	// Test Global Object
	engine.NewContext(nil).Scope(func(cs ContextScope) {
		global := cs.Global()

		if !global.SetProperty("println", functionTemplate.NewFunction()) {
		}

		global = cs.Global()

		if !global.HasProperty("println") {
			t.Fatal(`!global.HasProperty("println")`)
			return
		}

		cs.Eval(`println("Hello World!")`)
	})

	runtime.GC()
}

func TestUnderscoreJS(t *testing.T) {
	engine.NewContext(nil).Scope(func(cs ContextScope) {
		code, err := ioutil.ReadFile("samples/underscore.js")

		if err != nil {
			return
		}

		script := engine.Compile(code, nil)
		cs.Run(script)

		test := []byte(`
			_.find([1, 2, 3, 4, 5, 6], function(num) {
				return num % 2 == 0;
			});
		`)
		testScript := engine.Compile(test, nil)
		value := cs.Run(testScript)

		if value == nil || value.IsNumber() == false {
			t.FailNow()
		}

		result := value.ToNumber()

		if result != 2 {
			t.FailNow()
		}
	})

	runtime.GC()
}
