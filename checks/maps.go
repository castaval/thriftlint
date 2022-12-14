package checks

import (
	"thriftlint"

	"github.com/alecthomas/go-thrift/parser"
)

func CheckMapKeys() thriftlint.Check {
	return thriftlint.MakeCheck("map", checkMapKeys)
}

func checkMapKeys(file *parser.Thrift, t *parser.Type) (messages thriftlint.Messages) {
	if t.Name == "map" {
		kn := t.KeyType.Name
		if kn != "string" && kn != "i16" && kn != "i32" && kn != "i64" && kn != "double" {
			resolved := thriftlint.Resolve(kn, file)
			isEnum := false
			if resolved != nil {
				_, isEnum = resolved.(*parser.Enum)
			}
			if !isEnum {
				messages.Error(t, "map keys must be string, enum, integer or double, not %q", kn)
			}
		}
		return checkMapKeys(file, t.ValueType)
	}
	return
}
