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

type EscapableScope struct {
	ContextScope
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

//export go_escapable_scope_callback
func go_escapable_scope_callback(c unsafe.Pointer, callback unsafe.Pointer){
	(*(*func(EscapableScope))(callback))(EscapableScope{ContextScope{(*Context)(c)}})
}

func (c *Context) Scope(callback func(ContextScope)) {
	C.V8_Context_Scope(c.self, unsafe.Pointer(c), unsafe.Pointer(&callback))
}

func (c *Context) GetEngine() *Engine{
	return c.engine
}

func (c *Context) EscapableScope(callback func(EscapableScope)){
	C.V8_Escapable_Scope(c.self, unsafe.Pointer(c), unsafe.Pointer(&callback))
}

func (c *Context) SetSecurityToken(value *Value){
	C.V8_Context_SetSecurityToken(c.self, value.self)
}

func (c *Context) GetSecurityToken() *Value {
	return newValue(c.GetEngine(), C.V8_Context_GetSecurityToken(c.self))
}

func (c *Context) UseDefaultSecurityToken() {
	C.V8_Context_UseDefaultSecurityToken(c.self)
}

func (c *Context) GetEmbedderData(index int) *Value {
	return newValue(c.GetEngine(), C.V8_Context_GetEmbedderData(c.self,C.int(index)))
}

func (c *Context) SetEmbedderData(index int, value *Value) {
	C.V8_Context_SetEmbedderData(c.self, C.int(index), value.self)
}

func (c *Context) SetAlignedPointerInEmbedderData(index int, ptr unsafe.Pointer){
	C.V8_Context_SetAlignedPointerInEmbedderData(c.self, C.int(index), ptr)
}

func (c *Context) GetAlignedPointerFromEmbedderData(index int) unsafe.Pointer{
	return C.V8_Context_GetAlignedPointerFromEmbedderData(c.self, C.int(index))
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

func (es EscapableScope) Escape(escontext *Context) *Context {
	self := C.V8_Escapable_Escape(escontext.self)
	if self == nil {
		return nil
	}
	e := es.GetEngine()
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
	//return newValue(es.GetEngine(), C.V8_Escapable_Escape())
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

func (cs ContextScope) TryCatch(callback func()) *Message {
	msg := C.V8_Context_TryCatch(cs.context.self, unsafe.Pointer(&callback))
	if msg == nil {
		return nil
	}
	return (*Message)(msg)
}

type Exception struct {
	*Value
	*Message
}

func (cs ContextScope) TryCatchException(callback func()) *Exception {
	e := C.V8_Context_TryCatchException(cs.context.self, unsafe.Pointer(&callback))
	if e == nil {
		return nil
	}

	excep := (*exception)(e)
	val := newValue(cs.GetEngine(), excep.Pointer)
	if val == nil {
		return nil
	}

	return &Exception{val, excep.Message}
}

func (cs ContextScope) Global() *Object {
	return newValue(cs.GetEngine(), C.V8_Context_Global(cs.context.self)).ToObject()
}
