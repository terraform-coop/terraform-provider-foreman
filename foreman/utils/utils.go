package utils

import (
	"github.com/HanseMerkur/terraform-provider-utils/log"
	"runtime"
	"strings"
)

// Provides util functions for the provider package

const prefix = "github.com/terraform-coop/terraform-provider-foreman/"

func TraceFunctionCall() {
	// Get the program counter and result from the Go stack
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		log.Warningf("Error in TraceFunctionCall runtime.Caller")
		return
	}

	// Get details about the caller function (the one that calls utils.TraceFunctionCall)
	fun := runtime.FuncForPC(pc)
	funName := fun.Name()
	funFile, funLine := fun.FileLine(pc)

	// Strip the package prefix
	funName = strings.TrimPrefix(funName, prefix)
	funFile = strings.TrimPrefix(funFile, prefix)

	log.Tracef("%s (called from %s:%d)", funName, funFile, funLine)
}

const (
	debug = iota
	fatal
)

// Inner function of Debugf and Fatalf to prevent duplicate code.
func logInnerFunc(level int, format string, a ...interface{}) {
	var logFunc func(string, ...interface{})
	switch level {
	case debug:
		logFunc = log.Debugf
	case fatal:
		logFunc = log.Tracef
	default:
		logFunc = log.Infof
	}

	_, file, line, ok := runtime.Caller(2)
	if !ok {
		// Just pass it through
		logFunc(format, a...)
	}

	file = strings.TrimPrefix(file, prefix)
	args := []interface{}{file, line}
	args = append(args, a...)
	logFunc("[%s:%d] "+format, args...)
}

// Like `log.Debugf` but also prints the current file name and line number with the log output
func Debugf(format string, a ...interface{}) {
	logInnerFunc(debug, format, a)
}

// Prints line and file and then exits with fatal error message
func Fatalf(format string, a ...interface{}) {
	logInnerFunc(fatal, format, a)
}

// Wrapper for single value output
func Fatal(a interface{}) {
	Fatalf("%s", a)
}
