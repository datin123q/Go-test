package domain

import (
	"context"
	"path"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
)

const NEXWAY_APPLICANT_IMAGE_FILE_TYPE = "nexway_applicant_image"

// NexwayApplicantImage Nexway eKYC申請時に提出された本人確認書類の画像（元画像）
type NexwayApplicantImage struct {
	Id               int64              `json:"id" xorm:"pk autoincr"`
	EkycRequestLogId int64              `json:"ekyc_request_log_id" xorm:"notnull"`
	ImageType        NexwayImageType    `json:"image_type" xorm:"notnull"`
	S3Key            string             `json:"s3_key" xorm:"notnull"` // S3キー (例: /nexway_applicant_image/2026-01-27/abc123.jpg)
	Created          jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated          jsonutils.JSONTime `json:"updated" xorm:"updated"`
}

func (NexwayApplicantImage) TableName() string {
	return "nexway_applicant_image"
}

func (n NexwayApplicantImage) SelectAll() string {
	return n.TableName() + ".*"
}

// S3Object interface implementation
func (n *NexwayApplicantImage) GetFileType() string {
	return NEXWAY_APPLICANT_IMAGE_FILE_TYPE
}

func (n *NexwayApplicantImage) GetKey() string {
	return n.S3Key
}

func (n *NexwayApplicantImage) SetLocation(loc string) {
}

func (n *NexwayApplicantImage) GetFileName() string {
	return path.Base(n.S3Key)
}

func (n *NexwayApplicantImage) GetCDNUrl() string {
	return s3utils.MakeCDNURL(n.GetKey())
}

func (n *NexwayApplicantImage) GetCDNSignUrl() string {
	return s3utils.MakeCDNSignURL(n.GetKey())
}

// CRUD operations
func (n *NexwayApplicantImage) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(n)
}

func (NexwayApplicantImage) GetById(ctx context.Context, id int64) (*NexwayApplicantImage, error) {
	var item NexwayApplicantImage
	has, err := factory.DB(ctx).ID(id).Get(&item)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	return &item, nil
}

func (NexwayApplicantImage) FindByEkycRequestLogId(ctx context.Context, ekycRequestLogId int64) ([]*NexwayApplicantImage, error) {
	var items []*NexwayApplicantImage
	err := factory.DB(ctx).Where("ekyc_request_log_id = ?", ekycRequestLogId).Find(&items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (n *NexwayApplicantImage) Update(ctx context.Context) error {
	_, err := factory.DB(ctx).ID(n.Id).Update(n)
	return err
}

func (n *NexwayApplicantImage) Delete(ctx context.Context) error {
	_, err := factory.DB(ctx).ID(n.Id).Delete(n)
	return err
}
