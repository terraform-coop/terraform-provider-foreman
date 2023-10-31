package foreman

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

	_, fileName := path.Split(file)
	parts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	pl := len(parts)
	packageName := ""
	funcName := parts[pl-1]

	if parts[pl-2][0] == '(' {
		funcName = parts[pl-2] + "." + funcName
		packageName = strings.Join(parts[0:pl-2], ".")
	} else {
		packageName = strings.Join(parts[0:pl-1], ".")
	}

	_ = packageName
	// log.Tracef("TraceFunctionCall from %s the end:", pc, file, fileName, packageName, funcName, line)
	log.Tracef("TraceFunctionCall from %s:%d in func %s", fileName, line, funcName)
}
