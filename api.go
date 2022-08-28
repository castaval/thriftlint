package thriftlint

import (
	"fmt"
	"strings"

	"github.com/alecthomas/go-thrift/parser"
)

type Severity int

const (
	Warning Severity = iota
	Error
)

func (s Severity) String() string {
	if s == Warning {
		return "warning"
	}
	return "error"
}

type Message struct {
	File     *parser.Thrift
	Checker  string
	Severity Severity
	Object   interface{}
	Message  string
}

type Messages []*Message

func (w *Messages) Warning(object interface{}, msg string, args ...interface{}) Messages {
	message := &Message{Severity: Warning, Object: object, Message: fmt.Sprintf(msg, args...)}
	*w = append(*w, message)
	return *w
}

func (w *Messages) Error(object interface{}, msg string, args ...interface{}) Messages {
	message := &Message{Severity: Error, Object: object, Message: fmt.Sprintf(msg, args...)}
	*w = append(*w, message)
	return *w
}

type Checks []Check

func (c Checks) CloneAndDisable(prefixes ...string) Checks {
	out := Checks{}
skip:
	for _, check := range c {
		id := check.ID()
		for _, prefix := range prefixes {
			if prefix == id || strings.HasPrefix(id, prefix+".") {
				continue skip
			}
		}
		out = append(out, check)
	}
	return out
}

func (c Checks) Has(prefix string) bool {
	for _, check := range c {
		id := check.ID()
		if prefix == id || strings.HasPrefix(id, prefix+".") {
			return true
		}
	}
	return false
}

type Check interface {
	ID() string
	Checker() interface{}
}

func MakeCheck(id string, checker interface{}) Check {
	return &statelessCheck{
		id:      id,
		checker: checker,
	}
}

type statelessCheck struct {
	id      string
	checker interface{}
}

func (s *statelessCheck) ID() string {
	return s.id
}

func (s *statelessCheck) Checker() interface{} {
	return s.checker
}
