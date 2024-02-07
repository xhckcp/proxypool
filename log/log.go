package log

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

var (
	level = INFO
)

func init() {

	log.SetFormatter(&MyFormatter{})

	writer3, err := os.OpenFile("run.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		log.Fatalf("create file run.log failed: %v", err)
	}

	log.SetReportCaller(true)
	log.SetOutput(io.MultiWriter(os.Stderr, writer3))

}

func SetLevel(l LogLevel) {
	level = l
	log.SetLevel(levelMapping[level])
}

// func Traceln(format string, v ...interface{}) {
// 	log.Traceln(fmt.Sprintf(format, v...))
// }

// func Debugln(format string, v ...interface{}) {
// 	log.Debugln(fmt.Sprintf(format, v...))
// }

// func Infoln(format string, v ...interface{}) {
// 	log.Infoln(fmt.Sprintf(format, v...))
// }

// func Warnln(format string, v ...interface{}) {
// 	log.Warnln(fmt.Sprintf(format, v...))
// }

// func Errorln(format string, v ...interface{}) {
// 	log.Errorln(fmt.Sprintf(format, v...))
// }

// func Fileln(l LogLevel, data string) {
// 	if l >= level {
// 		if f := initFile(filepath.Join(logDir, logFile)); f != nil {
// 			fileMux.Lock()
// 			fileLogger.SetOutput(f)
// 			fileLogger.Logln(levelMapping[l], data)
// 			fileMux.Unlock()
// 			_ = f.Close()
// 		}
// 	}
// }

type MyFormatter struct {
	Prefix string
	Suffix string
}

// Format implement the Formatter interface
func (mf *MyFormatter) Format(entry *log.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	//  Get caller information
	if entry.HasCaller() {

		file := filepath.Base(entry.Caller.File)
		b.WriteString(fmt.Sprintf("[%s][%s][%s:%d %s]: %s\n", entry.Time.Format("2006-01-02T15:04:05.999-07:00"), entry.Level, file, entry.Caller.Line, entry.Caller.Function, entry.Message))

	} else {

		b.WriteString(fmt.Sprintf("[%s][%s]: %s\n", entry.Time.Format("2006-01-02T15:04:05.999-07:00"), entry.Level, entry.Message))

	}

	return b.Bytes(), nil
}

var (
	Trace   = log.Trace
	Traceln = log.Tracef // 将*ln重定向到*f, 解决原来日志调用代码中格式化字符串的问题
	Debug   = log.Debug
	Debugln = log.Debugf
	Info    = log.Info
	Infoln  = log.Infof
	Warn    = log.Warn
	Warnln  = log.Warnf
	Error   = log.Error
	Errorln = log.Errorf
	Fatal   = log.Fatal
	Fataln  = log.Fatalf
	Panic   = log.Panic
	Panicln = log.Panicf
	Printf  = log.Printf
	Print   = log.Print
	Println = log.Println
)
