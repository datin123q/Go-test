package mysql

import (
	"context"
	"time"

	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

// db factory mock
type dbFactoryMock struct{}

var (
	baseTime        = time.Now()
	loginTime1, _   = time.ParseInLocation(time.DateTime, "2021-06-23 12:14:25", time.Local)
	createTime1, _  = time.ParseInLocation(time.DateTime, "2021-06-23 12:14:25", time.Local)
	loginTime2, _   = time.ParseInLocation(time.DateTime, "2021-06-23 13:21:04", time.Local)
	createTime2, _  = time.ParseInLocation(time.DateTime, "2021-06-23 13:21:04", time.Local)
	loginTime3, _   = time.ParseInLocation(time.DateTime, "2021-06-23 13:21:05", time.Local)
	createTime3, _  = time.ParseInLocation(time.DateTime, "2021-06-23 13:21:05", time.Local)
	loginTime4, _   = time.ParseInLocation(time.DateTime, "2021-06-23 15:47:36", time.Local)
	createTime4, _  = time.ParseInLocation(time.DateTime, "2021-06-23 15:47:36", time.Local)
	loginTime5, _   = time.ParseInLocation(time.DateTime, "2021-06-23 16:07:38", time.Local)
	createTime5, _  = time.ParseInLocation(time.DateTime, "2021-06-23 16:07:38", time.Local)
	loginTime6, _   = time.ParseInLocation(time.DateTime, "2021-06-23 16:45:38", time.Local)
	createTime6, _  = time.ParseInLocation(time.DateTime, "2021-06-23 16:45:38", time.Local)
	loginTime7, _   = time.ParseInLocation(time.DateTime, "2021-06-23 19:00:35", time.Local)
	createTime7, _  = time.ParseInLocation(time.DateTime, "2021-06-23 19:00:35", time.Local)
	loginTime8, _   = time.ParseInLocation(time.DateTime, "2021-06-23 19:08:30", time.Local)
	createTime8, _  = time.ParseInLocation(time.DateTime, "2021-06-23 19:08:30", time.Local)
	loginTime9, _   = time.ParseInLocation(time.DateTime, "2021-06-23 19:08:45", time.Local)
	createTime9, _  = time.ParseInLocation(time.DateTime, "2021-06-23 19:08:45", time.Local)
	loginTime10, _  = time.ParseInLocation(time.DateTime, "2021-06-23 19:30:18", time.Local)
	createTime10, _ = time.ParseInLocation(time.DateTime, "2021-06-23 19:30:18", time.Local)
	loginTime11, _  = time.ParseInLocation(time.DateTime, "2021-06-23 20:15:31", time.Local)
	createTime11, _ = time.ParseInLocation(time.DateTime, "2021-06-23 20:15:31", time.Local)
	loginTime12, _  = time.ParseInLocation(time.DateTime, "2021-06-23 20:16:18", time.Local)
	createTime12, _ = time.ParseInLocation(time.DateTime, "2021-06-23 20:16:18", time.Local)
	loginTime13, _  = time.ParseInLocation(time.DateTime, "2021-06-23 20:35:04", time.Local)
	createTime13, _ = time.ParseInLocation(time.DateTime, "2021-06-23 20:35:04", time.Local)
	loginTime14, _  = time.ParseInLocation(time.DateTime, "2021-06-23 20:36:18", time.Local)
	createTime14, _ = time.ParseInLocation(time.DateTime, "2021-06-23 20:36:18", time.Local)
	loginTime15, _  = time.ParseInLocation(time.DateTime, "2021-06-23 20:37:53", time.Local)
	createTime15, _ = time.ParseInLocation(time.DateTime, "2021-06-23 20:37:53", time.Local)
	loginTime16, _  = time.ParseInLocation(time.DateTime, "2021-06-23 21:02:18", time.Local)
	createTime16, _ = time.ParseInLocation(time.DateTime, "2021-06-23 21:02:18", time.Local)
	loginTime17, _  = time.ParseInLocation(time.DateTime, "2021-06-23 21:06:23", time.Local)
	createTime17, _ = time.ParseInLocation(time.DateTime, "2021-06-23 21:06:23", time.Local)
	loginTime18, _  = time.ParseInLocation(time.DateTime, "2021-06-23 21:08:55", time.Local)
	createTime18, _ = time.ParseInLocation(time.DateTime, "2021-06-23 21:08:55", time.Local)
	loginTime19, _  = time.ParseInLocation(time.DateTime, "2021-06-23 21:10:14", time.Local)
	createTime19, _ = time.ParseInLocation(time.DateTime, "2021-06-23 21:10:14", time.Local)
	loginTime20, _  = time.ParseInLocation(time.DateTime, "2021-06-23 21:12:05", time.Local)
	createTime20, _ = time.ParseInLocation(time.DateTime, "2021-06-23 21:12:05", time.Local)

	mockLoginHistory1 = domain.LoginHistory{
		Id:          1,
		UserId:      6,
		Name:        "パク・ジュンヨン",
		Email:       "pjy1418@gmail.com",
		IpAddr:      "59.18.200.132",
		Agent:       "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Mobile Safari/537.36",
		LoginResult: domain.SUCCESS_LOGIN,
		LoginTime:   jsonutils.JSONTime(loginTime1),
		Created:     jsonutils.JSONTime(createTime1),
	}
	mockLoginHistory2 = domain.LoginHistory{
		Id:          2,
		UserId:      110,
		Name:        "やまだ タロ",
		Email:       "private@test.com",
		IpAddr:      "110.10.64.236",
		Agent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36",
		LoginResult: domain.SUCCESS_LOGIN,
		LoginTime:   jsonutils.JSONTime(loginTime2),
		Created:     jsonutils.JSONTime(createTime2),
	}
	mockLoginHistory3 = domain.LoginHistory{
		Id:          3,
		UserId:      1,
		Name:        "ADMIN",
		Email:       "admin@energyshares.jp",
		IpAddr:      "10.8.1.30",
		Agent:       "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Mobile Safari/537.36",
		LoginResult: domain.SUCCESS_LOGIN,
		LoginTime:   jsonutils.JSONTime(loginTime3),
		Created:     jsonutils.JSONTime(createTime3),
	}
	mockLoginHistory4 = domain.LoginHistory{
		Id:          4,
		UserId:      108,
		Name:        "ホソン・ジョン",
		Email:       "inbq1190@naver.com",
		IpAddr:      "127.0.0.1",
		Agent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36",
		LoginResult: domain.SUCCESS_LOGIN,
		LoginTime:   jsonutils.JSONTime(loginTime4),
		Created:     jsonutils.JSONTime(createTime4),
	}
	mockLoginHistory5 = domain.LoginHistory{
		Id:          5,
		UserId:      1,
		Name:        "ADMIN",
		Email:       "admin@energyshares.jp",
		IpAddr:      "10.8.1.26",
		Agent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36",
		LoginResult: domain.SUCCESS_LOGIN,
		LoginTime:   jsonutils.JSONTime(loginTime5),
		Created:     jsonutils.JSONTime(createTime5),
	}
	mockLoginHistory6 = domain.LoginHistory{
		Id:          6,
		UserId:      110,
		Name:        "やまだ タロ",
		Email:       "private@test.com",
		IpAddr:      "10.8.1.42",
		Agent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36",
		LoginResult: domain.SUCCESS_LOGIN,
		LoginTime:   jsonutils.JSONTime(loginTime6),
		Created:     jsonutils.JSONTime(createTime6),
	}
	mockLoginHistory7 = domain.LoginHistory{
		Id:          7,
		UserId:      110,
		Name:        "やまだ タロ",
		Email:       "private@test.com",
		IpAddr:      "10.8.1.42",
		Agent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36",
		LoginResult: domain.SUCCESS_LOGIN,
		LoginTime:   jsonutils.JSONTime(loginTime7),
		Created:     jsonutils.JSONTime(createTime7),
	}
	mockLoginHistory8 = domain.LoginHistory{
		Id:          8,
		UserId:      110,
		Name:        "やまだ タロ",
		Email:       "private@test.com",
		IpAddr:      "110.10.64.236",
		Agent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36",
		LoginResult: domain.SUCCESS_LOGIN,
		LoginTime:   jsonutils.JSONTime(loginTime8),
		Created:     jsonutils.JSONTime(createTime8),
	}
	mockLoginHistory9 = domain.LoginHistory{
		Id:          9,
		UserId:      1,
		Name:        "ADMIN",
		Email:       "admin@energyshares.jp",
		IpAddr:      "10.8.1.26",
		Agent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36",
		LoginResult: domain.SUCCESS_LOGIN,
		LoginTime:   jsonutils.JSONTime(loginTime9),
		Created:     jsonutils.JSONTime(createTime9),
	}
	mockLoginHistory10 = domain.LoginHistory{
		Id:          10,
		UserId:      110,
		Name:        "やまだ タロ",
		Email:       "private@test.com",
		IpAddr:      "10.8.1.42",
		Agent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36",
		LoginResult: domain.SUCCESS_LOGIN,
		LoginTime:   jsonutils.JSONTime(loginTime10),
		Created:     jsonutils.JSONTime(createTime10),
	}
	mockLoginHistory11 = domain.LoginHistory{
		Id:          11,
		UserId:      108,
		Name:        "ホソン・ジョン",
		Email:       "inbq1190@naver.com",
		IpAddr:      "127.0.0.1",
		Agent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36",
		LoginResult: domain.SUCCESS_LOGIN,
		LoginTime:   jsonutils.JSONTime(loginTime11),
		Created:     jsonutils.JSONTime(createTime11),
	}
	mockLoginHistory12 = domain.LoginHistory{
		Id:          12,
		UserId:      108,
		Name:        "ホソン・ジョン",
		Email:       "inbq1190@naver.com",
		IpAddr:      "127.0.0.1",
		Agent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36",
		LoginResult: domain.SUCCESS_LOGIN,
		LoginTime:   jsonutils.JSONTime(loginTime12),
		Created:     jsonutils.JSONTime(createTime12),
	}
	mockLoginHistory13 = domain.LoginHistory{
		Id:          13,
		UserId:      108,
		Name:        "ホソン・ジョン",
		Email:       "inbq1190@naver.com",
		IpAddr:      "127.0.0.1",
		Agent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36",
		LoginResult: domain.SUCCESS_LOGIN,
		LoginTime:   jsonutils.JSONTime(loginTime13),
		Created:     jsonutils.JSONTime(createTime13),
	}
	mockLoginHistory14 = domain.LoginHistory{
		Id:          14,
		UserId:      110,
		Name:        "やまだ タロ",
		Email:       "private@test.com",
		IpAddr:      "127.0.0.1",
		Agent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36",
		LoginResult: domain.SUCCESS_LOGIN,
		LoginTime:   jsonutils.JSONTime(loginTime14),
		Created:     jsonutils.JSONTime(createTime14),
	}
	mockLoginHistory15 = domain.LoginHistory{
		Id:          15,
		UserId:      110,
		Name:        "やまだ タロ",
		Email:       "private@test.com",
		IpAddr:      "127.0.0.1",
		Agent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36",
		LoginResult: domain.SUCCESS_LOGIN,
		LoginTime:   jsonutils.JSONTime(loginTime15),
		Created:     jsonutils.JSONTime(createTime15),
	}
	mockLoginHistory16 = domain.LoginHistory{
		Id:          16,
		UserId:      110,
		Name:        "やまだ タロ",
		Email:       "private@test.com",
		IpAddr:      "127.0.0.1",
		Agent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36",
		LoginResult: domain.SUCCESS_LOGIN,
		LoginTime:   jsonutils.JSONTime(loginTime16),
		Created:     jsonutils.JSONTime(createTime16),
	}
	mockLoginHistory17 = domain.LoginHistory{
		Id:          17,
		UserId:      110,
		Name:        "やまだ タロ",
		Email:       "private@test.com",
		IpAddr:      "127.0.0.1",
		Agent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36",
		LoginResult: domain.SUCCESS_LOGIN,
		LoginTime:   jsonutils.JSONTime(loginTime17),
		Created:     jsonutils.JSONTime(createTime17),
	}
	mockLoginHistory18 = domain.LoginHistory{
		Id:          18,
		UserId:      110,
		Name:        "やまだ タロ",
		Email:       "private@test.com",
		IpAddr:      "127.0.0.1",
		Agent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36",
		LoginResult: domain.SUCCESS_LOGIN,
		LoginTime:   jsonutils.JSONTime(loginTime18),
		Created:     jsonutils.JSONTime(createTime18),
	}
	mockLoginHistory19 = domain.LoginHistory{
		Id:          19,
		UserId:      110,
		Name:        "やまだ タロ",
		Email:       "private@test.com",
		IpAddr:      "127.0.0.1",
		Agent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36",
		LoginResult: domain.SUCCESS_LOGIN,
		LoginTime:   jsonutils.JSONTime(loginTime19),
		Created:     jsonutils.JSONTime(createTime19),
	}
	mockLoginHistory20 = domain.LoginHistory{
		Id:          20,
		UserId:      110,
		Name:        "やまだ タロ",
		Email:       "private@test.com",
		IpAddr:      "127.0.0.1",
		Agent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36",
		LoginResult: domain.SUCCESS_LOGIN,
		LoginTime:   jsonutils.JSONTime(loginTime20),
		Created:     jsonutils.JSONTime(createTime20),
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.LoginHistory))
	xormEngine.ShowSQL(true)

	// テストデータ挿入
	xormEngine.Insert(mockLoginHistory1, mockLoginHistory2, mockLoginHistory3, mockLoginHistory4, mockLoginHistory5,
		mockLoginHistory6, mockLoginHistory7, mockLoginHistory8, mockLoginHistory9, mockLoginHistory10,
		mockLoginHistory11, mockLoginHistory12, mockLoginHistory13, mockLoginHistory14, mockLoginHistory15,
		mockLoginHistory16, mockLoginHistory17, mockLoginHistory18, mockLoginHistory19, mockLoginHistory20)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
