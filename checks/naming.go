package checks

import (
	"reflect"
	"regexp"
	"strings"

	"thriftlint"
)

type NamingStyle struct {
	Name    string
	Pattern *regexp.Regexp
}

var (
	upperCamelCaseStyle = NamingStyle{
		Name:    "title case",
		Pattern: regexp.MustCompile(`^_?([A-Z][0-9a-z]*)*$`),
	}

	lowerCamelCaseStyle = NamingStyle{
		Name:    "camel case",
		Pattern: regexp.MustCompile(`^_?[a-z][A-Z0-9a-z]*$`),
	}

	upperSnakeCaseStyle = NamingStyle{
		Name:    "upper snake case",
		Pattern: regexp.MustCompile(`^_?[A-Z][A-Z0-9]*(_[A-Z0-9]+)*$`),
	}

	CheckNamesDefaults = map[reflect.Type]NamingStyle{
		thriftlint.ServiceType:   upperCamelCaseStyle,
		thriftlint.EnumType:      upperCamelCaseStyle,
		thriftlint.StructType:    upperCamelCaseStyle,
		thriftlint.EnumValueType: upperSnakeCaseStyle,
		thriftlint.FieldType:     lowerCamelCaseStyle,
		thriftlint.MethodType:    lowerCamelCaseStyle,
		thriftlint.ConstantType:  upperSnakeCaseStyle,
	}

	CheckNamesDefaultBlacklist = map[string]bool{
		"class": true,
		"int":   true,
	}
)

func CheckNames(matches map[reflect.Type]NamingStyle, blacklist map[string]bool) thriftlint.Check {
	if matches == nil {
		matches = CheckNamesDefaults
	}
	if blacklist == nil {
		blacklist = CheckNamesDefaultBlacklist
	}
	return thriftlint.MakeCheck("naming", func(v interface{}) (messages thriftlint.Messages) {
		rv := reflect.Indirect(reflect.ValueOf(v))
		nameField := rv.FieldByName("Name")
		if !nameField.IsValid() {
			return nil
		}
		name := nameField.Interface().(string)
		checker, ok := matches[rv.Type()]
		if !ok || strings.HasPrefix(name, "DEPRECATED_") {
			return nil
		}
		if blacklist[name] {
			messages.Warning(v, "%q is a disallowed name", name)
		}
		if ok := checker.Pattern.MatchString(name); !ok {
			messages.Warning(v, "name of %s %q should be %s", strings.ToLower(rv.Type().Name()),
				name, checker.Name)
		}
		return
	})
}
