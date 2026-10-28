package domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type MyNumberUsecase interface {
	Create(ctx context.Context, request MyNumberCreateRequest) (MyNumberCreateResponse, error)
	Get(ctx context.Context, request MyNumberGetRequest) (MyNumberGetResponse, error)
}

type MyNumberCreateRequest struct {
	JWTUser      jwthelper.User `json:"-" swaggerignore:"true"`
	MyNumber     string         `json:"my_number" form:"my_number" swagger:"desc(マイナンバー)" validate:"required"`
	FrontFileUrl string         `json:"front_file_url" form:"front_file" swagger:"desc(マイナンバー表面)" validate:"required"`
	BackFileUrl  string         `json:"back_file_url" form:"back_file" swagger:"desc(マイナンバー裏面)" validate:"required"`
}

type MyNumberCreateResponse struct {
	MyNumberBody `json:"my_number" swagger:"desc(マイナンバー)"`
}

type MyNumberGetRequest struct {
	JWTUser jwthelper.User `json:"-" swaggerignore:"true"`
}

type MyNumberGetResponse struct {
	*MyNumberBody `json:"my_number" swagger:"desc(マイナンバー)"`
}

type MyNumberBody struct {
	Id        int64                   `json:"id" swagger:"desc(My Number ID)"`
	MyNumber  string                  `json:"my_number" swagger:"desc(マイナンバー)"`
	FrontFile *UserFileBody           `json:"front_file" swagger:"desc(マイナンバー表面)"`
	BackFile  *UserFileBody           `json:"back_file" swagger:"desc(マイナンバー裏面)"`
	Status    domain.MY_NUMBER_STATUS `json:"status" swagger:"desc(マイナンバーステータス)"`
}

func (m MyNumberBody) FromMyNumber(myNumber domain.MyNumber) *MyNumberBody {
	if myNumber.IsEmpty() {
		return &MyNumberBody{
			Status: domain.DISAPPROVED,
		}
	}

	return &MyNumberBody{
		Id:        myNumber.Id,
		MyNumber:  myNumber.MyNumber,
		FrontFile: UserFileBody{}.FromUserFile(*myNumber.FrontFile),
		BackFile:  UserFileBody{}.FromUserFile(*myNumber.BackFile),
		Status:    myNumber.Status,
	}
}

type UserFileBody struct {
	Id            int64                 `json:"id" swagger:"desc(ID)"`
	OriginalName  string                `json:"original_name" swagger:"desc(オリジナルファイル名)"`
	FileType      domain.USER_FILE_TYPE `json:"file_type" swagger:"desc(添付ファイルタイプ、共通コード [USER_FILE_TYPE])"`
	FileSize      int64                 `json:"file_size" swagger:"desc(ファイルサイズ)"`
	FileExtension string                `json:"file_extension" swagger:"desc(ファイル拡張子)"`
	FilePath      string                `json:"file_path" swagger:"desc(ファイルパス)"`
	FileUrl       string                `json:"file_url" swagger:"desc(URL)"`
	OwnerId       int64                 `json:"owner_id" swagger:"desc(所有者ID)"`
	UpdateId      int64                 `json:"update_id" swagger:"desc(最新更新ユーザー)"`
	Created       jsonutils.JSONTime    `json:"created" swagger:"desc(作成日)"`
	Updated       jsonutils.JSONTime    `json:"updated" swagger:"desc(更新日)"`
}

func (m UserFileBody) FromUserFile(file domain.UserFile) *UserFileBody {
	if file.IsEmpty() {
		return nil
	}

	return &UserFileBody{
		Id:            file.Id,
		OriginalName:  file.OriginalName,
		FileType:      file.FileType,
		FileSize:      file.FileSize,
		FileExtension: file.FileExtension,
		FilePath:      file.FilePath,
		FileUrl:       file.FileUrl,
		OwnerId:       file.OwnerId,
		UpdateId:      file.UpdateId,
		Created:       jsonutils.JSONTime(file.Created),
		Updated:       jsonutils.JSONTime(file.Updated),
	}
}
