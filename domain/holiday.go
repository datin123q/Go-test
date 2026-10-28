package domain

import (
	"context"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type Holiday struct {
	Id          int64              `json:"id"`
	Title       string             `json:"title"`
	Date        string             `json:"date"`
	HolidayDate jsonutils.JSONTime `json:"holiday_date"`
	Created     jsonutils.JSONTime `json:"created" xorm:"created"`
}

func NewHoliday(Title string, date time.Time) *Holiday {
	return &Holiday{
		Title:       Title,
		Date:        date.Format("2006/01/02"),
		HolidayDate: jsonutils.JSONTime(date),
	}
}

func (j *Holiday) Create(ctx context.Context) (int64, error) {

	return factory.DB(ctx).Insert(j)
}

func (Holiday) GetByDate(ctx context.Context, date string) (*Holiday, error) {

	var j Holiday
	if has, err := factory.DB(ctx).Where("date = ?", date).Get(&j); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &j, nil
}

func (Holiday) GetAllAfterToday(ctx context.Context, startDate, endDate time.Time) (map[string]Holiday, error) {
	var items []Holiday
	err := factory.DB(ctx).Table("holiday").Select("holiday.*").Where("holiday.holiday_date > ? AND holiday.holiday_date <= ? ", startDate, endDate).OrderBy("holiday.holiday_date").Find(&items)
	if err != nil {
		return nil, err
	}

	hMap := make(map[string]Holiday)
	for _, item := range items {
		hMap[item.Date] = item
	}

	return hMap, nil
}

func (h *Holiday) IsExist(ctx context.Context) (bool, error) {
	return factory.DB(ctx).Where("date = ?", h.Date).Exist(&Holiday{})
}
