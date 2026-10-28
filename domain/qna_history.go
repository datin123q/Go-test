package domain

import (
	"context"
	"crypto/sha256"

	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
)

type QNA_CHANGED_CONTENT_TYPE string

const (
	QNA_CHANGED_CONTENT_TYPE_CODE                            = "QNA_CHANGED_CONTENT_TYPE"
	QNA_CHANGED_CONTENT_FIRST       QNA_CHANGED_CONTENT_TYPE = "QNA_CHANGED_CONTENT_FIRST"       // 답변 최초 등록
	QNA_CHANGED_CONTENT_ANSWER      QNA_CHANGED_CONTENT_TYPE = "QNA_CHANGED_CONTENT_ANSWER"      // 답변 변경
	QNA_CHANGED_CONTENT_BEFORE_TEXT QNA_CHANGED_CONTENT_TYPE = "QNA_CHANGED_CONTENT_BEFORE_TEXT" // (변경 전)답변내용
	QNA_CHANGED_CONTENT_AFTER_TEXT  QNA_CHANGED_CONTENT_TYPE = "QNA_CHANGED_CONTENT_AFTER_TEXT"  // (변경 후)답변내용
	QNA_CHANGED_CONTENT_TEXT        QNA_CHANGED_CONTENT_TYPE = "QNA_CHANGED_CONTENT_TEXT"        // 답변내용
)

func (QnaInfoHistory) GetQnaChangedContentTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(QNA_CHANGED_CONTENT_FIRST), Name: i18nutils.Translate("commoncode.QNA_CHANGED_CONTENT_TYPE.QNA_CHANGED_CONTENT_FIRST", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(QNA_CHANGED_CONTENT_ANSWER), Name: i18nutils.Translate("commoncode.QNA_CHANGED_CONTENT_TYPE.QNA_CHANGED_CONTENT_ANSWER", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(QNA_CHANGED_CONTENT_BEFORE_TEXT), Name: i18nutils.Translate("commoncode.QNA_CHANGED_CONTENT_TYPE.QNA_CHANGED_CONTENT_BEFORE_TEXT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(QNA_CHANGED_CONTENT_AFTER_TEXT), Name: i18nutils.Translate("commoncode.QNA_CHANGED_CONTENT_TYPE.QNA_CHANGED_CONTENT_AFTER_TEXT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(QNA_CHANGED_CONTENT_TEXT), Name: i18nutils.Translate("commoncode.QNA_CHANGED_CONTENT_TYPE.QNA_CHANGED_CONTENT_TEXT", loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: QNA_CHANGED_CONTENT_TYPE_CODE,
		Codes:     codes,
	}
}

type QnaInfoHistory struct {
	Id              int64              `json:"id"`                                // 고유아이디
	UserId          int64              `json:"user_id"`                           // 문의한 유저 아이디
	ProcessId       int64              `json:"process_id"`                        // 처리자 아이디
	QnaId           int64              `json:"qna_id"`                            // 상태변경된 QNA 아이디
	ContentType     string             `json:"content_type"`                      // 변경된 내용
	PrevContent     string             `json:"prev_content"`                      // 변경 전 문구
	CurrContent     string             `json:"curr_content"`                      // 변경 후 문구
	PrevContentHtml string             `json:"prev_content_html" xorm:"LONGTEXT"` // 변경 전 내용
	CurrContentHtml string             `json:"curr_content_html" xorm:"LONGTEXT"` // 변경 후 내용
	Name            string             `json:"name"`                              // 처리자
	Email           string             `json:"email"`                             // 처리자 이메일
	Created         jsonutils.JSONTime `json:"created" xorm:"created"`            // 처리일시
}

func (QnaInfoHistory) GetAll(ctx context.Context, sortby, order string, offset, limit int, qnaId *string) (int64, []QnaInfoHistory, error) {

	query := factory.DB(ctx).Table("qna_info_history").Select("qna_info_history.*")

	if err := modelhelper.SetSortOrder(query, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchEqQuery(query, "qna_info_history", "qna_id", qnaId); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var qnaInfoHistoryList []QnaInfoHistory
	var totalCount int64
	totalCount, err := query.Desc("created").Limit(limit, offset).FindAndCount(&qnaInfoHistoryList)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, qnaInfoHistoryList, nil
}

func (QnaInfoHistory) Create(ctx context.Context, before, after interface{}, processId int64) (int64, error) {

	prevQna := before.(*Qna)
	currQna := after.(*Qna)

	var prev string
	var curr string
	if len(prevQna.Answers) <= 0 {
		if len(currQna.Answers) <= 0 {
			return 0, nil
		}
		prev = ""
		curr = currQna.Answers[0].Answer
	} else if len(currQna.Answers) <= 0 {
		logrus.Println("Has not been answered.")
		return 0, nil
	} else {
		prev = prevQna.Answers[0].Answer
		curr = currQna.Answers[0].Answer
	}

	// hash 변경 후 비교
	s1 := sha256.Sum256([]byte(prev))
	s2 := sha256.Sum256([]byte(curr))

	// 변경된 내용 비교 후 히스토리를 남김
	if s1 != s2 {

		user, _ := AdminUser{}.GetById(ctx, processId)

		var qna QnaInfoHistory
		qna.QnaId = currQna.Id
		qna.UserId = currQna.UserId
		qna.ProcessId = processId
		qna.Name = user.Name
		qna.Email = user.Email
		if len(prev) <= 0 {
			qna.ContentType = string(QNA_CHANGED_CONTENT_FIRST)
			qna.PrevContent = ""
			qna.CurrContent = string(QNA_CHANGED_CONTENT_TEXT)
		} else {
			qna.ContentType = string(QNA_CHANGED_CONTENT_ANSWER)
			qna.PrevContent = string(QNA_CHANGED_CONTENT_BEFORE_TEXT)
			qna.CurrContent = string(QNA_CHANGED_CONTENT_AFTER_TEXT)
		}
		qna.PrevContentHtml = prev
		qna.CurrContentHtml = curr

		if _, err := factory.DB(ctx).Table("qna_info_history").Insert(&qna); err != nil {
			logrus.Println("Error saving qna information history")
		}
	}

	return 0, nil
}
