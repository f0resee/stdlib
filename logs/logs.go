package logs

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// fork from https://github.com/Kidsunbo/kie_toolbox_go/blob/master/logs/logs.go

var logger *Logger

func init() {
	logger = New()
	logger.SetCallerDepth(3)
}

type RequestID struct{}

type Level int8

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

var colorFormat = []string{
	"\033[1;37m%s\033[0m",
	"\033[1;36m%s\033[0m",
	"\033[1;33m%s\033[0m",
	"\033[1;35m%s\033[0m",
	"\033[1;31m%s\033[0m",
}

func New() *Logger {
	return &Logger{
		logger:      log.New(os.Stdout, " ", 0),
		level:       LevelInfo,
		ctxFunc:     defaultContextFunction,
		pathLength:  1,
		callerDepth: 2,
		msgPrefix:   "",
	}
}

type Logger struct {
	logger      *log.Logger
	level       Level
	ctxFunc     func(ctx context.Context) string
	pathLength  int8
	callerDepth int
	msgPrefix   string
}

func (l *Logger) SetLevel(level Level) {
	l.level = level
}

func (l *Logger) SetOutputFile(filename string) {
	if filename == "" {
		l.logger.SetOutput(os.Stderr)
	} else {
		file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal(err)
		}
		l.logger.SetOutput(file)
	}
}

func (l *Logger) SetCallerDepth(depth int) {
	l.callerDepth = depth
}

func (l *Logger) SetPathLength(length int8) {
	l.pathLength = length
}

func (l *Logger) SetMessagePrefix(prefix string) {
	l.msgPrefix = prefix
}

func (l *Logger) getLogMetaInfo() string {
	pc, file, line, ok := runtime.Caller(l.callerDepth)
	if !ok {
		return ""
	}

	currentDatetime := l.getCurrentDatetime()
	currentFilePosition := l.getFilePosition(file, line)
	currentFunctionName := l.getCurrentFunctionName(pc)

	return fmt.Sprintf("%v %v [%v]", currentDatetime, currentFilePosition, currentFunctionName)
}

func (l *Logger) getCurrentFunctionName(pc uintptr) string {
	functionName := runtime.FuncForPC(pc).Name()
	idx := strings.LastIndex(functionName, ".")
	if idx == -1 {
		return functionName
	} else if idx == len(functionName) {
		return functionName
	}
	return functionName[idx+1:]
}

func (l *Logger) getCurrentDatetime() string {
	return time.Now().Format("2006-01-02 15:04:05.000")
}

func (l *Logger) getFilePosition(file string, line int) string {
	idx := len(file) - 1
	var count int8
	for ; idx >= 0; idx-- {
		if file[idx] == '/' {
			count++
		}
		if count == l.pathLength {
			break
		}
	}
	var filename string
	if idx == -1 {
		filename = file
	} else {
		filename = file[idx+1:]
	}
	return fmt.Sprintf("%v:%v", filename, line)
}

func (l *Logger) SetContextFunction(f func(ctx context.Context) string) {
	l.ctxFunc = f
}

func (l *Logger) Debug(format string, values ...interface{}) {
	if l.level > LevelDebug {
		return
	}
	rawLog := fmt.Sprintf(colorFormat[LevelDebug], stringify("[DEBUG] ", l.getLogMetaInfo(), fmt.Sprintf(l.msgPrefix+format, values...)))
	l.logger.Println(rawLog)
}

func (l *Logger) CtxDebug(ctx context.Context, format string, values ...interface{}) {
	if l.level > LevelDebug {
		return
	}
	var ctxStr string
	if l.ctxFunc != nil {
		ctxStr = l.ctxFunc(ctx)
	}
	rawLog := fmt.Sprintf(colorFormat[LevelDebug], stringify("[DEBUG] ", l.getLogMetaInfo(), ctxStr, fmt.Sprintf(l.msgPrefix+format, values...)))
	l.logger.Println(rawLog)
}

func (l *Logger) Info(format string, values ...interface{}) {
	if l.level > LevelInfo {
		return
	}
	rawLog := fmt.Sprintf(colorFormat[LevelInfo], stringify("[INFO] ", l.getLogMetaInfo(), fmt.Sprintf(l.msgPrefix+format, values...)))
	l.logger.Println(rawLog)
}

func (l *Logger) CtxInfo(ctx context.Context, format string, values ...interface{}) {
	if l.level > LevelInfo {
		return
	}
	var ctxStr string
	if l.ctxFunc != nil {
		ctxStr = l.ctxFunc(ctx)
	}
	rawLog := fmt.Sprintf(colorFormat[LevelInfo], stringify("[INFO] ", l.getLogMetaInfo(), ctxStr, fmt.Sprintf(l.msgPrefix+format, values...)))
	l.logger.Println(rawLog)
}

func (l *Logger) Warn(format string, values ...interface{}) {
	if l.level > LevelWarn {
		return
	}
	rawLog := fmt.Sprintf(colorFormat[LevelWarn], stringify("[WARN] ", l.getLogMetaInfo(), fmt.Sprintf(l.msgPrefix+format, values...)))
	l.logger.Println(rawLog)
}

func (l *Logger) CtxWarn(ctx context.Context, format string, values ...interface{}) {
	if l.level > LevelWarn {
		return
	}
	var ctxStr string
	if l.ctxFunc != nil {
		ctxStr = l.ctxFunc(ctx)
	}
	rawLog := fmt.Sprintf(colorFormat[LevelWarn], stringify("[WARN] ", l.getLogMetaInfo(), ctxStr, fmt.Sprintf(l.msgPrefix+format, values...)))
	l.logger.Println(rawLog)
}

func (l *Logger) Error(format string, values ...interface{}) {
	if l.level > LevelError {
		return
	}
	rawLog := fmt.Sprintf(colorFormat[LevelError], stringify("[ERROR] ", l.getLogMetaInfo(), fmt.Sprintf(l.msgPrefix+format, values...)))
	l.logger.Println(rawLog)
}

func (l *Logger) CtxError(ctx context.Context, format string, values ...interface{}) {
	if l.level > LevelError {
		return
	}
	var ctxStr string
	if l.ctxFunc != nil {
		ctxStr = l.ctxFunc(ctx)
	}
	rawLog := fmt.Sprintf(colorFormat[LevelError], stringify("[ERROR] ", l.getLogMetaInfo(), ctxStr, fmt.Sprintf(l.msgPrefix+format, values...)))
	l.logger.Println(rawLog)
}

func (l *Logger) Fatal(format string, values ...interface{}) {
	if l.level > LevelFatal {
		return
	}
	rawLog := fmt.Sprintf(colorFormat[LevelFatal], stringify("[FATAL] ", l.getLogMetaInfo(), fmt.Sprintf(l.msgPrefix+format, values...)))
	l.logger.Println(rawLog)
	panic("panic happened because fatal is reported")
}

func (l *Logger) CtxFatal(ctx context.Context, format string, values ...interface{}) {
	if l.level > LevelFatal {
		return
	}
	var ctxStr string
	if l.ctxFunc != nil {
		ctxStr = l.ctxFunc(ctx)
	}
	rawLog := fmt.Sprintf(colorFormat[LevelFatal], stringify("[FATAL] ", l.getLogMetaInfo(), ctxStr, fmt.Sprintf(l.msgPrefix+format, values...)))
	l.logger.Println(rawLog)
	panic("panic happened because fatal is reported")
}

func defaultContextFunction(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	result := make([]string, 0, 2)
	if ctx.Value(RequestID{}) != nil {
		requestID, ok := ctx.Value(RequestID{}).(string)
		if ok {
			result = append(result, requestID)
		}
	}

	return strings.Join(result, " ")
}

func Default() *Logger {
	return logger
}

func Debug(format string, values ...interface{}) {
	Default().Debug(format, values...)
}

func CtxDebug(ctx context.Context, format string, values ...interface{}) {
	Default().CtxDebug(ctx, format, values...)
}

func Info(format string, values ...interface{}) {
	Default().Info(format, values...)
}

func CtxInfo(ctx context.Context, format string, values ...interface{}) {
	Default().CtxInfo(ctx, format, values...)
}

func Warn(format string, values ...interface{}) {
	Default().Warn(format, values...)
}

func CtxWarn(ctx context.Context, format string, values ...interface{}) {
	Default().CtxWarn(ctx, format, values...)
}

func Error(format string, values ...interface{}) {
	Default().Error(format, values...)
}

func CtxError(ctx context.Context, format string, values ...interface{}) {
	Default().CtxError(ctx, format, values...)
}

func Fatal(format string, values ...interface{}) {
	Default().Fatal(format, values...)
}

func CtxFatal(ctx context.Context, format string, values ...interface{}) {
	Default().CtxFatal(ctx, format, values...)
}

func stringify(ss ...string) string {
	nonEmpty := make([]string, 0, len(ss))
	for _, s := range ss {
		if s == "" {
			continue
		}
		nonEmpty = append(nonEmpty, s)
	}
	return strings.Join(nonEmpty, " ")
}
