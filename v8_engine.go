package v8

/*
#cgo pkg-config: v8.pc
#include "v8_wrap.h"
#include <stdlib.h>
*/
import "C"
import "unsafe"
import "runtime"
import "reflect"

var traceDispose = false

// Represents an isolated instance of the V8 engine.
// Objects from one engine must not be used in other engine.
// Not thred safe!
type Engine struct {
	embedable
	self unsafe.Pointer

	// those fields are used to cache special values
	_undefined *Value
	_null      *Value
	_true      *Value
	_false     *Value

	// those fields are used to keep reference
	// make GC can't destory objects
	funcTemplateId       int64
	funcTemplates        map[int64]*FunctionTemplate
	objectTemplateId     int64
	objectTemplates      map[int64]*ObjectTemplate
	fieldOwnerId         int64
	fieldOwners          map[int64]*Object
	messageListenerId    int64
	firstMessageListener *messageListener
	lastMessageListener  *messageListener

	bindTypes map[reflect.Type]*ObjectTemplate
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
		bindTypes:       make(map[reflect.Type]*ObjectTemplate),
	}

	runtime.SetFinalizer(result, func(e *Engine) {
		if traceDispose {
			println("v8.Engine.Dispose()", e.self)
		}
		C.V8_DisposeEngine(e.self)
	})

	return result
}

//export go_panic
func go_panic(message *C.char) {
	panic(C.GoString(message))
}

//export go_field_owner_weak_callback
func go_field_owner_weak_callback(engine unsafe.Pointer, ownerId C.int64_t) {
	delete((*Engine)(engine).fieldOwners, int64(ownerId))
}

func (engine *Engine) SetCaptureStackTraceForUncaughtExceptions(capture bool, frameLimit int) {
	icapture := 0
	if capture {
		icapture = 1
	}

	C.V8_SetCaptureStackTraceForUncaughtExceptions(engine.self, C.int(icapture), C.int(frameLimit))
}

type MessageCallback func(message *Message)

type messageListener struct {
	Next     *messageListener
	Id       int64
	Callback MessageCallback
}

func (engine *Engine) AddMessageListener(callback MessageCallback) int64 {
	listener := &messageListener{
		Id:       engine.messageListenerId,
		Callback: callback,
	}

	if engine.lastMessageListener == nil {
		engine.firstMessageListener = listener
		engine.lastMessageListener = listener
		C.V8_EnableMessageListener(engine.self, unsafe.Pointer(engine), 1)
	} else {
		engine.lastMessageListener.Next = listener
		engine.lastMessageListener = listener
	}

	return listener.Id
}

func (engine *Engine) RemoveMessageListener(id int64) {
	var p *messageListener
	for i := engine.firstMessageListener; i != nil; p, i = i, i.Next {
		if i.Id == id {
			if p == nil {
				engine.firstMessageListener = i.Next
			} else {
				p.Next = i.Next
			}
			if i == engine.lastMessageListener {
				engine.lastMessageListener = p
			}
			break
		}
	}

	if engine.firstMessageListener == nil {
		C.V8_EnableMessageListener(engine.self, unsafe.Pointer(engine), 0)
	}
}

//export go_message_callback
func go_message_callback(engine, message unsafe.Pointer) {
	for i := (*Engine)(engine).firstMessageListener; i != nil; i = i.Next {
		i.Callback((*Message)(message))
	}
}
