package domain

import (
	"context"
	"mime/multipart"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
)

const ATTACHMENT_FILE_TYPE = "attachment_file"

type Attachment struct {
	Id            int64              `json:"id" description:"첨부파일 ID" example:"262"`
	FileName      string             `json:"file_name" description:"파일이름" xorm:"UNIQUE NOT NULL" example:"cqhkh82nulri1lp6i7f0"`
	OriginalName  string             `json:"original_name" description:"원본파일이름" xorm:"NOT NULL" example:"f2cbdb972f8d68b0cd66dac681af6e9f.jpg"`
	FileSize      int64              `json:"file_size" description:"파일사이즈" xorm:"NOT NULL" example:"12243"`
	FileExtension string             `json:"file_extension" description:"파일 확장자" xorm:"NOT NULL" example:".jpg"`
	FilePath      string             `json:"file_path" description:"파일경로" xorm:"NOT NULL" example:"/attachment_file/2024-07-26/"`
	FileUrl       string             `json:"file_url" description:"첨부파일 URL" xorm:"NOT NULL" example:"https://cdn.dev.energyshares.jp/attachment_file/2024-07-26/cqhkh82nulri1lp6i7f0.jpg"`
	ResourceId    string             `json:"resource_id" description:"자원 ID (공용 ID)" xorm:"UUID NOT NULL INDEX" example:"aeee0ebc-cb8a-43bd-9eb6-94c34b8686ae"`
	Actor         ChangelogActor     `json:"actor" description:"행위자" xorm:"json"`
	Created       jsonutils.JSONTime `json:"created" description:"생성 일시" xorm:"created" example:"2024-07-26 15:56:32"`
	Updated       jsonutils.JSONTime `json:"updated" description:"수정 일시" xorm:"updated" example:"2024-07-26 15:56:32"`
	Deleted       jsonutils.JSONTime `json:"deleted" description:"삭제 일시" xorm:"deleted" example:""`
}

func (Attachment) TableName() string {
	return "attachment"
}

func (a Attachment) SelectAll() string {
	return a.TableName() + ".*"
}

func (c *Attachment) GetFileType() string {
	return ATTACHMENT_FILE_TYPE
}

func (c *Attachment) GetKey() string {
	return c.FilePath + c.FileName + c.FileExtension
}

func (c *Attachment) SetLocation(loc string) {
	c.FileUrl = loc
}

func (c *Attachment) GetFileName() string {
	return c.OriginalName
}

func (c *Attachment) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(c)
}

func (Attachment) GetById(ctx context.Context, id int64) (*Attachment, error) {
	var v Attachment
	if has, err := factory.DB(ctx).ID(id).Get(&v); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	return &v, nil
}

func (Attachment) GetAll(ctx context.Context, sortby, order string, offset, limit int) (int64, []Attachment, error) {
	q := factory.DB(ctx)
	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []Attachment
	totalCount, err := q.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}
	return totalCount, items, nil
}

func (Attachment) FindByResourceId(ctx context.Context, limit int, id string) ([]*Attachment, error) {
	q := factory.DB(ctx)

	items := make([]*Attachment, 0)
	_, err := q.Where("resource_id = ?", id).Limit(limit).FindAndCount(&items)
	if err != nil {
		return nil, err
	}

	for i := range items {
		items[i].FileUrl = items[i].GetCDNUrl()
	}

	return items, nil
}

func (c *Attachment) Update(ctx context.Context) (err error) {
	_, err = factory.DB(ctx).ID(c.Id).Update(c)
	return err
}

func (Attachment) Delete(ctx context.Context, id int64) (err error) {
	_, err = factory.DB(ctx).ID(id).Delete(&Attachment{})
	return err
}

// UploadAttachmentFileFunc は、添付をS3へ上げる。テストで差し替えるために変数にしている(MIR-1403)。
var UploadAttachmentFileFunc = s3utils.UploadAttachmentFile

func (Attachment) AddAttachments(ctx context.Context, commonid string, form *multipart.Form) ([]*Attachment, error) {
	var err error
	attachments := make([]*Attachment, 0)
	for _, value := range form.File {

		for _, file := range value {
			attachment := &Attachment{
				FileName:      s3utils.MakeS3FileName(),
				OriginalName:  file.Filename,
				FileSize:      file.Size,
				FileExtension: s3utils.GetExtension(file.Filename),
				FilePath:      s3utils.MakeS3Path(ATTACHMENT_FILE_TYPE),
				FileUrl:       "",
				ResourceId:    commonid,
			}
			attachment.FileUrl, err = UploadAttachmentFileFunc(ctx, s3utils.BUCKET_PUBLIC_NAME, attachment, file)
			if err != nil {
				return nil, err
			}
			if _, err = attachment.Create(ctx); err != nil {
				return nil, err
			}
			attachments = append(attachments, attachment)
		}
	}
	return attachments, nil
}

func (Attachment) AddAttachmentsExtended(ctx context.Context, commonid string, form *multipart.Form, actor ChangelogActor) ([]*Attachment, error) {
	var err error
	attachments := make([]*Attachment, 0)
	for _, value := range form.File {

		for _, file := range value {
			attachment := &Attachment{
				FileName:      s3utils.MakeS3FileName(),
				OriginalName:  file.Filename,
				FileSize:      file.Size,
				FileExtension: s3utils.GetExtension(file.Filename),
				FilePath:      s3utils.MakeS3Path(ATTACHMENT_FILE_TYPE),
				FileUrl:       "",
				ResourceId:    commonid,
				Actor: ChangelogActor{
					Id:    actor.Id,
					Name:  actor.Name,
					Email: actor.Email,
				},
			}
			attachment.FileUrl, err = UploadAttachmentFileFunc(ctx, s3utils.BUCKET_PUBLIC_NAME, attachment, file)
			if err != nil {
				return nil, err
			}
			if _, err = attachment.Create(ctx); err != nil {
				return nil, err
			}
			attachments = append(attachments, attachment)
		}
	}
	return attachments, nil
}

func (c *Attachment) GetCDNUrl() string {
	return s3utils.MakeCDNURL(c.GetKey())
}

func (c *Attachment) GetCDNSignUrl() string {
	return s3utils.MakeCDNSignURL(c.GetKey())
}

func (a *Attachment) CompareAttachment(b *Attachment) bool {
	if a.Id != b.Id {
		return false
	}
	if a.FileName != b.FileName {
		return false
	}
	if a.OriginalName != b.OriginalName {
		return false
	}
	if a.FileSize != b.FileSize {
		return false
	}
	if a.FileExtension != b.FileExtension {
		return false
	}
	if a.FilePath != b.FilePath {
		return false
	}
	if a.FileUrl != b.FileUrl {
		return false
	}
	if a.ResourceId != b.ResourceId {
		return false
	}
	return true
}

func (Attachment) CompareAttachmentSlice(a []Attachment, b []Attachment) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !(&a[i]).CompareAttachment(&b[i]) {
			return false
		}
	}
	return true
}
