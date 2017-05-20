package v8

/*
#include "v8_wrap.h"
*/
import "C"
import "unsafe"

type embedable struct {
	data interface{}
}

//GetPrivateData get private data
func (e embedable) GetPrivateData() interface{} {
	return e.data
}

//SetPrivateData set private data
func (e *embedable) SetPrivateData(data interface{}) {
	e.data = data
}

//GetVersion get v8 version
func GetVersion() string {
	return C.GoString(C.V8_GetVersion())
}

//GetIsolateNumberOfDataSlots get number of isolate data slots
func GetIsolateNumberOfDataSlots() uint {
	return uint(C.V8_Isolate_GetNumberOfDataSlots())
}

//SetFlagsFromString set flags from string.
func SetFlagsFromString(cmd string) {
	cs := C.CString(cmd)
	defer C.free(unsafe.Pointer(cs))
	C.V8_SetFlagsFromString(cs, C.int(len(cmd)))
}

//UseDefaultArrayBufferAllocator Set default array buffer allocator to V8 for
// ArrayBuffer, ArrayBufferView, Int8Array...
// If you want to use your own allocator. You can implement it in C++
// and invoke v8::SetArrayBufferAllocator by your self
func UseDefaultArrayBufferAllocator() {
	C.V8_UseDefaultArrayBufferAllocator()
}
