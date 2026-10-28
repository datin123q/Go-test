package mysql

import (
	"context"
	"time"

	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/tool"
)

// db factory mock
type dbFactoryMock struct{}

var (
	t1, _  = time.ParseInLocation("2006-01-02 15:04:05", "2021-07-09 17:41:45", time.Local)
	t2, _  = time.ParseInLocation("2006-01-02 15:04:05", "2021-07-29 13:42:41", time.Local)
	t3, _  = time.ParseInLocation("2006-01-02 15:04:05", "2021-07-29 13:43:11", time.Local)
	t4, _  = time.ParseInLocation("2006-01-02 15:04:05", "2021-07-29 13:43:27", time.Local)
	t5, _  = time.ParseInLocation("2006-01-02 15:04:05", "2021-07-29 13:43:45", time.Local)
	t6, _  = time.ParseInLocation("2006-01-02 15:04:05", "2021-08-08 13:23:47", time.Local)
	t7, _  = time.ParseInLocation("2006-01-02 15:04:05", "2021-08-18 14:54:09", time.Local)
	t8, _  = time.ParseInLocation("2006-01-02 15:04:05", "2021-08-18 14:54:31", time.Local)
	t9, _  = time.ParseInLocation("2006-01-02 15:04:05", "2021-08-18 14:54:49", time.Local)
	t10, _ = time.ParseInLocation("2006-01-02 15:04:05", "2021-08-18 14:55:07", time.Local)
	t11, _ = time.ParseInLocation("2006-01-02 15:04:05", "2021-09-23 09:56:46", time.Local)
	t12, _ = time.ParseInLocation("2006-01-02 15:04:05", "2022-02-15 13:22:32", time.Local)

	mockNoticeMsg1 = domain.NoticeMsg{
		Id:             1,
		Title:          "test",
		ContentHtml:    "<p>test</p>",
		NoticeDate:     jsonutils.JSONTime(t1),
		Created:        jsonutils.JSONTime(t1),
		NoticeUserType: domain.NOTICE_NEWSLETTER,
		ProjectId:      0,
	}
	mockNoticeMsg2 = domain.NoticeMsg{
		Id:             2,
		Title:          "뉴스레터 수신동의 테스트",
		ContentHtml:    "<p>123</p>",
		NoticeDate:     jsonutils.JSONTime(t2),
		Created:        jsonutils.JSONTime(t2),
		NoticeUserType: domain.NOTICE_NEWSLETTER,
		ProjectId:      0,
	}
	mockNoticeMsg3 = domain.NoticeMsg{
		Id:             3,
		Title:          "직접입력 테스트",
		ContentHtml:    "<p>직접입력 테스트</p>",
		NoticeDate:     jsonutils.JSONTime(t3),
		Created:        jsonutils.JSONTime(t3),
		NoticeUserType: domain.NOTICE_DIRECT,
		ProjectId:      0,
	}
	mockNoticeMsg4 = domain.NoticeMsg{
		Id:             4,
		Title:          "회원 구분 테스트",
		ContentHtml:    "<p>회원 구분 테스트</p>",
		NoticeDate:     jsonutils.JSONTime(t4),
		Created:        jsonutils.JSONTime(t4),
		NoticeUserType: domain.NOTICE_USER,
		ProjectId:      0,
	}
	mockNoticeMsg5 = domain.NoticeMsg{
		Id:             5,
		Title:          "투자상품별테스트",
		ContentHtml:    "<p>투자상품별테스트</p>",
		NoticeDate:     jsonutils.JSONTime(t5),
		Created:        jsonutils.JSONTime(t5),
		NoticeUserType: domain.NOTICE_PROJECT,
		ProjectId:      0,
	}
	mockNoticeMsg6 = domain.NoticeMsg{
		Id:             6,
		Title:          "123",
		ContentHtml:    "<p>123</p>",
		NoticeDate:     jsonutils.JSONTime(t6),
		Created:        jsonutils.JSONTime(t6),
		NoticeUserType: domain.NOTICE_PROJECT,
		ProjectId:      2,
	}
	mockNoticeMsg7 = domain.NoticeMsg{
		Id:             7,
		Title:          "test",
		ContentHtml:    "<p>test</p>",
		NoticeDate:     jsonutils.JSONTime(t7),
		Created:        jsonutils.JSONTime(t7),
		NoticeUserType: domain.NOTICE_NEWSLETTER,
		ProjectId:      0,
	}
	mockNoticeMsg8 = domain.NoticeMsg{
		Id:             8,
		Title:          "test2",
		ContentHtml:    "<p>test2</p>",
		NoticeDate:     jsonutils.JSONTime(t8),
		Created:        jsonutils.JSONTime(t8),
		NoticeUserType: domain.NOTICE_NEWSLETTER,
		ProjectId:      0,
	}
	mockNoticeMsg9 = domain.NoticeMsg{
		Id:             9,
		Title:          "test3",
		ContentHtml:    "<p>test3</p>",
		NoticeDate:     jsonutils.JSONTime(t9),
		Created:        jsonutils.JSONTime(t9),
		NoticeUserType: domain.NOTICE_NEWSLETTER,
		ProjectId:      0,
	}
	mockNoticeMsg10 = domain.NoticeMsg{
		Id:             10,
		Title:          "test4",
		ContentHtml:    "<p>test4</p>",
		NoticeDate:     jsonutils.JSONTime(t10),
		Created:        jsonutils.JSONTime(t10),
		NoticeUserType: domain.NOTICE_NEWSLETTER,
		ProjectId:      0,
	}
	mockNoticeMsg11 = domain.NoticeMsg{
		Id:             11,
		Title:          "123",
		ContentHtml:    "<p>123</p>",
		NoticeDate:     jsonutils.JSONTime(t11),
		Created:        jsonutils.JSONTime(t11),
		NoticeUserType: domain.NOTICE_PROJECT,
		ProjectId:      30,
	}
	mockNoticeMsg12 = domain.NoticeMsg{
		Id:             12,
		Title:          "test",
		ContentHtml:    "<p>test</p>",
		NoticeDate:     jsonutils.JSONTime(t12),
		Created:        jsonutils.JSONTime(t12),
		NoticeUserType: domain.NOTICE_DIRECT,
		ProjectId:      0,
	}
)

var (
	user107_SignUpDate, _        = time.ParseInLocation("2006-01-02 15:04:05", "2020-12-16 14:29:59", time.Local)
	user107_LastLoginDate, _     = time.ParseInLocation("2006-01-02 15:04:05", "2020-12-16 14:32:39", time.Local)
	user107_NewsletterAgreeAt, _ = time.ParseInLocation("2006-01-02 15:04:05", "2020-12-01 11:11:11", time.Local)
	user107_SmsCertAt, _         = time.ParseInLocation("2006-01-02 15:04:05", "2020-12-01 11:11", time.Local)
	user107                      = domain.User{
		Id:                107,
		Uid:               "10000107",
		LoginType:         domain.MAIL,
		CommonId:          "48fe34b8-726e-508c-8414-a453cb0fed26",
		InvestorType:      domain.PRIVATE_INVESTOR,
		Name:              "박 근석",
		Email:             "keunseok.park@illuminarean.com",
		Password:          "53d8b9379c8dac1546713dc61b5dff7a6e35dd6dbbaa3e418299bc15170b5f60ccbf15fd432da042c7ae2ee8029f87b97435",
		Salt:              "TRfPLosMtA",
		MobilePhone:       "01012345678",
		SignUpDate:        jsonutils.JSONTime(user107_SignUpDate),
		LastLoginDate:     jsonutils.JSONTime(user107_LastLoginDate),
		BankName:          "ｼﾞ-ｴﾑｵ-ｱｵｿﾞﾗﾈﾂﾄ",
		BankCode:          tool.StringToPointer("0310"),
		BankAccount:       "1001413",
		BranchName:        "ﾆｼﾞｼﾃﾝ",
		BranchCode:        tool.StringToPointer("302"),
		DepositType:       domain.ORDINARY_ACCOUNT,
		Depositor:         "ｺﾊﾞﾔｼ ｴﾘﾄ",
		ZipCode:           "12345678",
		Address:           "강남구 ",
		Address2:          tool.StringToPointer("미왕빌딩 301호"),
		Province:          "미소현 ",
		City:              "서울시",
		UserType:          domain.PRIVATE,
		AuditStatus:       domain.INVALIDITY_AUDIT,
		InvestmentPurpose: "SURPLUS_MONEY",
		IsSurplusFund:     tool.BoolToPointer(true),
		IsNewsletter:      true,
		NewsletterAgreeAt: jsonutils.JSONTime(user107_NewsletterAgreeAt),
		IsSmsCert:         true,
		SmsCertAt:         jsonutils.JSONTime(user107_SmsCertAt),
		PostalCode:        "213223",
		PostalCodeStatus:  true,
		InvalidReason:     tool.StringToPointer("없음"),
	}

	user111_SignUpDate, _        = time.ParseInLocation("2006-01-02 15:04:05", "2020-12-30 10:51:57", time.Local)
	user111_LastLoginDate, _     = time.ParseInLocation("2006-01-02 15:04:05", "2020-12-29 13:46:41", time.Local)
	user111_NewsletterAgreeAt, _ = time.ParseInLocation("2006-01-02 15:04:05", "2020-12-01 11:11:11", time.Local)
	user111_SmsCertAt, _         = time.ParseInLocation("2006-01-02 15:04:05", "2020-12-01 11:11:11", time.Local)
	user111                      = domain.User{
		Id:                111,
		Uid:               "10000111",
		LoginType:         domain.MAIL,
		CommonId:          "09dba0fc-fc0e-577e-afbc-3467b6aadaf3",
		InvestorType:      domain.CORPORATE_INVESTOR,
		Name:              "Test Company",
		Email:             "corporation@test.com",
		Password:          "efdd1fca37cd909b7f4a3e8f3e4b30afe5421acca503b0d9b44cf55819d23cd2ebadadfbadfd15cd2e5f39bc73e2fb2130ff",
		Salt:              "cv7mHmYU65",
		MobilePhone:       "01098467073",
		SignUpDate:        jsonutils.JSONTime(user111_SignUpDate),
		LastLoginDate:     jsonutils.JSONTime(user111_LastLoginDate),
		BankName:          "ｼﾞ-ｴﾑｵ-ｱｵｿﾞﾗﾈﾂﾄ",
		BankCode:          tool.StringToPointer("0310"),
		BankAccount:       "1001413",
		BranchName:        "ﾆｼﾞｼﾃﾝ",
		BranchCode:        tool.StringToPointer("302"),
		DepositType:       domain.ORDINARY_ACCOUNT,
		Depositor:         "ｺﾊﾞﾔｼ ｴﾘﾄ",
		ZipCode:           "1000002",
		Address:           "皇居外苑",
		Address2:          tool.StringToPointer("1310호"),
		Province:          "東京都 ",
		City:              "千代田区",
		UserType:          domain.CORPORATE,
		AuditStatus:       domain.WAIT_AUDIT,
		InvestmentPurpose: "SURPLUS_MONEY",
		IsSurplusFund:     tool.BoolToPointer(true),
		IsNewsletter:      true,
		NewsletterAgreeAt: jsonutils.JSONTime(user111_NewsletterAgreeAt),
		IsSmsCert:         true,
		SmsCertAt:         jsonutils.JSONTime(user111_SmsCertAt),
		PostalCode:        "213223",
		PostalCodeStatus:  false,
		InvalidReason:     tool.StringToPointer(""),
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.User), new(domain.NoticeMsg), new(domain.UserNoticeMapping))
	xormEngine.ShowSQL(true)

	// テストデータ挿入
	xormEngine.Insert(user107, user111,
		mockNoticeMsg1, mockNoticeMsg2, mockNoticeMsg3, mockNoticeMsg4, mockNoticeMsg5,
		mockNoticeMsg6, mockNoticeMsg7, mockNoticeMsg8, mockNoticeMsg9, mockNoticeMsg10,
		mockNoticeMsg11, mockNoticeMsg12,
	)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
