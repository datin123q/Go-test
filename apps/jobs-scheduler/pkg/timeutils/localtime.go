package timeutils

import "time"

const LOCAL_ZONE = "Asia/Tokyo"

func LocalZoneTime() time.Time {
	jst, _ := time.LoadLocation(LOCAL_ZONE)
	return time.Now().In(jst)
}
