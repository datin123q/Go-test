package jobs

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/sirupsen/logrus"
)

type UsersJobs struct{}

func (UsersJobs) DeletedUserJob(ctx context.Context) error {
	// 1. 탈퇴 후 30일 지난 이메일 처리
	// 2. 기존 이메일을 변경하여 동일한 이메일로도 가입 할수 있도록 변경
	// 3. 향후 메일 데이터 암호화가 진행될 경우 변경 될수 있음
	q := factory.DB(ctx)
	ret, err := q.Exec("UPDATE energysharesdb.user SET `email` = CONCAT(`email`,\"_\",DATE_FORMAT(`deleted`, \"%Y.%M.%D_deleted\")), `withdrawal_type` = 2 WHERE `deleted` IS NOT NULL AND `deleted` >= DATE_SUB(`deleted`, INTERVAL 30 DAY) AND (`withdrawal_type` = 0 OR `withdrawal_type` = 1)")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"desc": "Fail UPDATE User Table",
			"ret":  ret,
			"err":  err,
		}).Error("Deleted User Jobs")
		return err
	}

	logrus.WithFields(logrus.Fields{
		"desc": "success update",
		"ret":  ret,
	}).Info("Deleted User Jobs")

	return nil
}
