package formutils

import (
	"mime/multipart"
	"strings"
)

func FromToUrls(form multipart.Form, name string) []string {
	images := make([]string, 0)
	for key, values := range form.Value {
		if strings.Contains(key, name) {
			images = append(images, values[0])
		}
	}
	return images
}
