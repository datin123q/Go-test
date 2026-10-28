package domain

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/segmentio/ksuid"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type EventCategory string

const (
	EVENT_INVESTMENT_REQUEST      EventCategory = "EVENT_INVESTMENT_REQUEST" // 투자 성립
	EVENT_INVESTMENT_WAIT_REQUEST EventCategory = "EVENT_INVESTMENT_WAIT_REQUEST"
	EVENT_LOTTERY_REQUEST         EventCategory = "EVENT_LOTTERY_REQUEST"
	EVENT_INVESTMENT_COMPLETED    EventCategory = "EVENT_INVESTMENT_COMPLETED" // 투자 완료 (출자)
	EVENT_INVESTMENT_VOIDED       EventCategory = "EVENT_INVESTMENT_VOIDED"    // 청약 철회 (입금 전 취소)
	EVENT_INVESTMENT_CANCELED     EventCategory = "EVENT_INVESTMENT_CANCELED"  // 해약 (입금 후 취소)
	EVENT_FUND_FAILURE            EventCategory = "EVENT_FUND_FAILURE"         // 불성립
	EVENT_TRANSACTION             EventCategory = "EVENT_TRANSACTION"          // 고객원장

	EVENT_DIVIDEND_DECISION      EventCategory = "EVENT_DIVIDEND_DECISION"      // 분배 실시
	EVENT_USER_DIVIDEND_DECISION EventCategory = "EVENT_USER_DIVIDEND_DECISION" // 고객별 분배 실시
	EVENT_FUND_TRANSFER          EventCategory = "EVENT_FUND_TRANSFER"          // 출자금 이체
)

const (
	PRIVATE_OFFERING Recruitment = "私募の取扱い"
	PUBLIC_OFFERING  Recruitment = "募集の取扱い"
	CANCELLATION     Recruitment = "解約"
)

type Recruitment string

type TransactionEventType string

const (
	INVESTMENT TransactionEventType = "出資" // 출자
	DEPOSIT    TransactionEventType = "入金" // 입금
	WITHDRAWAL TransactionEventType = "出金" // 출금
	DIVIDEND   TransactionEventType = "配当" // 분배
	REPAYMENT  TransactionEventType = "償還" // 상환
)

type InvestmentCancellation string

const (
	INVESTMENT_VOIDED InvestmentCancellation = "不成立(申込みの撤回)"
	FUND_FAILURE      InvestmentCancellation = "不成立(ファンド不成立)"
)

type ContractStatus string

const (
	INVESTMENT_COMPLETED ContractStatus = "買"
	INVESTMENT_CANCELED  ContractStatus = "解約"
)

type StatementType string

type EventMessage struct {
	Category      EventCategory `json:"category" xorm:"event_category"`
	Content       any           `json:"content" xorm:"-"`
	ContentString string        `json:"-" xorm:"content"`
	Uuid          string        `json:"uuid" xorm:"uuid pk"`
	EventTime     time.Time     `json:"event_time" xorm:"event_time"`
	IsPublish     bool          `json:"-" xorm:"is_publish"`
	Created       time.Time     `json:"-" xorm:"created"`
}

func NewEventMessage(category EventCategory, content any) EventMessage {
	ksuid, _ := ksuid.NewRandom()
	eventId := ksuid.String()

	return EventMessage{
		Category:  category,
		Content:   content,
		Uuid:      eventId,
		EventTime: time.Now(),
	}
}

func (e *EventMessage) Create(ctx context.Context) (int64, error) {
	const ErrorCategory = "EventMessage.Create"
	if e.Uuid == "" {
		return 0, errors.Wrap(errors.New("uuid is empty"), ErrorCategory)
	}
	if e.EventTime.IsZero() {
		return 0, errors.Wrap(errors.New("event time is zero"), ErrorCategory)
	}
	if e.Category == "" {
		return 0, errors.Wrap(errors.New("category is empty"), ErrorCategory)
	}
	if e.Content == nil {
		return 0, errors.Wrap(errors.New("content is nil"), ErrorCategory)
	}
	var err error
	e.ContentString, err = jsonutils.MarshalString(e.Content)
	if err != nil {
		return 0, errors.Wrap(err, "EventMessage.Create")
	}

	return factory.DB(ctx).Insert(e)
}

func (e *EventMessage) Update(ctx context.Context) (int64, error) {
	var err error
	e.ContentString, err = jsonutils.MarshalString(e.Content)
	if err != nil {
		return 0, errors.Wrap(err, "EventMessage.Update")
	}

	return factory.DB(ctx).ID(e.Uuid).UseBool("is_publish").Update(e)
}

// UpdatePublishStatus は is_publish 列だけを更新する(MIR-1114)。Update と違い
// ContentString(元のイベント内容)を再マーシャルして書き換えない。RePublish は
// ContentString を Unmarshal した DTO へ変換してから送信するため、Update を
// そのまま使うと content 列が DTO 経由の再マーシャル結果で上書きされ、DTO に
// 無いフィールドが消えてしまう(送信の成否にかかわらず発生する)。
func (e EventMessage) UpdatePublishStatus(ctx context.Context, isPublish bool) (int64, error) {
	return factory.DB(ctx).ID(e.Uuid).Cols("is_publish").Update(&EventMessage{IsPublish: isPublish})
}

// eventIdentifierFieldNames は、content から識別子として拾ってよいフィールド名
// (Go のフィールド名) と、それを出すログキーの対応表。
//
// イベント種別ごとに「どの識別子を出すか」を手作業のリストで持つと、
// 種別が増えたときに更新漏れが起きる(MIR-1094 でも似た形の手作業リストが
// 2箇所に分かれて乖離したことがある)。DTO 側の識別子フィールドは共通の
// 名前しか使っていないため、リフレクションで機械的に拾えば種別ごとの
// 追加対応を増やさずに済む。
//
// 適用範囲の限界: ここに載っている名前と完全一致するフィールドしか拾わない。
// SendCustomerLedgerWithdrawalEvent の "account_id"(実体は TransactionDto.UserId)
// のように、意図的に別名でログへ出しているものは対象外(該当関数側で個別に
// コメントしている)。対応する型も int 系・string のみで、将来 *int64 のような
// ポインタ型の識別子フィールドが増えた場合はこの switch の拡張が要る。
var eventIdentifierFieldNames = []struct {
	field  string
	logKey string
}{
	{"InvestmentId", "investment_id"},
	{"InvestmentWaiterId", "investment_waiter_id"},
	{"ScheduleId", "schedule_id"},
	{"UserId", "user_id"},
	{"ProductId", "product_id"},
}

// EventIdentifierFields は content (イベントの dto) から、ログに残してよい
// 業務上の識別子だけを reflect で拾う。
//
// content には氏名・取引金額・残高が入るため丸ごとログへ渡してはいけないが、
// 識別子まで削ると障害調査で対象を特定できなくなる。RePublish はポインタ
// (GetMessageType が返す型)、PublishEvent の直接呼び出しは値を渡すため、
// 両方を受けられるようポインタは Elem() で剥がす。
//
// content が匿名フィールドとして nil のポインタを埋め込んでいる場合、
// reflect.Value.FieldByName はその埋め込み先を辿ろうとして panic する。
// 現状の DTO はすべてフラットで該当しないが、エラーログ経路(送信失敗時)
// から呼ばれる関数がここで panic すると影響が大きいため、そこまでに
// 拾えた識別子で打ち切って安全側に倒す。
func EventIdentifierFields(content any) (fields logrus.Fields) {
	fields = logrus.Fields{}
	if content == nil {
		return fields
	}
	v := reflect.ValueOf(content)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return fields
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return fields
	}
	defer func() {
		recover()
	}()
	for _, id := range eventIdentifierFieldNames {
		f := v.FieldByName(id.field)
		if !f.IsValid() {
			continue
		}
		switch f.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if f.Int() != 0 {
				fields[id.logKey] = f.Int()
			}
		case reflect.String:
			if f.String() != "" {
				fields[id.logKey] = f.String()
			}
		}
	}
	return fields
}

// LogFields はログに残してよい識別子だけを返す。
//
// e には content（氏名・取引金額・残高など）が入るため、e を丸ごとログへ
// 渡さないこと。呼び出しごとに新しい map を返すので、呼び出し元で
// "error" や "time" を追加で詰めても他の呼び出しに影響しない。
//
// domain 層・service 層(service/event_message)の両方からこの1箇所を
// 呼ぶことで、同じ機能の実装が2つに並存して片方だけ直すような乖離を防ぐ
// (2026-08-19、一括処理でイベントログが1件566〜870バイト・1,000件超と
// なりログ書き込みが詰まってヘルスチェックを巻き込んだ障害の再発防止)。
//
// investment_id のような一意な識別子を持たない dto(TransferFundDto 等)は
// product_id・user_id だけになり、一括処理内の1件を単独では特定できない
// ことがある。その場合は同じループで直前に出る別イベントとの時系列対応で
// 追跡する前提(既知の限界)。
func (e EventMessage) LogFields() logrus.Fields {
	fields := EventIdentifierFields(e.Content)
	fields["event_uuid"] = e.Uuid
	fields["event_category"] = e.Category
	fields["event_time"] = e.EventTime
	return fields
}

// publishToEventBus はテストでKafka送信の成否を差し替えるためのパッケージレベル変数。
var publishToEventBus = func(data []byte) error {
	return factory.EventBus().Publish(data)
}

// rePublishBatchSize は RePublish の1回の実行で処理する上限件数(MIR-1114)。
// 未送信行が大量に溜まった状態で無制限に読み込むと、1回の実行時間が長引き、
// cronの多重起動防止(skipIfStillRunning)の窓を広げてしまうため上限を設ける。
const rePublishBatchSize = 500

// rePublishMinAge は RePublish が拾う行の下限経過時間(MIR-1202のOutbox化に伴う
// 追加)。MIR-1202で出資実行のイベント発行がOutboxパターンになり、
// event_messageの行はトランザクションのコミット時点でis_publish=falseのまま
// 確定し、実際の送信(Kafka Publish)はコミット後(AfterCommit)に行われる。
// つまりコミットから送信完了までの間、「コミット済みだが送信中」という行が
// 一時的に存在する。この間にRePublishが同じ行を拾うと、AfterCommit側の送信と
// 合わせて二重送信になる。この窓を避けるため、直近に作られた行は対象外とする。
// 2026-09-08の本番実測で631件の出資実行(分割3回)が3分9秒で完走しており、
// Outbox化後にこの窓へ入るのはコミット後のKafka送信部分だけなのでこれより
// 短くなる。10分は十分な余裕を見た値だが、見直しが必要になるのは
// バッチサイズを変えたときではなく、出資実行の単一トランザクションが長引く
// ときである点に注意(createdはコミット時刻ではなくINSERT時刻なので、
// トランザクションが10分を超えて開いたままだと、コミットされた瞬間には
// 既に猶予がゼロになっている)。この時間条件は「送信中かもしれない行を
// 拾わない」ことしか保証せず、送信(またはトランザクション)が10分を
// 超えて長引いた場合の重複は防げない。
const rePublishMinAge = 10 * time.Minute

// RePublish は is_publish=false の行を拾って再送する(未送信の取りこぼしを埋める、MIR-1114)。
// 1件の失敗が残りの処理を止めないよう、Unmarshal・Publish・Update のどこで失敗しても
// ログに残したうえで次の行へ進む。以前はUpdate呼び出しがコメントアウトされており、
// 送信に成功してもis_publishがfalseのまま残るため、実行するたびに同じイベントを
// 際限なく再送していた。
func (EventMessage) RePublish(ctx context.Context) error {
	events := make([]EventMessage, 0)
	// .UTC()はMySQL(本番)では実質no-op。go-sql-driver/mysqlはtime.Time値を
	// bind直前に必ずDSNのloc=(config.local_sample.ymlはloc=Local)で
	// v.In(cfg.Loc)により変換し直すため、渡す前に.UTC()していても上書きされる。
	// 本番でcreated列(xormのDatabaseTZ、MySQLではtime.Local)との比較が
	// 一貫するのは、書き込み側・比較側のどちらも最終的に同じLocal基準に
	// 揃うためで、コンテナの実タイムゾーン設定には依存しない。
	// 一方sqlite(テスト環境)ではmattn/go-sqlite3がtime.Time値をそのまま
	// (呼び出し元のロケーションのまま)フォーマットしてしまい、xormが
	// createdへ書き込む値(DatabaseTZ、sqliteではtime.UTC固定)と噛み合わない。
	// このズレは開発機がJSTだと実際に再現する。.UTC()はこのsqlite側の
	// 不整合を埋めるために必要。
	if err := factory.DB(ctx).
		Where("is_publish = ? AND created < ?", false, time.Now().Add(-rePublishMinAge).UTC()).
		OrderBy("created ASC").
		Limit(rePublishBatchSize).
		Find(&events); err != nil {
		return errors.Wrap(err, "EventMessage.RePublish")
	}
	// 手動エンドポイント(POST /events/republish)は障害時に叩かれる想定だが、
	// rePublishMinAgeにより直近の行は対象から除外されるため、押しても0件で
	// 何も起きないことがある。件数を出しておけば、それが条件で除外されて
	// いるだけなのか判断できる。
	factory.Logger(ctx).WithField("count", len(events)).Info("EventMessage.RePublish target count")
	for _, e := range events {
		v := GetMessageType(e.Category)
		if err := jsonutils.Unmarshal([]byte(e.ContentString), v); err != nil {
			fields := e.LogFields()
			fields["error"] = err
			fields["time"] = time.Now()
			factory.Logger(ctx).WithFields(fields).Error("fail unmarshal EventMessage.ContentString")
			continue
		}

		// json.Marshal用にContentだけ差し替えたコピーを送信に使い、eのContentStringは
		// 元のまま保つ。UpdatePublishStatusはis_publishしか更新しないため実害は無いが、
		// eにContentを持たせたままにしておくと将来Updateを誤って呼んだときにcontent列を
		// 破壊しかねない。
		toSend := e
		toSend.Content = v

		data, err := json.Marshal(toSend)
		if err != nil {
			fields := toSend.LogFields()
			fields["error"] = err
			fields["time"] = time.Now()
			factory.Logger(ctx).WithFields(fields).Error("fail marshal EventMessage")
			continue
		}

		var isPublish bool
		if err := publishToEventBus(data); err != nil {
			fields := toSend.LogFields()
			fields["error"] = err
			fields["time"] = time.Now()
			factory.Logger(ctx).WithFields(fields).Error("fail published EventMessage.RePublish")
			isPublish = false
		} else {
			fields := toSend.LogFields()
			fields["time"] = time.Now()
			factory.Logger(ctx).WithFields(fields).Info("success published EventMessage.RePublish")
			isPublish = true
		}

		if _, err := e.UpdatePublishStatus(ctx, isPublish); err != nil {
			fields := toSend.LogFields()
			fields["error"] = err
			fields["time"] = time.Now()
			factory.Logger(ctx).WithFields(fields).Error("fail update EventMessage.RePublish")
			continue
		}
	}
	return nil
}

// PublishEvent is a function to send event message to event bus
func (EventMessage) PublishEvent(ctx context.Context, category EventCategory, content any) error {
	e := NewEventMessage(category, content)
	if _, err := e.Create(ctx); err != nil {
		return errors.Wrap(err, "EventMessage.PublishEvent")
	}
	data, err := json.Marshal(e)
	if err != nil {
		return errors.Wrap(err, "EventMessage.PublishEvent")
	}
	err = factory.EventBus().Publish(data)
	if err != nil {
		fields := e.LogFields()
		fields["error"] = err
		fields["time"] = time.Now()
		factory.Logger(ctx).WithFields(fields).Error("fail published EventMessage.PublishEvent")
		e.IsPublish = false
	} else {
		fields := e.LogFields()
		fields["time"] = time.Now()
		factory.Logger(ctx).WithFields(fields).Info("success published EventMessage.PublishEvent")
		e.IsPublish = true
	}

	_, err = e.Update(ctx)
	if err != nil {
		fields := e.LogFields()
		fields["error"] = err
		fields["time"] = time.Now()
		factory.Logger(ctx).WithFields(fields).Error("fail update EventMessage.PublishEvent")
		return errors.Wrap(err, "fail update EventMessage.PublishEvent")
	}
	return nil
}

type InvestmentRequestDto struct {
	UserId         int64       `json:"user_id,omitempty" `
	UserName       string      `json:"user_name,omitempty"`
	InvestmentId   string      `json:"investment_id,omitempty"` // 투자ID
	ProductName    string      `json:"product_name,omitempty"`
	ProductId      int64       `json:"product_id,omitempty"`
	Recruitment    Recruitment `json:"recruitment,omitempty"`
	OrderDateTime  time.Time   `json:"order_date_time,omitempty"` // 신청일자
	OrderQuantity  int64       `json:"order_quantity,omitempty"`  // 신청수량
	OrderUnitPrice int64       `json:"order_unit_price,omitempty"`
	OrderAmount    int64       `json:"order_amount,omitempty"`
}

type InvestmentWaitRequestDto struct {
	UserId             int64       `json:"user_id,omitempty"`
	UserName           string      `json:"user_name,omitempty"`
	InvestmentWaiterId int64       `json:"investment_waiter_id,omitempty"`
	ProductName        string      `json:"product_name,omitempty"`
	ProductId          int64       `json:"product_id,omitempty"`
	Recruitment        Recruitment `json:"recruitment"`                // 私募/募集/売り出しの取り扱い
	OrderDateTime      time.Time   `json:"order_date_time,omitempty"`  // 受注日時
	WaitingTurn        int64       `json:"waiting_turn,omitempty"`     // 待機順番
	OrderQuantity      int64       `json:"order_quantity,omitempty"`   // 受注数量
	OrderUnitPrice     int64       `json:"order_unit_price,omitempty"` // 受注単価
	OrderAmount        int64       `json:"order_amount,omitempty"`     // 受注金額
}

type LotteryRequestDto struct {
	UserId         int64       `json:"user_id,omitempty"`
	UserName       string      `json:"user_name,omitempty"`
	ProductName    string      `json:"product_name,omitempty"`
	ProductId      int64       `json:"product_id,omitempty"`
	Recruitment    Recruitment `json:"recruitment,omitempty"`
	OrderDateTime  time.Time   `json:"order_date_time,omitempty"`
	OrderQuantity  int64       `json:"order_quantity,omitempty"`
	OrderUnitPrice int64       `json:"order_unit_price,omitempty"`
	OrderAmount    int64       `json:"order_amount,omitempty"`
}

type InvestmentCompletedDto struct {
	Recruitment     Recruitment `json:"recruitment"`
	UserId          int64       `json:"user_id,omitempty"`
	CompanyName     string      `json:"company_name,omitempty"` // 합동회사 명칭
	ProductName     string      `json:"product_name,omitempty"`
	ProductId       int64       `json:"product_id,omitempty"`
	InvestmentId    string      `json:"investment_id,omitempty"`    // 투자ID
	CommitDateTime  time.Time   `json:"commit_date_time,omitempty"` // 약정일자
	CommitQuantity  int64       `json:"commit_quantity,omitempty"`  // 약정수량
	CommitUnitPrice int64       `json:"commit_unit_price,omitempty"`
	CommitAmount    int64       `json:"commit_amount,omitempty"`
	UserName        string      `json:"user_name,omitempty"` // 투자자 명칭
}

type InvestmentVoidDto struct {
	RequestDatetime time.Time `json:"request_datetime"`
	InvestmentId    string    `json:"investment_id"`
}

type FundFailureDto struct {
	ProductId        int64              `json:"product_id"`
	CancellationDate jsonutils.JSONTime `json:"cancellation_date"`
}

type InvestmentCancelDto struct {
	RequestDatetime time.Time `json:"request_datetime"`
	InvestmentId    string    `json:"investment_id"`
}

type TransactionDto struct {
	EventDateTime   *time.Time           `json:"event_date_time,omitempty"`  // <受渡年月日> PrimarySortKey ApplyDateTimeがある場合はApplyDateTimeが優先される
	CommitDateTime  *time.Time           `json:"commit_date_time,omitempty"` // <約定年月日>
	ApplyDateTime   *time.Time           `json:"apply_date_time,omitempty"`  // <受渡年月日> EventDateTimeよりも優先される
	TransactionType TransactionEventType `json:"transaction_type"`
	UserId          int64                `json:"user_id,omitempty"`       // 투자자 ID
	UserName        string               `json:"user_name,omitempty"`     // 투자자 명칭
	InvestmentId    string               `json:"investment_id,omitempty"` // share.KSUID SecondarySortKey
	ProductName     string               `json:"product_name,omitempty"`
	SpvId           int64                `json:"spv_id,omitempty"`
	ProductId       int64                `json:"product_id,omitempty"`
	Quantity        int64                `json:"quantity,omitempty"`
	UnitPrice       int64                `json:"unit_price,omitempty"`
	Amount          int64                `json:"amount"`          // 투자/상환일 겨우만
	Asset           int64                `json:"asset,omitempty"` // 고객의 자산으로 돌아가는 금액  자산( Asset ) : 투자금액
	Debt            int64                `json:"debt,omitempty"`  // 입금시 부채 발생 (출금시 부채 감소) 입금 금액
	Balance         int64                `json:"balance"`         // 현재 잔액 (입금시 증가, 출금시 감소) 예치금 잔액
}

// TODO: 레거시 제거 예정
type DividendDecisionDtoLegacy struct {
	SpvId                 int64   `json:"spv_id"`
	Year                  int     `json:"year"`
	Month                 int     `json:"month"`
	ProductId             int64   `json:"product_id"`
	InvestorCount         int64   `json:"investor_count"`
	InvestmentTotalAmount int64   `json:"investment_total_amount"`
	DividendRatio         float64 `json:"dividend_ratio"`
	WithholdingTaxAmount  float64 `json:"withholding_tax_amount"`
}

// EVENT_DIVIDEND_DECISION
type DividendDecisionDto struct {
	ScheduleId         int64     `json:"schedule_id"`          // 스케쥴 아이디 - 이슈 발생시 검증용도
	SpvId              int64     `json:"spv_id"`               // 개발사 아이디
	DecideDate         time.Time `json:"decide_date"`          // 결정 날짜
	ExecuteDate        time.Time `json:"execute_date"`         // 실행 날짜
	ProductId          int64     `json:"product_id"`           // 프로덕트 아이디, 법정장부 서버에서는 이 용어를 써서 그대로 둔다
	InvestCount        int64     `json:"invest_count"`         // (주의) 투자자 수가 아니라 투자횟수. 한 투자자가 중복투자할 수 있다.
	ProjectShareAmount int64     `json:"project_share_amount"` // 총 투자금액
	DividendAmount     int64     `json:"dividend_amount"`      // 영업 수수료를 뺀 순 배당금액임에 주의
	ReturnRate         float64   `json:"return_rate"`          // 수익률
	WithholdingTax     int64     `json:"withholding_tax"`      // 원천징수액
}

// EVENT_USER_DIVIDEND_DECISION
type UserDividendDecisionDto struct {
	ScheduleId             int64     `json:"schedule_id"` // 기록 추적용도
	SpvId                  int64     `json:"spv_id"`
	DecideDate             time.Time `json:"decide_date"`
	ExecuteDate            time.Time `json:"execute_date"`
	ProductId              int64     `json:"product_id"`
	UserId                 int64     `json:"user_id"`
	UserProjectShareAmount int64     `json:"user_project_share_amount"` // 투자자의 투자금액
	UserDividendAmount     int64     `json:"user_dividend_amount"`      // 투자자의 세전 배당금액
	ReturnRate             float64   `json:"return_rate"`
	WithholdingTax         int64     `json:"withholding_tax"`
}

type TransferFundDto struct {
	ProductId      int64     `json:"product_id"`
	ProductName    string    `json:"product_name"`
	Count          int64     `json:"count"`
	UnitPrice      int64     `json:"unit_price"`
	Amount         int64     `json:"amount"`
	UserId         int64     `json:"user_id"`
	CommitDateTime time.Time `json:"commit_date_time"`
}

// CreateCustomerLedgerDepositEvent 入金時に発火
func CreateCustomerLedgerDepositEvent(ctx context.Context, userId int64, userName string, amount int64, transferDate time.Time) error {
	balance, err := CalculateCustomerLedgerBalance(ctx, userId)
	if err != nil {
		return err
	}
	eventType := DEPOSIT
	dto := TransactionDto{
		EventDateTime:   &transferDate,
		CommitDateTime:  &transferDate,
		UserId:          userId,
		UserName:        userName,
		TransactionType: eventType,
		Amount:          amount,
		Asset:           amount,
		Balance:         balance,
	}
	// dto には氏名・入金額・残高が入るため、識別できる最小限だけを出す
	factory.Logger(ctx).WithFields(logrus.Fields{
		"user_id":          userId,
		"transaction_type": eventType,
		"time":             time.Now(),
	}).Info("CreateCustomerLedgerDepositEvent")
	return EventMessage{}.PublishEvent(ctx, EVENT_TRANSACTION, dto)
}

// 투자 신청 이벤트 발생
func CreateRequestInvestmentEvent(ctx context.Context, owner *User, s *Share, p *Project) error {
	dto := InvestmentRequestDto{
		UserId:         owner.Id,
		UserName:       owner.Name,
		InvestmentId:   s.Id,
		ProductName:    p.Name,
		ProductId:      p.Id,
		Recruitment:    p.OfferingType.ToRecruitment(),
		OrderDateTime:  s.InvestDate.Time(),
		OrderQuantity:  s.Qty,
		OrderUnitPrice: s.Price,
		OrderAmount:    s.Amount,
	}

	// dto には氏名・取引金額・残高が入るため出さない。product_id だけでは
	// 一括処理時に個々の申込を特定できないため、identifier も合わせて出す。
	fields := EventIdentifierFields(dto)
	fields["time"] = time.Now()
	factory.Logger(ctx).WithFields(fields).Info("CreateRequestInvestmentEvent")
	return EventMessage{}.PublishEvent(ctx, EVENT_INVESTMENT_REQUEST, dto)
}

func CreateInvestmentWaitRequestEvent(ctx context.Context, user *User, investmentWait *InvestmentWaiters, project *Project) error {
	dto := InvestmentWaitRequestDto{
		UserId:             user.Id,
		UserName:           user.Name,
		InvestmentWaiterId: investmentWait.Id,
		ProductName:        project.Name,
		ProductId:          project.Id,
		Recruitment:        project.OfferingType.ToRecruitment(),
		OrderDateTime:      investmentWait.WaitingDate.Time(),
		WaitingTurn:        investmentWait.WaitingTurn,
		OrderQuantity:      investmentWait.Qty,
		OrderUnitPrice:     project.SharePrice,
		OrderAmount:        investmentWait.Amount,
	}

	// dto には氏名・取引金額・残高が入るため出さない。product_id だけでは
	// 一括処理時に個々の待機申込を特定できないため、identifier も合わせて出す。
	fields := EventIdentifierFields(dto)
	fields["time"] = time.Now()
	factory.Logger(ctx).WithFields(fields).Info("CreateInvestmentWaitRequestEvent")
	return EventMessage{}.PublishEvent(ctx, EVENT_INVESTMENT_WAIT_REQUEST, dto)
}

func CreateLotteryRequestEvent(ctx context.Context, user *User, project *Project, application *LotteryApplication) error {
	dto := LotteryRequestDto{
		UserId:         user.Id,
		UserName:       user.Name,
		ProductName:    project.Name,
		ProductId:      project.Id,
		Recruitment:    project.OfferingType.ToRecruitment(),
		OrderDateTime:  application.AppliedAt.Time(),
		OrderQuantity:  application.Qty,
		OrderUnitPrice: application.UnitPrice,
		OrderAmount:    application.Amount,
	}

	// dto には氏名・取引金額・残高が入るため出さない
	fields := EventIdentifierFields(dto)
	fields["time"] = time.Now()
	factory.Logger(ctx).WithFields(fields).Info("CreateLotteryRequestEvent")
	return EventMessage{}.PublishEvent(ctx, EVENT_LOTTERY_REQUEST, dto)
}

func CreateCompletedInvestmentEvent(ctx context.Context, s *Share, owner *User, p *Project, dr *DepositRequest) error {
	dto := InvestmentCompletedDto{
		UserId:          s.OwnerId,
		UserName:        owner.Name,
		InvestmentId:    s.Id,
		CompanyName:     p.Spv.Name,
		ProductName:     p.Name,
		ProductId:       p.Id,
		Recruitment:     p.OfferingType.ToRecruitment(),
		CommitDateTime:  dr.DepositDate.Time(),
		CommitQuantity:  s.Qty,
		CommitUnitPrice: s.Price,
		CommitAmount:    s.Amount,
	}
	// dto には氏名・取引金額・残高が入るため出さない。product_id だけでは
	// 一括の出資成立処理で個々の投資を特定できないため、identifier も合わせて出す。
	fields := EventIdentifierFields(dto)
	fields["time"] = time.Now()
	factory.Logger(ctx).WithFields(fields).Info("CreateCompletedInvestmentEvent")
	return EventMessage{}.PublishEvent(ctx, EVENT_INVESTMENT_COMPLETED, dto)
}

// 투자를 취소하거나 투자 불성립이 발생시 호출된다.
func CreateInvestmentEventsByDepositStatus(ctx context.Context, requestDatetime time.Time, dr *DepositRequest) error {
	if dr.DepositCheckStatus == NO_DEPOSIT { // 투자금을 모두 넣지 않았을 때
		if err := CreateInvestmentVoidEvent(ctx, requestDatetime, dr.ShareId); err != nil {
			return errors.Wrap(err, "fail to send investment void event")
		}
	} else if dr.DepositCheckStatus == OK_DEPOSIT { // 투자금을 모두 넣었을 때
		if err := CreateInvestmentCancelEvent(ctx, requestDatetime, dr.ShareId); err != nil {
			return errors.Wrap(err, "fail to send investment cancel event")
		}
	}
	return nil
}

// 해약 (입금 후 취소) 이벤트 발생
func CreateInvestmentCancelEvent(ctx context.Context, requestDatetime time.Time, investmentId string) error {
	dto := InvestmentCancelDto{
		RequestDatetime: requestDatetime,
		InvestmentId:    investmentId,
	}
	factory.Logger(ctx).WithFields(logrus.Fields{
		// dto には氏名・取引金額・残高が入るため出さない
		"investment_id": investmentId,
		"time":          time.Now(),
	}).Info("CreateInvestmentCancelEvent")
	return EventMessage{}.PublishEvent(ctx, EVENT_INVESTMENT_CANCELED, dto)
}

// 청약 철회 (입금 전 취소)
func CreateInvestmentVoidEvent(ctx context.Context, requestDatetime time.Time, investmentId string) error {
	dto := InvestmentVoidDto{
		RequestDatetime: requestDatetime,
		InvestmentId:    investmentId,
	}

	factory.Logger(ctx).WithFields(logrus.Fields{
		// dto には氏名・取引金額・残高が入るため出さない
		"investment_id": investmentId,
		"time":          time.Now(),
	}).Info("CreateInvestmentVoidEvent")
	return EventMessage{}.PublishEvent(ctx, EVENT_INVESTMENT_VOIDED, dto)
}

// SendCustomerLedgerWithdrawalEvent 出金時に発火
func SendCustomerLedgerWithdrawalEvent(ctx context.Context, accountId int64, accountName string, transferAmount int64, transferTime time.Time) error {
	balance, err := CalculateCustomerLedgerBalance(ctx, accountId)
	if err != nil {
		return err
	}
	eventTime := transferTime
	dto := TransactionDto{
		EventDateTime:   &eventTime,
		CommitDateTime:  &eventTime,
		UserId:          accountId,
		UserName:        accountName,
		TransactionType: WITHDRAWAL,
		Debt:            transferAmount,
		Amount:          transferAmount,
		Balance:         balance,
	}

	factory.Logger(ctx).WithFields(logrus.Fields{
		// dto には氏名・取引金額・残高が入るため出さない。
		// dto.UserId には accountId を入れているが、この関数のログキーは
		// 既存の "account_id" のまま残す(EventIdentifierFields を使うと
		// "user_id" に変わり、この関数だけキー名が変わってしまうため)。
		// PublishEvent 内部のログは EventIdentifierFields 経由になるので
		// そちらは "user_id" として出る。同一 dto でも経路によってキー名が
		// 異なる点は既知の非対称。
		"account_id": accountId,
		"time":       time.Now(),
	}).Info("SendCustomerLedgerWithdrawalEvent")
	EventMessage{}.PublishEvent(ctx, EVENT_TRANSACTION, dto)
	return nil
}

// CreateCustomerLedgerDividendEvent 配当・償還の入金時に発火
func CreateCustomerLedgerDividendEvent(ctx context.Context, d *UserDividend, projectName string, account *UserAccount) error {
	balance, err := CalculateCustomerLedgerBalance(ctx, d.UserId)
	if err != nil {
		return err
	}
	var eventType TransactionEventType
	if d.Dividend.TransactionType == DEPOSIT_DIVIDEND {
		eventType = DIVIDEND
	} else if d.Dividend.TransactionType == DEPOSIT_REPAYMENT {
		eventType = REPAYMENT
	} else {
		err := errors.Wrap(errors.New("invalid transaction type"), "createCustomerLedgerEvent")
		if err != nil {
			return err
		}
	}
	eventDate := d.TransactionDate.Time()
	dto := TransactionDto{
		EventDateTime:   &eventDate,
		CommitDateTime:  &eventDate,
		TransactionType: eventType,
		UserId:          d.UserId,
		UserName:        account.UserName,
		InvestmentId:    d.TransactionId,
		ProductName:     projectName,
		SpvId:           d.SpvId,
		ProductId:       d.ProjectId,
		Quantity:        d.Qty,
		// 1口あたりの単価を設定
		// 整数除算により端数は切り捨て
		UnitPrice: d.DepositAmount / d.Qty,
		Amount:    d.DepositAmount, // 源泉徴収後の実入金額
		Asset:     d.DepositAmount,
		Balance:   balance,
	}

	// dto には氏名・取引金額・残高が入るため出さない。user_id だけでは
	// 一括の配当・償還処理で個々の投資/商品を特定できないため、
	// investment_id・product_id も合わせて出す。
	fields := EventIdentifierFields(dto)
	fields["time"] = time.Now()
	factory.Logger(ctx).WithFields(fields).Info("CreateCustomerLedgerDividendEvent")
	return EventMessage{}.PublishEvent(ctx, EVENT_TRANSACTION, dto)
}

func CreateFundFailureEvent(ctx context.Context, p *Project) error {
	// 펀드 실패 이벤트 발생
	dto := FundFailureDto{
		ProductId:        p.Id,
		CancellationDate: jsonutils.JSONTime(time.Now()),
	}
	factory.Logger(ctx).WithFields(logrus.Fields{
		// dto には氏名・取引金額・残高が入るため出さない
		"product_id": p.Id,
		"time":       time.Now(),
	}).Info("CreateFundFailureEvent")
	return EventMessage{}.PublishEvent(ctx, EVENT_FUND_FAILURE, dto)
}

// 원래는 배당 확정이나, 확정단계가 없어지면서 실행시 이벤트 발송
func CreateDividendExecutedEvent(ctx context.Context, schedule RepaymentDividendSchedule) error {
	// 배당 확정 이벤트 발생
	dto := DividendDecisionDto{
		ScheduleId:         schedule.Id,
		SpvId:              schedule.ProjectSpvId,
		DecideDate:         schedule.DecideDate.Time(),  // 결정일
		ExecuteDate:        schedule.ExecuteDate.Time(), // 실시일
		ProductId:          schedule.ProjectId,
		InvestCount:        schedule.NumInvestors,
		ProjectShareAmount: schedule.ProjectShareAmount,
		DividendAmount:     schedule.NetDividendAfterBrokerFee,
		ReturnRate:         schedule.ReturnRate,
		WithholdingTax:     schedule.WithholdingTax,
	}

	factory.Logger(ctx).WithFields(logrus.Fields{
		// DividendDecisionDto はファンド単位の集計で個人を特定しないが、
		// 他のイベントと粒度を揃えて識別子だけ出す
		"schedule_id": schedule.Id,
		"product_id":  schedule.ProjectId,
		"time":        time.Now(),
	}).Info("CreateDividendExecutedEvent")
	return EventMessage{}.PublishEvent(ctx, EVENT_DIVIDEND_DECISION, dto)
}

func CreateUserDividendExecutedEvent(ctx context.Context, ud UserDividend, returnRate float64) error {
	// 고객의 배당 실시 이벤트 발생
	dto := UserDividendDecisionDto{
		SpvId:                  ud.SpvId,
		DecideDate:             ud.ConfirmedDate.Time(), // ConfirmDate 는 DecideDate 이다. 결정일
		ExecuteDate:            ud.DividendDate.Time(),  // DividendDate 는 ExecuteDate 이다. 실시일
		ProductId:              ud.ProjectId,
		UserId:                 ud.UserId,
		UserProjectShareAmount: ud.InvestmentAmount, // 투자금액
		UserDividendAmount:     ud.Amount,           // 세전 배당금
		ReturnRate:             returnRate,          // 기존 Devidend 항목에 없기에 우선 별도 파라미터로 전달
		WithholdingTax:         ud.Tax,
	}

	// dto には氏名・取引金額・残高が入るため出さない。user_id だけでは
	// 一括の分配実施処理で個々の商品を特定できないため、product_id も合わせて出す。
	fields := EventIdentifierFields(dto)
	fields["time"] = time.Now()
	factory.Logger(ctx).WithFields(fields).Info("CreateUserDividendExecutedEvent")
	return EventMessage{}.PublishEvent(ctx, EVENT_USER_DIVIDEND_DECISION, dto)
}

func GetMessageType(category EventCategory) any {
	switch category {
	case EVENT_INVESTMENT_REQUEST:
		return &InvestmentRequestDto{}
	case EVENT_INVESTMENT_WAIT_REQUEST:
		return &InvestmentWaitRequestDto{}
	case EVENT_LOTTERY_REQUEST:
		return &LotteryRequestDto{}
	case EVENT_INVESTMENT_COMPLETED:
		return &InvestmentCompletedDto{}
	case EVENT_FUND_FAILURE:
		return &FundFailureDto{}
	case EVENT_INVESTMENT_CANCELED:
		return &InvestmentCancelDto{}
	case EVENT_INVESTMENT_VOIDED:
		return &InvestmentVoidDto{}
	case EVENT_TRANSACTION:
		return &TransactionDto{}
	case EVENT_DIVIDEND_DECISION:
		return &DividendDecisionDto{}
	case EVENT_USER_DIVIDEND_DECISION:
		return &UserDividendDecisionDto{}
	case EVENT_FUND_TRANSFER:
		return &TransferFundDto{}
	default:
		return ""
	}
}

// 1. 투자자가 예치금 입금시 회사 기준의 부채 발생 예치금 잔액은 100% 부채
// 2. 투자자가 투자시 회사 기준의 자산 발생 예치금 잔액은 100% 자산
// 3. 투자 완료시 부채 감소
// 3. SPC로 자금이 입금시 회사 기준의 부채 감소 정확한 시점은 출자 시점
// 상환 및 배당은 고객에게 입금 되는 시점에 부채로
// 투자 성립으로 변할때 약정일을 업데이트 하자
// 펀드 출자로 표시 하면됨?? /
// 펀드 양도일도 출자 일

// CalculateCustomerLedgerBalance 顧客勘定元帳イベント向けの「表示用残高」を算出する。
func CalculateCustomerLedgerBalance(ctx context.Context, userID int64) (int64, error) {
	// NOTE: 帳票表示では投資申請中・出金申請中の金額も含めた見かけ残高が必要なため、
	// user_asset_view を利用し user_account.balance + withdrawal_request_amount(INCOMPLETE/WAIT/FAIL) + invest_request_amount(REQUEST_INVEST/SUCCESS_INVEST) を都度計算して返す。
	assetView, err := UserAssetView{}.GetByUserId(ctx, userID)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get user asset view")
	}
	return assetView.Deposit + assetView.WithdrawalRequestAmount + assetView.InvestRequestAmount + assetView.InvestSuccessAmount, nil
}

// 약정인른 투자 완료시점

// 배당시점은 투자 구수 단가 필요 없음
// 배당금은 원천세 제외한 금액이 입금됨

// 배당금이 ESKK 로 입금시 ESKK 부채가 늘어남
// 상환금이 입금시 ESKK 부채가 늘어남
// 배당금/출자금 입금시 ESKK의 부채가 늘어남 근대 그 시점을 알수 없음
// 생성 요청시 전달하는 방식은 안되나?
