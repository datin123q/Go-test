package logger

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"xorm.io/xorm/log"
)

// xormLogger is the default implment of core.ILogger
type xormLogger struct {
	showSQL  bool
	logLevel log.LogLevel
}

func NewXorm() log.Logger {
	return &xormLogger{
		showSQL: true,
	}
}

// Error implement core.ILogger
func (x *xormLogger) Error(v ...interface{}) {
	logrus.Error(fmt.Sprint(v...))
}

func (x *xormLogger) Errorf(format string, v ...interface{}) {
	logrus.Errorf(format, v...)
}

func (x *xormLogger) Debug(v ...interface{}) {
	logrus.Debug(fmt.Sprint(v...))
}

func (x *xormLogger) Debugf(format string, v ...interface{}) {
	logrus.Debugf(format, v...)
}

func (x *xormLogger) Info(v ...interface{}) {
	logrus.Info(fmt.Sprint(v...))
}

func (x *xormLogger) Infof(format string, v ...interface{}) {

	argsReplacer := strings.NewReplacer("?", "'%v'")
	queryArgs := make([]interface{}, 0)
	queryFormat := argsReplacer.Replace(fmt.Sprintf("%v", v[1]))

	if args, ok := v[2].([]interface{}); ok {
		queryArgs = args
	}

	entry := logrus.WithFields(logrus.Fields{
		"query": fmt.Sprintf(queryFormat, queryArgs...),
	})

	requestId, ok := v[0].(string)
	if ok {
		replacer := strings.NewReplacer("[", "", "]", "")
		entry = entry.WithField("request_id", replacer.Replace(requestId))
	}

	if v[3] != nil {
		entry = entry.WithField("latency", fmt.Sprintf("%v", v[3]))
	}

	entry.Debug("query executed")
}

func (x *xormLogger) Warn(v ...interface{}) {
	logrus.Warn(fmt.Sprint(v...))
}

func (x *xormLogger) Warnf(format string, v ...interface{}) {
	logrus.Warnf(format, v...)
}

func (x *xormLogger) Level() log.LogLevel {
	return x.logLevel
}

func (x *xormLogger) SetLevel(l log.LogLevel) {
	x.logLevel = l
}

func (x *xormLogger) ShowSQL(show ...bool) {
	if len(show) == 0 {
		x.showSQL = true
		return
	}
	x.showSQL = show[0]
}

func (x *xormLogger) IsShowSQL() bool {
	return x.showSQL
}

func replaceUnicodeLiterals(input string) string {
	re := regexp.MustCompile(`\\u([0-9a-fA-F]{4})`)
	return re.ReplaceAllStringFunc(input, func(match string) string {
		// マッチされたグループからUnicodeポイントを抽出して整数に変換します。
		code, err := strconv.ParseUint(match[2:], 16, 32)
		if err != nil {
			// 変換に失敗した場合、マッチされた文字列をそのまま返します。
			return match
		}
		// 変換されたコードポイントに基づいてUnicode文字を生成して返します。
		return string(rune(code))
	})
}
