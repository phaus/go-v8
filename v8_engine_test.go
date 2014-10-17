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

func rand_sched(max int) {
	for j := rand.Intn(max); j > 0; j-- {
		runtime.Gosched()
	}
}

// Issue #40
//
func Test_EngineDispose(t *testing.T) {
	_ = NewEngine()
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
			object.SetProperty("abc", engine.NewInteger(5678))
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
