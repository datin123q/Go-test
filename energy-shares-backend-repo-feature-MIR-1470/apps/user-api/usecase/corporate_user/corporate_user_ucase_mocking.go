package corporate_user

import (
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"

	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
)

func mustParseCorporateUserDate(value string) jsonutils.JSONTime {
	t, err := jsonutils.ParseJSONTimeJST(time.DateOnly, value)
	if err != nil {
		panic(err)
	}
	return t
}

var (
	mockUserFile1 = domain.UserFile{
		FileName:      "cou7922nulrn470ulh4g",
		FileExtension: ".jpeg",
		FilePath:      "/user_file/2024-05-09/",
		FileUrl:       "https://cdn.dev.energyshares.jp/user_file/2024-05-09/cou7922nulrn470ulh4g.jpeg",
	}

	mockCertifiedCopyFile = domain.UserFile{
		Id:            0,
		UserId:        1000,
		FileName:      "kyc_certified_copy",
		OriginalName:  "test.jpeg",
		FileType:      domain.CERTIFIED_COPY,
		FileSize:      1000,
		FileExtension: ".jpeg",
		FilePath:      "/user_file/2024-05-09/",
		FileUrl:       "https://cdn.dev.energyshares.jp/user_file/2024-05-09/kyc_certified_copy.jpeg",
		OwnerId:       1000,
		UpdateId:      1000,
		UserType:      domain.CORPORATE,
		Etc:           "",
		Created:       jsonutils.JSONTime{},
		Updated:       jsonutils.JSONTime{},
		Deleted:       jsonutils.JSONTime{},
	}

	mockIdentification1FrontFile = domain.UserFile{
		Id:            0,
		UserId:        1000,
		FileName:      "kyc_identification1_front",
		OriginalName:  "test.jpeg",
		FileType:      domain.IDENTIFICATION1_FRONT,
		FileSize:      1000,
		FileExtension: ".jpeg",
		FilePath:      "/user_file/2024-05-09/",
		FileUrl:       "https://cdn.dev.energyshares.jp/user_file/2024-05-09/kyc_identification1_front.jpeg",
		OwnerId:       1000,
		UpdateId:      1000,
		UserType:      domain.CORPORATE,
	}

	mockIdentification2FrontFile = domain.UserFile{
		Id:            0,
		UserId:        1000,
		FileName:      "kyc_identification2_front",
		OriginalName:  "test.jpeg",
		FileType:      domain.IDENTIFICATION2_FRONT,
		FileSize:      1000,
		FileExtension: ".jpeg",
		FilePath:      "/user_file/2024-05-09/",
		FileUrl:       "https://cdn.dev.energyshares.jp/user_file/2024-05-09/kyc_identification2_front.jpeg",
		OwnerId:       1000,
		UpdateId:      1000,
		UserType:      domain.CORPORATE,
	}

	mockPOAFile = domain.UserFile{
		Id:            0,
		UserId:        1000,
		FileName:      "kyc_poa",
		OriginalName:  "test.jpeg",
		FileType:      domain.POA,
		FileSize:      1000,
		FileExtension: ".jpeg",
		FilePath:      "/user_file/2024-05-09/",
		FileUrl:       "https://cdn.dev.energyshares.jp/user_file/2024-05-09/kyc_poa.jpeg",
		OwnerId:       1000,
		UpdateId:      1000,
		UserType:      domain.CORPORATE,
	}

	mockIdentification1BackFile = domain.UserFile{
		Id:            0,
		UserId:        1000,
		FileName:      "kyc_identification1_back",
		OriginalName:  "test.jpeg",
		FileType:      domain.IDENTIFICATION1_BACK,
		FileSize:      1000,
		FileExtension: ".jpeg",
		FilePath:      "/user_file/2024-05-09/",
		FileUrl:       "https://cdn.dev.energyshares.jp/user_file/2024-05-09/kyc_identification1_back.jpeg",
		OwnerId:       1000,
		UpdateId:      1000,
		UserType:      domain.CORPORATE,
	}

	mockIdentification2BAckFile = domain.UserFile{
		Id:            0,
		UserId:        1000,
		FileName:      "kyc_identification2_back",
		OriginalName:  "test.jpeg",
		FileType:      domain.IDENTIFICATION2_BACK,
		FileSize:      1000,
		FileExtension: ".jpeg",
		FilePath:      "/user_file/2024-05-09/",
		FileUrl:       "https://cdn.dev.energyshares.jp/user_file/2024-05-09/kyc_identification2_back.jpeg",
		OwnerId:       1000,
		UpdateId:      1000,
		UserType:      domain.CORPORATE,
	}

	mockAOIFile = domain.UserFile{
		Id:            0,
		UserId:        1000,
		FileName:      "cou7922nulrn470ulh4g",
		OriginalName:  "test.jpeg",
		FileType:      domain.AOI,
		FileSize:      1000,
		FileExtension: ".jpeg",
		FilePath:      "/user_file/2024-05-09/",
		FileUrl:       "https://cdn.dev.energyshares.jp/user_file/2024-05-09/cou7922nulrn470ulh4g.jpeg",
		OwnerId:       1000,
		UpdateId:      1000,
		UserType:      domain.CORPORATE,
	}

	mockShareholdersFile = domain.UserFile{
		Id:            0,
		UserId:        1000,
		FileName:      "cou7922nulrn470ulh4g",
		OriginalName:  "test.jpeg",
		FileType:      domain.SHAREHOLDERS,
		FileSize:      1000,
		FileExtension: ".jpeg",
		FilePath:      "/user_file/2024-05-09/",
		FileUrl:       "https://cdn.dev.energyshares.jp/user_file/2024-05-09/cou7922nulrn470ulh4g.jpeg",
		OwnerId:       1000,
		UpdateId:      1000,
		UserType:      domain.CORPORATE,
	}

	mockOthersFile = domain.UserFile{
		Id:            0,
		UserId:        1000,
		FileName:      "cou7922nulrn470ulh4g",
		OriginalName:  "test.jpeg",
		FileType:      domain.USER_FILE_OTHERS,
		FileSize:      1000,
		FileExtension: ".jpeg",
		FilePath:      "/user_file/2024-05-09/",
		FileUrl:       "https://cdn.dev.energyshares.jp/user_file/2024-05-09/cou7922nulrn470ulh4g.jpeg",
		OwnerId:       1000,
		UpdateId:      1000,
		UserType:      domain.CORPORATE,
	}

	mockUserFilesForVerifyWithKyc = []domain.UserFile{
		mockCertifiedCopyFile, mockIdentification1FrontFile, mockIdentification2FrontFile, mockPOAFile, mockIdentification1BackFile, mockIdentification2BAckFile,
	}

	mockUser1 = domain.User{
		Email:       "gildong@email.com",
		Id:          1000,
		AuditStatus: domain.HOLD_AUDIT,
	}

	mockUser2 = domain.User{
		Id:          1001,
		Name:        "イルミナリアン",
		Email:       "gildong@email.com",
		MobilePhone: "01012345678",
		ZipCode:     "1234567",
		Address:     "詳細住所",
		Address2:    lo.ToPtr("詳細住所2"),
		Province:    "都道府県",
		City:        "市",
	}

	mockUser3 = domain.User{
		Id:                1001,
		Name:              "イルミナリアン",
		Email:             "gildong@email.com",
		MobilePhone:       "01012345678",
		ZipCode:           "1234567",
		Address:           "詳細住所",
		Address2:          lo.ToPtr("詳細住所2"),
		Province:          "都道府県",
		City:              "市",
		BankName:          "イルミ銀行",
		BranchName:        "イルミ支店",
		BankCode:          lo.ToPtr("123"),
		BranchCode:        lo.ToPtr("456"),
		DepositType:       domain.ORDINARY_ACCOUNT,
		BankAccount:       "1234567",
		Depositor:         "1001",
		InvestmentPurpose: domain.RAISING_FUTURE_MONEY,
		HasExperience:     lo.ToPtr(true),
		IsSurplusFund:     lo.ToPtr(true),
		InvestorType:      domain.CORPORATE_INVESTOR,
		IsNewsletter:      true,
	}

	mockCorporateProfile1 = domain.CorporateProfile{
		UserId:          mockUser1.Id,
		ManagerBirthday: mustParseCorporateUserDate("2000-01-01"),
		EmployeeCount:   "100",
		CapitalAmount:   "1000000",
		Sales:           "500000",
	}

	mockCorporateProfile2 = domain.CorporateProfile{
		Id:                           2001,
		UserId:                       mockUser1.Id,
		CorpNo:                       "1234567890123",
		CorpType:                     "IT企業",
		RepresentativeFamilyName:     "ホン",
		RepresentativeFirstName:      "ギルドン",
		RepresentativeFamilyNameKana: "ホン",
		RepresentativeFirstNameKana:  "ギルドン",
		RepresentativeMobilePhone:    "01012345678",
		RepresentativeEmail:          "gildong@email.com",
		ManagerFamilyName:            lo.ToPtr("ホン"),
		ManagerFirstName:             lo.ToPtr("ギルドン"),
		ManagerFamilyNameKana:        lo.ToPtr("ホン"),
		ManagerFirstNameKana:         lo.ToPtr("ギルドン"),
		ManagerEmail:                 lo.ToPtr("gildong@email.com"),
		ManagerZipCode:               "1234567",
		ManagerProvince:              "都道府県",
		ManagerCity:                  "市",
		ManagerAddress:               "詳細住所",
		ManagerAddress2:              lo.ToPtr("詳細住所2"),
		ManagerMobilePhone:           lo.ToPtr("01012345678"),
		ManagerBirthday:              mustParseCorporateUserDate("2000-01-01"),
		Department:                   lo.ToPtr("部署"),
		Position:                     lo.ToPtr("職責"),
		EmployeeCount:                "100",
		CapitalAmount:                "1000000",
		Sales:                        "500000",
	}

	mockCorporateProfile3 = domain.CorporateProfile{
		Id:            2001,
		UserId:        mockUser1.Id,
		EmployeeCount: "100",
		CapitalAmount: "1000000",
		Sales:         "500000",
	}

	mockCorporateShareholder1 = domain.CorporateShareholder{
		UserId:         mockUser2.Id,
		Uid:            "0001",
		FamilyName:     "ホン",
		FirstName:      "ギルドン",
		FamilyNameKana: "ホン",
		FirstNameKana:  "ギルドン",
		ZipCode:        "1234567",
		Address:        "詳細住所",
		Address2:       "詳細住所2",
		Province:       "都道府県",
		City:           "市",
		Birthday:       mustParseCorporateUserDate("2000-01-01"),
	}

	mockS3ObjectInfo1 = s3utils.S3ObjectInfo{
		FileSize:         1000,
		OriginalFileName: "test.jpeg",
	}
)
