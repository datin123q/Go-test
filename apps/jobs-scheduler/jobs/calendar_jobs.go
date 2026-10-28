package jobs

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/calender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type CalendarJobs struct {
}

// 배치 시점 기준으로 1년치의 공휴일을 조회하여 저장한다. (일본)
func (c CalendarJobs) CreateHolidayCalendarJob(ctx context.Context) error {
	return runJobWithExecutionHistory(ctx, JobExecutionHistoryInput{
		JobType: domain.JOB_EXECUTION_TYPE_CREATE_HOLIDAY_CALENDAR,
		JobName: jobNameCreateHolidayCalendar,
	}, createHolidayCalendarJob)
}

func createHolidayCalendarJob(ctx context.Context) error {
	startDate := time.Now()
	endDate := time.Now().AddDate(1, 0, 0)
	logrus.WithFields(logrus.Fields{
		"desc":       "start create holiday calendar job",
		"start_date": startDate,
		"end_date":   endDate,
	}).Info("CreateHolidayCalendarJob")
	items, err := RequestHolidayCalendar(ctx, startDate, endDate)
	if err != nil {
		return errors.New("failed to create holiday calendar job")
	}

	for i := range items {
		start, err := jsonutils.ParseDateTimeJST(time.DateOnly, items[i].Start.Date)
		if err != nil {
			return errors.New("failed to parse start date")
		}

		end, err := jsonutils.ParseDateTimeJST(time.DateOnly, items[i].End.Date)
		if err != nil {
			return errors.New("failed to parse end date")
		}

		for date := start; date.Before(end); date = date.AddDate(0, 0, 1) {
			h := domain.NewHoliday(items[i].Summary, date)
			done, err := h.IsExist(ctx)
			if err != nil {
				return errors.Wrap(err, "failed to check holiday exist")
			}
			if done {
				continue
			}
			_, err = h.Create(ctx)
			if err != nil {
				return errors.Wrap(err, "failed to create holiday")
			}
		}
	}
	logrus.WithFields(logrus.Fields{
		"desc":       "success create holiday calendar job",
		"start_date": startDate,
		"end_date":   endDate,
	}).Info("CreateHolidayCalendarJob")
	return nil
}

// RequestHolidayCalendar is a function that requests holiday calendar
func RequestHolidayCalendar(ctx context.Context, startDate, endDate time.Time) ([]calender.GoogleCalenderEvent, error) {
	calender := calender.NewCalendarClient()
	items, err := calender.GetPublicHolidays(startDate, endDate)
	if err != nil {
		return items, errors.Wrap(err, "failed to get public holidays with google api")
	}

	return items, nil
}
