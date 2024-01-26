package utils

// Provides util functions for the provider package

func TraceFunctionCall() {
	// Removed in branch feat/job_templates, to be filled in separate branch
}

// Like `log.Debugf` but also prints the current file name and line number with the log output
func Debug(format string, a ...interface{}) {
	// Removed in branch feat/job_templates, to be filled in separate branch
}

func Debugf(format string, a ...interface{}) {
	// Placeholder in branch feat/katello_lifecycleenvs, to be filled by PR #146
	// https://github.com/terraform-coop/terraform-provider-foreman/pull/146
}

// Prints line and file and then exits with fatal error message
func Fatalf(format string, a ...interface{}) {
	// Removed in branch feat/job_templates, to be filled in separate branch
}

// Wrapper for single value output
func Fatal(a interface{}) {
	Fatalf("%s", a)
}
