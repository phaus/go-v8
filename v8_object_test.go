package v8

import "testing"
import "runtime"

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

		if !object.SetProperty("b", engine.True()) {
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
		if !object.SetProperty("中文字段", engine.False()) {
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

		// Test ForceSetProperty
		if !object.ForceSetProperty("x", engine.False(), PA_DontDelete|PA_ReadOnly) {
			t.Fatal("could't force set property 'x'")
		}

		if prop := object.GetProperty("x"); prop != nil {
			if !prop.IsBoolean() || !prop.IsFalse() {
				t.Fatal("property 'x' value not match")
			}
		} else {
			t.Fatal("could't get property 'x'")
		}

		// Test GetPropertyAttributes
		attris := object.GetPropertyAttributes("x")

		if attris&(PA_DontDelete|PA_ReadOnly) != PA_DontDelete|PA_ReadOnly {
			t.Fatal("property attributes not match")
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
