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

// Pre-compiles the specified script (context-independent).
//
func (e *Engine) PreCompile(code []byte) *ScriptData {
	codePtr := unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&code)).Data)
	return newScriptData(C.V8_PreCompile(
		e.self, (*C.char)(codePtr), C.int(len(code)),
	))
}

// Compiles the specified script (context-independent).
// 'data' is the Pre-parsing data, as obtained by PreCompile()
// using pre_data speeds compilation if it's done multiple times.
//
func (e *Engine) Compile(code []byte, origin *ScriptOrigin, data *ScriptData) *Script {
	var dataPtr unsafe.Pointer

	if data != nil {
		dataPtr = data.self
	}

	codePtr := unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&code)).Data)
	self := C.V8_Compile(e.self, (*C.char)(codePtr), C.int(len(code)), unsafe.Pointer(origin), dataPtr)

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

// Pre-compilation data that can be associated with a script.  This
// data can be calculated for a script in advance of actually
// compiling it, and can be stored between compilations.  When script
// data is given to the compile method compilation will be faster.
//
type ScriptData struct {
	self unsafe.Pointer
}

func newScriptData(self unsafe.Pointer) *ScriptData {
	if self == nil {
		return nil
	}

	result := &ScriptData{
		self: self,
	}

	runtime.SetFinalizer(result, func(s *ScriptData) {
		if traceDispose {
			println("v8.ScriptData.Dispose()")
		}
		C.V8_DisposeScriptData(s.self)
	})

	return result
}

// Load previous pre-compilation data.
//
func NewScriptData(data []byte) *ScriptData {
	return newScriptData(C.V8_NewScriptData(
		(*C.char)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&data)).Data)),
		C.int(len(data)),
	))
}

// Returns the length of Data().
//
func (sd *ScriptData) Length() int {
	return int(C.V8_ScriptData_Length(sd.self))
}

// Returns a serialized representation of this ScriptData that can later be
// passed to New(). NOTE: Serialized data is platform-dependent.
//
func (sd *ScriptData) Data() []byte {
	return C.GoBytes(
		unsafe.Pointer(C.V8_ScriptData_Data(sd.self)),
		C.V8_ScriptData_Length(sd.self),
	)
}

// Returns true if the source code could not be parsed.
//
func (sd *ScriptData) HasError() bool {
	return C.V8_ScriptData_HasError(sd.self) == 1
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
