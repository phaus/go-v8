package v8

/*
#include "v8_wrap.h"
*/
import "C"
import "unsafe"
import "reflect"
import "sync"

type AccessControl int

// Access control specifications.
//
// Some accessors should be accessible across contexts.  These
// accessors have an explicit access control parameter which specifies
// the kind of cross-context access that should be allowed.
//
// Additionally, for security, accessors can prohibit overwriting by
// accessors defined in JavaScript.  For objects that have such
// accessors either locally or in their prototype chain it is not
// possible to overwrite the accessor by using __defineGetter__ or
// __defineSetter__ from JavaScript code.
//
const (
	AC_DEFAULT               AccessControl = 0
	AC_ALL_CAN_READ                        = 1
	AC_ALL_CAN_WRITE                       = 1 << 1
	AC_PROHIBITS_OVERWRITING               = 1 << 2
)

type ObjectTemplate struct {
	sync.Mutex
	id                 int64
	engine             *Engine
	accessors          map[string]*accessorInfo
	namedInfo          *namedPropertyInfo
	indexedInfo        *indexedPropertyInfo
	properties         map[string]*propertyInfo
	self               unsafe.Pointer
	internalFieldCount int
}

type namedPropertyInfo struct {
	getter     NamedPropertyGetterCallback
	setter     NamedPropertySetterCallback
	deleter    NamedPropertyDeleterCallback
	query      NamedPropertyQueryCallback
	enumerator NamedPropertyEnumeratorCallback
	data       interface{}
}

type indexedPropertyInfo struct {
	getter     IndexedPropertyGetterCallback
	setter     IndexedPropertySetterCallback
	deleter    IndexedPropertyDeleterCallback
	query      IndexedPropertyQueryCallback
	enumerator IndexedPropertyEnumeratorCallback
	data       interface{}
}

type accessorInfo struct {
	key     string
	getter  AccessorGetterCallback
	setter  AccessorSetterCallback
	data    interface{}
	attribs PropertyAttribute
}

type NamedPropertyGetterCallback func(string, PropertyCallbackInfo)
type NamedPropertySetterCallback func(string, *Value, PropertyCallbackInfo)
type NamedPropertyDeleterCallback func(string, PropertyCallbackInfo)
type NamedPropertyQueryCallback func(string, PropertyCallbackInfo)
type NamedPropertyEnumeratorCallback func(PropertyCallbackInfo)

type IndexedPropertyGetterCallback func(uint32, PropertyCallbackInfo)
type IndexedPropertySetterCallback func(uint32, *Value, PropertyCallbackInfo)
type IndexedPropertyDeleterCallback func(uint32, PropertyCallbackInfo)
type IndexedPropertyQueryCallback func(uint32, PropertyCallbackInfo)
type IndexedPropertyEnumeratorCallback func(PropertyCallbackInfo)

type propertyInfo struct {
	key     string
	value   *Value
	attribs PropertyAttribute
}

func newObjectTemplate(e *Engine, self unsafe.Pointer) *ObjectTemplate {
	if self == nil {
		return nil
	}

	ot := &ObjectTemplate{
		id:         e.objectTemplateId + 1,
		engine:     e,
		accessors:  make(map[string]*accessorInfo),
		properties: make(map[string]*propertyInfo),
		self:       self,
	}

	e.objectTemplateId += 1
	e.objectTemplates[ot.id] = ot

	return ot
}

func (e *Engine) NewObjectTemplate() *ObjectTemplate {
	self := C.V8_NewObjectTemplate(e.self)

	return newObjectTemplate(e, self)
}

func (ot *ObjectTemplate) Dispose() {
	ot.Lock()
	defer ot.Unlock()

	if ot.id > 0 {
		delete(ot.engine.objectTemplates, ot.id)
		ot.id = 0
		ot.engine = nil
		C.V8_DisposeObjectTemplate(ot.self)
	}
}

func (e *Engine) NewInstanceOf(ot *ObjectTemplate) *Value {
	ot.Lock()
	defer ot.Unlock()

	if ot.engine == nil {
		return nil
	}

	return newValue(e, C.V8_ObjectTemplate_NewInstance(e.self, ot.self))
}

func (ot *ObjectTemplate) Plugin(pluginInit unsafe.Pointer) {
	C.V8_ObjectTemplate_Plugin(ot.self, pluginInit)
}

func (ot *ObjectTemplate) WrapObject(value *Value) {
	ot.Lock()
	defer ot.Unlock()

	object := value.ToObject()

	for _, info := range ot.accessors {
		object.setAccessor(info)
	}

	for _, info := range ot.properties {
		object.SetProperty(info.key, info.value)
	}
}

func (ot *ObjectTemplate) SetProperty(key string, value *Value, attribs PropertyAttribute) {
	info := &propertyInfo{
		key:     key,
		value:   value,
		attribs: attribs,
	}

	ot.properties[key] = info

	keyPtr := unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&info.key)).Data)

	C.V8_ObjectTemplate_SetProperty(
		ot.self, (*C.char)(keyPtr), C.int(len(key)), value.self, C.int(attribs),
	)
}

func (ot *ObjectTemplate) SetInternalFieldCount(count int) {
	C.V8_ObjectTemplate_SetInternalFieldCount(ot.self, C.int(count))
	ot.internalFieldCount = count
}

func (ot *ObjectTemplate) InternalFieldCount() int {
	return ot.internalFieldCount
}

func (ot *ObjectTemplate) SetAccessor(
	key string,
	getter AccessorGetterCallback,
	setter AccessorSetterCallback,
	data interface{},
	attribs PropertyAttribute,
) {
	info := &accessorInfo{
		key:     key,
		getter:  getter,
		setter:  setter,
		data:    data,
		attribs: attribs,
	}

	ot.accessors[key] = info

	var getterPointer, setterPointer unsafe.Pointer
	if info.getter != nil {
		getterPointer = unsafe.Pointer(&info.getter)
	}

	if info.setter != nil {
		setterPointer = unsafe.Pointer(&info.setter)
	}

	keyPtr := unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&info.key)).Data)

	C.V8_ObjectTemplate_SetAccessor(
		ot.self,
		(*C.char)(keyPtr), C.int(len(info.key)),
		getterPointer,
		setterPointer,
		unsafe.Pointer(&info.data),
		C.int(info.attribs),
	)
}

func (ot *ObjectTemplate) SetNamedPropertyHandler(
	getter NamedPropertyGetterCallback,
	setter NamedPropertySetterCallback,
	query NamedPropertyQueryCallback,
	deleter NamedPropertyDeleterCallback,
	enumerator NamedPropertyEnumeratorCallback,
	data interface{},
) {
	info := &namedPropertyInfo{
		getter:     getter,
		setter:     setter,
		query:      query,
		deleter:    deleter,
		enumerator: enumerator,
		data:       data,
	}

	ot.namedInfo = info
	var getterPointer, setterPointer, queryPointer, deleterPointer, enumeratorPointer unsafe.Pointer
	if info.getter != nil {
		getterPointer = unsafe.Pointer(&info.getter)
	}

	if info.setter != nil {
		setterPointer = unsafe.Pointer(&info.setter)
	}

	if info.query != nil {
		queryPointer = unsafe.Pointer(&info.query)
	}

	if info.deleter != nil {
		deleterPointer = unsafe.Pointer(&info.deleter)
	}

	if info.enumerator != nil {
		enumeratorPointer = unsafe.Pointer(&info.enumerator)
	}

	C.V8_ObjectTemplate_SetNamedPropertyHandler(
		ot.self,
		getterPointer,
		setterPointer,
		queryPointer,
		deleterPointer,
		enumeratorPointer,
		unsafe.Pointer(&data))
}

func (ot *ObjectTemplate) SetIndexedPropertyHandler(
	getter IndexedPropertyGetterCallback,
	setter IndexedPropertySetterCallback,
	query IndexedPropertyQueryCallback,
	deleter IndexedPropertyDeleterCallback,
	enumerator IndexedPropertyEnumeratorCallback,
	data interface{},
) {
	info := &indexedPropertyInfo{
		getter:     getter,
		setter:     setter,
		query:      query,
		deleter:    deleter,
		enumerator: enumerator,
		data:       data,
	}

	var getterPointer, setterPointer, queryPointer, deleterPointer, enumeratorPointer unsafe.Pointer
	if info.getter != nil {
		getterPointer = unsafe.Pointer(&info.getter)
	}

	if info.setter != nil {
		setterPointer = unsafe.Pointer(&info.setter)
	}

	if info.query != nil {
		queryPointer = unsafe.Pointer(&info.query)
	}

	if info.deleter != nil {
		deleterPointer = unsafe.Pointer(&info.deleter)
	}

	if info.enumerator != nil {
		enumeratorPointer = unsafe.Pointer(&info.enumerator)
	}

	ot.indexedInfo = info

	C.V8_ObjectTemplate_SetIndexedPropertyHandler(
		ot.self,
		getterPointer,
		setterPointer,
		queryPointer,
		deleterPointer,
		enumeratorPointer,
		unsafe.Pointer(&data))
}

type PropertyCallbackInfo struct {
	self        unsafe.Pointer
	typ         C.PropertyDataEnum
	data        interface{}
	returnValue ReturnValue
	context     *Context
}

func (p PropertyCallbackInfo) CurrentScope() ContextScope {
	return ContextScope{p.context}
}

func (p PropertyCallbackInfo) This() *Object {
	return newValue(p.context.engine, C.V8_PropertyCallbackInfo_This(p.self, p.typ)).ToObject()
}

func (p PropertyCallbackInfo) Holder() *Object {
	return newValue(p.context.engine, C.V8_PropertyCallbackInfo_Holder(p.self, p.typ)).ToObject()
}

func (p PropertyCallbackInfo) Data() interface{} {
	return p.data
}

func (p PropertyCallbackInfo) ReturnValue() ReturnValue {
	if p.returnValue.self == nil {
		p.returnValue.self = C.V8_PropertyCallbackInfo_ReturnValue(p.self, p.typ)
	}
	return p.returnValue
}

// Property getter callback info
//
type AccessorCallbackInfo struct {
	self        unsafe.Pointer
	data        interface{}
	returnValue ReturnValue
	context     *Context
	typ         C.AccessorDataEnum
}

func (ac AccessorCallbackInfo) CurrentScope() ContextScope {
	return ContextScope{ac.context}
}

func (ac AccessorCallbackInfo) This() *Object {
	return newValue(ac.context.engine, C.V8_AccessorCallbackInfo_This(ac.self, ac.typ)).ToObject()
}

func (ac AccessorCallbackInfo) Holder() *Object {
	return newValue(ac.context.engine, C.V8_AccessorCallbackInfo_Holder(ac.self, ac.typ)).ToObject()
}

func (ac AccessorCallbackInfo) Data() interface{} {
	return ac.data
}

func (ac *AccessorCallbackInfo) ReturnValue() ReturnValue {
	if ac.returnValue.self == nil {
		ac.returnValue.self = C.V8_AccessorCallbackInfo_ReturnValue(ac.self, ac.typ)
	}
	return ac.returnValue
}

type AccessorGetterCallback func(name string, info AccessorCallbackInfo)

type AccessorSetterCallback func(name string, value *Value, info AccessorCallbackInfo)

//export go_accessor_callback
func go_accessor_callback(typ C.AccessorDataEnum, info *C.V8_AccessorCallbackInfo, context unsafe.Pointer) {
	name := reflect.StringHeader{
		Data: uintptr(unsafe.Pointer(info.key)),
		Len:  int(info.key_length),
	}
	gname := *(*string)(unsafe.Pointer(&name))
	gcontext := (*Context)(context)
	switch typ {
	case C.OTA_Getter:
		(*(*AccessorGetterCallback)(info.callback))(
			gname,
			AccessorCallbackInfo{unsafe.Pointer(info), *(*interface{})(info.data), ReturnValue{}, gcontext, typ})
	case C.OTA_Setter:
		(*(*AccessorSetterCallback)(info.callback))(
			gname,
			newValue(gcontext.engine, info.setValue),
			AccessorCallbackInfo{unsafe.Pointer(info), *(*interface{})(info.data), ReturnValue{}, gcontext, typ})
	default:
		panic("impossible type")
	}
}

//export go_named_property_callback
func go_named_property_callback(typ C.PropertyDataEnum, info *C.V8_PropertyCallbackInfo, context unsafe.Pointer) {
	gname := ""
	if info.key != nil {
		gname = C.GoString(info.key)
	}
	gcontext := (*Context)(context)
	switch typ {
	case C.OTP_Getter:
		(*(*NamedPropertyGetterCallback)(info.callback))(
			gname, PropertyCallbackInfo{unsafe.Pointer(info), typ, *(*interface{})(info.data), ReturnValue{}, gcontext})
	case C.OTP_Setter:
		(*(*NamedPropertySetterCallback)(info.callback))(
			gname,
			newValue(gcontext.engine, info.setValue),
			PropertyCallbackInfo{unsafe.Pointer(info), typ, *(*interface{})(info.data), ReturnValue{}, gcontext})
	case C.OTP_Deleter:
		(*(*NamedPropertyDeleterCallback)(info.callback))(
			gname, PropertyCallbackInfo{unsafe.Pointer(info), typ, *(*interface{})(info.data), ReturnValue{}, gcontext})
	case C.OTP_Query:
		(*(*NamedPropertyQueryCallback)(info.callback))(
			gname, PropertyCallbackInfo{unsafe.Pointer(info), typ, *(*interface{})(info.data), ReturnValue{}, gcontext})
	case C.OTP_Enumerator:
		(*(*NamedPropertyEnumeratorCallback)(info.callback))(
			PropertyCallbackInfo{unsafe.Pointer(info), typ, *(*interface{})(info.data), ReturnValue{}, gcontext})
	}
}

//export go_indexed_property_callback
func go_indexed_property_callback(typ C.PropertyDataEnum, info *C.V8_PropertyCallbackInfo, context unsafe.Pointer) {
	gcontext := (*Context)(context)
	switch typ {
	case C.OTP_Getter:
		(*(*IndexedPropertyGetterCallback)(info.callback))(
			uint32(info.index), PropertyCallbackInfo{unsafe.Pointer(info), typ, *(*interface{})(info.data), ReturnValue{}, gcontext})
	case C.OTP_Setter:
		(*(*IndexedPropertySetterCallback)(info.callback))(
			uint32(info.index),
			newValue(gcontext.engine, info.setValue),
			PropertyCallbackInfo{unsafe.Pointer(info), typ, *(*interface{})(info.data), ReturnValue{}, gcontext})
	case C.OTP_Deleter:
		(*(*IndexedPropertyDeleterCallback)(info.callback))(
			uint32(info.index), PropertyCallbackInfo{unsafe.Pointer(info), typ, *(*interface{})(info.data), ReturnValue{}, gcontext})
	case C.OTP_Query:
		(*(*IndexedPropertyQueryCallback)(info.callback))(
			uint32(info.index), PropertyCallbackInfo{unsafe.Pointer(info), typ, *(*interface{})(info.data), ReturnValue{}, gcontext})
	case C.OTP_Enumerator:
		(*(*IndexedPropertyEnumeratorCallback)(info.callback))(
			PropertyCallbackInfo{unsafe.Pointer(info), typ, *(*interface{})(info.data), ReturnValue{}, gcontext})
	}
}

type FunctionTemplate struct {
	sync.Mutex
	id       int64
	engine   *Engine
	callback FunctionCallback
	data     interface{}
	self     unsafe.Pointer
}

func (e *Engine) NewFunctionTemplate(callback FunctionCallback, data interface{}) *FunctionTemplate {
	ft := &FunctionTemplate{
		id:       e.funcTemplateId + 1,
		engine:   e,
		callback: callback,
		data:     data,
	}

	var callbackPtr unsafe.Pointer

	if callback != nil {
		callbackPtr = unsafe.Pointer(&ft.callback)
	}

	self := C.V8_NewFunctionTemplate(e.self, callbackPtr, unsafe.Pointer(&data))
	if self == nil {
		return nil
	}
	ft.self = self

	e.funcTemplateId += 1
	e.funcTemplates[ft.id] = ft

	return ft
}

func (ft *FunctionTemplate) Dispose() {
	ft.Lock()
	defer ft.Unlock()

	if ft.id > 0 {
		delete(ft.engine.funcTemplates, ft.id)
		ft.id = 0
		ft.engine = nil
		C.V8_DisposeFunctionTemplate(ft.self)
	}
}

func (ft *FunctionTemplate) NewFunction() *Value {
	ft.Lock()
	defer ft.Unlock()

	if ft.engine == nil {
		return nil
	}

	return newValue(ft.engine, C.V8_FunctionTemplate_GetFunction(ft.self))
}

func (ft *FunctionTemplate) SetClassName(name string) {
	ft.Lock()
	defer ft.Unlock()

	if ft.engine == nil {
		panic("engine can't be nil")
	}

	namePtr := unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&name)).Data)
	C.V8_FunctionTemplate_SetClassName(ft.self, (*C.char)(namePtr), C.int(len(name)))
}

func (ft *FunctionTemplate) InstanceTemplate() *ObjectTemplate {
	ft.Lock()
	defer ft.Unlock()

	if ft.engine == nil {
		panic("engine can't be nil")
	}

	self := C.V8_FunctionTemplate_InstanceTemplate(ft.self)
	return newObjectTemplate(ft.engine, self)
}
