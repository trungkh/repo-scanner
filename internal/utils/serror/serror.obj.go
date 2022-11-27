package serror

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	gerr "github.com/go-errors/errors"

	"repo-scanner/internal/utils/utint"
	"repo-scanner/internal/utils/utstring"
)

type (
	serrorObj struct {
		err      error
		key      string
		comments []string
		code     int

		frames []gerr.StackFrame
		stack  []uintptr
	}
)

func construct(stack []uintptr, code int, key string, err error, skip int, note string) SError {
	var (
		res    *serrorObj
		isSErr bool
	)

	if err != nil {
		if errx, ok := err.(SError); ok {
			isSErr = true
			res = &serrorObj{
				err:      errx.Cause(),
				key:      errx.Key(),
				code:     errx.Code(),
				comments: errx.CommentStack(),
			}

			for _, v := range errx.StackFrames() {
				res.stack = append(res.stack, v.ProgramCounter)
			}

			res.key = utstring.Chains(key, res.key)
			if code != 0 {
				res.code = code
			}
		}
	}

	if res == nil {
		res = &serrorObj{
			err:   err,
			key:   key,
			code:  code,
			stack: stack,
		}
	}

	if !isSErr && err != nil && note == "@" {
		note = err.Error()
	}

	if note != "" {
		res.AddCommentsx(skip+2, note)
	}
	return res
}

// Error function
func (ox serrorObj) Error() string {
	return fmt.Sprintf("%v", ox.err)
}

// Cause function
func (ox serrorObj) Cause() error {
	return ox.err
}

// Code to get error code
func (ox serrorObj) Code() int {
	return ox.code
}

// Key to get error key
func (ox serrorObj) Key() string {
	return ox.key
}

// Callers to get callers
func (ox serrorObj) Callers() []uintptr {
	return ox.stack
}

// StackFrames to get stack frames
func (ox *serrorObj) StackFrames() []gerr.StackFrame {
	if ox.frames == nil {
		ox.frames = make([]gerr.StackFrame, len(ox.stack))

		for i, pc := range ox.stack {
			item := gerr.NewStackFrame(pc)
			item.File = getPath(item.File)
			ox.frames[i] = item
		}
	}

	return ox.frames
}

// StackTraces to get stack traces
func (ox *serrorObj) StackTraces(max int) []string {
	var (
		traces = make([]string, max)
		frames = ox.StackFrames()
	)

	if max <= 0 {
		max = len(frames)
	}

	for i := utint.MaxInt(len(frames)-max, 0); i < len(frames); i++ {
		frame := frames[i]
		traces = append(traces, fmt.Sprintf("%d: %s [%s:%d]", i, frame.Package, frame.File, frame.LineNumber))
	}

	return traces
}

// Type to get error type
func (ox serrorObj) Type() string {
	return reflect.TypeOf(ox.err).String()
}

// File to get file
func (ox *serrorObj) File() string {
	frames := ox.StackFrames()
	if len(frames) > 0 {
		return frames[0].File
	}
	return ""
}

// Line to get file line
func (ox *serrorObj) Line() int {
	frames := ox.StackFrames()
	if len(frames) > 0 {
		return frames[0].LineNumber
	}
	return 0
}

// FN to get function
func (ox *serrorObj) FN() string {
	frames := ox.StackFrames()
	if len(frames) > 0 {
		return frames[0].Name
	}
	return ""
}

// Title to get error title
func (ox serrorObj) Title() string {
	if len(ox.comments) > 0 {
		return ox.comments[0]
	}

	return ox.Error()
}

// Comments to get error comments
func (ox serrorObj) Comments() string {
	return strings.Join(ox.comments, "， ")
}

// CommentStack to get error comment stacks
func (ox serrorObj) CommentStack() []string {
	return ox.comments
}

// SetKey function
func (ox *serrorObj) SetKey(key string) {
	ox.key = key
}

func (ox *serrorObj) setRawComment(note string, skip int) {
	if ox.comments == nil && len(ox.comments) <= 0 {
		ox.comments = []string{}
	}

	if utstring.Length(note) <= 0 {
		return
	}

	if len(ox.comments) <= 0 {
		ox.comments = append(ox.comments, strings.ToUpper(string(note[0]))+string(note[1:]))
		return
	}

	_, file, line, _ := runtime.Caller(skip + 1)
	ox.comments = append(ox.comments, fmt.Sprintf("%s on [%s:%d]", note, getPath(file), line))
}

// SetComments function
//
// Deprecated: Use AddComments instead.
func (ox *serrorObj) SetComments(note string) {
	ox.setRawComment(note, 1)
}

// AddComments function
func (ox *serrorObj) AddComments(msg ...string) {
	for _, v := range msg {
		ox.setRawComment(v, 1)
	}
}

// AddCommentf function
func (ox *serrorObj) AddCommentf(msg string, opts ...interface{}) {
	ox.setRawComment(fmt.Sprintf(msg, opts...), 1)
}

// AddCommentsx function
func (ox *serrorObj) AddCommentsx(skip int, msg ...string) {
	for _, v := range msg {
		ox.setRawComment(v, 1+skip)
	}
}

// Sign function
func (ox *serrorObj) Sign(ctx HCtx) {
	_ = ctx.SignError(ox)
}

// AddCommentfx function
func (ox *serrorObj) AddCommentfx(skip int, msg string, opts ...interface{}) {
	ox.setRawComment(fmt.Sprintf(msg, opts...), 1+skip)
}

// String function
func (ox serrorObj) String() string {
	return fmt.Sprintf(StandardFormat(), ox.fParams()...)
}

// ColoredString function
func (ox serrorObj) ColoredString() string {
	return fmt.Sprintf(StandardColorFormat(), ox.fParams()...)
}

// SimpleString function
func (ox serrorObj) SimpleString() string {
	msg := ox.Error()
	if len(ox.comments) > 0 {
		msg = fmt.Sprintf("%s， detail: %s", ox.Comments(), msg)
	}

	return msg
}

// Panic to throw error to panic
func (ox serrorObj) Panic() {
	defer exit()
	if isLocal() {
		ox.PrintWithColor()
		return
	}

	ox.Print()
}

// Print function
func (ox serrorObj) Print() {
	printErr(ox.String())
}

// PrintWithColor function
func (ox serrorObj) PrintWithColor() {
	printErr(ox.ColoredString())
}

// IsEqual to check are error same or not
func (ox serrorObj) IsEqual(err error) bool {
	return IsEqual(ox.Cause(), err)
}

// private

func (ox serrorObj) fParams() []interface{} {
	pars := []interface{}{
		ox.FN(),
		ox.File(),
		ox.Line(),
		"",
		ox.Error(),
	}
	if ox.comments != nil && len(ox.comments) > 0 {
		pars[3] = fmt.Sprintf("%s， details: ", ox.Comments())
	}
	return pars
}
