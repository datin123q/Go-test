package jobs

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

const CONTRACT_MAX_COUNT int = 40

type ContractJobs struct{}

func (ContractJobs) ContractCreateJob(ctx context.Context) error {
	return runJobWithExecutionHistory(ctx, JobExecutionHistoryInput{
		JobType: domain.JOB_EXECUTION_TYPE_CONTRACT_CREATE,
		JobName: jobNameContractCreate,
	}, contractCreateJob)
}

func contractCreateJob(ctx context.Context) error {
	var con []domain.Contract
	err := factory.DB(ctx).Table("contract").Select("contract.*").
		Where("contract.status = ?", domain.CONTRACT_NONE).Limit(CONTRACT_MAX_COUNT).Find(&con)
	if err != nil {
		return err
	}

	if len(con) <= 0 {
		return nil
	}

	for i, _ := range con {

		session := factory.DB(ctx).(*xorm.Session)
		if err := session.Begin(); err != nil {
			logrus.WithFields(logrus.Fields{
				"err":  err,
				"time": time.Now(),
			}).Error("fail contract session begin")
			continue
		}

		con[i].Status = domain.CONTRACT_PROCESSING
		if err := con[i].Update(ctx); err != nil {
			con[i].Status = domain.CONTRACT_NONE
			if err2 := session.Rollback(); err2 != nil {
				logrus.WithFields(logrus.Fields{
					"err":  err2,
					"time": time.Now(),
				}).Error("fail contract session rollback")
			}

			logrus.WithFields(logrus.Fields{
				"err":  err,
				"time": time.Now(),
			}).Error(" fail contract update")
			continue
		}

		if err := session.Commit(); err != nil {
			con[i].Status = domain.CONTRACT_NONE
			if err2 := session.Rollback(); err2 != nil {
				logrus.WithFields(logrus.Fields{
					"err":  err2,
					"time": time.Now(),
				}).Error(" fail contract session rollback")
			}

			logrus.WithFields(logrus.Fields{
				"err":  err,
				"time": time.Now(),
			}).Error(" fail contract session commit")
		}
	}

	for i, _ := range con {

		if con[i].Status != domain.CONTRACT_PROCESSING {
			continue
		}

		if err := con[i].UpdateContractAfterPage(ctx); err != nil {
			logrus.WithFields(logrus.Fields{
				"category": "UpdateContractAfterPage",
				"err":      err,
				"time":     time.Now(),
			}).Error(" fail contract update contract after page")

			session := factory.DB(ctx).(*xorm.Session)
			if err := session.Begin(); err != nil {
				logrus.WithFields(logrus.Fields{
					"category": "UpdateContractAfterPage",
					"err":      err,
					"time":     time.Now(),
				}).Error(" fail contract session begin")
				continue
			}

			con[i].Status = domain.CONTRACT_NONE
			con[i].ContractAfterPageUrl = ""
			if err := con[i].Update(ctx); err != nil {
				if err2 := session.Rollback(); err2 != nil {
					logrus.WithFields(logrus.Fields{
						"category": "UpdateContractAfterPage",
						"err":      err2,
						"time":     time.Now(),
					}).Error("fail contract session rollback")
				}

				logrus.WithFields(logrus.Fields{
					"category": "UpdateContractAfterPage",
					"err":      err,
					"time":     time.Now(),
				}).Error(" fail contract update")
				continue
			}

			if err := session.Commit(); err != nil {
				err2 := session.Rollback()
				if err2 != nil {
					logrus.WithFields(logrus.Fields{
						"category": "UpdateContractAfterPage",
						"err":      err2,
						"time":     time.Now(),
					}).Error(" fail contract session rollback")
				}

				logrus.WithFields(logrus.Fields{
					"category": "UpdateContractAfterPage",
					"err":      err,
					"time":     time.Now(),
				}).Error(" fail contract session commit")
			}
		}
	}

	logrus.WithFields(logrus.Fields{
		"category": "create contract",
		"time":     time.Now(),
	}).Info("ContractCreateJob")

	return nil
}
