package utils

import (
	//	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"

	//	"sync/atomic"
	"time"
)

type LoggerProxy struct {
	Logger *log.Logger
	Tag    string
}

var (
	// key为module-time，
	mLogInstance map[string]*LoggerProxy
)

var LEVEL_FATAL int = 1
var LEVEL_PANIC int = 2
var LEVEL_ERROR int = 3
var LEVEL_WARNING int = 4
var LEVEL_INFO int = 5
var LEVEL_DEBUG int = 6

var LEVEL_MAP map[int]string = map[int]string{
	LEVEL_FATAL:   "FATAL",
	LEVEL_PANIC:   "PANIC",
	LEVEL_ERROR:   "ERROR",
	LEVEL_WARNING: "WARNING",
	LEVEL_INFO:    "INFO",
	LEVEL_DEBUG:   "DEBUG",
}

var LEVEL_DEFAULT int = 7

var LOG_PATH = "log"

var enableDebug = false

func SetLogDebug(enable bool) {
	enableDebug = enable
}

func GetLogger(module string) *LoggerProxy {

	// TODO 做清理

	if mLogInstance == nil {
		mLogInstance = map[string]*LoggerProxy{}
	}

	// 获取当前的日期时间
	date := time.Now().Format("2006-01-02")

	//
	key := fmt.Sprintf("%s-%s", module, date)

	result := mLogInstance[key]
	// 如果包含则直接返回
	if result != nil {

		return result
	}

	// 否则建立新的log.Logger对象返回

	var out io.Writer
	file := fmt.Sprintf("%s/%s.log", LOG_PATH, key)

	if !CheckFileExist(LOG_PATH) {
		MakeDir(LOG_PATH)
	}

	out, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		return &LoggerProxy{Logger: nil}
	}

	logger := &LoggerProxy{Logger: log.New(out, "", 0), Tag: module}

	mLogInstance[key] = logger

	return logger
}

func Logger() *LoggerProxy {
	return GetLogger("default")
}

func (l *LoggerProxy) Log(level int, tag string, message string) {
	// 设置日志输出级别
	if level > LEVEL_DEFAULT {
		return
	}

	// 判断是否控关闭Log功能
	if !enableDebug {
		return
	}
	output := l.generateMessage(level, tag, message)
	// 判断是否控制台输出
	if IsDebug() {
		fmt.Println(output)
		return
	}

	// 写入到文件中
	l.Logger.Output(2, output)
}

// 生成每行的数据
func (l *LoggerProxy) generateMessage(level int, tag string, message string) string {
	//TODO  组织message，生成： [time][level][tag][code]message格式

	timeNow := time.Now().Format("2006-01-02 15:04:05")

	levelStr := LEVEL_MAP[level]

	_, file, line, _ := runtime.Caller(3)
	fileSplits := strings.Split(file, "/")

	fileName := fileSplits[len(fileSplits)-1]

	lenMsg := len(message)
	if lenMsg > 2 && message[0:1] == "[" && message[len(message)-1:] == "]" {
		message = message[1 : lenMsg-1]
	}

	outPut := fmt.Sprintf("[%s][%s][%s][%s %d]%s", timeNow, levelStr, tag, fileName, line, message)

	return outPut
}

func (l *LoggerProxy) Fatal(message ...interface{}) {

	l.Log(LEVEL_FATAL, l.Tag, fmt.Sprint(message))
}

func (l *LoggerProxy) Panic(message ...interface{}) {

	l.Log(LEVEL_PANIC, l.Tag, fmt.Sprint(message))
}

func (l *LoggerProxy) Error(message ...interface{}) {

	l.Log(LEVEL_ERROR, l.Tag, fmt.Sprint(message))
}

func (l *LoggerProxy) Warn(message ...interface{}) {

	l.Log(LEVEL_WARNING, l.Tag, fmt.Sprint(message))
}

func (l *LoggerProxy) Info(message ...interface{}) {

	l.Log(LEVEL_INFO, l.Tag, fmt.Sprint(message))
}

func (l *LoggerProxy) Debug(message ...interface{}) {
	l.Log(LEVEL_DEBUG, l.Tag, fmt.Sprint(message))
}
