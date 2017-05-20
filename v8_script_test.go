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
		script := engine.Compile(code, nil)

		retVal := cs.Run(script)
		fmt.Println(retVal)
	})

	runtime.GC()
}
