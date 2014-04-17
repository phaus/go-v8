package v8

/*
#include "v8_wrap.h"
*/
import "C"
import "unsafe"
import "runtime"
import (
	"sync"
)

var (
	gAllocator *ArrayBufferAllocator
	gMutex     sync.Mutex
)

func init() {
	gAllocator = newArrayBufferAllocator()
}

type embedable struct {
	data interface{}
}

func (this embedable) GetPrivateData() interface{} {
	return this.data
}

func (this *embedable) SetPrivateData(data interface{}) {
	this.data = data
}

type ArrayBufferAllocateCallback func(int, bool) unsafe.Pointer
type ArrayBufferFreeCallback func(unsafe.Pointer, int)

type ArrayBufferAllocator struct {
	self unsafe.Pointer
}

func GetVersion() string {
	return C.GoString(C.V8_GetVersion())
}

func ForceGC() {
	C.V8_ForceGC()
}

func SetFlagsFromString(cmd string) {
	cs := C.CString(cmd)
	defer C.free(unsafe.Pointer(cs))
	C.V8_SetFlagsFromString(cs, C.int(len(cmd)))
}

// Call this to get a new ArrayBufferAllocator
func newArrayBufferAllocator() *ArrayBufferAllocator {
	allocator := &ArrayBufferAllocator{}
	runtime.SetFinalizer(allocator, func(allocator *ArrayBufferAllocator) {
		if allocator.self == nil {
			return
		}
		if traceDispose {
			println("dispose array buffer allocator", allocator.self)
		}
		C.V8_Dispose_Allocator(allocator.self)
	})
	return allocator
}

// Call SetArrayBufferAllocator first if you want use any of
// ArrayBuffer, ArrayBufferView, Int8Array...
// Please be sure to call this function once and keep allocator
// Please set ac and fc to nil if you don't want a custom one
func SetArrayBufferAllocator(
	ac ArrayBufferAllocateCallback,
	fc ArrayBufferFreeCallback,
) {
	var acPointer, fcPointer unsafe.Pointer
	if ac != nil {
		acPointer = unsafe.Pointer(&ac)
	}
	if fc != nil {
		fcPointer = unsafe.Pointer(&fc)
	}

	gMutex.Lock()
	defer gMutex.Unlock()
	gAllocator.self = C.V8_SetArrayBufferAllocator(
		gAllocator.self,
		acPointer,
		fcPointer)
}

//export go_array_buffer_allocate
func go_array_buffer_allocate(callback unsafe.Pointer, length C.size_t, initialized C.int) unsafe.Pointer {
	return (*(*ArrayBufferAllocateCallback)(callback))(int(length), initialized != 0)
}

//export go_array_buffer_free
func go_array_buffer_free(callback unsafe.Pointer, data unsafe.Pointer, length C.size_t) {
	(*(*ArrayBufferFreeCallback)(callback))(data, int(length))
}

func (engine *Engine) SetCaptureStackTraceForUncaughtExceptions(capture bool, frameLimit int) {
	icapture := 0
	if capture {
		icapture = 1
	}

	C.V8_SetCaptureStackTraceForUncaughtExceptions(engine.self, C.int(icapture), C.int(frameLimit))
}
