package thriftlint

import (
	"reflect"

	"github.com/alecthomas/go-thrift/parser"
)

func Annotation(node interface{}, key, dflt string) string {
	annotations := reflect.Indirect(reflect.ValueOf(node)).
		FieldByName("Annotations").
		Interface().([]*parser.Annotation)
	for _, annotation := range annotations {
		if annotation.Name == key {
			return annotation.Value
		}
	}
	return dflt
}

func AnnotationExists(node interface{}, key string) bool {
	return Annotation(node, key, "\000") != "\000"
}
