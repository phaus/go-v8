#include "v8.h"
#include "v8_plugin.h"

using namespace v8;

extern "C" {

static void LogCallback(const FunctionCallbackInfo<Value>& args) {
	if (args.Length() < 1) return;

	HandleScope scope(args.GetIsolate());
	Handle<Value> arg = args[0];
	String::Utf8Value value(arg);

	printf("%s\n", (char*)*value);
}

v8_export_plugin(log, {
	global->Set(isolate, "log",
		FunctionTemplate::New(isolate, LogCallback)
	);
});

}