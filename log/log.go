// provides a flexible logging system that supports both shell-based
// and standard logging outputs with different severity levels and colored output
package log

import (
	"fmt"
	"net/http"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

// logging level constants
const (
	FATAL = 0 // only fatal messages will be logged
	WARN  = 1 // warning and fatal messages will be logged
	INFO  = 2 // all messages (info, warning, fatal) will be logged
)

var (
	// IsShellMode determines whether to output colored text to shell (true)
	// or use structured logging via logrus (false)
	IsShellMode = false

	// color functions for different types of messages
	successCol = color.New(color.FgGreen).SprintFunc()
	infoCol    = color.New(color.FgWhite).SprintFunc()
	warnCol    = color.New(color.FgYellow).SprintFunc()
	errCol     = color.New(color.FgRed).SprintFunc()
)

// configures the minimum logging level for the application, maps the custom log levels to logrus levels
func SetLoggingLevel(l int) {
	switch l {
	case FATAL:
		logrus.SetLevel(logrus.FatalLevel)
	case WARN:
		logrus.SetLevel(logrus.WarnLevel)
	case INFO:
		logrus.SetLevel(logrus.InfoLevel)
	}
}

// logs a success message. In shell mode, it prints in green color,
// in normal mode, it logs as an info message through logrus
func Success(format string, args ...interface{}) {
	if IsShellMode {
		s := fmt.Sprintf(successCol(format), args...)
		fmt.Println(s)
		return
	}
	logrus.Info(fmt.Sprintf(format, args...))
}

// prints a prompt message to stdout in the info color
func Prompt(p string) {
	fmt.Print(infoCol(p))
}

// logs an informational message. In shell mode, it prints in white color,
// in normal mode, it logs through logrus at info level
func Info(format string, args ...interface{}) {
	if IsShellMode {
		s := fmt.Sprintf(infoCol(format), args...)
		fmt.Println(s)
		return
	}
	logrus.Info(fmt.Sprintf(format, args...))
}

// writes an info message to both an http.ResponseWriter and the log,
// useful for HTTP handlers that need to log their responses
func WInfo(w http.ResponseWriter, format string, args ...interface{}) {
	fmt.Fprintf(w, format, args...)
	Info(format, args...)
}

// logs a warning message. In shell mode, it prints in yellow color,
// in normal mode, it logs through logrus at warning level
func Warn(format string, args ...interface{}) {
	if IsShellMode {
		s := fmt.Sprintf(warnCol(format), args...)
		fmt.Println(s)
		return
	}
	logrus.Warnf(format, args...)
}

// writes a warning message to both an http.ResponseWriter and the log,
// useful for HTTP handlers that need to log their warning responses
func WWarn(w http.ResponseWriter, format string, args ...interface{}) {
	fmt.Fprintf(w, format, args...)
	Warn(format, args...)
}

// logs a fatal error and terminates the program. In shell mode, it prints in red color and panics,
// in normal mode, it logs through logrus at fatal level
func Fatal(err error) {
	if IsShellMode {
		s := fmt.Sprintf(errCol("fatal: %s"), err.Error())
		fmt.Println(s)
		panic(err)
	}
	logrus.Fatal(err)
}
