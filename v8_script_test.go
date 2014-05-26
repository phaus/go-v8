package v8

import "testing"
import "runtime"

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
