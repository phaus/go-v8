package v8

/*
#include "v8_wrap.h"
#include <stdlib.h>
*/
import "C"
import "unsafe"
import "strings"
import "fmt"

type StackTraceOptions uint

const (
	kLineNumber            = StackTraceOptions(1)
	kColumnOffset          = StackTraceOptions(1<<1) | kLineNumber
	kScriptName            = StackTraceOptions(1 << 2)
	kFunctionName          = StackTraceOptions(1 << 3)
	kIsEval                = StackTraceOptions(1 << 4)
	kIsConstructor         = StackTraceOptions(1 << 5)
	kScriptNameOrSourceURL = StackTraceOptions(1 << 6)
	kScriptId              = StackTraceOptions(1 << 7)
	kOverview              = kLineNumber | kColumnOffset | kScriptName | kFunctionName
	kDetailed              = kOverview | kIsEval | kIsConstructor | kScriptNameOrSourceURL
)

type Message struct {
	Message            string
	SourceLine         string
	ScriptResourceName string
	StackTrace         StackTrace
	Line               int
	StartPosition      int
	EndPosition        int
	StartColumn        int
	EndColumn          int
}

func (m *Message) Error() string {
	return m.Message
}

type StackTrace []*StackFrame

func (s StackTrace) String() string {
	l := make([]string, len(s))
	for i, f := range s {
		l[i] = fmt.Sprintf("%+v", f)
	}
	return strings.Join(l, "\n")
}

type StackFrame struct {
	Line                  int
	Column                int
	ScriptId              int
	ScriptName            string
	ScriptNameOrSourceURL string
	FunctionName          string
	IsEval                bool
	IsConstructor         bool
}

//export go_make_message
func go_make_message(
	message, source_line, script_resource_name *C.char,
	stack_trace unsafe.Pointer,
	line, start_pos, end_pos, start_col, end_col int,
) unsafe.Pointer {

	go_message := &Message{
		C.GoString(message),
		C.GoString(source_line),
		C.GoString(script_resource_name),
		nil,
		line,
		start_pos,
		end_pos,
		start_col,
		end_col,
	}

	if stack_trace != nil {
		go_message.StackTrace = *(*StackTrace)(stack_trace)
	}

	if go_message.ScriptResourceName == "undefined" {
		go_message.ScriptResourceName = ""
	}

	maybe_free(unsafe.Pointer(message))
	maybe_free(unsafe.Pointer(source_line))
	maybe_free(unsafe.Pointer(script_resource_name))

	return unsafe.Pointer(go_message)
}

//export go_make_stacktrace
func go_make_stacktrace() unsafe.Pointer {
	return unsafe.Pointer(&StackTrace{})
}

//export go_make_stackframe
func go_make_stackframe(line, column, script_id int, script_name, script_name_or_url, function_name *C.char, is_eval, is_constructor bool) unsafe.Pointer {
	frame := &StackFrame{
		line, column, script_id,
		C.GoString(script_name),
		C.GoString(script_name_or_url),
		C.GoString(function_name),
		is_eval,
		is_constructor,
	}

	maybe_free(unsafe.Pointer(script_name))
	maybe_free(unsafe.Pointer(script_name_or_url))
	maybe_free(unsafe.Pointer(function_name))

	return unsafe.Pointer(frame)
}

//export go_push_stackframe
func go_push_stackframe(ptr_s, ptr_f unsafe.Pointer) {
	if ptr_s == nil || ptr_f == nil {
		return
	}

	s := (*StackTrace)(ptr_s)
	f := (*StackFrame)(ptr_f)

	*s = append(*s, f)
}

func maybe_free(p unsafe.Pointer) {
	if p != nil {
		C.free(p)
	}
}

type exception struct {
	p unsafe.Pointer
	*Message
}

//export go_make_exception
func go_make_exception(value, message unsafe.Pointer) unsafe.Pointer {

	msg := (*Message)(message)

	go_exception := &exception{value, msg}

	return go_exception.p
}
