package v8

/*
#include "v8_wrap.h"
#include <stdlib.h>
*/
import "C"
import "unsafe"
import "runtime"

//import "reflect"

// A sandboxed execution context with its own set of built-in objects
// and functions.
type Context struct {
	embedable
	self   unsafe.Pointer
	engine *Engine
}

type ContextScope struct {
	context *Context
}

func (cs ContextScope) GetEngine() *Engine {
	return cs.context.engine
}

func (cs ContextScope) GetPrivateData() interface{} {
	return cs.context.GetPrivateData()
}

func (cs ContextScope) SetPrivateData(data interface{}) {
	cs.context.SetPrivateData(data)
}

func (e *Engine) NewContext(globalTemplate *ObjectTemplate) *Context {
	var globalTemplatePtr unsafe.Pointer
	if globalTemplate != nil {
		globalTemplatePtr = globalTemplate.self
	}
	self := C.V8_NewContext(e.self, globalTemplatePtr)

	if self == nil {
		return nil
	}

	result := &Context{
		self:   self,
		engine: e,
	}

	runtime.SetFinalizer(result, func(c *Context) {
		if traceDispose {
			println("v8.Context.Dispose()", c.self)
		}
		C.V8_DisposeContext(c.self)
	})

	return result
}

//export go_context_scope_callback
func go_context_scope_callback(c unsafe.Pointer, callback unsafe.Pointer) {
	(*(*func(ContextScope))(callback))(ContextScope{(*Context)(c)})
}

func (c *Context) Scope(callback func(ContextScope)) {
	C.V8_Context_Scope(c.self, unsafe.Pointer(c), unsafe.Pointer(&callback))
}

//export go_try_catch_callback
func go_try_catch_callback(callback unsafe.Pointer) {
	(*(*func())(callback))()
}

func escape(s string) string {
	output := ""
	for _, r := range s {
		switch r {
		case '"':
			output += "\\\""
		case '\\':
			output += "\\\\"
		case '/':
			output += "\\/"
		case '\n':
			output += "\\n"
		case '\r':
			output += "\\r"
		case '\t':
			output += "\\t"
		case '\b':
			output += "\\b"
		case '\f':
			output += "\\f"
		default:
			output += string(r)
		}
	}
	return output
}

func (cs ContextScope) ThrowException(err string) {
	cs.Eval(`throw "` + escape(err) + `"`)
	//
	// TODO: use Isolate::ThrowException() will make FunctionTemplate::GetFunction() returns NULL, why?
	//
	//errPtr := unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&err)).Data)
	//C.V8_Context_ThrowException(cs.context.self, (*C.char)(errPtr), C.int(len(err)))
}

func (cs ContextScope) ThrowException2(value *Value) {
	C.V8_Context_ThrowException2(value.self)
}

func (cs ContextScope) TryCatch(callback func()) error {
	msg := C.V8_Context_TryCatch(cs.context.self, unsafe.Pointer(&callback))
	if msg == nil {
		return nil
	}
	return (*Message)(msg)
}

func (cs ContextScope) Global() *Object {
	return newValue(cs.GetEngine(), C.V8_Context_Global(cs.context.self)).ToObject()
}
