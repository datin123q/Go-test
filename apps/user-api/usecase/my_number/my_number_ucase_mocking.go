package my_number

import (
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

var (
	// Test_myNumberUsecase_Create
	mockMyNumber = domain.MyNumber{
		Id:       int64(1),
		UserId:   int64(1),
		MyNumber: "123456789012",
		CommonId: "commonID",
		Status:   domain.WAIT_APPROVE,
	}
	mockMyNumberFront = domain.UserFile{
		Id:       int64(1),
		UserId:   int64(1),
		OwnerId:  int64(1),
		UpdateId: int64(1),
		UserType: domain.PRIVATE,
		FileType: domain.MY_NUMBER_FRONT,
		FileUrl:  "https://es-jp-devv2-private-file-storage.s3.ap-northeast-2.amazonaws.com/user_file/2024-05-09/cou7922nulrn470ulh4g.jpeg",
		Etc:      "1",
	}
	mockMyNumberBack = domain.UserFile{
		Id:       int64(2),
		UserId:   int64(1),
		OwnerId:  int64(1),
		UpdateId: int64(1),
		UserType: domain.PRIVATE,
		FileType: domain.MY_NUMBER_BACK,
		FileUrl:  "https://es-jp-devv2-private-file-storage.s3.ap-northeast-2.amazonaws.com/user_file/2024-05-09/cou793qnulrn470ulh50.jpeg",
		Etc:      "1",
	}

	// Test_myNumberUsecase_Get
	mockMyNumber2 = domain.MyNumber{
		Id:       int64(2),
		UserId:   int64(2),
		MyNumber: "123456789012",
		CommonId: "commonID",
		Status:   domain.APPROVE,
		FrontFile: &domain.UserFile{
			UserId:   int64(2),
			OwnerId:  int64(2),
			UpdateId: int64(2),
			UserType: domain.PRIVATE,
			FileType: domain.MY_NUMBER_FRONT,
			FileUrl:  "https://es-jp-devv2-private-file-storage.s3.ap-northeast-2.amazonaws.com/user_file/2024-05-09/cou793qnulrn470ulh50.jpeg",
			Etc:      "2",
		},
		BackFile: &domain.UserFile{
			UserId:   int64(2),
			OwnerId:  int64(2),
			UpdateId: int64(2),
			UserType: domain.PRIVATE,
			FileType: domain.MY_NUMBER_BACK,
			FileUrl:  "https://es-jp-devv2-private-file-storage.s3.ap-northeast-2.amazonaws.com/user_file/2024-05-09/cou793qnulrn470ulh50.jpeg",
			Etc:      "2",
		},
	}
)
