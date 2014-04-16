#ifndef _V8_PLUGIN_H_
#define _V8_PLUGIN_H_

#define v8_import_plugin(Name) \
	extern void v8_plugin_##Name(void*, void*)

#define v8_export_plugin(Name, Body) \
void v8_plugin_##Name(void *isolate_ptr, void* global_ptr) { \
	Isolate* isolate = (Isolate*)isolate_ptr; \
	Locker locker(isolate); \
	Isolate::Scope isolate_scope(isolate); \
	HandleScope scope(isolate); \
	Local<ObjectTemplate> global = *((Local<ObjectTemplate>*)global_ptr); \
	Body \
}

#endif