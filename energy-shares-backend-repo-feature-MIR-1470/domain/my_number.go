package domain

import (
	"context"
	"strconv"

	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/cipher"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
)

type MY_NUMBER_STATUS string

const (
	MY_NUMBER_STATUS_CODE                  = "MY_NUMBER_STATUS"
	APPROVE               MY_NUMBER_STATUS = "APPROVE"      // 承認
	DISAPPROVED           MY_NUMBER_STATUS = "DISAPPROVED"  // 非承認
	WAIT_APPROVE          MY_NUMBER_STATUS = "WAIT_APPROVE" // 承認待機中
	UNREGISTERED          MY_NUMBER_STATUS = "UNREGISTERED" // 未登録
)

type MyNumber struct {
	Id           int64              `json:"id" swagger:"desc(ID)"`
	CommonId     string             `json:"common_id" xorm:"notnull UUID" swagger:"desc(공용 ID)"`
	UserId       int64              `json:"user_id" xorm:"notnull" swagger:"desc(유저 ID)"`
	Status       MY_NUMBER_STATUS   `json:"status" xorm:"notnull" swagger:"desc(마이넘버 승인 상태 : 공통 코드 MY_NUMBER_STATUS)"`
	ManagerName  string             `json:"manager_name" swagger:"desc(담당자 명)"`
	ApprovalDate jsonutils.JSONTime `json:"approval_date" swagger:"desc(승인 날짜)"`
	MyNumber     string             `json:"my_number" xorm:"-" swagger:"desc(마이넘버 조회용)"`
	FrontFile    *UserFile          `json:"front_file" xorm:"-"`
	BackFile     *UserFile          `json:"back_file" xorm:"-"`
	Created      jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated      jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Deleted      jsonutils.JSONTime `json:"deleted" xorm:"deleted"`
}

func (m MyNumber) IsEmpty() bool {
	return m.Id == 0
}

func (MyNumber) TableName() string {
	return "my_number"
}

func (m MyNumber) SelectAll() string {
	return m.TableName() + ".*"
}

func (MyNumber) GetMyNumberStatusGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(APPROVE), Name: i18nutils.Translate("commoncode.MY_NUMBER_STATUS_CODE.APPROVE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(DISAPPROVED), Name: i18nutils.Translate("commoncode.MY_NUMBER_STATUS_CODE.DISAPPROVED", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(WAIT_APPROVE), Name: i18nutils.Translate("commoncode.MY_NUMBER_STATUS_CODE.WAIT_APPROVE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(UNREGISTERED), Name: i18nutils.Translate("commoncode.MY_NUMBER_STATUS_CODE.UNREGISTERED", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: MY_NUMBER_STATUS_CODE,
		Codes:     codes,
	}
}

func (myNumber *MyNumber) EncryptMyNumber() (string, error) {

	crypto, err := cipher.NewNiceCrypto(CIPHER_KEY, CIPHER_IVKEY)
	if err != nil {
		return "", errors.Wrap(err, "failed create cipher crypto")
	}

	encryptMyNumber, err := crypto.Encrypt(myNumber.MyNumber)
	if err != nil {
		return "", errors.Wrap(err, "failed encrypt MyNumber")
	}

	myNumber.MyNumber = encryptMyNumber

	return encryptMyNumber, nil
}

func (MyNumber) GetByUserId(ctx context.Context, userId int64) (*MyNumber, error) {
	var myNumber MyNumber
	var err error

	if _, err = factory.DB(ctx).Where("user_id = ?", userId).Get(&myNumber); err != nil {
		return nil, err
	}

	if myNumber.Id != 0 {
		myNumber.MyNumber, err = LookUpMyNumber{}.GetDecryptedByMyNumberIdIncludeDeleted(ctx, myNumber.Id)
		if err != nil {
			return nil, err
		}

		myNumber.FrontFile, err = UserFile{}.GetOneMyNumberFile(ctx, userId, MY_NUMBER_FRONT, strconv.Itoa(int(myNumber.Id)))
		if err != nil {
			return nil, err
		}
		myNumber.BackFile, err = UserFile{}.GetOneMyNumberFile(ctx, userId, MY_NUMBER_BACK, strconv.Itoa(int(myNumber.Id)))
		if err != nil {
			return nil, err
		}
	} else {
		return &MyNumber{
			Status: DISAPPROVED,
		}, nil
	}

	if err != nil {
		return nil, err
	}

	return &myNumber, nil
}

func (MyNumber) CountWaitApprove(ctx context.Context) (int64, error) {
	var myNumber MyNumber
	var count int64
	var err error
	if count, err = factory.DB(ctx).Table("my_number").
		Join("INNER", "user", "my_number.user_id = user.id").
		Where("status = ? AND user.user_type = ? AND user.audit_status = ?", WAIT_APPROVE, PRIVATE, COMPLETED_AUDIT).
		Count(&myNumber); err != nil {
		return 0, err
	}

	return count, nil
}

type MyNumberSearchOption struct {
	Status            *string
	UserId            *string
	MyNumberIdList    []int64
	ApprovalStartDate string
	ApprovalEndDate   string
	RequestStartDate  string
	RequestEndDate    string
}

func (MyNumber) GetAll(ctx context.Context, option *MyNumberSearchOption, sortBy, order string, offset, limit int) (int64, []MyNumber, error) {
	q := factory.DB(ctx).Unscoped().Table("my_number").
		Select("my_number.id, my_number.status, my_number.common_id, my_number.status, my_number.manager_name, my_number.approval_date, my_number.created, user.id as user_id").
		Join("LEFT", "user", "my_number.user_id = user.id").
		Where("user.user_type = ? AND user.audit_status != ?", PRIVATE, INVALIDITY_AUDIT)

	if err := modelhelper.SetSortOrder(q, sortBy, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if len(option.MyNumberIdList) > 0 {
		q.In("my_number.id", option.MyNumberIdList)
	}
	if err := modelhelper.SetBetweenDate(q, "my_number.approval_date", option.ApprovalStartDate, option.ApprovalEndDate); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SetBetweenDate(q, "my_number.created", option.RequestStartDate, option.RequestEndDate); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(q, "my_number", "status", option.Status); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(q, "my_number", "user_id", option.UserId); err != nil {
		factory.Logger(ctx).Error(err)
	}

	items := make([]MyNumber, 0)
	var totalCount int64
	totalCount, err := q.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}
	for i, v := range items {
		if v.Id == 0 {
			v.Status = UNREGISTERED
		} else {
			v.MyNumber, err = LookUpMyNumber{}.GetDecryptedByMyNumberIdIncludeDeleted(ctx, v.Id)
			if err != nil {
				return 0, nil, err
			}
			v.FrontFile, err = UserFile{}.GetOneMyNumberFile(ctx, v.UserId, MY_NUMBER_FRONT, strconv.Itoa(int(v.Id)))
			if err != nil {
				return 0, nil, err
			}
			v.BackFile, err = UserFile{}.GetOneMyNumberFile(ctx, v.UserId, MY_NUMBER_BACK, strconv.Itoa(int(v.Id)))
			if err != nil {
				return 0, nil, err
			}
		}

		items[i] = v
	}

	return totalCount, items, nil
}

func (MyNumber) GetById(ctx context.Context, id int64) (*MyNumber, error) {
	var myNumber MyNumber
	has, err := factory.DB(ctx).ID(id).Get(&myNumber)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	return &myNumber, nil
}

func (myNumber *MyNumber) Update(ctx context.Context) (err error) {
	_, err = factory.DB(ctx).ID(myNumber.Id).Update(myNumber)
	if err != nil {
		return err
	}
	return
}

func (MyNumber) CheckMyNumberStatusByUserId(ctx context.Context, userId int64) (bool, error) {
	// NOTE: マイナンバー登録を復活させるまでtrue固定で返す
	// マイナンバー登録スキップの場合、マイナンバー未登録による制限は全て外すため。
	return true, nil

	// var myNumber MyNumber
	// if has, err := factory.DB(ctx).Table("my_number").Where("user_id = ?", userId).Get(&myNumber); err != nil {
	// 	return false, err
	// } else if !has {
	// 	return false, nil
	// }
	// if myNumber.Status != APPROVE {
	// 	return false, nil
	// }
	// return true, nil
}

type LookUpMyNumber struct {
	Id         int64              `json:"id" swagger:"desc(ID)"`
	MyNumberId int64              `json:"my_number_id" xorm:"notnull" swagger:"desc(energy shares 마이넘버 ID)"`
	MyNumber   string             `json:"my_number" xorm:"varchar(255) notnull" swagger:"desc(마이넘버)"`
	Created    jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated    jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Deleted    jsonutils.JSONTime `json:"deleted" xorm:"deleted"`
}

func (LookUpMyNumber) TableName() string {
	return "look_up_my_number"
}

func (l LookUpMyNumber) SelectAll() string {
	return l.TableName() + ".*"
}

func (LookUpMyNumber) GetDecryptedByMyNumberId(ctx context.Context, id int64) (string, error) {
	var lookUpMyNumber LookUpMyNumber
	if _, err := factory.MyNumberDB(ctx).Where("my_number_id = ?", id).Get(&lookUpMyNumber); err != nil {
		return "", err
	}
	crypto, err := cipher.NewNiceCrypto(CIPHER_KEY, CIPHER_IVKEY)
	if err != nil {
		return "", errors.Wrap(err, "failed create cipher crypto")
	}
	decryptMyNumber, err := crypto.Decrypt(lookUpMyNumber.MyNumber)
	if err != nil {
		return "", errors.Wrap(err, "failed decrypt MyNumber")
	}

	return decryptMyNumber, nil
}

func (LookUpMyNumber) GetDecryptedByMyNumberIdIncludeDeleted(ctx context.Context, id int64) (string, error) {
	var lookUpMyNumber LookUpMyNumber

	q := factory.MyNumberDB(ctx)
	if _, err := q.SQL("SELECT * FROM look_up_my_number WHERE my_number_id = ?", id).Get(&lookUpMyNumber); err != nil {
		return "", err
	}

	crypto, err := cipher.NewNiceCrypto(CIPHER_KEY, CIPHER_IVKEY)
	if err != nil {
		return "", errors.Wrap(err, "failed create cipher crypto")
	}
	decryptMyNumber, err := crypto.Decrypt(lookUpMyNumber.MyNumber)
	if err != nil {
		return "", errors.Wrap(err, "failed decrypt MyNumber")
	}

	return decryptMyNumber, nil
}

func (LookUpMyNumber) GetAllByMyNumberIncludeDeleted(ctx context.Context, myNumber string) ([]LookUpMyNumber, error) {
	var lookUpMyNumbers []LookUpMyNumber

	err := factory.MyNumberDB(ctx).
		SQL("SELECT * FROM look_up_my_number WHERE my_number = ?", myNumber).
		Find(&lookUpMyNumbers)
	if err != nil {
		return nil, err
	}

	return lookUpMyNumbers, nil
}

func (LookUpMyNumber) GetOthersWithSameMyNumber(ctx context.Context, id int64) ([]LookUpMyNumber, error) {
	var lookUpMyNumbers []LookUpMyNumber

	err := factory.MyNumberDB(ctx).
		Where("my_number_id != ?", id).
		And("my_number = (SELECT my_number FROM look_up_my_number WHERE my_number_id = ?)", id).
		Find(&lookUpMyNumbers)
	if err != nil {
		return nil, err
	}

	return lookUpMyNumbers, nil
}

func (LookUpMyNumber) DeleteByMyNumberId(ctx context.Context, id int64) error {
	_, err := factory.MyNumberDB(ctx).Where("my_number_id = ?", id).Delete(&LookUpMyNumber{})
	if err != nil {
		return err
	}
	return nil
}
