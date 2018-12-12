// Package log extends the capabilities of the Golang stdlib "log"
// package.
//
// It defines a type, LevelLogger, which embeds the log.Logger type. This
// allows all of the exported log.Logger functions to be exported for the
// LevelLogger.
//
// Like the log package, this package exports logger helper functions
// (Print[f|ln], Fatal[f|ln], Panic[f|ln]) through a standard logger.  In this
// way, developers can import the logger and utilize the logging functionality
// without creating a LevelLogger reference.
//
// This package adds the concept of log levels to the standard logger. This is
// exposed through the exported 'LogLevel' type. The logger will filter out log
// messages based on the log level- only displaying messages that are of the
// logger's level or higher.  LevelLogger also tags each log message with
// the LogLevel as a prefix.
//
// This package exports a Printf-style function for each of the defined
// LogLevels - Debugf, Tracef, Infof, Warningf, and Errorf.
//
// This package defines the following log levels (from most verbose to least
// verbose):
//   LevelDebug - Intermediate values, calculations
//   LevelTrace - Function enter/exit notifications
//   LevelInfo - Informational messages, not used in debugging or runtime
//   LevelWarning - Errors; The system could recover gracefully
//   LevelError - Errors; The system could not recover gracefully
//   LevelNone - Suppresses all log output, no messages commited to output
package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

// -----------------------------------------------------------------------------
// Logging Level Constants / Enum Representation
// -----------------------------------------------------------------------------

// LogLevel type definition - see LogLevel constants defined below for
// more information.
type LogLevel int

// String - Implements the Stringer interface.  Given a log level, return the
// string representation. If an invalid LogLevel is provided, an empty string
// is returned.
func (l LogLevel) String() string {
	lInt := int(l)
	if lInt < 0 || lInt > len(logLevels) {
		return ""
	}
	return logLevels[lInt]
}

const LevelInvalid LogLevel = -1

// LogLevelFromStrings converts the given the string representation of the
// LogLevel, return the LogLevel. If an invalid string representation is
// provided, an error is returned and the LogLevel return corresponds to an
// invalid LogLevel.
func LogLevelFromString(name string) (LogLevel, error) {
	// List of log levels isn't too large - use linear search.
	// strings.EqualFold is used to perform case-insensitive search.
	for key, value := range logLevels {
		if strings.EqualFold(value, name) {
			return LogLevel(key), nil
		}
	}
	return LevelInvalid, fmt.Errorf("No LogLevel exists for string [%s]", name)
}

// Logging levels (ordered from most verbose, to least verbose).  Once a
// logger's level is set, messages with a log level greater than or equal to
// the logger's level are shown. Otherwise, they are ignored.
//
// For example, if the logger's log level is set to 'LevelWarning', only
// warning (LevelWarning) and error (LevelError) messages are shown.  The
// rest are discarded. Therefore, if a logger is set to 'LevelNone', *all*
// log messages are discarded - including errors.
const (
	LevelDebug LogLevel = iota
	LevelTrace
	LevelInfo
	LevelWarning
	LevelError
	LevelNone
)

// String representations of the above log levels.
//
// NOTE(ALL): The index of the string should correspond with the iota value
//   of the constant for easy conversion/mapping in LogLevelFromString()
var logLevels = [...]string{
	"DEBUG",
	"TRACE",
	"INFO",
	"WARNING",
	"ERROR",
	"NONE",
}

// -----------------------------------------------------------------------------
// Level Logger Instance
// -----------------------------------------------------------------------------

// LevelLogger extends the Logger type from the log package with the addition
// of log levels and level-based logging methods.
type LevelLogger struct {
	// Embedded Logger from 'pkg/log'.  This allows the level logger to
	// intercept all of the functionality of the Go stdlib logger package
	// and let us embed our own behavior.
	*log.Logger
	// Mutex for setting/modifying the log level.  Write operation synchronicity
	// are already handled by *log.Logger's mutex.
	mutex *sync.Mutex
	// The log level - see the constants defined above for more information.
	level LogLevel
}

// NewLeveLogger creates a new reference to a 'LevelLogger' with the provided
// output writer, logging flags, and log level threshold.
func NewLevelLogger(out io.Writer, flags int, logLevel LogLevel) *LevelLogger {
	return &LevelLogger{
		// initialize LevelLogger unique members
		level: logLevel,
		mutex: &sync.Mutex{},
		// pass the rest over to the Logger
		Logger: log.New(out, "", flags),
	}
}

// Level returns the LogLevel of the logger
func (logger *LevelLogger) Level() LogLevel {
	logger.mutex.Lock()
	defer logger.mutex.Unlock()
	return logger.level
}

// SetLevel sets the underlying LogLevel for the LevelLogger. The LogLevel
// should be a 'LogLevel' constant.
func (logger *LevelLogger) SetLevel(level LogLevel) {
	logger.mutex.Lock()
	defer logger.mutex.Unlock()
	logger.level = level
}

// -----------------------------------------------------------------------------
// LevelLogger Logging Functions - Logger Specific
// -----------------------------------------------------------------------------

// Debugf writes a debug message to the logs.  The message will only be written
// if the logger's level is less than or equal to LevelDebug.
//
// Debug messages can include intermediate values and calculations for
// stepping through the program.
func (logger *LevelLogger) Debugf(format string, a ...interface{}) {
	logger.mutex.Lock()
	defer logger.mutex.Unlock()
	if logger.level > LevelDebug {
		return
	}
	logger.Logger.Printf("[DEBUG] "+format, a...)
}

// Tracef writes a trace message to the logs.  The message will only be written
// if the logger's level is less than or equal to LevelTrace.
//
// Trace messages are used to log the flow of execution of the application such
// as function enter/exit notifications.
func (logger *LevelLogger) Tracef(format string, a ...interface{}) {
	logger.mutex.Lock()
	defer logger.mutex.Unlock()
	if logger.level > LevelTrace {
		return
	}
	logger.Logger.Printf("[TRACE] "+format, a...)
}

// Infof writes an info message to the logs.  The message will only be written
// if the logger's level is less than or equal to LevelInfo.
//
// Info messages are important system messages that do not require further
// action or communicate any type of error.
func (logger *LevelLogger) Infof(format string, a ...interface{}) {
	logger.mutex.Lock()
	defer logger.mutex.Unlock()
	if logger.level > LevelInfo {
		return
	}
	logger.Logger.Printf("[INFO ] "+format, a...)
}

// Warningf writes a warning message to the logs.  The message will only be
// written if the logger's level is less than or equal to LevelWarning.
//
// Warning messages are not as critical as an error and are used to communicate
// unsafe or potentially hazardous behavior that can be avoided or recovered.
func (logger *LevelLogger) Warningf(format string, a ...interface{}) {
	logger.mutex.Lock()
	defer logger.mutex.Unlock()
	if logger.level > LevelWarning {
		return
	}
	logger.Logger.Printf("[WARN ] "+format, a...)
}

// Errorf writes an error message to the logs.  The message will only be
// written if the logger's level is less than or equal to LevelError.
func (logger *LevelLogger) Errorf(format string, a ...interface{}) {
	logger.mutex.Lock()
	defer logger.mutex.Unlock()
	if logger.level > LevelError {
		return
	}
	logger.Logger.Printf("[ERROR] "+format, a...)
}

// -----------------------------------------------------------------------------
// LevelLogger Logging Functions - From Golang "log"
// -----------------------------------------------------------------------------

// Fatal writes a fatal message to the log using custom fatal prefix
//
// See 'pkg/log#Logger.Fatal()'
func (logger *LevelLogger) Fatal(a ...interface{}) {
	logger.Fatalf("%s", fmt.Sprint(a...))
}

// Fatalf writes a fatal message to the log using custom fatal prefix
//
// See 'pkg/log#Logger.Fatalf()'
func (logger *LevelLogger) Fatalf(format string, a ...interface{}) {
	logger.Logger.Fatalf("[FATAL] "+format, a...)
}

// Fataln writes a fatal message to the log using custom fatal prefix
//
// See 'pkg/log#Logger.Fatalln()'
func (logger *LevelLogger) Fatalln(a ...interface{}) {
	logger.Fatalf("%s", fmt.Sprintln(a...))
}

// Panic writes a panic message to the log using custom panic prefix
//
// See 'pkg/log#Logger.Panic()'
func (logger *LevelLogger) Panic(a ...interface{}) {
	logger.Panicf("%s", fmt.Sprint(a...))
}

// Panicf writes a panic message to the log using custom panic prefix
//
// See 'pkg/log#Logger.Panicf()'
func (logger *LevelLogger) Panicf(format string, a ...interface{}) {
	logger.Logger.Panicf("[PANIC] "+format, a...)
}

// Panicln writes a panic message to the log using custom panic prefix
//
// See 'pkg/log#Logger.Panicln()'
func (logger *LevelLogger) Panicln(a ...interface{}) {
	logger.Panicf("%s", fmt.Sprintln(a...))
}

// -----------------------------------------------------------------------------
// Built-in, Standard Logger
// -----------------------------------------------------------------------------

// Standard, package logger.  Functions similarly to the Go stdlib standard
// logger but with a few tweaks.
//
// Defaults to logging 'LevelInfo' messages or higher with standard flags
// out to StdErr.
var stdLog = NewLevelLogger(os.Stderr, log.LstdFlags, LevelInfo)

// -----------------------------------------------------------------------------
// Standard Log Utils - Logger Specific
// -----------------------------------------------------------------------------

// Level returns the LogLevel for the package-wide standard logger
func Level() LogLevel {
	return stdLog.Level()
}

// SetLevel sets the LogLevel for the package-wide standard logger. The
// LogLevel should be a 'LogLevel' constant.
func SetLevel(level LogLevel) {
	stdLog.SetLevel(level)
}

// -----------------------------------------------------------------------------
// Standard Log Utils - From Golang "log"
// -----------------------------------------------------------------------------

// SetOutput sets the output of the standard logger
//
// See 'pkg/log#SetOutput()'
func SetOutput(w io.Writer) {
	stdLog.Logger.SetOutput(w)
}

// Flags retrieves the flags of the standard logger
//
// See 'pkg/log#Flags()'
func Flags() int {
	return stdLog.Flags()
}

// SetFlags sets the flags of the standard logger
//
// See 'pkg/log#SetFlags()'
func SetFlags(flags int) {
	stdLog.SetFlags(flags)
}

// Prefix retrieves the prefix of the standard logger
//
// See 'pkg/log#Prefix()'
func Prefix() string {
	return stdLog.Prefix()
}

// SetPrefix sets the prefix of the standard logger
//
// See 'pkg/log#SetPrefix()'
func SetPrefix(prefix string) {
	stdLog.SetPrefix(prefix)
}

// -----------------------------------------------------------------------------
// Package Logging Functions - Logger Specific
// -----------------------------------------------------------------------------

// Debugf writes a debug message using the standard logger.
//
// See LevelLogger.Debugf()
func Debugf(format string, a ...interface{}) {
	stdLog.Debugf(format, a...)
}

// Tracef writes a trace message using the standard logger.
//
// See LevelLogger.Tracef()
func Tracef(format string, a ...interface{}) {
	stdLog.Tracef(format, a...)
}

// Infof writes an info message using the standard logger.
//
// See LevelLogger.Infof()
func Infof(format string, a ...interface{}) {
	stdLog.Infof(format, a...)
}

// Warningf writes a warning message using the standard logger.
//
// See LevelLogger.Warningf()
func Warningf(format string, a ...interface{}) {
	stdLog.Warningf(format, a...)
}

// Errorf writes an error message using the standard logger.
//
// See LevelLogger.Errorf()
func Errorf(format string, a ...interface{}) {
	stdLog.Errorf(format, a...)
}

// -----------------------------------------------------------------------------
// Package Logging Functions - From Golang "log"
// -----------------------------------------------------------------------------

// Fatal
//
// See 'pkg/log#Fatal()'
func Fatal(a ...interface{}) {
	stdLog.Fatal(a...)
}

// Fatalf
//
// See 'pkg/log#Fatalf()'
func Fatalf(format string, a ...interface{}) {
	stdLog.Fatalf(format, a...)
}

// Fatalln
//
// See 'pkg/log#Fatalln()'
func Fatalln(a ...interface{}) {
	stdLog.Fatalln(a...)
}

// Panic
//
// See 'pkg/log#Panic()'
func Panic(a ...interface{}) {
	stdLog.Panic(a...)
}

// Panicf
//
// See 'pkg/log#Panicf()'
func Panicf(format string, a ...interface{}) {
	stdLog.Panicf(format, a...)
}

// Panicln
//
// See 'pkg/log#Panicln()'
func Panicln(a ...interface{}) {
	stdLog.Panicln(a...)
}

// Print
//
// See 'pkg/log#Print()'
func Print(a ...interface{}) {
	stdLog.Print(a...)
}

// Printf
//
// See 'pkg/log#Printf()'
func Printf(format string, a ...interface{}) {
	stdLog.Printf(format, a...)
}

// Println
//
// See 'pkg/log#Println()'
func Println(a ...interface{}) {
	stdLog.Println(a...)
}
