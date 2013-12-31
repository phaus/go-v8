package v8

import (
	"errors"
	"reflect"
)

func (template *ObjectTemplate) Bind(typeName string, target interface{}) error {
	e := template.engine

	typeInfo := reflect.TypeOf(target)

	if typeInfo.Kind() == reflect.Ptr {
		typeInfo = typeInfo.Elem()
	}

	if typeInfo.Kind() != reflect.Struct {
		return errors.New("target not a struct or struct pointer")
	}

	constructor := e.NewFunctionTemplate(func(info FunctionCallbackInfo) {
		value := reflect.New(typeInfo)
		info.This().SetInternalField(0, &value)
	}, nil)
	constructor.SetClassName(typeName)

	obj_template := constructor.InstanceTemplate()
	obj_template.SetNamedPropertyHandler(
		// get
		func(name string, info PropertyCallbackInfo) {
			value := info.This().ToObject().GetInternalField(0).(*reflect.Value)

			field := value.Elem().FieldByName(name)

			if field.IsValid() {
				switch field.Kind() {
				case reflect.Bool:
					info.ReturnValue().SetBoolean(field.Bool())
				case reflect.String:
					info.ReturnValue().SetString(field.String())
				case reflect.Int8, reflect.Int16, reflect.Int32:
					info.ReturnValue().SetInt32(int32(field.Int()))
				case reflect.Uint8, reflect.Uint16, reflect.Uint32:
					info.ReturnValue().SetUint32(uint32(field.Uint()))
				case reflect.Int, reflect.Uint, reflect.Int64, reflect.Uint64:
					info.ReturnValue().SetNumber(float64(field.Int()))
				case reflect.Float32, reflect.Float64:
					info.ReturnValue().SetNumber(field.Float())
				case reflect.Func:
					// TODO:
				default:
					info.CurrentScope().ThrowException("unsupported property type '" + typeName + "." + name + "'")
				}
				return
			}

			method := value.MethodByName(name)

			if method.IsValid() {
				methodType := method.Type()
				info.ReturnValue().Set(e.NewFunctionTemplate(func(methodCallbackInfo FunctionCallbackInfo) {
					numIn := methodType.NumIn()

					if numIn != methodCallbackInfo.Length() {
						info.CurrentScope().ThrowException("argument number not match when calling '" + typeName + "." + name + "()'")
						return
					}

					in := make([]reflect.Value, numIn)

					for i := 0; i < len(in); i++ {
						jsvalue := methodCallbackInfo.Get(i)
						switch methodType.In(i).Kind() {
						case reflect.Bool:
							in[i] = reflect.ValueOf(jsvalue.ToBoolean())
						case reflect.String:
							in[i] = reflect.ValueOf(jsvalue.ToString())
						case reflect.Int8, reflect.Int16, reflect.Int32:
							in[i] = reflect.ValueOf(int64(jsvalue.ToInt32()))
						case reflect.Uint8, reflect.Uint16, reflect.Uint32:
							in[i] = reflect.ValueOf(uint64(jsvalue.ToUint32()))
						case reflect.Int, reflect.Uint, reflect.Int64, reflect.Uint64:
							in[i] = reflect.ValueOf(jsvalue.ToInteger())
						case reflect.Float32, reflect.Float64:
							in[i] = reflect.ValueOf(jsvalue.ToNumber())
						default:
							info.CurrentScope().ThrowException("unsupported function argument type at '" + typeName + "." + name + "()'")
							return
						}
					}

					method.Call(in)
				}, nil).NewFunction())
				return
			}

			info.CurrentScope().ThrowException("could't found property or method '" + typeName + "." + name + "'")
		},
		// set
		func(name string, jsvalue *Value, info PropertyCallbackInfo) {
			value := info.This().ToObject().GetInternalField(0).(*reflect.Value).Elem()

			field := value.FieldByName(name)

			if !field.IsValid() {
				info.CurrentScope().ThrowException("could't found property '" + typeName + "." + name + "'")
				return
			}

			switch field.Kind() {
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
			default:
				info.CurrentScope().ThrowException("unsupported property type '" + typeName + "." + name + "'")
			}
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
	obj_template.SetInternalFieldCount(1)

	template.SetAccessor(typeName, func(name string, info AccessorCallbackInfo) {
		info.ReturnValue().Set(constructor.NewFunction())
	}, nil, nil, PA_None)

	return nil
}
