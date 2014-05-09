package v8

import (
	"io/ioutil"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
	"testing"
	"time"
)

var engine = NewEngine()

func init() {
	// traceDispose = true
	rand.Seed(time.Now().UnixNano())
	go func() {
		for {
			input, err := ioutil.ReadFile("test.cmd")

			if err == nil && len(input) > 0 {
				ioutil.WriteFile("test.cmd", []byte(""), 0744)

				cmd := strings.Trim(string(input), " \n\r\t")

				var p *pprof.Profile

				switch cmd {
				case "lookup goroutine":
					p = pprof.Lookup("goroutine")
				case "lookup heap":
					p = pprof.Lookup("heap")
				case "lookup threadcreate":
					p = pprof.Lookup("threadcreate")
				default:
					println("unknow command: '" + cmd + "'")
				}

				if p != nil {
					file, err := os.Create("test.out")
					if err != nil {
						println("couldn't create test.out")
					} else {
						p.WriteTo(file, 2)
					}
				}
			}

			time.Sleep(2 * time.Second)
		}
	}()
}

func Test_InternalField(t *testing.T) {
	iCache := make([]interface{}, 0)
	ot := engine.NewObjectTemplate()
	ot.SetInternalFieldCount(11)
	context := engine.NewContext(nil)
	context.SetPrivateData(iCache)
	context.Scope(func(cs ContextScope) {
		cache := cs.GetPrivateData().([]interface{})
		str1 := "hello"
		cache = append(cache, str1)
		cs.SetPrivateData(cache)
		obj := engine.NewInstanceOf(ot).ToObject()
		obj.SetInternalField(0, str1)
		str2 := obj.GetInternalField(0).(string)
		if str1 != str2 {
			t.Fatal("data not match")
		}
	})
	context.SetPrivateData(nil)
}

func Test_GetVersion(t *testing.T) {
	t.Log(GetVersion())
}

func Test_Allocator(t *testing.T) {
	UseDefaultArrayBufferAllocator()

	script := engine.Compile([]byte(`
		var data = new ArrayBuffer(10);
		data[0]='a';
		data[0];
	`), nil)

	engine.NewContext(nil).Scope(func(cs ContextScope) {
		exception := cs.TryCatch(func() {
			value := cs.Run(script)
			if value.ToString() != "a" {
				t.Fatal("value failed")
			}
		})
		if exception != nil {
			t.Fatalf("exception found: %s", exception)
		}
	})
}

func Test_MessageListener(t *testing.T) {
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

func Test_HelloWorld(t *testing.T) {
	engine.NewContext(nil).Scope(func(cs ContextScope) {
		if cs.Eval("'Hello ' + 'World!'").ToString() != "Hello World!" {
			t.Fatal("result not match")
		}
	})

	runtime.GC()
}

func Test_ReturnValue(t *testing.T) {
	template := engine.NewObjectTemplate()

	template.Bind("Call", func() *Value {
		val := engine.NewObject()
		obj := val.ToObject()
		obj.SetProperty("name", engine.NewString("test object"), PA_None)
		obj.SetProperty("id", engine.NewInteger(1234), PA_None)
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

func Test_ThrowException(t *testing.T) {
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

func Test_ThrowException2(t *testing.T) {
	template := engine.NewObjectTemplate()
	template.Bind("Call", func() {
		val := engine.NewObject()
		obj := val.ToObject()
		obj.SetProperty("name", engine.NewString("test object"), PA_None)
		obj.SetProperty("id", engine.NewInteger(1234), PA_None)
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

func Test_ThrowException2_Error(t *testing.T) {
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

func Test_TryCatch(t *testing.T) {
	engine.SetCaptureStackTraceForUncaughtExceptions(true, 1)
	defer engine.SetCaptureStackTraceForUncaughtExceptions(false, 0)

	engine.NewContext(nil).Scope(func(cs ContextScope) {
		err := cs.TryCatch(func() {
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

func Test_TryCatch_WithScriptOrigin(t *testing.T) {
	engine.SetCaptureStackTraceForUncaughtExceptions(true, 1)
	defer engine.SetCaptureStackTraceForUncaughtExceptions(false, 0)

	engine.NewContext(nil).Scope(func(cs ContextScope) {
		var script *Script

		err := cs.TryCatch(func() {
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

func Test_TryCatchException(t *testing.T) {
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

func Test_TryCatchException_Custom(t *testing.T) {
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

func Test_TryCatchException_Primitive(t *testing.T) {
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

func Test_Values(t *testing.T) {
	engine.NewContext(nil).Scope(func(cs ContextScope) {

		if !engine.Undefined().IsUndefined() {
			t.Fatal("Undefined() not match")
		}

		if !engine.Null().IsNull() {
			t.Fatal("Null() not match")
		}

		if !engine.True().IsTrue() {
			t.Fatal("True() not match")
		}

		if !engine.False().IsFalse() {
			t.Fatal("False() not match")
		}

		if engine.Undefined() != engine.Undefined() {
			t.Fatal("Undefined() != Undefined()")
		}

		if engine.Null() != engine.Null() {
			t.Fatal("Null() != Null()")
		}

		if engine.True() != engine.True() {
			t.Fatal("True() != True()")
		}

		if engine.False() != engine.False() {
			t.Fatal("False() != False()")
		}

		var (
			maxInt32  = int64(0x7FFFFFFF)
			maxUint32 = int64(0xFFFFFFFF)
			maxUint64 = uint64(0xFFFFFFFFFFFFFFFF)
			maxNumber = int64(maxUint64)
		)

		if engine.NewBoolean(true).ToBoolean() != true {
			t.Fatal(`NewBoolean(true).ToBoolean() != true`)
		}

		if engine.NewNumber(12.34).ToNumber() != 12.34 {
			t.Fatal(`NewNumber(12.34).ToNumber() != 12.34`)
		}

		if engine.NewNumber(float64(maxNumber)).ToInteger() != maxNumber {
			t.Fatal(`NewNumber(float64(maxNumber)).ToInteger() != maxNumber`)
		}

		if engine.NewInteger(maxInt32).IsInt32() == false {
			t.Fatal(`NewInteger(maxInt32).IsInt32() == false`)
		}

		if engine.NewInteger(maxUint32).IsInt32() != false {
			t.Fatal(`NewInteger(maxUint32).IsInt32() != false`)
		}

		if engine.NewInteger(maxUint32).IsUint32() == false {
			t.Fatal(`NewInteger(maxUint32).IsUint32() == false`)
		}

		if engine.NewInteger(maxNumber).ToInteger() != maxNumber {
			t.Fatal(`NewInteger(maxNumber).ToInteger() != maxNumber`)
		}

		if engine.NewString("Hello World!").ToString() != "Hello World!" {
			t.Fatal(`NewString("Hello World!").ToString() != "Hello World!"`)
		}

		if engine.NewObject().IsObject() == false {
			t.Fatal(`NewObject().IsObject() == false`)
		}

		if engine.NewArray(5).IsArray() == false {
			t.Fatal(`NewArray(5).IsArray() == false`)
		}

		if engine.NewArray(5).ToArray().Length() != 5 {
			t.Fatal(`NewArray(5).Length() != 5`)
		}

		if engine.NewRegExp("foo", RF_None).IsRegExp() == false {
			t.Fatal(`NewRegExp("foo", RF_None).IsRegExp() == false`)
		}

		if engine.NewRegExp("foo", RF_Global).ToRegExp().Pattern() != "foo" {
			t.Fatal(`NewRegExp("foo", RF_Global).ToRegExp().Pattern() != "foo"`)
		}

		if engine.NewRegExp("foo", RF_Global).ToRegExp().Flags() != RF_Global {
			t.Fatal(`NewRegExp("foo", RF_Global).ToRegExp().Flags() != RF_Global`)
		}
	})

	runtime.GC()
}

func Test_Object(t *testing.T) {
	engine.NewContext(nil).Scope(func(cs ContextScope) {
		script := engine.Compile([]byte("a={};"), nil)
		value := cs.Run(script)
		object := value.ToObject()

		// Test get/set property
		if prop := object.GetProperty("a"); prop != nil {
			if !prop.IsUndefined() {
				t.Fatal("property 'a' value not match")
			}
		} else {
			t.Fatal("could't get property 'a'")
		}

		if !object.SetProperty("b", engine.True(), PA_None) {
			t.Fatal("could't set property 'b'")
		}

		if prop := object.GetProperty("b"); prop != nil {
			if !prop.IsBoolean() || !prop.IsTrue() {
				t.Fatal("property 'b' value not match")
			}
		} else {
			t.Fatal("could't get property 'b'")
		}

		// Test get/set non-ascii property
		if !object.SetProperty("中文字段", engine.False(), PA_None) {
			t.Fatal("could't set non-ascii property")
		}

		if prop := object.GetProperty("中文字段"); prop != nil {
			if !prop.IsBoolean() || !prop.IsFalse() {
				t.Fatal("non-ascii property value not match")
			}
		} else {
			t.Fatal("could't get non-ascii property")
		}

		// Test get/set element
		if elem := object.GetElement(0); elem != nil {
			if !elem.IsUndefined() {
				t.Fatal("element 0 value not match")
			}
		} else {
			t.Fatal("could't get element 0")
		}

		if !object.SetElement(0, engine.True()) {
			t.Fatal("could't set element 0")
		}

		if elem := object.GetElement(0); elem != nil {
			if !elem.IsTrue() {
				t.Fatal("element 0 value not match")
			}
		} else {
			t.Fatal("could't get element 0")
		}

		// Test GetPropertyAttributes
		if !object.SetProperty("x", engine.True(), PA_DontDelete|PA_ReadOnly) {
			t.Fatal("could't set property with attributes")
		}

		attris := object.GetPropertyAttributes("x")

		if attris&(PA_DontDelete|PA_ReadOnly) != PA_DontDelete|PA_ReadOnly {
			t.Fatal("property attributes not match")
		}

		// Test ForceSetProperty
		if !object.ForceSetProperty("x", engine.False(), PA_None) {
			t.Fatal("could't force set property 'x'")
		}

		if prop := object.GetProperty("x"); prop != nil {
			if !prop.IsBoolean() || !prop.IsFalse() {
				t.Fatal("property 'x' value not match")
			}
		} else {
			t.Fatal("could't get property 'x'")
		}

		// Test HasProperty and DeleteProperty
		if object.HasProperty("a") {
			t.Fatal("property 'a' exists")
		}

		if !object.HasProperty("b") {
			t.Fatal("property 'b' not exists")
		}

		if !object.DeleteProperty("b") {
			t.Fatal("could't delete property 'b'")
		}

		if object.HasProperty("b") {
			t.Fatal("delete property 'b' failed")
		}

		// Test ForceDeleteProperty
		if !object.ForceDeleteProperty("x") {
			t.Fatal("could't delete property 'x'")
		}

		if object.HasProperty("x") {
			t.Fatal("delete property 'x' failed")
		}

		// Test HasElement and DeleteElement
		if object.HasElement(1) {
			t.Fatal("element 1 exists")
		}

		if !object.HasElement(0) {
			t.Fatal("element 0 not exists")
		}

		if !object.DeleteElement(0) {
			t.Fatal("could't delete element 0")
		}

		if object.HasElement(0) {
			t.Fatal("delete element 0 failed")
		}

		// Test GetPropertyNames
		script = engine.Compile([]byte("a={x:10,y:20,z:30};"), nil)
		value = cs.Run(script)
		object = value.ToObject()

		names := object.GetPropertyNames()

		if names.Length() != 3 {
			t.Fatal(`names.Length() != 3`)
		}

		if names.GetElement(0).ToString() != "x" {
			t.Fatal(`names.GetElement(0).ToString() != "x"`)
		}

		if names.GetElement(1).ToString() != "y" {
			t.Fatal(`names.GetElement(1).ToString() != "y"`)
		}

		if names.GetElement(2).ToString() != "z" {
			t.Fatal(`names.GetElement(2).ToString() != "z"`)
		}
	})

	runtime.GC()
}

func Test_Array(t *testing.T) {
	engine.NewContext(nil).Scope(func(cs ContextScope) {
		script := engine.Compile([]byte("[1,2,3]"), nil)
		value := cs.Run(script)
		result := value.ToArray()

		if result.Length() != 3 {
			t.Fatal("array length not match")
		}

		if elem := result.GetElement(0); elem != nil {
			if !elem.IsNumber() || elem.ToNumber() != 1 {
				t.Fatal("element 0 value not match")
			}
		} else {
			t.Fatal("could't get element 0")
		}

		if elem := result.GetElement(1); elem != nil {
			if !elem.IsNumber() || elem.ToNumber() != 2 {
				t.Fatal("element 1 value not match")
			}
		} else {
			t.Fatal("could't get element 1")
		}

		if elem := result.GetElement(2); elem != nil {
			if !elem.IsNumber() || elem.ToNumber() != 3 {
				t.Fatal("element 2 value not match")
			}
		} else {
			t.Fatal("could't get element 2")
		}

		if !result.SetElement(0, engine.True()) {
			t.Fatal("could't set element")
		}

		if elem := result.GetElement(0); elem != nil {
			if !elem.IsTrue() {
				t.Fatal("element 0 value not match")
			}
		} else {
			t.Fatal("could't get element 0")
		}
	})

	runtime.GC()
}

func Test_Function(t *testing.T) {
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

		function := engine.NewFunctionTemplate(func(info FunctionCallbackInfo) {
			if info.Get(0).ToString() != "Hello World!" {
				t.Fatal(`info.Get(0).ToString() != "Hello World!"`)
			}
			info.ReturnValue().SetBoolean(true)
		}, nil).NewFunction()

		if function == nil {
			t.Fatal("function == nil")
		}

		if function.ToFunction().Call(
			engine.NewString("Hello World!"),
		).IsTrue() == false {
			t.Fatal("callback return not match")
		}
	})

	runtime.GC()
}

func Test_Accessor(t *testing.T) {
	// Object
	engine.NewContext(nil).Scope(func(cs ContextScope) {
		var propertyValue int32 = 1234

		object := engine.NewObject().ToObject()

		object.SetAccessor("abc",
			func(name string, info AccessorCallbackInfo) {
				data := info.Data().(*int32)
				info.ReturnValue().SetInt32(*data)
			},
			func(name string, value *Value, info AccessorCallbackInfo) {
				data := info.Data().(*int32)
				*data = value.ToInt32()
			},
			&propertyValue,
			PA_None,
		)

		if object.GetProperty("abc").ToInt32() != 1234 {
			t.Fatal(`object.GetProperty("abc").ToInt32() != 1234`)
		}

		object.SetProperty("abc", engine.NewInteger(5678), PA_None)

		if propertyValue != 5678 {
			t.Fatal(`propertyValue != 5678`)
		}
	})

	// ObjectTemplate
	engine.NewContext(nil).Scope(func(cs ContextScope) {
		template := engine.NewObjectTemplate()
		var propertyValue int32

		template.SetAccessor(
			"abc",
			func(name string, info AccessorCallbackInfo) {
				data := info.Data().(*int32)
				info.ReturnValue().SetInt32(*data)
			},
			func(name string, value *Value, info AccessorCallbackInfo) {
				data := info.Data().(*int32)
				*data = value.ToInt32()
			},
			&propertyValue,
			PA_None,
		)

		template.SetProperty("def", engine.NewInteger(8888), PA_None)

		values := []*Value{
			engine.NewInstanceOf(template), // Make
			engine.NewObject(),             // Wrap
		}
		template.WrapObject(values[1])

		for i := 0; i < 2; i++ {
			value := values[i]

			propertyValue = 1234

			object := value.ToObject()

			if object.GetProperty("abc").ToInt32() != 1234 {
				t.Fatal(`object.GetProperty("abc").ToInt32() != 1234`)
			}

			object.SetProperty("abc", engine.NewInteger(5678), PA_None)

			if propertyValue != 5678 {
				t.Fatal(`propertyValue != 5678`)
			}

			if object.GetProperty("abc").ToInt32() != 5678 {
				t.Fatal(`object.GetProperty("abc").ToInt32() != 5678`)
			}

			if object.GetProperty("def").ToInt32() != 8888 {
				t.Fatal(`object.GetProperty("def").ToInt32() != 8888`)
			}
		}
	})

	runtime.GC()
}

func Test_NamedPropertyHandler(t *testing.T) {
	obj_template := engine.NewObjectTemplate()

	var (
		get_called    = false
		set_called    = false
		query_called  = false
		delete_called = false
		enum_called   = false
	)

	obj_template.SetNamedPropertyHandler(
		func(name string, info PropertyCallbackInfo) {
			//t.Logf("get %s", name)
			get_called = get_called || name == "abc"
		},
		func(name string, value *Value, info PropertyCallbackInfo) {
			//t.Logf("set %s", name)
			set_called = set_called || name == "abc"
		},
		func(name string, info PropertyCallbackInfo) {
			//t.Logf("query %s", name)
			query_called = query_called || name == "abc"
		},
		func(name string, info PropertyCallbackInfo) {
			//t.Logf("delete %s", name)
			delete_called = delete_called || name == "abc"
		},
		func(info PropertyCallbackInfo) {
			//t.Log("enumerate")
			enum_called = true
		},
		nil,
	)

	func_template := engine.NewFunctionTemplate(func(info FunctionCallbackInfo) {
		info.ReturnValue().Set(engine.NewInstanceOf(obj_template))
	}, nil)

	global_template := engine.NewObjectTemplate()

	global_template.SetAccessor("GetData", func(name string, info AccessorCallbackInfo) {
		info.ReturnValue().Set(func_template.NewFunction())
	}, nil, nil, PA_None)

	engine.NewContext(global_template).Scope(func(cs ContextScope) {
		object := engine.NewInstanceOf(obj_template).ToObject()

		object.GetProperty("abc")
		object.SetProperty("abc", engine.NewInteger(123), PA_None)
		object.GetPropertyAttributes("abc")

		cs.Eval(`
			var data = GetData();

			delete data.abc;

			for (var p in data) {
			}
		`)
	})

	if !(get_called && set_called && query_called && delete_called && enum_called) {
		t.Fatal(get_called, set_called, query_called, delete_called, enum_called)
	}

	runtime.GC()
}

func Test_IndexedPropertyHandler(t *testing.T) {
	obj_template := engine.NewObjectTemplate()

	var (
		get_called    = false
		set_called    = false
		query_called  = true // TODO
		delete_called = true // TODO
		enum_called   = true // TODO
	)

	obj_template.SetIndexedPropertyHandler(
		func(index uint32, info PropertyCallbackInfo) {
			//t.Logf("get %d", index)
			get_called = get_called || index == 1
		},
		func(index uint32, value *Value, info PropertyCallbackInfo) {
			//t.Logf("set %d", index)
			set_called = set_called || index == 1
		},
		func(index uint32, info PropertyCallbackInfo) {
			//t.Logf("query %d", index)
			query_called = query_called || index == 1
		},
		func(index uint32, info PropertyCallbackInfo) {
			//t.Logf("delete %d", index)
			delete_called = delete_called || index == 1
		},
		func(info PropertyCallbackInfo) {
			//t.Log("enumerate")
			enum_called = true
		},
		nil,
	)

	func_template := engine.NewFunctionTemplate(func(info FunctionCallbackInfo) {
		info.ReturnValue().Set(engine.NewInstanceOf(obj_template))
	}, nil)

	global_template := engine.NewObjectTemplate()

	global_template.SetAccessor("GetData", func(name string, info AccessorCallbackInfo) {
		info.ReturnValue().Set(func_template.NewFunction())
	}, nil, nil, PA_None)

	engine.NewContext(global_template).Scope(func(cs ContextScope) {
		object := engine.NewInstanceOf(obj_template).ToObject()

		object.GetElement(1)
		object.SetElement(1, engine.NewInteger(123))

		cs.Eval(`
			var data = GetData();

			delete data[1];

			for (var p in data) {
			}
		`)
	})

	if !(get_called && set_called && query_called && delete_called && enum_called) {
		t.Fatal(get_called, set_called, query_called, delete_called, enum_called)
	}

	runtime.GC()
}

func Test_ObjectConstructor(t *testing.T) {
	type MyClass struct {
		name string
	}

	ftConstructor := engine.NewFunctionTemplate(func(info FunctionCallbackInfo) {
		info.This().SetInternalField(0, new(MyClass))
	}, nil)
	ftConstructor.SetClassName("MyClass")

	obj_template := ftConstructor.InstanceTemplate()

	var (
		get_called    = false
		set_called    = false
		query_called  = false
		delete_called = false
		enum_called   = false
	)

	obj_template.SetNamedPropertyHandler(
		func(name string, info PropertyCallbackInfo) {
			//t.Logf("get %s", name)
			get_called = get_called || name == "abc"
			data := info.This().GetInternalField(0).(*MyClass)
			info.ReturnValue().Set(engine.NewString(data.name))
		},
		func(name string, value *Value, info PropertyCallbackInfo) {
			//t.Logf("set %s", name)
			set_called = set_called || name == "abc"
			data := info.This().GetInternalField(0).(*MyClass)
			data.name = value.ToString()
			info.ReturnValue().Set(value)
		},
		func(name string, info PropertyCallbackInfo) {
			//t.Logf("query %s", name)
			query_called = query_called || name == "abc"
		},
		func(name string, info PropertyCallbackInfo) {
			//t.Logf("delete %s", name)
			delete_called = delete_called || name == "abc"
		},
		func(info PropertyCallbackInfo) {
			//t.Log("enumerate")
			enum_called = true
		},
		nil,
	)
	obj_template.SetInternalFieldCount(1)

	engine.NewContext(nil).Scope(func(cs ContextScope) {
		cs.Global().SetProperty("MyClass", ftConstructor.NewFunction(), PA_None)

		if !cs.Eval("(new MyClass) instanceof MyClass").IsTrue() {
			t.Fatal("(new MyClass) instanceof MyClass == false")
		}

		object := cs.Eval(`
			var data = new MyClass;
			var temp = data.abc;
			data.abc = 1;
			delete data.abc;
			for (var p in data) {
			}
			data;
		`).ToObject()

		object.GetPropertyAttributes("abc")
		data := object.GetInternalField(0).(*MyClass)
		if data.name != "1" {
			t.Fatal("InternalField failed")
		}

		if !(get_called && set_called && query_called && delete_called && enum_called) {
			t.Fatal(get_called, set_called, query_called, delete_called, enum_called)
		}
	})

	runtime.GC()
}

func Test_Context(t *testing.T) {
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

		if !global.SetProperty("println", functionTemplate.NewFunction(), PA_None) {
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

func Test_UnderscoreJS(t *testing.T) {
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

func Test_JSON(t *testing.T) {
	engine.NewContext(nil).Scope(func(cs ContextScope) {
		json := `{"a":1,"b":2,"c":"xyz","e":true,"f":false,"g":null,"h":[4,5,6]}`

		value := cs.ParseJSON(json)

		if value == nil {
			t.Fatal(`value == nil`)
		}

		if value.IsObject() == false {
			t.Fatal(`value == false`)
		}

		if string(ToJSON(value)) != json {
			t.Fatal(`string(ToJSON(value)) != json`)
		}

		object := value.ToObject()

		if object.GetProperty("a").ToInt32() != 1 {
			t.Fatal(`object.GetProperty("a").ToInt32() != 1`)
		}

		if object.GetProperty("b").ToInt32() != 2 {
			t.Fatal(`object.GetProperty("b").ToInt32() != 2`)
		}

		if object.GetProperty("c").ToString() != "xyz" {
			t.Fatal(`object.GetProperty("c").ToString() != "xyz"`)
		}

		if object.GetProperty("e").IsTrue() == false {
			t.Fatal(`object.GetProperty("e").IsTrue() == false`)
		}

		if object.GetProperty("f").IsFalse() == false {
			t.Fatal(`object.GetProperty("f").IsFalse() == false`)
		}

		if object.GetProperty("g").IsNull() == false {
			t.Fatal(`object.GetProperty("g").IsNull() == false`)
		}

		array := object.GetProperty("h").ToArray()

		if array.Length() != 3 {
			t.Fatal(`array.Length() != 3`)
		}

		if array.GetElement(0).ToInt32() != 4 {
			t.Fatal(`array.GetElement(0).ToInt32() != 4`)
		}

		if array.GetElement(1).ToInt32() != 5 {
			t.Fatal(`array.GetElement(1).ToInt32() != 5`)
		}

		if array.GetElement(2).ToInt32() != 6 {
			t.Fatal(`array.GetElement(2).ToInt32() != 6`)
		}

		json = `"\"\/\r\n\t\b\\"`

		if string(ToJSON(cs.ParseJSON(json))) != json {
			t.Fatal(`ToJSON(cs.ParseJSON(json)) != json`)
		}
	})

	runtime.GC()
}

func rand_sched(max int) {
	for j := rand.Intn(max); j > 0; j-- {
		runtime.Gosched()
	}
}

// use one engine in different threads
//
func Test_ThreadSafe1(t *testing.T) {
	fail := false

	wg := new(sync.WaitGroup)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			engine.NewContext(nil).Scope(func(cs ContextScope) {
				script := engine.Compile([]byte("'Hello ' + 'World!'"), nil)
				value := cs.Run(script)
				result := value.ToString()
				fail = fail || result != "Hello World!"
				runtime.GC()
				wg.Done()
			})
		}()
	}
	wg.Wait()
	runtime.GC()

	if fail {
		t.FailNow()
	}
}

// use one context in different threads
//
func Test_ThreadSafe2(t *testing.T) {
	fail := false
	context := engine.NewContext(nil)

	wg := new(sync.WaitGroup)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			context.Scope(func(cs ContextScope) {
				rand_sched(200)

				script := engine.Compile([]byte("'Hello ' + 'World!'"), nil)
				value := cs.Run(script)
				result := value.ToString()
				fail = fail || result != "Hello World!"
				runtime.GC()
				wg.Done()
			})
		}()
	}
	wg.Wait()
	runtime.GC()

	if fail {
		t.FailNow()
	}
}

// use one script in different threads
//
func Test_ThreadSafe3(t *testing.T) {
	fail := false
	script := engine.Compile([]byte("'Hello ' + 'World!'"), nil)

	wg := new(sync.WaitGroup)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			engine.NewContext(nil).Scope(func(cs ContextScope) {
				rand_sched(200)

				value := cs.Run(script)
				result := value.ToString()
				fail = fail || result != "Hello World!"
				runtime.GC()
				wg.Done()
			})
		}()
	}
	wg.Wait()
	runtime.GC()

	if fail {
		t.FailNow()
	}
}

// use one context and one script in different threads
//
func Test_ThreadSafe4(t *testing.T) {
	fail := false
	script := engine.Compile([]byte("'Hello ' + 'World!'"), nil)
	context := engine.NewContext(nil)

	wg := new(sync.WaitGroup)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			context.Scope(func(cs ContextScope) {
				rand_sched(200)

				value := cs.Run(script)
				result := value.ToString()
				fail = fail || result != "Hello World!"
				runtime.GC()
				wg.Done()
			})
		}()
	}
	wg.Wait()
	runtime.GC()

	if fail {
		t.FailNow()
	}
}

// ....
//
func Test_ThreadSafe6(t *testing.T) {
	var (
		fail        = false
		gonum       = 100
		scriptChan  = make(chan *Script, gonum)
		contextChan = make(chan *Context, gonum)
	)

	for i := 0; i < gonum; i++ {
		go func() {
			rand_sched(200)

			scriptChan <- engine.Compile([]byte("'Hello ' + 'World!'"), nil)
		}()
	}

	for i := 0; i < gonum; i++ {
		go func() {
			rand_sched(200)

			contextChan <- engine.NewContext(nil)
		}()
	}

	for i := 0; i < gonum; i++ {
		go func() {
			rand_sched(200)

			context := <-contextChan
			script := <-scriptChan

			context.Scope(func(cs ContextScope) {
				result := cs.Run(script).ToString()
				fail = fail || result != "Hello World!"
			})
		}()
	}

	runtime.GC()

	if fail {
		t.FailNow()
	}
}

func Benchmark_NewContext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		engine.NewContext(nil)
	}

	b.StopTimer()
	runtime.GC()
	b.StartTimer()
}

func Benchmark_NewInteger(b *testing.B) {
	engine.NewContext(nil).Scope(func(cs ContextScope) {
		for i := 0; i < b.N; i++ {
			engine.NewInteger(int64(i))
		}
	})

	b.StopTimer()
	runtime.GC()
	b.StartTimer()
}

func Benchmark_NewString(b *testing.B) {
	engine.NewContext(nil).Scope(func(cs ContextScope) {
		for i := 0; i < b.N; i++ {
			engine.NewString("Hello World!")
		}
	})

	b.StopTimer()
	runtime.GC()
	b.StartTimer()
}

func Benchmark_NewObject(b *testing.B) {
	engine.NewContext(nil).Scope(func(cs ContextScope) {
		for i := 0; i < b.N; i++ {
			engine.NewObject()
		}
	})

	b.StopTimer()
	runtime.GC()
	b.StartTimer()
}

func Benchmark_NewArray0(b *testing.B) {
	engine.NewContext(nil).Scope(func(cs ContextScope) {
		for i := 0; i < b.N; i++ {
			engine.NewArray(0)
		}
	})

	b.StopTimer()
	runtime.GC()
	b.StartTimer()
}

func Benchmark_NewArray5(b *testing.B) {
	engine.NewContext(nil).Scope(func(cs ContextScope) {
		for i := 0; i < b.N; i++ {
			engine.NewArray(5)
		}
	})

	b.StopTimer()
	runtime.GC()
	b.StartTimer()
}

func Benchmark_NewArray20(b *testing.B) {
	engine.NewContext(nil).Scope(func(cs ContextScope) {
		for i := 0; i < b.N; i++ {
			engine.NewArray(20)
		}
	})

	b.StopTimer()
	runtime.GC()
	b.StartTimer()
}

func Benchmark_NewArray100(b *testing.B) {
	engine.NewContext(nil).Scope(func(cs ContextScope) {
		for i := 0; i < b.N; i++ {
			engine.NewArray(100)
		}
	})

	b.StopTimer()
	runtime.GC()
	b.StartTimer()
}

func Benchmark_Compile(b *testing.B) {
	b.StopTimer()
	code, err := ioutil.ReadFile("samples/underscore.js")
	if err != nil {
		return
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		engine.Compile(code, nil)
	}

	b.StopTimer()
	runtime.GC()
	b.StartTimer()
}

func Benchmark_RunScript(b *testing.B) {
	b.StopTimer()
	context := engine.NewContext(nil)
	script := engine.Compile([]byte("1+1"), nil)
	b.StartTimer()

	context.Scope(func(cs ContextScope) {
		for i := 0; i < b.N; i++ {
			cs.Run(script)
		}
	})

	b.StopTimer()
	runtime.GC()
	b.StartTimer()
}

func Benchmark_JsFunction(b *testing.B) {
	b.StopTimer()

	script := engine.Compile([]byte(`
		a = function(){
			return 1;
		}
	`), nil)

	engine.NewContext(nil).Scope(func(cs ContextScope) {
		value := cs.Run(script)
		b.StartTimer()

		for i := 0; i < b.N; i++ {
			value.ToFunction().Call()
		}
	})

	b.StopTimer()
	runtime.GC()
	b.StartTimer()
}

func Benchmark_GoFunction(b *testing.B) {
	engine.NewContext(nil).Scope(func(cs ContextScope) {
		b.StopTimer()
		value := engine.NewFunctionTemplate(func(info FunctionCallbackInfo) {
			info.ReturnValue().SetInt32(123)
		}, nil).NewFunction()
		function := value.ToFunction()
		b.StartTimer()

		for i := 0; i < b.N; i++ {
			function.Call()
		}
	})

	b.StopTimer()
	runtime.GC()
	b.StartTimer()
}

func Benchmark_Getter(b *testing.B) {
	engine.NewContext(nil).Scope(func(cs ContextScope) {
		b.StopTimer()
		var propertyValue int32 = 1234

		template := engine.NewObjectTemplate()

		template.SetAccessor(
			"abc",
			func(name string, info AccessorCallbackInfo) {
				data := info.Data().(*int32)
				info.ReturnValue().SetInt32(*data)
			},
			func(name string, value *Value, info AccessorCallbackInfo) {
				data := info.Data().(*int32)
				*data = value.ToInt32()
			},
			&propertyValue,
			PA_None,
		)

		object := engine.NewInstanceOf(template).ToObject()

		b.StartTimer()

		for i := 0; i < b.N; i++ {
			object.GetProperty("abc")
		}
	})

	b.StopTimer()
	runtime.GC()
	b.StartTimer()
}

func Benchmark_Setter(b *testing.B) {
	engine.NewContext(nil).Scope(func(cs ContextScope) {
		b.StopTimer()

		var propertyValue int32 = 1234

		template := engine.NewObjectTemplate()

		template.SetAccessor(
			"abc",
			func(name string, info AccessorCallbackInfo) {
				data := info.Data().(*int32)
				info.ReturnValue().SetInt32(*data)
			},
			func(name string, value *Value, info AccessorCallbackInfo) {
				data := info.Data().(*int32)
				*data = value.ToInt32()
			},
			&propertyValue,
			PA_None,
		)

		object := engine.NewInstanceOf(template).ToObject()

		b.StartTimer()

		for i := 0; i < b.N; i++ {
			object.SetProperty("abc", engine.NewInteger(5678), PA_None)
		}
	})

	b.StopTimer()
	runtime.GC()
	b.StartTimer()
}

func Benchmark_TryCatch(b *testing.B) {
	engine.NewContext(nil).Scope(func(cs ContextScope) {
		for i := 0; i < b.N; i++ {
			cs.TryCatch(func() {
				cs.Eval("a[=1;")
			})
		}
	})
}

func Benchmark_TryCatchException(b *testing.B) {
	engine.NewContext(nil).Scope(func(cs ContextScope) {
		for i := 0; i < b.N; i++ {
			cs.TryCatchException(func() {
				cs.Eval("a[=1;")
			})
		}
	})
}
