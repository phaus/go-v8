package v8

import (
	"errors"
	"reflect"
	"time"
)

type BindObject struct {
	Target     reflect.Value
	Properties []BindObjectProperty
}

type BindObjectProperty struct {
	Name  string
	Value *Value
}

func (bo *BindObject) Set(name string, jsvalue *Value) {
	for i := 0; i < len(bo.Properties); i++ {
		if bo.Properties[i].Name == name {
			bo.Properties[i].Value = jsvalue
			return
		}
	}

	bo.Properties = append(bo.Properties, BindObjectProperty{
		Name:  name,
		Value: jsvalue,
	})
}

func (bo *BindObject) Get(name string) *Value {
	for i := 0; i < len(bo.Properties); i++ {
		if bo.Properties[i].Name == name {
			return bo.Properties[i].Value
		}
	}

	return nil
}

func bindFuncCallback(callbackInfo FunctionCallbackInfo) {
	engine := callbackInfo.CurrentScope().GetEngine()

	gofunc := callbackInfo.Data().(reflect.Value)
	funcType := gofunc.Type()

	numIn := funcType.NumIn()
	numArgs := callbackInfo.Length()

	var out []reflect.Value

	in := make([]reflect.Value, numIn)
	for i := 0; i < numIn-1; i++ {
		jsvalue := callbackInfo.Get(i)
		in[i] = reflect.Indirect(reflect.New(funcType.In(i)))
		engine.SetJsValueToGo(in[i], jsvalue)
	}
	if funcType.IsVariadic() {
		sliceLen := numArgs - (numIn - 1)
		in[numIn-1] = reflect.MakeSlice(funcType.In(numIn-1), sliceLen, sliceLen)

		for i := 0; i < sliceLen; i++ {
			jsvalue := callbackInfo.Get(numIn - 1 + i)
			engine.SetJsValueToGo(in[numIn-1].Index(i), jsvalue)
		}

		out = gofunc.CallSlice(in)
	} else {
		if numIn > 0 {
			jsvalue := callbackInfo.Get(numIn - 1)
			in[numIn-1] = reflect.Indirect(reflect.New(funcType.In(numIn - 1)))
			engine.SetJsValueToGo(in[numIn-1], jsvalue)
		}

		out = gofunc.Call(in)
	}

	if out == nil {
		callbackInfo.CurrentScope().ThrowException("argument number not match")
		return
	}

	switch {
	// when Go function returns only one value
	case len(out) == 1:
		callbackInfo.ReturnValue().Set(engine.GoValueToJsValue(out[0]))
	// when Go function returns multi-value, put them in a JavaScript array
	case len(out) > 1:
		jsResults := engine.NewArray(len(out))
		jsResultsArray := jsResults.ToArray()

		for i := 0; i < len(out); i++ {
			jsResultsArray.SetElement(i, engine.GoValueToJsValue(out[i]))
		}

		callbackInfo.ReturnValue().Set(jsResults)
	}
}

//
// Fast bind Go type or function to JS. Note, The function template and object template
// created in fast bind internal are never destroyed. The JS class map to Go type use a
// internal field to reference a Go object when it instanced. All of the internal field
// keep reference by engine. So, may be you don't like to create too many instance of them.
//
func (template *ObjectTemplate) Bind(typeName string, target interface{}) error {
	engine := template.engine

	typeInfo := reflect.TypeOf(target)

	if typeInfo.Kind() == reflect.Ptr {
		typeInfo = typeInfo.Elem()
	}

	if typeInfo.Kind() == reflect.Func {
		goFunc := reflect.ValueOf(target)
		template.SetAccessor(typeName, func(name string, info AccessorCallbackInfo) {
			info.ReturnValue().Set(engine.NewFunction(bindFuncCallback, goFunc).Value)
		}, nil, nil, PA_None)
		return nil
	}

	if typeInfo.Kind() == reflect.Struct {
		if _, exists := engine.bindTypes[typeInfo]; exists {
			return errors.New("duplicate type binding")
		}

		constructor := engine.NewFunctionTemplate(func(info FunctionCallbackInfo) {
			info.This().SetInternalField(0, &BindObject{
				Target: reflect.New(typeInfo),
			})
		}, nil)
		constructor.SetClassName(typeName)

		objTemplate := constructor.InstanceTemplate()
		objTemplate.SetInternalFieldCount(1)
		objTemplate.SetNamedPropertyHandler(
			// get
			func(name string, info PropertyCallbackInfo) {
				bindObj := info.This().GetInternalField(0).(*BindObject)
				value := bindObj.Target

				field := reflect.Indirect(value).FieldByName(name)

				if field.IsValid() {
					info.ReturnValue().Set(engine.GoValueToJsValue(field))
					return
				}

				method := value.MethodByName(name)

				if method.IsValid() {
					info.ReturnValue().Set(engine.NewFunction(bindFuncCallback, method).Value)
					return
				}

				jsvalue := bindObj.Get(name)

				if jsvalue != nil {
					info.ReturnValue().Set(jsvalue)
					return
				}

				info.CurrentScope().ThrowException("Could't found property or method '" + typeName + "." + name + "' when set value")
			},
			// set
			func(name string, jsvalue *Value, info PropertyCallbackInfo) {
				bindObj := info.This().GetInternalField(0).(*BindObject)
				value := bindObj.Target

				field := reflect.Indirect(value).FieldByName(name)

				if field.IsValid() {
					engine.SetJsValueToGo(field, jsvalue)
					return
				}

				bindObj.Set(name, jsvalue)
			},
			// query
			func(name string, info PropertyCallbackInfo) {
				bindObj := info.This().ToObject().GetInternalField(0).(*BindObject)
				value := bindObj.Target

				if reflect.Indirect(value).FieldByName(name).IsValid() || value.MethodByName(name).IsValid() {
					info.ReturnValue().SetBoolean(true)
					return
				}
			},
			// delete
			nil,
			// enum
			nil,
			nil,
		)
		engine.bindTypes[typeInfo] = objTemplate

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
		return engine.NewInteger(int64(value.Uint()))
	case reflect.Int, reflect.Int64:
		return engine.NewNumber(float64(value.Int()))
	case reflect.Uint, reflect.Uint64:
		return engine.NewNumber(float64(value.Uint()))
	case reflect.Float32, reflect.Float64:
		return engine.NewNumber(value.Float())
	// TODO: avoid data copy
	case reflect.Array, reflect.Slice:
		arrayLen := value.Len()
		jsArrayVal := engine.NewArray(value.Len())
		jsArray := jsArrayVal.ToArray()
		for i := 0; i < arrayLen; i++ {
			jsArray.SetElement(i, engine.GoValueToJsValue(value.Index(i)))
		}
		return jsArrayVal
	// TODO: avoid data copy
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
		return engine.NewFunction(bindFuncCallback, value).Value
	case reflect.Interface:
		return engine.GoValueToJsValue(reflect.ValueOf(value.Interface()))
	case reflect.Ptr:
		valType := value.Type()
		if valType == typeOfValue {
			return value.Interface().(*Value)
		}
		elemType := valType.Elem()
		if elemType.Kind() == reflect.Struct {
			if objectTemplate, exits := engine.bindTypes[elemType]; exits {
				objectVal := engine.NewInstanceOf(objectTemplate)
				object := objectVal.ToObject()
				object.SetInternalField(0, &BindObject{
					Target: value,
				})
				return objectVal
			}
		}
	case reflect.Struct:
		switch value.Interface().(type) {
		case time.Time:
			return engine.NewDate(value.Interface().(time.Time))
		default:
			if objectTemplate, exits := engine.bindTypes[value.Type()]; exits {
				objectVal := engine.NewInstanceOf(objectTemplate)
				object := objectVal.ToObject()
				object.SetInternalField(0, &BindObject{
					Target: value,
				})
				return objectVal
			}
		}
	}
	return engine.Undefined()
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
