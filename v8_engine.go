package v8

/*
#cgo pkg-config: v8.pc
#include "v8_wrap.h"
#include <stdlib.h>
*/
import "C"
import "unsafe"
import "runtime"

var traceDispose = false

// Represents an isolated instance of the V8 engine.
// Objects from one engine must not be used in other engine.
type Engine struct {
	embedable
	self unsafe.Pointer

	// those fields are used to cache special values
	_undefined *Value
	_null      *Value
	_true      *Value
	_false     *Value

	// those fields are used to keep reference
	// make object can't destroy by GC
	funcTemplateId   int64
	funcTemplates    map[int64]*FunctionTemplate
	objectTemplateId int64
	objectTemplates  map[int64]*ObjectTemplate
	fieldOwnerId     int64
	fieldOwners      map[int64]*Object
}

func NewEngine() *Engine {
	self := C.V8_NewEngine()

	if self == nil {
		return nil
	}

	result := &Engine{
		self:            self,
		funcTemplates:   make(map[int64]*FunctionTemplate),
		objectTemplates: make(map[int64]*ObjectTemplate),
		fieldOwners:     make(map[int64]*Object),
	}

	runtime.SetFinalizer(result, func(e *Engine) {
		if traceDispose {
			println("v8.Engine.Dispose()", e.self)
		}
		C.V8_DisposeEngine(e.self)
	})

	return result
}

//export v8_panic
func v8_panic(message *C.char) {
	panic(C.GoString(message))
}

//export v8_field_owner_weak_callback
func v8_field_owner_weak_callback(engine unsafe.Pointer, ownerId C.int64_t) {
	delete((*Engine)(engine).fieldOwners, int64(ownerId))
}
