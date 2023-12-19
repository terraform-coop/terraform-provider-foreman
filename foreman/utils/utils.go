package utils

import (
	"github.com/HanseMerkur/terraform-provider-utils/log"
	"runtime"
	"strings"
)

// Provides util functions for the provider package

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
	const prefix = "github.com/terraform-coop/terraform-provider-foreman/"
	funName = strings.TrimPrefix(funName, prefix)
	funFile = strings.TrimPrefix(funFile, prefix)

	log.Tracef("%s (called from %s:%d)",
		funName, funFile, funLine)
}

// Like `log.Debugf` but also prints the current file name and line number with the log output
func Debug(format string, a ...interface{}) {
	// Removed in branch feat/job_templates, to be filled in separate branch
}

// Prints line and file and then exits with fatal error message
func Fatalf(format string, a ...interface{}) {
	// Removed in branch feat/job_templates, to be filled in separate branch
}

// Wrapper for single value output
func Fatal(a interface{}) {
	Fatalf("%s", a)
}
