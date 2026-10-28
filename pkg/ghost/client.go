package ghost

import (
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type RespError struct {
	Error string `json:"error"`
}

func RequestTopPosts(URL, key string) (Posts, error) {
	const TOP_POSTS_URL = "/ghost/api/v3/content/posts"
	client := resty.New()
	params := map[string]string{}
	params["key"] = key
	params["limit"] = "10"
	params["filter"] = "visibility:public"

	var respBody Posts
	var respError RespError

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetError(&respError).
		SetQueryParams(params).
		SetResult(&respBody).
		Get(URL + TOP_POSTS_URL)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"resp":  resp,
		}).Error("fail post http request")
		return respBody, errors.Wrap(err, "fail post http request")
	}
	return respBody, nil
}
