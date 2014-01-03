package v8

import (
	"errors"
	"reflect"
)

func (template *ObjectTemplate) Bind(typeName string, target interface{}) error {
	engine := template.engine

	typeInfo := reflect.TypeOf(target)

	if typeInfo.Kind() == reflect.Ptr {
		typeInfo = typeInfo.Elem()
	}

	if typeInfo.Kind() == reflect.Func {
		jsfunc := engine.GoFuncToJsFunc(reflect.ValueOf(target))
		template.SetAccessor(typeName, func(name string, info AccessorCallbackInfo) {
			info.ReturnValue().Set(jsfunc.NewFunction())
		}, nil, nil, PA_None)
		return nil
	}

	if typeInfo.Kind() == reflect.Struct {
		constructor := engine.NewFunctionTemplate(func(info FunctionCallbackInfo) {
			value := reflect.New(typeInfo)
			info.This().SetInternalField(0, &value)
		}, nil)
		constructor.SetClassName(typeName)

		objTemplate := constructor.InstanceTemplate()
		objTemplate.SetInternalFieldCount(1)
		objTemplate.SetNamedPropertyHandler(
			// get
			func(name string, info PropertyCallbackInfo) {
				value := info.This().GetInternalField(0).(*reflect.Value)

				field := value.Elem().FieldByName(name)

				if field.IsValid() {
					info.ReturnValue().Set(engine.GoValueToJsValue(field))
					return
				}

				method := value.MethodByName(name)

				if !method.IsValid() {
					info.CurrentScope().ThrowException("could't found property or method '" + typeName + "." + name + "'")
					return
				}

				info.ReturnValue().Set(engine.GoFuncToJsFunc(method).NewFunction())
			},
			// set
			func(name string, jsvalue *Value, info PropertyCallbackInfo) {
				value := info.This().GetInternalField(0).(*reflect.Value).Elem()

				field := value.FieldByName(name)

				if !field.IsValid() {
					info.CurrentScope().ThrowException("could't found property '" + typeName + "." + name + "'")
					return
				}

				engine.SetJsValueToGo(field, jsvalue)
			},
			// query
			func(name string, info PropertyCallbackInfo) {
				value := info.This().ToObject().GetInternalField(0).(*reflect.Value).Elem()
				info.ReturnValue().SetBoolean(value.FieldByName(name).IsValid() || value.MethodByName(name).IsValid())
			},
			// delete
			nil,
			// enum
			nil,
			nil,
		)

		template.SetAccessor(typeName, func(name string, info AccessorCallbackInfo) {
			info.ReturnValue().Set(constructor.NewFunction())
		}, nil, nil, PA_None)

		return nil
	}

	return errors.New("unsupported target type")
}

func (engine *Engine) GoValueToJsValue(value reflect.Value) *Value {
	switch value.Kind() {
	case reflect.Bool:
		return engine.NewBoolean(value.Bool())
	case reflect.String:
		return engine.NewString(value.String())
	case reflect.Int8, reflect.Int16, reflect.Int32:
		return engine.NewInteger(value.Int())
	case reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return engine.NewInteger(value.Int())
	case reflect.Int, reflect.Uint, reflect.Int64, reflect.Uint64:
		return engine.NewNumber(float64(value.Int()))
	case reflect.Float32, reflect.Float64:
		return engine.NewNumber(value.Float())
	case reflect.Array, reflect.Slice:
		arrayLen := value.Len()
		jsArrayVal := engine.NewArray(value.Len())
		jsArray := jsArrayVal.ToArray()
		for i := 0; i < arrayLen; i++ {
			jsArray.SetElement(i, engine.GoValueToJsValue(value.Index(i)))
		}
		return jsArrayVal
	case reflect.Map:
		jsObjectVal := engine.NewObject()
		jsObject := jsObjectVal.ToObject()
		for _, key := range value.MapKeys() {
			switch key.Kind() {
			case reflect.String:
				jsObject.SetProperty(key.String(), engine.GoValueToJsValue(value.MapIndex(key)), PA_None)
			case reflect.Int8, reflect.Int16, reflect.Int32,
				reflect.Uint8, reflect.Uint16, reflect.Uint32,
				reflect.Int, reflect.Uint, reflect.Int64, reflect.Uint64:
				jsObject.SetElement(int(key.Int()), engine.GoValueToJsValue(value.MapIndex(key)))
			}
		}
		return jsObjectVal
	case reflect.Func:
		return engine.GoFuncToJsFunc(value).NewFunction()
	}
	return engine.Undefined()
}

func (engine *Engine) GoFuncToJsFunc(gofunc reflect.Value) *FunctionTemplate {
	funcType := gofunc.Type()
	return engine.NewFunctionTemplate(func(callbackInfo FunctionCallbackInfo) {
		numIn := funcType.NumIn()
		numArgs := callbackInfo.Length()

		var out []reflect.Value

		if funcType.IsVariadic() {
			in := make([]reflect.Value, 1)
			in[0] = reflect.MakeSlice(funcType.In(0), numArgs, numArgs)

			for i := 0; i < numArgs; i++ {
				jsvalue := callbackInfo.Get(i)
				engine.SetJsValueToGo(in[0].Index(i), jsvalue)
			}

			out = gofunc.CallSlice(in)
		} else {
			in := make([]reflect.Value, numIn)

			for i := 0; i < len(in); i++ {
				jsvalue := callbackInfo.Get(i)
				in[i] = reflect.Indirect(reflect.New(funcType.In(i)))
				engine.SetJsValueToGo(in[i], jsvalue)
			}

			out = gofunc.Call(in)
		}

		if out == nil {
			callbackInfo.CurrentScope().ThrowException("argument number not match")
			return
		}

		if len(out) > 0 {
			jsResults := engine.NewArray(len(out))
			jsResultsArray := jsResults.ToArray()

			for i := 0; i < len(out); i++ {
				jsResultsArray.SetElement(i, engine.GoValueToJsValue(out[i]))
			}

			callbackInfo.ReturnValue().Set(jsResults)
		}
	}, nil)
}

var (
	typeOfValue    = reflect.TypeOf(new(Value))
	typeOfObject   = reflect.TypeOf(new(Object))
	typeOfArray    = reflect.TypeOf(new(Array))
	typeOfRegExp   = reflect.TypeOf(new(RegExp))
	typeOfFunction = reflect.TypeOf(new(Function))
)

func (engine *Engine) SetJsValueToGo(field reflect.Value, jsvalue *Value) {
	goType := field.Type()
	switch goType.Kind() {
	case reflect.Bool:
		field.SetBool(jsvalue.ToBoolean())
	case reflect.String:
		field.SetString(jsvalue.ToString())
	case reflect.Int8, reflect.Int16, reflect.Int32:
		field.SetInt(int64(jsvalue.ToInt32()))
	case reflect.Uint8, reflect.Uint16, reflect.Uint32:
		field.SetUint(uint64(jsvalue.ToUint32()))
	case reflect.Int, reflect.Uint, reflect.Int64, reflect.Uint64:
		field.SetInt(jsvalue.ToInteger())
	case reflect.Float32, reflect.Float64:
		field.SetFloat(jsvalue.ToNumber())
	case reflect.Slice:
		jsArray := jsvalue.ToArray()
		jsArrayLen := jsArray.Length()
		field.Set(reflect.MakeSlice(goType, jsArrayLen, jsArrayLen))
		fallthrough
	case reflect.Array:
		jsArray := jsvalue.ToArray()
		jsArrayLen := jsArray.Length()
		for i := 0; i < jsArrayLen; i++ {
			engine.SetJsValueToGo(field.Index(i), jsArray.GetElement(i))
		}
	case reflect.Map:
		jsObject := jsvalue.ToObject()
		jsObjectKeys := jsObject.GetPropertyNames()
		jsObjectKeysLen := jsObjectKeys.Length()
		field.Set(reflect.MakeMap(goType))
		itemType := goType.Elem()
		for i := 0; i < jsObjectKeysLen; i++ {
			mapKey := jsObjectKeys.GetElement(i).ToString()
			mapValue := reflect.Indirect(reflect.New(itemType))
			engine.SetJsValueToGo(mapValue, jsObject.GetProperty(mapKey))
			field.SetMapIndex(reflect.ValueOf(mapKey), mapValue)
		}
	case reflect.Interface:
		field.Set(reflect.ValueOf(jsvalue))
	case reflect.Func:
		function := jsvalue.ToFunction()
		field.Set(reflect.MakeFunc(goType, func(args []reflect.Value) []reflect.Value {
			jsargs := make([]*Value, len(args))
			for i := 0; i < len(args); i++ {
				jsargs[i] = engine.GoValueToJsValue(args[i])
			}
			jsresult := function.Call(jsargs...)

			outNum := goType.NumOut()

			if outNum == 1 {
				var result = reflect.Indirect(reflect.New(goType.Out(0)))
				engine.SetJsValueToGo(result, jsresult)
				return []reflect.Value{result}
			}

			results := make([]reflect.Value, outNum)
			jsresultArray := jsresult.ToArray()

			for i := 0; i < outNum; i++ {
				results[i] = reflect.Indirect(reflect.New(goType.Out(i)))
				engine.SetJsValueToGo(results[i], jsresultArray.GetElement(i))
			}

			return results
		}))
	default:
		switch goType {
		case typeOfValue:
			field.Set(reflect.ValueOf(jsvalue))
		case typeOfObject:
			field.Set(reflect.ValueOf(jsvalue.ToObject()))
		case typeOfArray:
			field.Set(reflect.ValueOf(jsvalue.ToArray()))
		case typeOfRegExp:
			field.Set(reflect.ValueOf(jsvalue.ToRegExp()))
		case typeOfFunction:
			field.Set(reflect.ValueOf(jsvalue.ToFunction()))
		}
	}
}
