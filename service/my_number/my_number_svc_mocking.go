package my_number

import (
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

var (
	mockMyNumber1 = domain.MyNumber{
		Id:       1000,
		CommonId: "2000",
		UserId:   3000,
		Status:   domain.APPROVE,
		MyNumber: "", // no data will be saved in my_number table
	}

	// decrypted "bBBbNyUqDZTowZNAcdPDgzlj9GsnJhMR/sxC4HGT5EoLcK2omgsgOXWZLOkZPun5",
	// encrypted "410d3741-7180-4024-a969-648723a57bbe",
	decryptedMyNumber = "410d3741-7180-4024-a969-648723a57bbe"

	mockUserFile3 = domain.UserFile{
		Id:            1002,
		UserId:        2000,
		FileName:      "c3a0ggd1u6bfi2k1njs0",
		OriginalName:  "dd.pdf",
		FileType:      domain.IDENTIFICATION1_FRONT,
		FileSize:      317576,
		FileExtension: ".pdf",
		FilePath:      "/user_file/2021-06-24/",
		FileUrl:       "",
		OwnerId:       3000,
		UpdateId:      4000,
		UserType:      domain.PRIVATE,
		Etc:           "3",
	}

	mockMyNumberFinal = domain.MyNumber{
		Id:        1000,
		CommonId:  "2000",
		UserId:    3000,
		Status:    domain.APPROVE,
		MyNumber:  decryptedMyNumber,
		FrontFile: &mockUserFile3,
		BackFile:  &mockUserFile3,
	}

	mockApproveMyNumber = domain.MyNumber{
		Id:       1001,
		CommonId: "2001",
		UserId:   3001,
		Status:   domain.APPROVE,
		MyNumber: "myNumber",
	}

	mockDisapprovedMyNumber = domain.MyNumber{
		Id:       1002,
		CommonId: "2002",
		UserId:   3002,
		Status:   domain.DISAPPROVED,
		MyNumber: "myNumber",
	}

	mockWaitApproveMyNumber = domain.MyNumber{
		Id:       1003,
		CommonId: "2003",
		UserId:   3003,
		Status:   domain.WAIT_APPROVE,
		MyNumber: "myNumber",
	}

	// Test_MyNumberService_RegisterWithFiles
	mockMyNumber2 = domain.MyNumber{
		Id:       int64(1),
		UserId:   int64(1),
		MyNumber: "123456789012",
		CommonId: "commonID",
		Status:   domain.WAIT_APPROVE,
	}
	mockMyNumber2Front = domain.UserFile{
		Id:       int64(1),
		UserId:   int64(1),
		OwnerId:  int64(1),
		UpdateId: int64(1),
		UserType: domain.PRIVATE,
		FileType: domain.MY_NUMBER_FRONT,
		FileUrl:  "https://es-jp-devv2-private-file-storage.s3.ap-northeast-2.amazonaws.com/user_file/2024-05-09/cou7922nulrn470ulh4g.jpeg",
		Etc:      "1",
	}
	mockMyNumber2Back = domain.UserFile{
		Id:       int64(2),
		UserId:   int64(1),
		OwnerId:  int64(1),
		UpdateId: int64(1),
		UserType: domain.PRIVATE,
		FileType: domain.MY_NUMBER_BACK,
		FileUrl:  "https://es-jp-devv2-private-file-storage.s3.ap-northeast-2.amazonaws.com/user_file/2024-05-09/cou793qnulrn470ulh50.jpeg",
		Etc:      "1",
	}
)
