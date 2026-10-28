package controllers

import "github.com/illuminarean/energy-shares-backend-repo/domain"

type FundingRequestSearchInput struct {
	SearchInput
	QuerySearchInput
	Id            *string `query:"id" swagger:"desc(融資要求ID)"`
	Email         *string `query:"email" swagger:"desc(融資要求者メール)"`
	Phone         *string `query:"phone" swagger:"desc(融資要求者電話番号)"`
	RequestStatus *string `query:"request_status" swagger:"desc(融資要求状態)"`
	StartDate     string  `query:"start_date"`
	EndDate       string  `query:"end_date"`
}

type FundingRequestCorpNoInput struct {
	CorpNo string `json:"corp_no" query:"CorpNo" swagger:"desc(法人番号)" validate:"min=1,max=50"`
}

type FundingRequestInput struct {
	Name                string                `json:"name" form:"name" swagger:"desc(会社名)" validate:"min=1,max=200"`
	ZipCode             string                `json:"zip_code" form:"zip_code" swagger:"desc(郵便番号)"` // 郵便番号
	Address             string                `json:"address" form:"address" swagger:"desc(住所)" validate:"min=1,max=200"`
	Address2            string                `json:"address2" form:"address2" swagger:"desc(詳細住所)"`
	CorpNo              string                `json:"corp_no" form:"corp_no" swagger:"desc(法人番号)"`   // 法人番号
	Province            string                `json:"province" form:"province" swagger:"desc(県名)"` // 県名
	City                string                `json:"city" form:"city" swagger:"desc(市または区、郡)"`
	Email               string                `json:"email" form:"email" swagger:"desc(メール)" validate:"email"`                 // メールかどうか確認
	Phone               string                `json:"phone" form:"phone" swagger:"desc(電話番号(数字のみ入力))" validate:"min=1,max=200"` // 数字のみを含むかどうか確認
	EquipmentName       string                `json:"equipment_name" form:"equipment_name" swagger:"desc(設備名)"  `
	EquipmentAddress    string                `json:"equipment_address" form:"equipment_address" swagger:"desc(設備住所)" `
	EquipmentAddress2   string                `json:"equipment_address2" form:"equipment_address2" swagger:"desc(設備詳細住所)"`
	EquipmentInfo       string                `json:"equipment_info" form:"equipment_info" swagger:"desc(プロジェクト概要)"`
	EquipmentConsulting string                `json:"equipment_consulting" form:"equipment_consulting" swagger:"desc(相談内容)"`
	RequestStatus       domain.REQUEST_STATUS `json:"request_status" form:"request_status" swagger:"enum(STANDBY_REQUEST|PROCESSING_REQUEST|PASS_REQUEST|FAIL_REQUEST),default(STANDBY_REQUEST),desc(1:審査待機中 2:審査処理中 3:[適合]審査完了 4:[不適合]審査完了)"`
}

func (fri *FundingRequestInput) ToModel() (*domain.FundingRequest, error) {
	res := &domain.FundingRequest{
		Name:                fri.Name,
		ZipCode:             fri.ZipCode,
		Address:             fri.Address,
		Address2:            fri.Address2,
		CorpNo:              fri.CorpNo,
		Province:            fri.Province,
		City:                fri.City,
		Email:               fri.Email,
		Phone:               fri.Phone,
		RequestStatus:       fri.RequestStatus,
		EquipmentName:       fri.EquipmentName,
		EquipmentAddress:    fri.EquipmentAddress,
		EquipmentAddress2:   fri.EquipmentAddress2,
		EquipmentInfo:       fri.EquipmentInfo,
		EquipmentConsulting: fri.EquipmentConsulting,
	}

	return res, nil
}
