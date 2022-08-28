package checks

import (
	"github.com/alecthomas/go-thrift/parser"

	"thriftlint"
)

func CheckTypeReferences() thriftlint.Check {
	return thriftlint.MakeCheck("types", func(file *parser.Thrift, t *parser.Type) (messages thriftlint.Messages) {
		if !thriftlint.BuiltinThriftTypes[t.Name] && !thriftlint.BuiltinThriftCollections[t.Name] &&
			thriftlint.Resolve(t.Name, file) == nil {
			messages.Error(t, "unknown type %q", t.Name)
		}
		return
	})
}
