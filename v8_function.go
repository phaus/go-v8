package v8

/*
#include "v8_wrap.h"
*/
import "C"
import "unsafe"
import "reflect"

// A JavaScript function object (ECMA-262, 15.3).
//
type Function struct {
	*Object
	callback FunctionCallback
	data     interface{}
}

type FunctionCallback func(FunctionCallbackInfo)

//export go_function_callback
func go_function_callback(info, callback, context, data unsafe.Pointer) {
	callbackFunc := *(*func(FunctionCallbackInfo))(callback)
	callbackFunc(FunctionCallbackInfo{info, ReturnValue{}, (*Context)(context), *(*interface{})(data)})
}

func (e *Engine) NewFunction(callback FunctionCallback, data interface{}) *Function {
	function := new(Function)
	function.data = data
	function.callback = callback

	function.Object = newValue(e, C.V8_NewFunction(
		e.self, unsafe.Pointer(&function.callback), unsafe.Pointer(&function.data),
	)).ToObject()

	function.setOwner(function)

	return function
}

func (f *Function) Call(args ...*Value) *Value {
	argv := make([]unsafe.Pointer, len(args))
	for i, arg := range args {
		argv[i] = arg.self
	}
	return newValue(f.engine, C.V8_Function_Call(
		f.self, C.int(len(args)),
		unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&argv)).Data),
	))
}

func (f *Function) NewInstance() *Value {
	return newValue(f.engine, C.V8_Function_NewInstance(f.self))
}

// Function and property return value
//
type ReturnValue struct {
	self unsafe.Pointer
}

func (rv ReturnValue) Set(value *Value) {
	C.V8_ReturnValue_Set(rv.self, value.self)
}

func (rv ReturnValue) SetBoolean(value bool) {
	valueInt := 0
	if value {
		valueInt = 1
	}
	C.V8_ReturnValue_SetBoolean(rv.self, C.int(valueInt))
}

func (rv ReturnValue) SetNumber(value float64) {
	C.V8_ReturnValue_SetNumber(rv.self, C.double(value))
}

func (rv ReturnValue) SetInt32(value int32) {
	C.V8_ReturnValue_SetInt32(rv.self, C.int32_t(value))
}

func (rv ReturnValue) SetUint32(value uint32) {
	C.V8_ReturnValue_SetUint32(rv.self, C.uint32_t(value))
}

func (rv ReturnValue) SetString(value string) {
	valuePtr := unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&value)).Data)
	C.V8_ReturnValue_SetString(rv.self, (*C.char)(valuePtr), C.int(len(value)))
}

func (rv ReturnValue) SetNull() {
	C.V8_ReturnValue_SetNull(rv.self)
}

func (rv ReturnValue) SetUndefined() {
	C.V8_ReturnValue_SetUndefined(rv.self)
}

// Function callback info
//
type FunctionCallbackInfo struct {
	self        unsafe.Pointer
	returnValue ReturnValue
	context     *Context
	data        interface{}
}

func (fc FunctionCallbackInfo) CurrentScope() ContextScope {
	return ContextScope{fc.context}
}

func (fc FunctionCallbackInfo) Get(i int) *Value {
	return newValue(fc.context.engine, C.V8_FunctionCallbackInfo_Get(fc.self, C.int(i)))
}

func (fc FunctionCallbackInfo) Length() int {
	return int(C.V8_FunctionCallbackInfo_Length(fc.self))
}

func (fc FunctionCallbackInfo) Callee() *Function {
	return newValue(fc.context.engine, C.V8_FunctionCallbackInfo_Callee(fc.self)).ToFunction()
}

func (fc FunctionCallbackInfo) This() *Object {
	return newValue(fc.context.engine, C.V8_FunctionCallbackInfo_This(fc.self)).ToObject()
}

func (fc FunctionCallbackInfo) Holder() *Object {
	return newValue(fc.context.engine, C.V8_FunctionCallbackInfo_Holder(fc.self)).ToObject()
}

func (fc FunctionCallbackInfo) Data() interface{} {
	return fc.data
}

func (fc *FunctionCallbackInfo) ReturnValue() ReturnValue {
	if fc.returnValue.self == nil {
		fc.returnValue.self = C.V8_FunctionCallbackInfo_ReturnValue(fc.self)
	}
	return fc.returnValue
}
