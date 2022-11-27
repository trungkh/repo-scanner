package serror

import (
	"errors"
	"fmt"
	"path"
	"runtime"
	"strings"

	"github.com/gearintellix/serr"
	gerr "github.com/go-errors/errors"

	"repo-scanner/internal/utils/utstring"
)

// public

type (
	// SError object
	SError interface {
		Error() string
		Cause() error

		Key() string
		Code() int
		Title() string
		Comments() string
		CommentStack() []string

		Callers() []uintptr
		StackFrames() []gerr.StackFrame
		StackTraces(max int) []string

		Type() string
		File() string
		Line() int
		FN() string

		SetKey(key string)
		AddComments(msg ...string)
		AddCommentf(msg string, opts ...interface{})
		AddCommentsx(skip int, msg ...string)
		AddCommentfx(skip int, msg string, opts ...interface{})
		Sign(ctx HCtx)

		// Deprecated: Use AddComments instead.
		SetComments(note string)

		String() string
		SimpleString() string
		ColoredString() string

		Panic()
		Print()
		PrintWithColor()
		IsEqual(err error) bool
	}

	HCtx interface {
		CreateError(msg string, notes ...string) (errx SError)
		CreateErrorEx(err error, notes ...string) (errx SError)
		SignError(errx SError) SError
	}
)

var (
	rootPaths []string
)

func RegisterRootPath(paths []string) {
	rootPaths = append(rootPaths, paths...)
}

func RegisterThisAsRoot(cskip int, pskip int) SError {
	_, file, _, ok := runtime.Caller(cskip + 1)
	if !ok {
		return Newc("Failed to get path", "@")
	}

	sep := "/"
	if runtime.GOOS == "windows" {
		sep = "\\"
	}

	file = path.Dir(file)
	paths := strings.Split(file, sep)
	if len(paths) > pskip {
		paths = paths[:len(paths)-pskip]
	}
	RegisterRootPath([]string{strings.Join(paths, sep)})

	return nil
}

// New serror from error message
func New(message string) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2, stack[:])

	return construct(stack[:length], 0, "-", errors.New(message), 0, "@")
}

// Newk serror from error message and error key
func Newk(key string, message string) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2, stack[:])

	return construct(stack[:length], 0, key, errors.New(message), 0, "@")
}

// Newf serror from error message with function
func Newf(message string, args ...interface{}) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2, stack[:])

	return construct(stack[:length], 0, "-", fmt.Errorf(message, args...), 0, "@")
}

// Newkf serror from error message with function
func Newkf(key string, message string, args ...interface{}) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2, stack[:])

	return construct(stack[:length], 0, key, fmt.Errorf(message, args...), 0, "@")
}

// Newc serror from error message and comments
func Newc(message string, note string) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2, stack[:])

	return construct(stack[:length], 0, "-", errors.New(message), 0, note)
}

// Newkc serror from error message, error key, and comments
func Newkc(key string, message string, note string) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2, stack[:])

	return construct(stack[:length], 0, key, errors.New(message), 0, note)
}

// Newi serror from error code, and error message
func Newi(code int, message string) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2, stack[:])

	return construct(stack[:length], code, "-", errors.New(message), 0, "@")
}

// Newic serror from error code, error message, and comments
func Newic(code int, message string, note string) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2, stack[:])

	return construct(stack[:length], code, "-", errors.New(message), 0, note)
}

// Newif serror from error code, and error message with function
func Newif(code int, message string, args ...interface{}) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2, stack[:])

	return construct(stack[:length], code, "-", fmt.Errorf(message, args...), 0, "@")
}

// Newik serror from error code, error key, and error message
func Newik(code int, key string, message string) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2, stack[:])

	return construct(stack[:length], code, "-", errors.New(message), 0, "@")
}

// Newikf serror from error code, error key, and error message with function
func Newikf(code int, key string, message string, args ...interface{}) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2, stack[:])

	return construct(stack[:length], code, key, fmt.Errorf(message, args...), 0, "@")
}

// Newikc serror from error code, error key, error message, and comments
func Newikc(code int, key string, message string, note string) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2, stack[:])

	return construct(stack[:length], code, key, errors.New(message), 0, note)
}

// News serror from error message and skip stacktrace
func News(skip int, message string) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2+skip, stack[:])

	return construct(stack[:length], 0, "-", errors.New(message), skip, "@")
}

// Newsf serror from error message and skip stacktrace with function
func Newsf(skip int, message string, args ...interface{}) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2+skip, stack[:])

	return construct(stack[:length], 0, "-", fmt.Errorf(message, args...), skip, "@")
}

// Newsk serror from skip stacktrace, error key, and error message
func Newsk(skip int, key string, message string) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2+skip, stack[:])

	return construct(stack[:length], 0, key, errors.New(message), skip, "@")
}

// Newskf serror from skip stacktrace, error key, and error message with function
func Newskf(skip int, key string, message string, args ...interface{}) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2+skip, stack[:])

	return construct(stack[:length], 0, key, fmt.Errorf(message, args...), skip, "@")
}

// Newsc serror from skip stacktrace and comments
func Newsc(skip int, message string, note string) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2+skip, stack[:])

	return construct(stack[:length], 0, "-", errors.New(message), skip, note)
}

// Newskc serror from skip stacktrace, error key and comments
func Newskc(skip int, key string, message string, note string) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2+skip, stack[:])

	return construct(stack[:length], 0, key, errors.New(message), skip, note)
}

// Newsi serror from skip stacktrace and error code
func Newsi(skip int, code int, message string) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2+skip, stack[:])

	return construct(stack[:length], code, "-", errors.New(message), skip, "@")
}

// Newsic serror from skip stacktrace and comments
func Newsic(skip int, code int, message string, note string) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2+skip, stack[:])

	return construct(stack[:length], code, "-", errors.New(message), skip, note)
}

// Newsik serror from skip stacktrace, error code and error key
func Newsik(skip int, code int, key string, message string) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2+skip, stack[:])

	return construct(stack[:length], code, key, errors.New(message), skip, "@")
}

// Newsikc serror from skip stacktrace, error code, error key and comments
func Newsikc(skip int, code int, key string, message string, note string) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2+skip, stack[:])

	return construct(stack[:length], code, key, errors.New(message), skip, note)
}

// NewFromError function
func NewFromError(err error) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2, stack[:])

	return construct(stack[:length], 0, "-", err, 0, "@")
}

// NewFromErrork function
func NewFromErrork(key string, err error) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2, stack[:])

	return construct(stack[:length], 0, key, err, 0, "@")
}

// NewFromErrorc serror from error, and comments
func NewFromErrorc(err error, note string) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2, stack[:])

	return construct(stack[:length], 0, "-", err, 0, note)
}

// NewFromErrorkc serror from error, error key, and comments
func NewFromErrorkc(key string, err error, note string) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2, stack[:])

	return construct(stack[:length], 0, key, err, 0, note)
}

// NewFromErrori serror from error code, and error
func NewFromErrori(code int, err error) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2, stack[:])

	return construct(stack[:length], code, "-", err, 0, "@")
}

// NewFromErroric serror from error code, error, and comments
func NewFromErroric(code int, err error, note string) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2, stack[:])

	return construct(stack[:length], code, "-", err, 0, note)
}

// NewFromErrorik serror from error code, error key, and error
func NewFromErrorik(code int, key string, err error) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2, stack[:])

	return construct(stack[:length], code, key, err, 0, "@")
}

// NewFromErrorikc serror from error code, error key, error, and comments
func NewFromErrorikc(code int, key string, err error, note string) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2, stack[:])

	return construct(stack[:length], code, key, err, 0, note)
}

// NewFromErrors serror from skip stacktrace
func NewFromErrors(skip int, err error) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2+skip, stack[:])

	return construct(stack[:length], 0, "-", err, skip, "@")
}

// NewFromErrorsi serror from skip stacktrace and error code
func NewFromErrorsi(skip int, code int, err error) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2+skip, stack[:])

	return construct(stack[:length], code, "-", err, skip, "@")
}

// NewFromErrorsic serror from skip stacktrace and error code with comments
func NewFromErrorsic(skip int, code int, err error, note string) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2+skip, stack[:])

	return construct(stack[:length], code, "-", err, skip, note)
}

// NewFromErrorsk serror from skip stacktrace and error key
func NewFromErrorsk(skip int, key string, err error) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2+skip, stack[:])

	return construct(stack[:length], 0, key, err, skip, "@")
}

// NewFromErrorskc serror from skip stacktrace and error key with comments
func NewFromErrorskc(skip int, key string, err error, note string) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2+skip, stack[:])

	return construct(stack[:length], 0, key, err, skip, note)
}

// NewFromErrorsc serror from skip stacktrace with comments
func NewFromErrorsc(skip int, err error, note string) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2+skip, stack[:])

	return construct(stack[:length], 0, "-", err, skip, note)
}

// NewFromSErr serror from serr
func NewFromSErr(err serr.SErr) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2, stack[:])

	res := construct(stack[:length], err.Code(), err.Key(), err.Cause(), 0, "")

	resx := res.(*serrorObj)
	resx.comments = err.CommentStack()
	return res
}

// NewFromSErrc serror from serr, and comments
func NewFromSErrc(err serr.SErr, note string) SError {
	stack := make([]uintptr, 50)
	length := runtime.Callers(2, stack[:])

	res := construct(stack[:length], err.Code(), err.Key(), err.Cause(), 0, "")

	resx := res.(*serrorObj)
	resx.comments = err.CommentStack()

	if note == "@" {
		note = err.String()
	}
	res.AddComments(note)

	return res
}

// StandardFormat function
func StandardFormat() string {
	return "In %s[%s:%d] %s%s"
}

// StandardColorFormat function
func StandardColorFormat() string {
	frmt := ""
	frmt += utstring.ApplyForeColor("In", utstring.DARK_GRAY) + " "
	frmt += utstring.ApplyForeColor("%s", utstring.LIGHT_YELLOW)
	frmt += utstring.ApplyForeColor("[", utstring.DARK_GRAY)
	frmt += utstring.ApplyForeColor("%s:%d", utstring.MAGENTA)
	frmt += utstring.ApplyForeColor("]", utstring.DARK_GRAY)
	frmt += " %s%s"
	return frmt
}

// IsEqual to check are error same or not
func IsEqual(a error, b error) bool {
	if a == nil || b == nil {
		return (a == b)
	}

	if errx, ok := a.(SError); ok {
		a = errx.Cause()
	}

	if errx, ok := b.(SError); ok {
		b = errx.Cause()
	}

	return (a.Error() == b.Error())
}
