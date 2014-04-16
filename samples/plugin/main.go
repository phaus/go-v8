package main

/*
#cgo pkg-config: ../../v8.pc

#include "v8_plugin.h"

v8_import_plugin(log);
*/
import "C"
import "github.com/idada/v8.go"

func main() {
	engine := v8.NewEngine()
	global := engine.NewObjectTemplate()

	// C.v8_plugin_log generate by v8_import_plugin(log)
	global.Plugin(C.v8_plugin_log)

	engine.NewContext(global).Scope(func(cs v8.ContextScope) {
		cs.Eval(`
			log("Hello Plugin!")
		`)
	})
}
