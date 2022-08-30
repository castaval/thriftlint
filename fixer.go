package thriftlint

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/alecthomas/go-thrift/parser"
)

type Fixer struct {
	Message []*Message
}

func NewFixer(message []*Message) *Fixer {
	return &Fixer{Message: message}
}

func (f *Fixer) FixWarning() {
	for _, msg := range f.Message {
		file, err := os.ReadFile(msg.File.Filename)
		if err != nil {
			return
		}
		pos := Pos(msg.Object)
		fileString := string(file)
		temp := strings.Split(fileString, "\n")
		var fixFileSlice []string
		fixedRow := f.fixRow(msg)
		for index, item := range temp {

			if pos.Line == index+1 {
				fixFileSlice = append(fixFileSlice, fixedRow)
				continue
			}

			fixFileSlice = append(fixFileSlice, item)
		}

		fixedFile := []byte(strings.Join(fixFileSlice, "\n"))

		os.WriteFile(msg.File.Filename, fixedFile, 0666)
	}

}

func (f *Fixer) fixRow(message *Message) (fixedRow string) {
	messageType := message.Checker
	rv := reflect.Indirect(reflect.ValueOf(message.Object))
	pos := rv.FieldByName("Pos")
	col := pos.FieldByName("Col").Interface().(int)
	name := rv.FieldByName("Name").Interface().(string)
	fmt.Println(rv.Type())

	values, row := f.createRow(rv)
	switch messageType {
	case "naming":
		rightName := f.fixNaming(name, message)
		values = append(values, rightName)
		fmt.Println(values...)
		rowWithPos := fmt.Sprint(strings.Repeat(" ", col-1) + row)
		fixedRow = fmt.Sprintf(rowWithPos, values...)
	case "optional":
		values = append(values, name)
		fmt.Println(values...)
		rowWithPos := fmt.Sprint(strings.Repeat(" ", col-1) + row)
		fixedRow = fmt.Sprintf(rowWithPos, values...)
	}
	return
}

func (f *Fixer) createRow(construct reflect.Value) (values []any, row string) {
	switch construct.Interface().(type) {
	case parser.Field:
		values = []any{construct.FieldByName("ID").Interface().(int), reflect.ValueOf("optional").Interface().(string), construct.FieldByName("Type")}
		row = "%d: %s %s %s"
	case parser.EnumValue:
		values = []any{}
		row = "%s"
		// case parser.Union
	}
	return
}

func (f *Fixer) fixNaming(name string, message *Message) (rightName string) {
	if strings.Contains(message.Message, "camel case") {
		rightName = LowerCamelCase(name)
	} else if strings.Contains(message.Message, "title case") {
		rightName = UpperCamelCase(name)
	} else {
		rightName = UpperSnakeCase(name)
	}
	return
}
