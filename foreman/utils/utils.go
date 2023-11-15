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
		log.Fatal("not ok")
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

func Debug(format string, a ...interface{}) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		log.Fatal("not ok")
	}

	_, fileName := path.Split(file)

	args := []interface{}{fileName, line}
	args = append(args, a...)
	log.Debugf("%s:%d \n"+format, args...)
}
