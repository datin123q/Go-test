package genhelper

type CodeDescription struct {
	Ja string `yaml:"ja"`
	Ko string `yaml:"ko"`
	En string `yaml:"en"`
}

type CodeItem struct {
	Code         string          `yaml:"code"`
	Descriptions CodeDescription `yaml:"descriptions"`
}

type GroupCodes struct {
	Codes []CodeItem `yaml:"codes"`
}

type DomainConfig map[string]map[string]GroupCodes
