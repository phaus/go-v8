package v8

import "testing"
import "runtime"

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

func Test_FunctionTemplate(t *testing.T) {
	engine.NewContext(nil).Scope(func(cs ContextScope) {
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
