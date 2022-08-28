package thriftlint

import (
	"reflect"

	"github.com/alecthomas/go-thrift/parser"
)

var (
	TypeType = reflect.TypeOf(parser.Type{})

	ThriftType = reflect.TypeOf(parser.Thrift{})

	ServiceType = reflect.TypeOf(parser.Service{})
	MethodType  = reflect.TypeOf(parser.Method{})

	EnumType      = reflect.TypeOf(parser.Enum{})
	EnumValueType = reflect.TypeOf(parser.EnumValue{})

	StructType = reflect.TypeOf(parser.Struct{})
	FieldType  = reflect.TypeOf(parser.Field{})

	ConstantType = reflect.TypeOf(parser.Constant{})
	TypedefType  = reflect.TypeOf(parser.Typedef{})
)

func Pos(v interface{}) parser.Pos {
	rv := reflect.Indirect(reflect.ValueOf(v))
	if f := rv.FieldByName("Pos"); f.IsValid() {
		return f.Interface().(parser.Pos)
	}
	return parser.Pos{}
}
