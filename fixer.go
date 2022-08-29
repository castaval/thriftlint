package thriftlint

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/alecthomas/go-thrift/parser"
)

type Fixer struct {
	Pos     parser.Pos
	Message *Message
}

func NewFixer(pos parser.Pos, message *Message) *Fixer {
	return &Fixer{Pos: pos, Message: message}
}

func (f *Fixer) FixWarning() {
	file, err := os.ReadFile(f.Message.File.Filename)
	if err != nil {
		return
	}

	fileString := string(file)
	temp := strings.Split(fileString, "\n")
	var fixFileSlice []string

	fixedRow := f.fixRow()
	for index, item := range temp {

		if f.Pos.Line == index+1 {
			fixFileSlice = append(fixFileSlice, fixedRow)
			continue
		}

		fixFileSlice = append(fixFileSlice, item)
	}

	fixedFile := []byte(strings.Join(fixFileSlice, "\n"))

	os.WriteFile(f.Message.File.Filename, fixedFile, 0666)

}

func (f *Fixer) fixRow() (fixedRow string) {
	messageType := f.Message.Checker
	switch messageType {
	case "naming":
		rv := reflect.Indirect(reflect.ValueOf(f.Message.Object))
		pos := rv.FieldByName("Pos")
		col := pos.FieldByName("Col").Interface().(int)
		name := rv.FieldByName("Name")
		id := rv.FieldByName("ID").Int()
		types := rv.FieldByName("Type")

		optional := "optional"

		fmt.Println(pos, col, id, optional, name, types)
		row := fmt.Sprint(strings.Repeat(" ", col-1) + "%d: %s %s %s")
		fixedRow = fmt.Sprintf(row, id, optional, types, name)
		return
	}
	return
}
