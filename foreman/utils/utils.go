package utils

// Provides util functions for the provider package

import (
	"path"
	"runtime"
	"strings"

	"github.com/HanseMerkur/terraform-provider-utils/log"
)

// Source: https://www.reddit.com/r/golang/comments/f04v0l/debug_and_current_function_name/
func TraceFunctionCall() {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		log.Fatal("TraceFunctionCall failed")
	}

	dirName, fileName := path.Split(file)

	parts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	pl := len(parts)
	funcName := parts[pl-1]
	if parts[pl-2][0] == '(' {
		funcName = parts[pl-2] + "." + funcName
	}

	prefix_path := strings.TrimPrefix(dirName, "github.com/terraform-coop/terraform-provider-foreman/")

	log.Tracef("%s (%s%s:%d)", funcName, prefix_path, fileName, line)
}

// Like `log.Debugf` but also prints the current file name and line number with the log output
func Debug(format string, a ...interface{}) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		log.Fatal("Debug() failed")
	}

	_, fileName := path.Split(file)

	args := []interface{}{fileName, line}
	args = append(args, a...)
	log.Debugf("%s:%d \n"+format, args...)
}

// Prints line and file and then exits with fatal error message
func Fatalf(format string, a ...interface{}) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		log.Fatal("Fatal() failed")
	}

	_, fileName := path.Split(file)

	args := []interface{}{fileName, line}
	args = append(args, a...)
	log.Fatalf("%s:%d \n"+format, args...)
}

// Wrapper for single value output
func Fatal(a interface{}) {
	Fatalf("%s", a)
}
