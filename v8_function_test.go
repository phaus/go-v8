package v8

import "testing"

func TestNewFunction(t *testing.T) {
	engine.NewContext(nil).Scope(func(cs ContextScope) {

		result := engine.NewFunction(func(info FunctionCallbackInfo) {
			result := info.Get(0).ToInteger() + info.Get(1).ToInteger() + info.Get(2).ToInteger()
			info.ReturnValue().Set(engine.NewInteger(result))
		}, nil).Call(
			engine.NewInteger(1),
			engine.NewInteger(2),
			engine.NewInteger(3),
		)

		if result.IsNumber() == false {
			t.Fatal("result not a number")
		}

		if result.ToInteger() != 6 {
			t.Fatal("result != 6")
		}
	})
}

func TestFunction(t *testing.T) {
	engine.NewContext(nil).Scope(func(cs ContextScope) {
		script := engine.Compile([]byte(`
			a = function(x,y,z){
				return x+y+z;
			}
		`), nil)

		value := cs.Run(script)

		if value.IsFunction() == false {
			t.Fatal("value not a function")
		}

		result := value.ToFunction().Call(
			engine.NewInteger(1),
			engine.NewInteger(2),
			engine.NewInteger(3),
		)

		if result.IsNumber() == false {
			t.Fatal("result not a number")
		}

		if result.ToInteger() != 6 {
			t.Fatal("result != 6")
		}
	})
}

func TestNewInstance(t *testing.T) {
	engine.NewContext(nil).Scope(func(cs ContextScope) {
		script := engine.Compile([]byte(`
			MyClass = function(x) {
				this.x = x
			}
		`), nil)

		value := cs.Run(script)
		if value.IsFunction() == false {
			t.Fatal("value not a function")
		}

		result := value.ToFunction().NewInstance(engine.NewInteger(42))
		if result.IsObject() == false {
			t.Fatal("result not an object")
		}

		x := result.ToObject().GetProperty("x")
		if x.ToInteger() != 42 {
			t.Fatal("result != 42")
		}
	})
}
