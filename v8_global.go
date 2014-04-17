package v8

/*
#include "v8_wrap.h"
*/
import "C"
import "unsafe"

type embedable struct {
	data interface{}
}

func (this embedable) GetPrivateData() interface{} {
	return this.data
}

func (this *embedable) SetPrivateData(data interface{}) {
	this.data = data
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

// Use the default array buffer allocator of
// ArrayBuffer, ArrayBufferView, Int8Array...
// If you want to use your allocator you can implement it in C++
// and invoke v8::SetArrayBufferAllocator by your self
func UseDefaultArrayBufferAllocator() {
	C.V8_UseDefaultArrayBufferAllocator()
}
