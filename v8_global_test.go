package v8

import "testing"

func TestGetVersion(t *testing.T) {
	t.Log(GetVersion())
}

func TestAllocator(t *testing.T) {
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
