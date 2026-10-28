package user_file

import (
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

var (
	mockUserFile1 = domain.UserFile{
		Id:            1000,
		UserId:        2000,
		FileName:      "c3a0ggd1u6bfi2k1njs0",
		OriginalName:  "dd.pdf",
		FileType:      domain.IDENTIFICATION1_FRONT,
		FileSize:      317576,
		FileExtension: ".pdf",
		FilePath:      "/user_file/2021-06-24/",
		FileUrl:       "https://ilm-stg-eskk-file-storage.s3.ap-northeast-2.amazonaws.com/user_file/2021-06-24/c3a0ggd1u6bfi2k1njs0.pdf",
		OwnerId:       3000,
		UpdateId:      4000,
		UserType:      domain.PRIVATE,
		Etc:           "3",
	}

	mockUserFile2 = domain.UserFile{
		Id:            1000,
		UserId:        2000,
		FileName:      "c3a0ggd1u6bfi2k1njs0",
		OriginalName:  "dd.pdf",
		FileType:      domain.IDENTIFICATION1_FRONT,
		FileSize:      317576,
		FileExtension: ".pdf",
		FilePath:      "/user_file/2021-06-24/",
		FileUrl:       "", // cause of s3utils.MakeCDNSignURL(getKey()) in user_file_svc.go
		OwnerId:       3000,
		UpdateId:      4000,
		UserType:      domain.PRIVATE,
		Etc:           "3",
	}

	mockCertifiedCopyFile = domain.UserFile{
		Id:            0,
		UserId:        1000,
		FileName:      "cou7922nulrn470ulh4g",
		OriginalName:  "test.jpeg",
		FileType:      domain.CERTIFIED_COPY,
		FileSize:      1000,
		FileExtension: ".jpeg",
		FilePath:      "/user_file/2024-05-09/",
		FileUrl:       "https://cdn.dev.energyshares.jp/user_file/2024-05-09/cou7922nulrn470ulh4g.jpeg",
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
		FileName:      "cou7922nulrn470ulh4g",
		OriginalName:  "test.jpeg",
		FileType:      domain.IDENTIFICATION1_FRONT,
		FileSize:      1000,
		FileExtension: ".jpeg",
		FilePath:      "/user_file/2024-05-09/",
		FileUrl:       "https://cdn.dev.energyshares.jp/user_file/2024-05-09/cou7922nulrn470ulh4g.jpeg",
		OwnerId:       1000,
		UpdateId:      1000,
		UserType:      domain.CORPORATE,
	}

	mockIdentification2FrontFile = domain.UserFile{
		Id:            0,
		UserId:        1000,
		FileName:      "cou7922nulrn470ulh4g",
		OriginalName:  "test.jpeg",
		FileType:      domain.IDENTIFICATION2_FRONT,
		FileSize:      1000,
		FileExtension: ".jpeg",
		FilePath:      "/user_file/2024-05-09/",
		FileUrl:       "https://cdn.dev.energyshares.jp/user_file/2024-05-09/cou7922nulrn470ulh4g.jpeg",
		OwnerId:       1000,
		UpdateId:      1000,
		UserType:      domain.CORPORATE,
	}

	mockPOAFile = domain.UserFile{
		Id:            0,
		UserId:        1000,
		FileName:      "cou7922nulrn470ulh4g",
		OriginalName:  "test.jpeg",
		FileType:      domain.POA,
		FileSize:      1000,
		FileExtension: ".jpeg",
		FilePath:      "/user_file/2024-05-09/",
		FileUrl:       "https://cdn.dev.energyshares.jp/user_file/2024-05-09/cou7922nulrn470ulh4g.jpeg",
		OwnerId:       1000,
		UpdateId:      1000,
		UserType:      domain.CORPORATE,
	}

	mockIdentification1BackFile = domain.UserFile{
		Id:            0,
		UserId:        1000,
		FileName:      "cou7922nulrn470ulh4g",
		OriginalName:  "test.jpeg",
		FileType:      domain.IDENTIFICATION1_BACK,
		FileSize:      1000,
		FileExtension: ".jpeg",
		FilePath:      "/user_file/2024-05-09/",
		FileUrl:       "https://cdn.dev.energyshares.jp/user_file/2024-05-09/cou7922nulrn470ulh4g.jpeg",
		OwnerId:       1000,
		UpdateId:      1000,
		UserType:      domain.CORPORATE,
	}

	mockIdentification2BAckFile = domain.UserFile{
		Id:            0,
		UserId:        1000,
		FileName:      "cou7922nulrn470ulh4g",
		OriginalName:  "test.jpeg",
		FileType:      domain.IDENTIFICATION2_BACK,
		FileSize:      1000,
		FileExtension: ".jpeg",
		FilePath:      "/user_file/2024-05-09/",
		FileUrl:       "https://cdn.dev.energyshares.jp/user_file/2024-05-09/cou7922nulrn470ulh4g.jpeg",
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

	mockUserFiles = domain.UserFiles{
		mockUserFile1,
		mockUserFile2,
	}

	mockPrivateUserIdentificationFiles = cdomain.PrivateUserIdentificationFiles{
		Front1: mockUserFile1,
		Back1:  mockUserFile2,
		Front2: mockUserFile1,
		Back2:  mockUserFile2,
	}

	mockCorporateUserIdentificationFiles = cdomain.CorporateUserIdentificationFiles{
		Front1:        &mockIdentification1FrontFile,
		Back1:         &mockIdentification1BackFile,
		Front2:        &mockIdentification2FrontFile,
		Back2:         &mockIdentification2BAckFile,
		CertifiedCopy: &mockCertifiedCopyFile,
		POA:           &mockPOAFile,
		AOI:           &mockAOIFile,
		Shareholders:  &mockShareholdersFile,
		Others:        &mockOthersFile,
	}
)
