package calender

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

const BASE_CALENDAR_URL = "https://www.googleapis.com/calendar/v3/calendars"
const CALENDER_REGION = "ja.japanese.official"
const JP_CALENDAR_ID_FOR_PUBLIC_HOLIDAY = "holiday@group.v.calendar.google.com"
const API_KEY = "AIzaSyA3NVVGCmVs_CgKONZd9b8WcJC5XiMiUCk"

type calendarClient struct {
	Service *calendar.Service
}

func NewCalendarClient() *calendarClient {
	ctx := context.Background()
	calendarService, err := calendar.NewService(ctx, option.WithAPIKey(API_KEY))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Error("failed to create calendar service")

		return nil
	}
	return &calendarClient{
		Service: calendarService,
	}
}

type GoogleCalendarResponse struct {
	Kind          string                `json:"kind"`
	Etag          string                `json:"etag"`
	Summary       string                `json:"summary"`
	Description   string                `json:"description"`
	Updated       string                `json:"updated"`
	TimeZone      string                `json:"timeZone"`
	AccessRole    string                `json:"accessRole"`
	NextPageToken string                `json:"nextPageToken"`
	NextSyncToken string                `json:"nextSyncToken"`
	Items         []GoogleCalenderEvent `json:"items"`
}

/*
	'   "start": {
	    "date": "2024-01-01"
	   },
	   "end": {
	    "date": "2024-01-02"
	   },
*/
type GoogleCalenderEvent struct {
	Kind        string `json:"kind"`
	Etag        string `json:"etag"`
	Id          string `json:"id"`
	Status      string `json:"status"`
	HtmlLink    string `json:"htmlLink"`
	Created     string `json:"created"`
	Updated     string `json:"updated"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
	Location    string `json:"location"`
	Start       struct {
		Date string `json:"date"`
	}
	End struct {
		Date string `json:"date"`
	}
}

type ErrorGoogleCalender struct {
}

func (c *calendarClient) GetPublicHolidays(startDate, endDate time.Time) ([]GoogleCalenderEvent, error) {
	// 祝日カレンダーから日本の祝日を取得する。
	var items []GoogleCalenderEvent
	events, err := c.Service.Events.List(CALENDER_REGION + "#" + JP_CALENDAR_ID_FOR_PUBLIC_HOLIDAY).
		TimeMin(startDate.Format(time.RFC3339)).TimeMax(endDate.Format(time.RFC3339)).Do()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Error("failed to get public holidays")

		return nil, errors.Wrap(err, "failed to get public holidays")
	}
	for _, item := range events.Items {
		items = append(items, GoogleCalenderEvent{
			Kind:        item.Kind,
			Etag:        item.Etag,
			Id:          item.Id,
			Status:      item.Status,
			HtmlLink:    item.HtmlLink,
			Created:     item.Created,
			Updated:     item.Updated,
			Summary:     item.Summary,
			Description: item.Description,
			Location:    item.Location,
			Start: struct {
				Date string `json:"date"`
			}{
				Date: item.Start.Date,
			},
			End: struct {
				Date string `json:"date"`
			}{
				Date: item.End.Date,
			},
		})
		logrus.WithFields(logrus.Fields{
			"item": item,
		}).Info("success to get public holidays")
	}

	return items, nil
}
