package tableset

import (
	"io/ioutil"

	"github.com/gofrs/uuid"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"gopkg.in/yaml.v2"
	"xorm.io/xorm"
)

type Policy_File struct {
	PolicyData []Policy_Data `yaml:"policy_data"`
}

type Policy_Data struct {
	PolicyViewType string `yaml:"policy_view_type"`
	PolicyType     string `yaml:"policy_type"`
	ContentHtml    string `yaml:"content_html"`
}

func InitPolicyTable(engine *xorm.Engine, path string) error {

	count, err := engine.Table("policy").Where("id > 0").Count(domain.Policy{})
	if err != nil {
		return err
	} else if count == 0 {

		if err := readPolicy(engine, path); err != nil {
			return err
		}
	}

	return nil
}

func readPolicy(engine *xorm.Engine, path string) error {
	filepath := path + "/tabledata/policy.yml"

	pf := &Policy_File{}
	if config, err := ioutil.ReadFile(filepath); err != nil {
		return err
	} else {
		if err := yaml.Unmarshal(config, &pf); err != nil {
			return err
		}
	}

	for _, data := range pf.PolicyData {
		p := domain.Policy{
			PolicyViewType: domain.POLICY_VIEW_TYPE(data.PolicyViewType),
			PolicyType:     domain.POLICY_TYPE(data.PolicyType),
			ContentHtml:    data.ContentHtml,
		}

		id, err := uuid.NewV4()
		if err != nil {
			return err
		}
		p.CommonId = id.String()

		if _, err := engine.Table("policy").Insert(p); err != nil {
			return err
		}
	}

	return nil
}
