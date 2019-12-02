package logger

import (
	"fmt"
)

var DelayDayFormat = "20060102"
var DelayHourFormat = "15"
var DelayMinuteFormat = "04"

const (
	FileLoggerDriver string = "file"
	MQLoggerDriver          = "mq"
)

const (
	SplitAsSize int = iota
	SplitAsDelayDay
	SplitAsDelayHour
	SplitAsDelayMinute
)

const (
	Ldate         = 1 << iota     // 日期:  2009/01/23
	Ltime                         // 时间:  01:23:23
	Lmicroseconds                 // 微秒:  01:23:23.123123.
	Llongfile                     // 路径+文件名+行号: /a/b/c/d.go:23
	Lshortfile                    // 文件名+行号:   d.go:23
	LUTC                          // 使用标准的UTC时间格式
	LstdFlags     = Ldate | Ltime // 默认
)

const (
	LevelInfo = iota
	LevelWarning
	LevelDebug
	LevelError
)

type LoggerDriver interface {
	Info(format string, v ...interface{})
	Warning(format string, v ...interface{})
	Debug(format string, v ...interface{})
	Error(format string, v ...interface{})
	Close() error
	Initialize(logDir, logName string, loggerSplit, level, flag int, args ...int)
}

var drivers = make(map[string]LoggerDriver)

func newDriver(name, logDir, logName string, loggerSplit, level, flag int, args ...int) LoggerDriver {
	driver, ok := drivers[name]
	if !ok {
		panic("unknow logger driver")
	}
	driver.Initialize(logDir, logName, loggerSplit, level, flag, args...)
	return driver
}

func New(name, logDir, logName string, loggerSplit, level, flag int, args ...int) LoggerDriver {
	return newDriver(name, logDir, logName, loggerSplit, level, flag, args...)
}

func regDrivers(name string, driver LoggerDriver) {
	if driver == nil {
		panic("logger driver is nil")
	}

	if _, ok := drivers[name]; ok {
		panic(fmt.Sprintf("driver %s had registered", name))
	}

	drivers[name] = driver
}
