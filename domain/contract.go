package domain

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/contract"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
)

type CONTRACT_STATUS string

const (
	CONTRACT_STATUS_CODE                 = "CONTRACT_STATUS"
	CONTRACT_NONE        CONTRACT_STATUS = "CONTRACT_NONE"
	CONTRACT_PROCESSING  CONTRACT_STATUS = "CONTRACT_PROCESSING" // 계약 처리중
	CONTRACT_SIGNED      CONTRACT_STATUS = "CONTRACT_SIGNED"     // 계약 완료
)

// TODO : 테스트 코드 작성
func (Contract) GetContractStatusGroup(loc string) commoncode.CodeGroup {

	codes := make([]commoncode.Code, 0, 0)

	codes = append(codes, commoncode.Code{Code: string(CONTRACT_NONE), Name: i18nutils.Translate("commoncode.CONTRACT_STATUS.CONTRACT_NONE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(CONTRACT_PROCESSING), Name: i18nutils.Translate("commoncode.CONTRACT_STATUS.CONTRACT_PROCESSING", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(CONTRACT_SIGNED), Name: i18nutils.Translate("commoncode.CONTRACT_STATUS.CONTRACT_SIGNED", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: CONTRACT_STATUS_CODE,
		Codes:     codes,
	}
}

type Contracts []Contract

func (c Contracts) IsEmpty() bool {
	return len(c) == 0
}

func (c Contracts) First() Contract {
	if c.IsEmpty() {
		return Contract{}
	}
	return c[0]

}

type Contract struct {
	Id                   int64              `json:"id" xorm:"pk autoincr"`
	ShareId              string             `json:"share_id" xorm:"index"`
	ShareQty             int64              `json:"share_qty"`
	ShareAmount          int64              `json:"share_amount"`
	Status               CONTRACT_STATUS    `json:"status"`
	ProjectId            int64              `json:"project_id"`
	UserName             string             `json:"user_name"`
	ManagerName          string             `json:"manager_name"`
	Address              string             `json:"address"`
	ContractAfterPageUrl string             `json:"contract_after_page_url"`
	InvestDate           jsonutils.JSONTime `json:"invest_date"`
	Created              jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated              jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Deleted              jsonutils.JSONTime `json:"deleted" xorm:"deleted"`
}

func (c Contract) TableName() string {
	return "contract"
}

func (c Contract) SelectAll() string {
	return c.TableName() + ".*"
}

func (c *Contract) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(c)
}

func (Contract) GetByShareId(ctx context.Context, sortby, order, shareId string, offset, limit int) (int64, []Contract, error) {
	q := factory.DB(ctx)
	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []Contract
	totalCount, err := q.Where("share_id = ?", shareId).Limit(limit, offset).FindAndCount(&items)
	if items == nil {
		items = make([]Contract, 0, 0)
	}
	if err != nil {
		return 0, nil, err
	}
	return totalCount, items, nil
}

func (Contract) GetById(ctx context.Context, id int64) (*Contract, error) {

	var c Contract
	if has, err := factory.DB(ctx).ID(id).Get(&c); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	return &c, nil
}

func (c *Contract) Update(ctx context.Context) (err error) {

	_, err = factory.DB(ctx).ID(c.Id).Update(c)
	return
}

func (Contract) Delete(ctx context.Context, id int64) (err error) {

	_, err = factory.DB(ctx).ID(id).Delete(&Contract{})
	return
}

func (c *Contract) UpdateContractAfterPage(ctx context.Context) error {

	project, err := Project{}.GetById(ctx, c.ProjectId)
	if err != nil {
		return errors.Wrap(err, "fail project get by id")
	}
	if project.ContractAfterPage == nil {
		logrus.WithFields(
			logrus.Fields{
				"project_id": c.ProjectId,
				"share_id":   c.ShareId,
				"user_id":    c.Status,
				"time":       jsonutils.JSONTime(time.Now()).ToString(),
			}).Error("fail get contract template file")
		return nil
	}
	body, err := GetContractTemplateFile(project.ContractAfterPage.FileUrl)
	if err != nil {
		return errors.Wrap(err, "fail get contract template file")
	}

	buf, err := contract.FillFields(body, modelhelper.ConvertNumberToPrice(strconv.FormatInt(project.SharePrice, 10)), modelhelper.ConvertNumberToPrice(strconv.FormatInt(c.ShareAmount, 10)), modelhelper.ConvertNumberToPrice(strconv.FormatInt(c.ShareQty, 10)), c.UserName, project.Spv.Name, project.Name, c.Address, c.ManagerName, c.InvestDate)

	if err != nil {
		return errors.Wrap(err, "fail contract fill fields")
	}
	if buf == nil {
		return errors.New("Contract PDF not created!!!")
	}

	// 업로드
	fileName := s3utils.MakeS3FileName()
	originalName := c.ShareId + "_" + c.Created.Time().Format("20060102150405") + ".pdf"

	shareFile := &ShareFile{
		ShareId:       c.ShareId,
		FileName:      fileName,
		OriginalName:  originalName,
		FileSize:      int64(buf.Len()),
		FileExtension: ".pdf",
		FilePath:      s3utils.MakeS3Path(SHARE_FILE_TYPE),
	}
	loc, err := s3utils.UploadBuffFile(ctx, s3utils.BUCKET_PRIVATE_NAME, shareFile, buf)
	if err != nil {
		return errors.Wrap(err, "fail upload file")
	}
	shareFile.SetLocation(loc)
	_, err = shareFile.Create(ctx)
	if err != nil {
		return errors.Wrap(err, "fail create share file")
	}

	c.Status = CONTRACT_SIGNED
	c.ContractAfterPageUrl = shareFile.GetCDNUrl()

	if err := c.Update(ctx); err != nil {
		return errors.Wrap(err, "fail update contract")
	}

	return nil
}

func GetContractTemplateFile(fileUrl string) ([]byte, error) {
	resp, err := http.Get(fileUrl)
	if err != nil {
		return nil, errors.Wrap(err, "fail contract file http.Get")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "fail contract file ioutil.ReadAll")
	}
	return body, nil
}
