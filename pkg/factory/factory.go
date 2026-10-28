package factory

import (
	"context"
	"reflect"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hillfolk/goutils/echomiddleware"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/eventbus"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/sessionutils"
)

var ConextDBType echomiddleware.ContextDBType = "DB"
var ConextMyNumberDBType echomiddleware.ContextDBType = "MY_NUMBER_DB"

var logger = logrus.New()

type eventBusFactory struct{}

type EventBusFactory interface {
	StatutoryLedgerEventBus(ctx context.Context) (eventbus.IEventBus, error)
}

func NewEventBusFactory() EventBusFactory {
	return &eventBusFactory{}
}

func (e eventBusFactory) StatutoryLedgerEventBus(ctx context.Context) (eventbus.IEventBus, error) {
	v := ctx.Value(eventbus.ContextStatutoryLedgerEventBus)
	if v == nil {
		return nil, errors.New("err event bus is not exist")
	}

	switch v.(type) {
	case eventbus.IEventBus:
		return v.(eventbus.IEventBus), nil
	default:
		return nil, errors.New("err event bus interface conversion")
	}
}

type loggerFactory struct{}

type LoggerFactory interface {
	Entry(ctx context.Context) *logrus.Entry
}

func (l loggerFactory) Entry(ctx context.Context) *logrus.Entry {
	requestId := ctx.Value("request_id")
	return logrus.NewEntry(logger).WithField("request_id", requestId)
}

func NewLoggerFactory() LoggerFactory {
	logger.SetFormatter(&logrus.JSONFormatter{
		DisableHTMLEscape: true,
	})
	logger.SetReportCaller(true)

	if logrus.GetLevel() == logrus.DebugLevel {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	return &loggerFactory{}
}

type dbFactory struct{}

type DBFactory interface {
	EnergySharesDB(ctx context.Context) (xorm.Interface, error)
	MyNumberDB(ctx context.Context) (xorm.Interface, error)
}

func NewDBFactory() DBFactory {
	return &dbFactory{}
}

func (d dbFactory) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	v := ctx.Value(ConextDBType)

	if v == nil {
		return nil, errors.New("err db is not exist")
	}

	switch v.(type) {
	case *xorm.Session, *xorm.Engine:
		return v.(xorm.Interface), nil
	default:
		return nil, errors.New("err xorm interface conversion")
	}
}

func (d dbFactory) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	v := ctx.Value(ConextMyNumberDBType)

	if v == nil {
		return nil, errors.New("err db is not exist")
	}

	switch v.(type) {
	case *xorm.Session, *xorm.Engine:
		return v.(xorm.Interface), nil
	default:
		return nil, errors.New("err xorm interface conversion")
	}
}

func DB(ctx context.Context) xorm.Interface {
	v := ctx.Value(ConextDBType)
	if v == nil {
		panic("DB is not exist")
	}
	if db, ok := v.(*xorm.Session); ok {
		return db
	}
	if db, ok := v.(*xorm.Engine); ok {
		return db
	}
	panic("DB is not exist")
}

// Engine は context から EnergyShares DB の *xorm.Engine を取り出す。
// Session の場合はその親 Engine を返す。
func Engine(ctx context.Context) (*xorm.Engine, error) {
	switch db := ctx.Value(ConextDBType).(type) {
	case *xorm.Engine:
		return db, nil
	case *xorm.Session:
		return db.Engine(), nil
	default:
		return nil, errors.New("db engine is not available")
	}
}

// InTransaction はEnergyShares DBのコンテキストが開始済みトランザクションを保持しているかを返す。
func InTransaction(ctx context.Context) bool {
	session, ok := ctx.Value(ConextDBType).(*xorm.Session)
	return ok && session.IsInTx()
}

func MyNumberDB(ctx context.Context) xorm.Interface {
	v := ctx.Value(ConextMyNumberDBType)
	if v == nil {
		panic("MY_NUMBER_DB is not exist")
	}
	if db, ok := v.(*xorm.Session); ok {
		return db
	}
	if db, ok := v.(*xorm.Engine); ok {
		return db
	}
	panic("MY_NUMBER_DB is not exist")
}

func Logger(ctx context.Context) *logrus.Entry {
	v := ctx.Value(echomiddleware.ContextLoggerName)
	if v == nil {
		return logrus.WithFields(logrus.Fields{})
	}
	if logger, ok := v.(*logrus.Entry); ok {
		return logger
	}
	return logrus.WithFields(logrus.Fields{})
}

func S3Uploader(ctx context.Context) *s3manager.Uploader {
	v := ctx.Value(echomiddleware.ContextAWSName)
	if v == nil {
		panic("Amazon Session is not exist")
	}

	if sess, ok := v.(*session.Session); ok {
		uploader := s3manager.NewUploader(sess)

		return uploader
	}
	// s3 Manager

	return nil
}

func S3Downloader(ctx context.Context) *s3manager.Downloader {
	v := ctx.Value(echomiddleware.ContextAWSName)
	if v == nil {
		panic("Amazon Session is not exist")
	}

	if sess, ok := v.(*session.Session); ok {
		downloader := s3manager.NewDownloader(sess)

		return downloader
	}
	// s3 Manager

	return nil
}

func S3(ctx context.Context) *s3.S3 {
	v := ctx.Value(echomiddleware.ContextAWSName)
	if v == nil {
		panic("Amazon Session is not exist")
	}

	if sess, ok := v.(*session.Session); ok {
		svc := s3.New(sess)
		return svc
	}
	return nil
}

func MailSender(ctx context.Context) *ses.SES {
	v := ctx.Value(sessionutils.ContextAwsSesName)
	if v == nil {
		panic("Amazon Mail Session is not exist")
	}

	if sess, ok := v.(*session.Session); ok {

		return ses.New(sess)
	}

	return nil
}

type SnsFactory interface {
	SmsMsgSender(ctx context.Context) *sns.SNS
}

type snsFactory struct{}

func NewSnsFactory() SnsFactory {
	return &snsFactory{}
}

func (s snsFactory) SmsMsgSender(ctx context.Context) *sns.SNS {
	v := ctx.Value(sessionutils.ContextAwsSnsName)
	if v == nil {
		panic("Amazon SNS Session is not exist")
	}

	if sess, ok := v.(*session.Session); ok {
		return sns.New(sess)
	}

	return nil
}

// Deprecated: Use SnsFactory interface instead
func SmsMsgSender(ctx context.Context) *sns.SNS {
	v := ctx.Value(sessionutils.ContextAwsSnsName)
	if v == nil {
		panic("Amazon SNS Session is not exist")
	}

	if sess, ok := v.(*session.Session); ok {

		return sns.New(sess)
	}

	return nil
}

func ReflectJsonFields(val reflect.Value) []string {
	var cols []string
	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		tag := typeField.Tag

		str := i18nutils.DefaultTranslate(tag.Get("i18n"))
		cols = append(cols, str)
	}

	return cols
}

func DataWriteToCsv(val reflect.Value) []string {

	record := make([]string, 0)
	for i := 0; i < val.NumField(); i++ {

		valueField := val.Field(i)

		jsonTime, ok := valueField.Interface().(jsonutils.JSONTime)
		if ok {
			if jsonTime.Time().IsZero() {
				record = append(record, "")
			} else {
				record = append(record, jsonTime.Time().Format(time.DateTime))
			}
		} else {
			switch t := valueField.Interface().(type) {
			case string:
				transStr := i18nutils.DefaultTranslate(t)
				record = append(record, transStr)
			case int64:
				record = append(record, strconv.FormatInt(valueField.Interface().(int64), 10))
			default: // Unsupported type
				record = append(record, "ERROR#INVALID_TYPE")
			}
		}
	}

	return record
}

func EventBus() eventbus.IEventBus {
	v := eventbus.GetEventBus()
	if v == nil {
		panic("EventBus is not exist")
	}

	if eventBus, ok := v.(eventbus.IEventBus); ok {
		return eventBus
	}

	return nil
}

func HasElem(s interface{}, elem interface{}) bool {
	arrV := reflect.ValueOf(s)

	if arrV.Kind() == reflect.Slice {
		for i := 0; i < arrV.Len(); i++ {

			if arrV.Index(i).Interface() == elem {
				return true
			}
		}
	}

	return false
}
