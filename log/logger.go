/*
 * Copyright 2020 sqlpump Author. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package log

import (
	"runtime"
	"strings"
	"github.com/astaxie/beego/logs"
	"runtime/debug"
	"os"
	"fmt"
)

// Log 使用 beego 的 log 库
var Log *logs.BeeLogger

func LogSetting(jsonConfig string) {
	Log = logs.NewLogger(1000)
	Log.EnableFuncCallDepth(true)
	Log.SetLogger("file", jsonConfig)
	Log.SetLevel(logs.LevelDebug)
	Log.Async()
}

// fileName get filename from path
func fileName(original string) string {
	i := strings.LastIndex(original, "/")
	if i == -1 {
		return original
	}
	return original[i+1:]
}

// 打印error日志
func LogIfError(err error, format string, v ...interface{}) {
	if err != nil {
		_, fn, line, _ := runtime.Caller(1)
		if format == "" {
			format = "[%s:%d] %s"
			errInfo := err.Error()
			stackInfo := "\n" + string(debug.Stack()) + "\n"
			Log.Error(format, fileName(fn), line, errInfo + stackInfo)
			Log.Flush()
			fmt.Println("{\n\"resultCode\": 1,\n\"sqlPath\": \"\",\n\"errorInfo\": \"" + errInfo + "\",\n\"panicInfo\": \"\",\n\"stackInfo\": \"" + stackInfo + "\"\n}")
			os.Exit(1)
		} else {
			format = "[%s:%d] %s" + format
			errInfo := err.Error()
			stackInfo := "\n" + string(debug.Stack()) + "\n"
			Log.Error(format, fileName(fn), line, errInfo + stackInfo, v)
			fmt.Println("{\n\"resultCode\": 1,\n\"sqlPath\": \"\",\n\"errorInfo\": \"" + errInfo + "\",\n\"panicInfo\": \"\",\n\"stackInfo\": \"" + stackInfo + "\"\n}")
			Log.Flush()
			os.Exit(1)
		}
	}
}

// 打印warn日志
func LogIfWarn(err error, format string, v ...interface{}) {
	if err != nil {
		_, fn, line, _ := runtime.Caller(1)
		if format == "" {
			format = "[%s:%d] %s"
			Log.Warn(format, fileName(fn), line, err.Error())
		} else {
			format = "[%s:%d] " + format + " Warn: %s"
			Log.Warn(format, fileName(fn), line, v, err.Error())
		}
	}
}

// 打印info日志
func LogIfInfo(info string, format string, v ...interface{}) {
	_, fn, line, _ := runtime.Caller(1)
	if format == "" {
		format = "[%s:%d] %s"
		Log.Info(format, fileName(fn), line, info)
	} else {
		format = "[%s:%d] " + format + " Info: %s"
		Log.Info(format, fileName(fn), line, info, v)
	}
}
