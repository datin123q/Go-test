package tableset

import (
	"io/ioutil"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"gopkg.in/yaml.v2"
	"xorm.io/xorm"
)

type User_Roles_File struct {
	UserRolesData []User_Roles_Data `yaml:"user_roles_data"`
}

type User_Roles_Data struct {
	RoleCode string `yaml:"role_code"`
	RoleName string `yaml:"role_name"`
}

func InitUserRolesTable(engine *xorm.Engine, path string) error {

	count, err := engine.Table("user_roles").Count(domain.UserRoles{})
	if err != nil {
		return err
	} else if count == 0 {

		if err := readUserRoles(engine, path); err != nil {
			return err
		}
	}

	return nil
}

func readUserRoles(engine *xorm.Engine, path string) error {

	filepath := path + "/tabledata/user_roles.yml"

	urf := &User_Roles_File{}
	if config, err := ioutil.ReadFile(filepath); err != nil {
		return err
	} else {
		if err := yaml.Unmarshal(config, &urf); err != nil {
			return err
		}
	}

	for _, data := range urf.UserRolesData {
		ur := domain.UserRoles{
			RoleCode: domain.ROLE_TYPE(data.RoleCode),
			RoleName: domain.ROLE_TYPE(data.RoleName),
		}

		if _, err := engine.Table("user_roles").Insert(ur); err != nil {
			return err
		}
	}

	return nil
}
