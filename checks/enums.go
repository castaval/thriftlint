package checks

import (
	"sort"

	"thriftlint"

	"github.com/alecthomas/go-thrift/parser"
)

func CheckEnumSequence() thriftlint.Check {
	return thriftlint.MakeCheck("enum", func(e *parser.Enum) (messages thriftlint.Messages) {
		values := []int{}
		for _, v := range e.Values {
			values = append(values, v.Value)
		}
		sort.Sort(sort.IntSlice(values))
		for i := 0; i < len(values); i++ {
			if values[i] != i {
				messages.Warning(e,
					"enum values for %s do not start at 0 and increase monotonically", e.Name)
				break
			}
		}
		return
	})
}
