package common

import (
    "runtime"
    "strings"

    "github.com/astaxie/beego/logs"
)

// Log 使用 beego 的 log 库
var Log *logs.BeeLogger

// BaseDir 日志打印在binary的根路径
var BaseDir string

func init() {
    Log = logs.NewLogger(0)
    Log.EnableFuncCallDepth(true)
}

// fileName get filename from path
func fileName(original string) string {
    i := strings.LastIndex(original, "/")
    if i == -1 {
        return original
    }
    return original[i+1:]
}

// LogIfError 简化if err != nil 打 Error 日志代码长度
func LogIfError(err error, format string, v ...interface{}) {
    if err != nil {
        _, fn, line, _ := runtime.Caller(1)
        if format == "" {
            format = "[%s:%d] %s"
            Log.Error(format, fileName(fn), line, err.Error())
        } else {
            format = "[%s:%d] " + " Error: %s, %s" + format
            Log.Error(format, fileName(fn), line, err.Error(), v)
        }
    }
}

// LogIfWarn 简化if err != nil 打 Warn 日志代码长度
func LogIfWarn(err error, format string, v ...interface{}) {
    if err != nil {
        _, fn, line, _ := runtime.Caller(1)
        if format == "" {
            format = "[%s:%d] %s"
            Log.Warn(format, fileName(fn), line, err.Error())
        } else {
            format = "[%s:%d] " + format + " Error: %s"
            Log.Warn(format, fileName(fn), line, v, err.Error())
        }
    }
}

