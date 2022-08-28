package checks

import (
	"thriftlint"

	"github.com/alecthomas/go-thrift/parser"
)

func CheckDefaultValues() thriftlint.Check {
	return thriftlint.MakeCheck("defaults", func(field *parser.Field) (messages thriftlint.Messages) {
		if field.Default != nil {
			messages.Warning(field, "default values are not allowed")
		}
		return
	})
}
