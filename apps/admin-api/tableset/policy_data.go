package tableset

import (
	"fmt"
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
	PolicyViewType domain.POLICY_VIEW_TYPE `yaml:"policy_view_type"`
	PolicyType     domain.POLICY_TYPE      `yaml:"policy_type"`
	ContentHtml    string                  `yaml:"content_html"`
}

var (
	// 必須規約一覧
	requiredPolicyKinds = []Policy_Data{
		{PolicyType: domain.POLICY_CAUTION, PolicyViewType: domain.POLICY_VIEW_FOOTER},
		{PolicyType: domain.POLICY_CAUTION, PolicyViewType: domain.POLICY_VIEW_JOIN},
		{PolicyType: domain.POLICY_PROTECTION, PolicyViewType: domain.POLICY_VIEW_FOOTER},
		{PolicyType: domain.POLICY_PROTECTION, PolicyViewType: domain.POLICY_VIEW_JOIN},
		{PolicyType: domain.POLICY_ANTISOCIAL_FORCES_BAGIC, PolicyViewType: domain.POLICY_VIEW_FOOTER},
		{PolicyType: domain.POLICY_NON_ANTISOCIAL_FORCES_AGREEMENT, PolicyViewType: domain.POLICY_VIEW_JOIN},
		{PolicyType: domain.POLICY_SUBSTANTIVE_RULER, PolicyViewType: domain.POLICY_VIEW_JOIN},
		{PolicyType: domain.POLICY_SOLICITATION, PolicyViewType: domain.POLICY_VIEW_FOOTER},
		{PolicyType: domain.POLICY_COMMISSION_RISK, PolicyViewType: domain.POLICY_VIEW_FOOTER},
		{PolicyType: domain.POLICY_CUSTOMER_CENTERED_WORK, PolicyViewType: domain.POLICY_VIEW_FOOTER},
		{PolicyType: domain.POLICY_ELECTRONIC_TEACHING, PolicyViewType: domain.POLICY_VIEW_JOIN},
		{PolicyType: domain.POLICY_NON_PEPS, PolicyViewType: domain.POLICY_VIEW_JOIN},
		{PolicyType: domain.POLICY_ADULT_GUARDIAN, PolicyViewType: domain.POLICY_VIEW_JOIN},
		{PolicyType: domain.POLICY_SPECIFIC_INVESTORS, PolicyViewType: domain.POLICY_VIEW_JOIN},
		{PolicyType: domain.POLICY_COOLING_OFF, PolicyViewType: domain.POLICY_VIEW_FOOTER},
		{PolicyType: domain.POLICY_CUSTOMER_HARASSMENT, PolicyViewType: domain.POLICY_VIEW_FOOTER},
	}
)

func InitPolicyTable(engine *xorm.Engine, path string) error {

	existingPolicies := make([]domain.Policy, 0, len(requiredPolicyKinds))
	if err := engine.Table("policy").Cols("policy_type", "policy_view_type").Find(&existingPolicies); err != nil {
		return err
	}

	existingKinds := make(map[Policy_Data]struct{}, len(existingPolicies))
	for _, policy := range existingPolicies {
		key := Policy_Data{
			PolicyType:     policy.PolicyType,
			PolicyViewType: policy.PolicyViewType,
		}
		existingKinds[key] = struct{}{}
	}

	if err := readPolicy(engine, path, existingKinds); err != nil {
		return err
	}

	for _, required := range requiredPolicyKinds {
		if _, ok := existingKinds[required]; !ok {
			return fmt.Errorf("required policy not prepared: %s/%s", required.PolicyType, required.PolicyViewType)
		}
	}

	return nil
}

func readPolicy(engine *xorm.Engine, path string, existingKinds map[Policy_Data]struct{}) error {
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
		key := Policy_Data{
			PolicyViewType: data.PolicyViewType,
			PolicyType:     data.PolicyType,
		}

		if _, exists := existingKinds[key]; exists {
			continue
		}

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

		existingKinds[key] = struct{}{}
	}

	return nil
}
