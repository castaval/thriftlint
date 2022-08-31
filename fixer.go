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
		fixedRow := f.fixRow(msg, temp[pos.Line-1])

		for i := 0; i < len(temp); i++ {

			if msg.Checker == "enum" && i == pos.Line {
				fixFileSlice = append(fixFileSlice, fixedRow...)
				i += len(fixedRow) - 1
				continue
			} else if pos.Line == i+1 && msg.Checker != "enum" {
				fixFileSlice = append(fixFileSlice, fixedRow...)
				continue
			}

			fixFileSlice = append(fixFileSlice, temp[i])
		}

		fixedFile := []byte(strings.Join(fixFileSlice, "\n"))

		os.WriteFile(msg.File.Filename, fixedFile, 0666)
	}

}

func (f *Fixer) fixRow(message *Message, line string) (fixedRow []string) {
	messageType := message.Checker
	rv := reflect.Indirect(reflect.ValueOf(message.Object))
	pos := rv.FieldByName("Pos")
	col := pos.FieldByName("Col").Interface().(int)
	name := rv.FieldByName("Name").Interface().(string)

	switch messageType {
	case "naming":
		values, row := f.createRow(rv, line)
		rightName := f.fixNaming(name, message)
		values = values[:len(values)-1]
		values = append(values, rightName)
		fmt.Println(values...)
		rowWithPos := fmt.Sprint(strings.Repeat(" ", col-1) + row)
		fixedRow = append(fixedRow, fmt.Sprintf(rowWithPos, values...))
	case "optional":
		values, row := f.createRow(rv, line)
		fmt.Println(values...)
		rowWithPos := fmt.Sprint(strings.Repeat(" ", col-1) + row)
		fixedRow = append(fixedRow, fmt.Sprintf(rowWithPos, values...))
	case "enum":
		fmt.Println(rv.FieldByName("Values"))
		for _, value := range rv.FieldByName("Values").MapKeys() {
			values, row := f.createRow(reflect.Indirect(rv.FieldByName("Values").MapIndex(value)), line)
			rowWithPos := fmt.Sprint(strings.Repeat(" ", 0) + row)
			fixedRow = append(fixedRow, fmt.Sprintf(rowWithPos, values...))
		}
	}
	return
}

func (f *Fixer) createRow(construct reflect.Value, line string) (values []any, row string) {
	switch construct.Interface().(type) {
	case parser.Field:
		values = []any{construct.FieldByName("ID").Interface().(int), reflect.ValueOf("optional").Interface().(string), construct.FieldByName("Type"), construct.FieldByName("Name").Interface().(string)}
		row = "%d: %s %s %s"
	case parser.EnumValue:
		values = []any{construct.FieldByName("Name").Interface().(string)}
		row = "%s"
	case parser.Struct:
		values = []any{construct.FieldByName("Name").Interface().(string)}
		if strings.Contains(line, "struct") {
			row = "struct %s {"
		} else if strings.Contains(line, "union") {
			row = "union %s {"
		} else {
			row = "exception %s {"
		}
	case parser.Service:
		values = []any{construct.FieldByName("Name").Interface().(string)}
		row = "service %s {"
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
