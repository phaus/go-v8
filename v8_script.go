package v8

/*
#include "v8_wrap.h"
#include <stdlib.h>
*/
import "C"
import "unsafe"
import "reflect"
import "runtime"

// A compiled JavaScript script.
//
type Script struct {
	self unsafe.Pointer
}

// Compiles the specified script (context-independent).
// 'data' is the Pre-parsing data, as obtained by PreCompile()
// using pre_data speeds compilation if it's done multiple times.
//
func (e *Engine) Compile(code []byte, origin *ScriptOrigin) *Script {
	codePtr := unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&code)).Data)
	self := C.V8_Compile(e.self, (*C.char)(codePtr), C.int(len(code)), unsafe.Pointer(origin))

	if self == nil {
		return nil
	}

	result := &Script{
		self: self,
	}

	runtime.SetFinalizer(result, func(s *Script) {
		if traceDispose {
			println("v8.Script.Dispose()", s.self)
		}
		C.V8_DisposeScript(s.self)
	})

	return result
}

// Runs the script returning the resulting value.
//
func (cs ContextScope) Run(s *Script) *Value {
	return newValue(cs.GetEngine(), C.V8_Script_Run(s.self))
}

// The origin, within a file, of a script.
//
type ScriptOrigin struct {
	Name         string
	LineOffset   int
	ColumnOffset int
}

func (e *Engine) NewScriptOrigin(name string, lineOffset, columnOffset int) *ScriptOrigin {
	return &ScriptOrigin{
		Name:         name,
		LineOffset:   lineOffset,
		ColumnOffset: columnOffset,
	}
}

//export go_script_origin_get_name
func go_script_origin_get_name(p unsafe.Pointer) *C.char {
	if p == nil {
		return C.CString("")
	}
	o := (*ScriptOrigin)(p)
	return C.CString(o.Name)
}

//export go_script_origin_get_line
func go_script_origin_get_line(p unsafe.Pointer) int {
	if p == nil {
		return 0
	}
	o := (*ScriptOrigin)(p)
	return o.LineOffset
}

//export go_script_origin_get_column
func go_script_origin_get_column(p unsafe.Pointer) int {
	if p == nil {
		return 0
	}
	o := (*ScriptOrigin)(p)
	return o.ColumnOffset
}
