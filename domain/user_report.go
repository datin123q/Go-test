package domain

import "github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"

type UserBalanceReport struct {
	Id            int64              `json:"id"`
	UserId        int64              `json:"user_id"`
	Title         string             `json:"title"`
	FileName      string             `json:"file_name" xorm:"UNIQUE NOT NULL"`
	OriginalName  string             `json:"original_name" xorm:"NOT NULL"`
	FileSize      int64              `json:"file_size" xorm:"NOT NULL"`
	FileExtension string             `json:"file_extension" xorm:"NOT NULL"`
	FilePath      string             `json:"file_path" xorm:"NOT NULL"`
	FileUrl       string             `json:"file_url" xorm:"NOT NULL"`
	Created       jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated       jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Deleted       jsonutils.JSONTime `json:"-" xorm:"deleted"`
}

type UserDividendReport struct {
	Id            int64              `json:"id"`
	UserId        int64              `json:"user_id"`
	Title         string             `json:"title"`
	FileName      string             `json:"file_name" xorm:"UNIQUE NOT NULL"`
	OriginalName  string             `json:"original_name" xorm:"NOT NULL"`
	FileSize      int64              `json:"file_size" xorm:"NOT NULL"`
	FileExtension string             `json:"file_extension" xorm:"NOT NULL"`
	FilePath      string             `json:"file_path" xorm:"NOT NULL"`
	FileUrl       string             `json:"file_url" xorm:"NOT NULL"`
	Created       jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated       jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Deleted       jsonutils.JSONTime `json:"-" xorm:"deleted"`
}
