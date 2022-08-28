package thriftlint

import (
	"io/ioutil"
	"log"
	"reflect"
	"strings"

	"github.com/alecthomas/go-thrift/parser"
)

type logger interface {
	Printf(format string, args ...interface{})
}

type Linter struct {
	checkers    Checks
	includeDirs []string
	log         logger
}

type Option func(*Linter)

func WithIncludeDirs(dirs ...string) Option {
	return func(l *Linter) { l.includeDirs = dirs }
}

func WithLogger(logger logger) Option {
	return func(l *Linter) { l.log = logger }
}

func Disable(checks ...string) Option {
	return func(l *Linter) {
		l.checkers = l.checkers.CloneAndDisable(checks...)
	}
}

func New(checks []Check, options ...Option) (*Linter, error) {
	ids := []string{}
	for _, check := range checks {
		ids = append(ids, check.ID())
	}
	l := &Linter{
		checkers: Checks(checks),
		log:      log.New(ioutil.Discard, "", 0),
	}
	for _, option := range options {
		option(l)
	}
	l.log.Printf("Linting with: %s", strings.Join(ids, ", "))
	return l, nil
}

func (l *Linter) Lint(sources []string) (Messages, error) {
	l.log.Printf("Parsing %d files", len(sources))
	files, err := Parse(l.includeDirs, sources)
	if err != nil {
		return nil, err
	}
	messages := Messages{}
	for _, file := range files {
		l.log.Printf("Linting %s", file.Filename)
		v := reflect.ValueOf(file)
		enabledChecks := l.checkers.CloneAndDisable()
		ancestors := []interface{}{file.Imports}
		messages = append(messages, l.walk(file, ancestors, v, enabledChecks)...)
	}
	return messages, nil
}

func (l *Linter) walk(file *parser.Thrift, ancestors []interface{}, v reflect.Value,
	enabledChecks Checks) (messages Messages) {
	originalNode := v
	v = reflect.Indirect(v)
	switch v.Kind() {
	case reflect.Struct:
		var annotations []*parser.Annotation
		if annotationsField := v.FieldByName("Annotations"); annotationsField.IsValid() {
			annotations = annotationsField.Interface().([]*parser.Annotation)
			for _, a := range annotations {
				if a.Name == "nolint" {
					if a.Value == "" {
						return
					}
					enabledChecks = enabledChecks.CloneAndDisable(strings.Fields(a.Value)...)
				}
			}
		}

		ancestors = append(ancestors, originalNode.Interface())
		for _, checker := range enabledChecks {
			id := checker.ID()
			for _, msg := range callChecker(checker.Checker(), ancestors) {
				msg.File = file
				msg.Checker = id
				messages = append(messages, msg)
			}
		}

		for i := 0; i < v.NumField(); i++ {
			ft := v.Type().Field(i)
			if ft.Name == "Pos" || (ft.Name == "Imports" && v.Type() == reflect.TypeOf(parser.Thrift{})) {
				continue
			}
			messages = append(messages, l.walk(file, ancestors, v.Field(i), enabledChecks)...)
		}

	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			messages = append(messages, l.walk(file, ancestors, v.Index(i), enabledChecks)...)
		}

	case reflect.Map:
		for _, key := range v.MapKeys() {
			messages = append(messages, l.walk(file, ancestors, v.MapIndex(key), enabledChecks)...)
		}
	}
	return
}

var emptyInterfaceValue interface{}
var emptyInterfaceType = reflect.TypeOf(&emptyInterfaceValue).Elem()

func callChecker(checker interface{}, ancestors []interface{}) Messages {
	l := reflect.TypeOf(checker)
	if l.Kind() != reflect.Func {
		panic("checker must be a function but is " + l.String())
	}
	if l.NumOut() != 1 || l.Out(0) != reflect.TypeOf(Messages{}) {
		panic("checkers must return exactly Messages")
	}

	args := []reflect.Value{}
	switch {
	case l.NumIn() == 1 && l.In(0) == emptyInterfaceType:
		args = append(args, reflect.ValueOf(ancestors[len(ancestors)-1]))

	case l.NumIn() == 2 && l.In(0) == emptyInterfaceType && l.In(1) == emptyInterfaceType:
		if len(ancestors) < 2 {
			return nil
		}
		args = append(args,
			reflect.ValueOf(ancestors[len(ancestors)-2]),
			reflect.ValueOf(ancestors[len(ancestors)-1]),
		)

	default:
		if reflect.TypeOf(ancestors[len(ancestors)-1]) != l.In(l.NumIn()-1) {
			return nil
		}

		ancestorsValues := []reflect.Value{}
		for _, a := range ancestors {
			ancestorsValues = append(ancestorsValues, reflect.ValueOf(a))
		}

		ancestorIndex := len(ancestorsValues) - 1
		for parameterIndex := l.NumIn() - 1; ancestorIndex >= 0 && parameterIndex >= 0; parameterIndex-- {
			for ancestorIndex >= 0 {
				arg := ancestorsValues[ancestorIndex]
				if arg.Type().ConvertibleTo(l.In(parameterIndex)) {
					args = append(args, arg)
					break
				}
				ancestorIndex--
			}
		}

		if len(args) != l.NumIn() {
			return nil
		}

		for i, j := 0, len(args)-1; i < j; i, j = i+1, j-1 {
			args[i], args[j] = args[j], args[i]
		}

	}
	out := reflect.ValueOf(checker).Call(args)
	return out[0].Interface().(Messages)
}
