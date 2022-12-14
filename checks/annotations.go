package checks

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/alecthomas/go-thrift/parser"

	"thriftlint"
)

type AnnotationPattern struct {
	Nodes      []reflect.Type
	Annotation string
	Regex      string
}

type annotationsCheck struct {
	patterns map[reflect.Type]map[string]string
	checks   thriftlint.Checks
}

func CheckAnnotations(patterns []*AnnotationPattern, checks thriftlint.Checks) thriftlint.Check {
	patternsLUT := map[reflect.Type]map[string]string{}
	for _, pattern := range patterns {
		for _, node := range pattern.Nodes {
			mapping, ok := patternsLUT[node]
			if !ok {
				mapping = map[string]string{}
				patternsLUT[node] = mapping
			}
			mapping[pattern.Annotation] = pattern.Regex
		}
	}
	return &annotationsCheck{
		patterns: patternsLUT,
		checks:   checks,
	}
}

func (c *annotationsCheck) ID() string {
	return "annotations"
}

func (c *annotationsCheck) Checker() interface{} {
	return c.checker
}

func (c *annotationsCheck) checker(self interface{}) (messages thriftlint.Messages) {
	v := reflect.Indirect(reflect.ValueOf(self))
	var annotations []*parser.Annotation
	if annotationsField := v.FieldByName("Annotations"); annotationsField.IsValid() {
		annotations = annotationsField.Interface().([]*parser.Annotation)
	}
	if checks, ok := c.patterns[v.Type()]; ok {
		for _, annotation := range annotations {
			if pattern, ok := checks[annotation.Name]; ok {
				re := regexp.MustCompile("^(?:" + pattern + ")$")
				if !re.MatchString(annotation.Value) {
					messages.Warning(annotation, "invalid value %q for annotation %q (should match %q)",
						annotation.Value, annotation.Name, pattern)
				}
			} else if annotation.Name != "nolint" {
				messages.Warning(annotation, "unsupported annotation %q", annotation.Name)
			}
		}
	} else {
		for _, annotation := range annotations {
			if annotation.Name != "nolint" {
				messages.Warning(annotation, "unsupported annotation %q", annotation.Name)
			}
		}
	}

	for _, annotation := range annotations {
		if annotation.Name == "nolint" && annotation.Value != "" {
			lints := strings.Fields(annotation.Value)
			for _, l := range lints {
				if !c.checks.Has(l) {
					messages.Warning(annotation, "%q is not a known linter check", l)
				}
			}
		}
	}
	return
}
