package domain

import (
	"context"

	"github.com/gofrs/uuid"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
)

type QNA_STATUS_TYPE string

const (
	QNA_STATUS_TYPE_CODE                   = "QNA_STATUS_TYPE"
	QNA_RECEIVED           QNA_STATUS_TYPE = "RECEIVED"           // 문의접수완료
	QNA_RESPONSE_PENDING   QNA_STATUS_TYPE = "RESPONSE_PENDING"   // 답변대기중
	QNA_RESPONSE_COMPLETED QNA_STATUS_TYPE = "RESPONSE_COMPLETED" // 답변완료
)

func (Qna) GetQnaStatusTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(QNA_RECEIVED), Name: i18nutils.Translate("commoncode.QNA_STATUS_TYPE.QNA_RECEIVED", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(QNA_RESPONSE_PENDING), Name: i18nutils.Translate("commoncode.QNA_STATUS_TYPE.QNA_RESPONSE_PENDING", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(QNA_RESPONSE_COMPLETED), Name: i18nutils.Translate("commoncode.QNA_STATUS_TYPE.QNA_RESPONSE_COMPLETED", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: QNA_STATUS_TYPE_CODE,
		Codes:     codes,
	}
}

var qnaColumnString = `
        qna.id,
        qna.common_id,
        qna.user_id,
        qna.title,
        qna.name,
        qna.mobile_phone,
        qna.email,
        qna.content,
        qna.answer_count,
        qna.read,
        qna.status,
        qna.created,
        qna.updated
    `

type Qna struct {
	Id          int64              `json:"id"`
	CommonId    string             `json:"common_id" xorm:"not null UUID"`
	UserId      int64              `json:"user_id" xorm:"not null"`
	Title       string             `json:"title" xorm:"not null" swagger:"desc(문의 답변 제목)"`
	Name        string             `json:"name" xorm:"not null" swagger:"desc(이름/법인명)"`
	MobilePhone string             `json:"mobile_phone" swagger:"desc(전화번호)"`
	Email       string             `json:"email" xorm:"not null" swagger:"desc(이메일)"`
	Content     string             `json:"content" xorm:"not null" swagger:"desc(문의내용)"`
	Answers     []*Answer          `json:"answers" xorm:"-" swagger:"desc(답변내용)"`
	AnswerCount int64              `json:"answer_count" xorm:"not null default 0" swagger:"desc(답변 수)"`
	Read        bool               `json:"read" xorm:"not null default 0"`
	Status      QNA_STATUS_TYPE    `json:"status" xorm:"not null"`
	Actor       ChangelogActor     `json:"actor" description:"행위자" xorm:"json"`
	Created     jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated     jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Attachments []*Attachment      `json:"attachments" xorm:"-"`
}

type Answer struct {
	Id          int64              `json:"id"`
	QnaId       int64              `json:"qna_id" xorm:"not null" swagger:"desc(문의 답변 고유 ID)"`
	CommonId    string             `json:"common_id" xorm:"not null UUID"`
	Answer      string             `json:"answer" swagger:"desc(답변)"`
	Actor       ChangelogActor     `json:"actor" description:"행위자" xorm:"json"`
	Created     jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated     jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Attachments []*Attachment      `json:"attachments" xorm:"-"`
}

func (Qna) GetAll(ctx context.Context, sortby, order string, offset, limit int, userId, name, title, status *string, startDate, endDate string) (int64, []Qna, error) {
	q := factory.DB(ctx)

	q.Table("qna").
		Select(qnaColumnString).
		Join("LEFT", "answer", "qna.id = answer.qna_id")

	if status != nil && *status != "" {
		q = q.Where("status = ?", *status)
	}

	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchEqQuery(q, "qna", "user_id", userId); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchLikeQuery(q, "qna", "title", title); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchLikeQuery(q, "qna", "name", name); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SetBetweenDate(q, "qna.created", startDate, endDate); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []Qna
	var totalCount int64

	totalCount, err := q.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return totalCount, items, err
	}

	return totalCount, items, nil
}

func (Qna) GetNoAnswerQna(ctx context.Context, sortby, order string, offset, limit int) (int64, []Qna, error) {
	q := factory.DB(ctx)
	q = q.Table("qna").
		Select(qnaColumnString).
		OrderBy("answer_count ASC, qna.created DESC")
	q = q.Where("answer_count = ?", 0)
	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []Qna
	var totalCount int64

	totalCount, err := q.Limit(limit, offset).FindAndCount(&items)

	if err != nil {
		return totalCount, items, err
	}

	if len(items) == 0 {
		items = make([]Qna, 0)
	}
	return totalCount, items, nil
}

func (Qna) GetById(ctx context.Context, id int64) (*Qna, error) {
	var q Qna
	has, err := factory.DB(ctx).ID(id).Get(&q)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	sortBy := "created"
	orderBy := "desc"

	_, answers, err := Answer{}.GetAllAnswerById(ctx, q.Id, sortBy, orderBy)
	if err != nil {
		return &q, err
	}

	q.Answers = answers

	return &q, nil
}

func (Answer) GetAllAnswerById(ctx context.Context, qnaId int64, sortby, order string) (int64, []*Answer, error) {
	a := factory.DB(ctx)

	if err := modelhelper.SetSortOrder(a, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items = make([]*Answer, 0, 0)
	totalCount, err := a.Where("qna_id = ?", qnaId).FindAndCount(&items)
	if err != nil {
		return totalCount, items, err
	}

	return totalCount, items, nil
}

func (a *Answer) CreateAnswer(ctx context.Context) (int64, error) {
	uuid, err := uuid.NewV4()
	if err != nil {
		return 0, err
	}
	a.CommonId = uuid.String()
	return factory.DB(ctx).Insert(a)
}

func (a *Answer) UpdateAnswer(ctx context.Context) (int64, error) {

	return factory.DB(ctx).ID(a.Id).Update(a)
}

func (Answer) DeleteByQnaId(ctx context.Context, qnaId int64) error {

	_, answers, err := Answer{}.GetAllAnswerById(ctx, qnaId, "", "")
	if err != nil {
		return err
	}

	for _, a := range answers {
		err := Answer{}.Delete(ctx, a.Id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (Answer) Delete(ctx context.Context, id int64) error {

	_, err := factory.DB(ctx).ID(id).Delete(&Answer{})
	return err
}

func (q *Qna) Create(ctx context.Context) (int64, error) {
	uuid, err := uuid.NewV4()
	if err != nil {
		return 0, err
	}
	q.CommonId = uuid.String()

	u, err := User{}.GetById(ctx, q.UserId)
	if err != nil {
		return 0, err
	}
	if u == nil {
		return 0, nil
	}

	q.Email = u.Email
	q.MobilePhone = u.MobilePhone
	q.Name = u.Name

	return factory.DB(ctx).Insert(q)
}

func (q *Qna) Update(ctx context.Context) (int64, error) {
	query := factory.DB(ctx).UseBool("read")
	return query.ID(q.Id).Update(q)
}

// UpdateByUser は、会員が自分のQnAを編集するときの更新(MIR-1403)。件名・本文・行為者だけを書き、readはfalseに
// 戻す(管理者に未読として見せるため。今までの動きと同じ)。common_id・user_id・状態・回答件数はリクエストから
// 書き換えさせない。共通のUpdateは、admin-apiが状態や回答件数の更新に使うので変えない。
func (q *Qna) UpdateByUser(ctx context.Context) (int64, error) {
	q.Read = false
	return factory.DB(ctx).ID(q.Id).Cols("title", "content", "actor", "read").UseBool("read").Update(q)
}

func (Qna) Delete(ctx context.Context, id int64) error {

	err := Answer{}.DeleteByQnaId(ctx, id)
	if err != nil {
		return err
	}

	_, err = factory.DB(ctx).ID(id).Delete(&Qna{})
	return err
}
