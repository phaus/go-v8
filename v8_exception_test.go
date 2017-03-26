package v8

import "testing"
import "runtime"

func TestMessageListener(t *testing.T) {
	id1 := engine.AddMessageListener(func(message *Message) {
		t.Log("MessageListener(1):", message)
	})
	engine.Compile([]byte(`var test[ = ;`), nil)
	// MessageListener(1)

	id2 := engine.AddMessageListener(func(message *Message) {
		t.Log("MessageListener(2):", message)
	})
	engine.Compile([]byte(`var test] = ;`), nil)
	// MessageListener(1)
	// MessageListener(2)

	engine.RemoveMessageListener(id1)
	engine.Compile([]byte(`var test] = ;`), nil)
	// MessageListener(2)

	engine.RemoveMessageListener(id2)
	engine.Compile([]byte(`var test] = ;`), nil)
	// nothing

	engine.NewContext(nil).Scope(func(cs ContextScope) {
		exception := cs.TryCatch(func() {
			engine.Compile([]byte(`var test[] = ;`), nil)
		})
		t.Log("Exception:", exception)
		// Exception
	})
}

func TestThrowException(t *testing.T) {
	template := engine.NewObjectTemplate()
	template.Bind("Call", func() {
		engine.NewContext(nil).Scope(func(cs ContextScope) {
			cs.ThrowException("a \"nice\" error")
		})
	})

	engine.NewContext(template).Scope(func(cs ContextScope) {
		script := engine.Compile([]byte(`
		try {
			Call()
		} catch(e) {
			e
		}
		`), nil)

		value := cs.Run(script)
		if !value.IsString() {
			t.Fatalf("expected string")
		} else if value.ToString() != "a \"nice\" error" {
			t.Fatalf("value should be %q not %q", "a \"nice\" error", value.ToString())
		}
	})

	runtime.GC()
}

func TestThrowException2(t *testing.T) {
	template := engine.NewObjectTemplate()
	template.Bind("Call", func() {
		val := engine.NewObject()
		obj := val.ToObject()
		obj.SetProperty("name", engine.NewString("test object"))
		obj.SetProperty("id", engine.NewInteger(1234))
		engine.NewContext(nil).Scope(func(cs ContextScope) {
			cs.ThrowException2(val)
		})
	})

	engine.NewContext(template).Scope(func(cs ContextScope) {
		script := engine.Compile([]byte(`
		try {
			Call();
		} catch(e) {
			e;
		}
		`), nil)

		retVal := cs.Run(script)
		if !retVal.IsObject() {
			t.Fatalf("expected object")
		}
		retObj := retVal.ToObject()
		if !retObj.HasProperty("name") || retObj.GetProperty("name").ToString() != "test object" {
			t.Fatalf("name should be %s not %s", "test object", retObj.GetProperty("name").ToString())
		} else if !retObj.HasProperty("id") || retObj.GetProperty("id").ToNumber() != 1234 {
			t.Fatalf("id should be %d not %d", 1234, retObj.GetProperty("id").ToNumber())
		}
	})

	runtime.GC()
}

func TestThrowException2Error(t *testing.T) {
	template := engine.NewObjectTemplate()
	template.Bind("RangeCall", func() {
		engine.NewContext(nil).Scope(func(cs ContextScope) {
			cs.ThrowException2(engine.NewRangeError("abcde"))
		})
	})
	template.Bind("ReferenceCall", func() {
		engine.NewContext(nil).Scope(func(cs ContextScope) {
			cs.ThrowException2(engine.NewReferenceError("abcde"))
		})
	})
	template.Bind("SyntaxCall", func() {
		engine.NewContext(nil).Scope(func(cs ContextScope) {
			cs.ThrowException2(engine.NewSyntaxError("abcde"))
		})
	})
	template.Bind("TypeCall", func() {
		engine.NewContext(nil).Scope(func(cs ContextScope) {
			cs.ThrowException2(engine.NewTypeError("abcde"))
		})
	})
	template.Bind("Call", func() {
		engine.NewContext(nil).Scope(func(cs ContextScope) {
			cs.ThrowException2(engine.NewError("abcde"))
		})
	})

	for _, prefix := range []string{"Range", "Reference", "Syntax", "Type", ""} {
		engine.NewContext(template).Scope(func(cs ContextScope) {
			script := engine.Compile([]byte(`
			try {
				`+prefix+`Call();
			} catch(e) {
				e;
			}
			`), nil)

			retVal := cs.Run(script)
			if !retVal.IsObject() {
				t.Fatalf("expected object")
			}
			retObj := retVal.ToObject()
			if !retObj.HasProperty("name") || retObj.GetProperty("name").ToString() != prefix+"Error" {
				t.Fatalf("name should be %s not %s", prefix+"Error", retObj.GetProperty("name").ToString())
			} else if !retObj.HasProperty("message") || retObj.GetProperty("message").ToString() != "abcde" {
				t.Fatalf("message should be %s not %s", "abcde", retObj.GetProperty("message").ToString())
			}
		})
	}

	runtime.GC()
}

func TestTryCatch(t *testing.T) {
	engine.SetCaptureStackTraceForUncaughtExceptions(true, 1)
	defer engine.SetCaptureStackTraceForUncaughtExceptions(false, 0)

	engine.NewContext(nil).Scope(func(cs ContextScope) {
		var err error
		err = cs.TryCatch(func() {
			cs.Eval(`
				function a() {
					1+2;
					b();
					1+2;
				}

				function b() {
					throw new Error("a nice error");
				}

				a();
      		`)
		})
		if err == nil {
			t.Fatal("expected an error")
		}

		msg := err.(*Message)
		if msg.Message != "Uncaught Error: a nice error" {
			t.Fatalf("msg.Message: should be %q not %q", "Uncaught Error: a nice error", msg.Message)
		}
		if msg.ScriptResourceName != "" {
			t.Fatalf("msg.ScriptResourceName: should be %q not %q", "", msg.ScriptResourceName)
		}
		if len(msg.StackTrace) != 1 {
			t.Fatalf("len(msg.StackTrace): should be %d not %d", 1, len(msg.StackTrace))
		}
	})

	runtime.GC()
}

func TestTryCatchWithScriptOrigin(t *testing.T) {
	engine.SetCaptureStackTraceForUncaughtExceptions(true, 1)
	defer engine.SetCaptureStackTraceForUncaughtExceptions(false, 0)

	engine.NewContext(nil).Scope(func(cs ContextScope) {
		var script *Script
		var err error
		err = cs.TryCatch(func() {
			script = engine.Compile([]byte(`
            	function a(c) {
            		c();
            	}

            	function b() {
            		throw new Error("a nice error");
            	}

            	a(b);
            `), engine.NewScriptOrigin("/test.js", 1, 0))
			cs.Run(script)
		})
		if err == nil {
			t.Fatal("expected an error")
		}

		msg := err.(*Message)
		if msg.Message != "Uncaught Error: a nice error" {
			t.Fatalf("msg.Message: should be %q not %q", "Uncaught Error: a nice error", msg.Message)
		}
		if msg.ScriptResourceName != "/test.js" {
			t.Fatalf("msg.ScriptResourceName: should be %q not %q", "/test.js", msg.ScriptResourceName)
		}
		if len(msg.StackTrace) != 1 {
			t.Fatalf("len(msg.StackTrace): should be %d not %d", 1, len(msg.StackTrace))
		}
	})

	runtime.GC()
}

func TestTryCatchException(t *testing.T) {
	engine.SetCaptureStackTraceForUncaughtExceptions(true, 1)
	defer engine.SetCaptureStackTraceForUncaughtExceptions(false, 0)

	engine.NewContext(nil).Scope(func(cs ContextScope) {
		script := engine.Compile([]byte(`
		function a(c) {
			c();
		}

		function b() {
			throw new Error("a nice error");
		}

		a(b);
		`), engine.NewScriptOrigin("/test.js", 1, 0))

		err := cs.TryCatchException(func() {
			cs.Run(script)
		})
		if err == nil {
			t.Fatal("expected an error")
		} else if !err.IsObject() {
			t.Fatal("expected an object")
		}
		obj := err.ToObject()
		if !obj.HasProperty("stack") {
			t.Fatal("expected stacktrace")
		} else if len(obj.GetProperty("stack").ToString()) == 0 {
			t.Fatal("expected stacktrace")
		}

		msg := err.Message
		if msg.Message != "Uncaught Error: a nice error" {
			t.Fatalf("msg.Message: should be %q not %q", "Uncaught Error: a nice error", msg.Message)
		}
		if msg.ScriptResourceName != "/test.js" {
			t.Fatalf("msg.ScriptResourceName: should be %q not %q", "/test.js", msg.ScriptResourceName)
		}
		if len(msg.StackTrace) != 1 {
			t.Fatalf("len(msg.StackTrace): should be %d not %d", 1, len(msg.StackTrace))
		}
	})

	runtime.GC()
}

func TestTryCatchExceptionCustom(t *testing.T) {
	engine.SetCaptureStackTraceForUncaughtExceptions(true, 1)
	defer engine.SetCaptureStackTraceForUncaughtExceptions(false, 0)

	engine.NewContext(nil).Scope(func(cs ContextScope) {
		script := engine.Compile([]byte(`
		function a(c) {
			c();
		}

		function b() {
			e = new Error("a nice error");
			e.customProperty = "custom property"
			throw e;
		}

		a(b);
		`), engine.NewScriptOrigin("/test.js", 1, 0))

		err := cs.TryCatchException(func() {
			cs.Run(script)
		})
		if err == nil {
			t.Fatal("expected an error")
		} else if !err.IsObject() {
			t.Fatal("expected an object")
		}
		obj := err.ToObject()
		if !obj.HasProperty("stack") {
			t.Fatal("expected stacktrace")
		} else if len(obj.GetProperty("stack").ToString()) == 0 {
			t.Fatal("expected stacktrace")
		} else if !obj.HasProperty("customProperty") {
			t.Fatal("expected customProperty")
		} else if obj.GetProperty("customProperty").ToString() != "custom property" {
			t.Fatalf("err.customProperty: should be %q not %q", "customProperty", obj.GetProperty("customProperty").ToString())
		}

		msg := err.Message
		if msg.Message != "Uncaught Error: a nice error" {
			t.Fatalf("msg.Message: should be %q not %q", "Uncaught Error: a nice error", msg.Message)
		}
		if msg.ScriptResourceName != "/test.js" {
			t.Fatalf("msg.ScriptResourceName: should be %q not %q", "/test.js", msg.ScriptResourceName)
		}
		if len(msg.StackTrace) != 1 {
			t.Fatalf("len(msg.StackTrace): should be %d not %d", 1, len(msg.StackTrace))
		}
	})

	runtime.GC()
}

func TestTryCatchExceptionPrimitive(t *testing.T) {
	engine.SetCaptureStackTraceForUncaughtExceptions(true, 1)
	defer engine.SetCaptureStackTraceForUncaughtExceptions(false, 0)

	engine.NewContext(nil).Scope(func(cs ContextScope) {
		script := engine.Compile([]byte(`
		function a(c) {
			c();
		}

		function b() {
			throw 1111;
		}

		a(b);
		`), engine.NewScriptOrigin("/test.js", 1, 0))

		err := cs.TryCatchException(func() {
			cs.Run(script)
		})
		if err == nil {
			t.Fatal("expected an error")
		} else if !err.IsNumber() {
			t.Fatal("expected an integer")
		} else if err.ToInteger() != 1111 {
			t.Fatalf("err: should be %q not %q", 1111, err.ToInteger())
		}

		msg := err.Message
		if msg.Message != "Uncaught 1111" {
			t.Fatalf("msg.Message: should be %q not %q", "Uncaught 1111", msg.Message)
		}
		if msg.ScriptResourceName != "/test.js" {
			t.Fatalf("msg.ScriptResourceName: should be %q not %q", "/test.js", msg.ScriptResourceName)
		}
		if len(msg.StackTrace) != 1 {
			t.Fatalf("len(msg.StackTrace): should be %d not %d", 1, len(msg.StackTrace))
		}
	})

	runtime.GC()
}
