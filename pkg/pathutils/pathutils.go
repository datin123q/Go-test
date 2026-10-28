package pathutils

import (
	"os"
	"strings"
)

func GetPathOfAssets() string {
	return getRootPath() + "/assets"
}

func GetPathOfApp(appName string) string {
	return getRootPath() + "/apps/" + appName
}

func getRootPath() string {
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	patterns := []string{
		"/apps", "/assets", "/domain", "/pkg", "/service", "/test",
	}

	for _, pattern := range patterns {
		index := strings.Index(workingDir, pattern)
		if index != -1 {
			workingDir = workingDir[:index]
			break // パターンを見つけたらループを終了します。
		}
	}

	return workingDir
}
