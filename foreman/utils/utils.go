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
	err = iota
	warning
	info
	debug
	fatal
)

// Inner function of wrapper log functions below to prevent duplicate code.
func logInnerFunc(level int, format string, a ...interface{}) {
	var logFunc func(string, ...interface{})
	switch level {
	case err:
		logFunc = log.Errorf
	case warning:
		logFunc = log.Warningf
	case info:
		logFunc = log.Infof
	case debug:
		logFunc = log.Debugf
	case fatal:
		logFunc = log.Fatalf
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

// Like `log.Errorf` but also prints the current file name and line number with the log output
func Errorf(format string, a ...interface{}) {
	logInnerFunc(err, format, a...)
}

// Like `log.Warningf` but also prints the current file name and line number with the log output
func Warningf(format string, a ...interface{}) {
	logInnerFunc(warning, format, a...)
}

// Like `log.Infof` but also prints the current file name and line number with the log output
func Infof(format string, a ...interface{}) {
	logInnerFunc(info, format, a...)
}

// Like `log.Debugf` but also prints the current file name and line number with the log output
func Debugf(format string, a ...interface{}) {
	logInnerFunc(debug, format, a...)
}

// Like `log.Fatalf` but also prints the current file name and line number with the log output.
// Exits with fatal error.
func Fatalf(format string, a ...interface{}) {
	logInnerFunc(fatal, format, a...)
}

// Wrapper for Fatalf with only a single value output
func Fatal(a interface{}) {
	Fatalf("%s", a)
}
