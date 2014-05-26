package v8

import "testing"
import "runtime"

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
