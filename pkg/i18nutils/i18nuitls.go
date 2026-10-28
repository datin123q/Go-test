package i18nutils

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type Translator interface {
	Translate(key string, langs ...string) string
	Translates(keys []string, langs ...string) []string
}

type translator struct {
	bundle *i18n.Bundle
}

func New(path string) Translator {
	_bundle := i18n.NewBundle(language.Japanese)
	_bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	for _, lang := range []string{"en", "ja", "ko"} {
		f := fmt.Sprintf("%s/i18n/active.%v.json", path, lang)

		if _, err := os.Stat(f); err != nil {
			panic(err)
		}
		bundle.MustLoadMessageFile(f)
	}

	return &translator{
		bundle: _bundle,
	}
}

func (t translator) Translate(key string, langs ...string) string {
	language := serviceutils.GetServiceLocale()

	if len(langs) != 0 {
		language = langs[0]
	}

	localizer := i18n.NewLocalizer(bundle, language)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: key})
	if err != nil {
		return key
	}

	return msg
}

func (t translator) Translates(keys []string, langs ...string) []string {
	msgs := make([]string, 0)

	language := serviceutils.GetServiceLocale()

	if len(langs) != 0 {
		language = langs[0]
	}

	localizer := i18n.NewLocalizer(bundle, language)

	for _, key := range keys {
		msg, err := localizer.Localize(
			&i18n.LocalizeConfig{
				MessageID: key,
			},
		)

		if err != nil {
			msgs = append(msgs, key)
		} else {
			msgs = append(msgs, msg)
		}
	}

	return msgs
}

var bundle *i18n.Bundle

func init() {

	bundle = i18n.NewBundle(language.Japanese)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

}

func InitI18nFiles(path string) error {

	for _, lang := range []string{"en", "ja", "ko"} {
		f := fmt.Sprintf("%s/i18n/active.%v.json", path, lang)
		if _, err := os.Stat(f); err != nil {
			return err
		}
		bundle.MustLoadMessageFile(f)
	}

	return nil
}

func Translate(key, locale string) string {
	localizer := i18n.NewLocalizer(bundle, locale, locale)
	msg, err := localizer.Localize(
		&i18n.LocalizeConfig{
			MessageID: key,
		},
	)
	if err != nil {
		return key
	}
	return msg
}

func DefaultTranslate(key string) string {
	localizer := i18n.NewLocalizer(bundle, serviceutils.GetServiceLocale())
	msg, err := localizer.Localize(
		&i18n.LocalizeConfig{
			MessageID: key,
		},
	)
	if err != nil {
		return key
	}
	return msg
}
