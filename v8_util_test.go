package v8

import "testing"
import "runtime"

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
