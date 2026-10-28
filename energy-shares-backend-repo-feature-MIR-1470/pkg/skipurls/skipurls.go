package skipurls

import (
	"os"
	"path"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/pkg/errors"
)

type RoleSkipUrl struct {
	Path   string `yaml:"path"`
	Method string `yaml:"method"`
}

func JwtSkipUrlsFromYaml(filepath string, fileName string) (map[string][]string, error) {
	filePath := path.Join(filepath, fileName)
	ymlData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var jwtSkipUrl map[string][]struct {
		Path   string `yaml:"path"`
		Method string `yaml:"method"`
	}
	if err := yaml.Unmarshal(ymlData, &jwtSkipUrl); err != nil {
		return nil, err
	}
	jwtSkipUrlList, exist := jwtSkipUrl["jwt_skip_data"]
	if !exist {
		return nil, errors.New("jwt_skip_data is not exist")
	}

	skipMap := make(map[string][]string, len(jwtSkipUrlList))
	for _, skipUrl := range jwtSkipUrlList {
		methodList := strings.Split(skipUrl.Method, ",")
		skipMap[skipUrl.Path] = methodList
	}
	return skipMap, nil
}
