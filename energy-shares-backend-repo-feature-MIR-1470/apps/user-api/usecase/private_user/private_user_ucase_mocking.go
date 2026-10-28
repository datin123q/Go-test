package private_user

import (
	"time"

	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

var (
	// Test_privateUserUsecase_RegisterMyNumber
	mockMyNumber = domain.MyNumber{
		Id:       int64(1),
		UserId:   int64(1),
		MyNumber: "123456789012",
		CommonId: "commonID",
		Status:   domain.WAIT_APPROVE,
	}
	mockMyNumberFront = domain.UserFile{
		UserId:   int64(1),
		OwnerId:  int64(1),
		UpdateId: int64(1),
		UserType: domain.PRIVATE,
		FileType: domain.MY_NUMBER_FRONT,
		FileUrl:  "https://es-jp-devv2-private-file-storage.s3.ap-northeast-2.amazonaws.com/user_file/2024-05-09/cou7922nulrn470ulh4g.jpeg",
		Etc:      "1",
	}
	mockMyNumberBack = domain.UserFile{
		UserId:   int64(1),
		OwnerId:  int64(1),
		UpdateId: int64(1),
		UserType: domain.PRIVATE,
		FileType: domain.MY_NUMBER_BACK,
		FileUrl:  "https://es-jp-devv2-private-file-storage.s3.ap-northeast-2.amazonaws.com/user_file/2024-05-09/cou793qnulrn470ulh50.jpeg",
		Etc:      "1",
	}

	mockMyNumber2 = domain.MyNumber{
		Id:       int64(2),
		UserId:   int64(2),
		MyNumber: "123456789012",
		CommonId: "commonID",
		Status:   domain.WAIT_APPROVE,
	}
	mockMyNumber2Front = domain.UserFile{
		UserId:   int64(2),
		OwnerId:  int64(2),
		UpdateId: int64(2),
		UserType: domain.PRIVATE,
		FileType: domain.MY_NUMBER_FRONT,
		FileUrl:  "https://cdn.dev.energyshares.jp/user_file/2024-05-09/cou7922nulrn470ulh4g.jpeg",
		Etc:      "2",
	}
	mockMyNumber2Back = domain.UserFile{
		UserId:   int64(2),
		OwnerId:  int64(2),
		UpdateId: int64(2),
		UserType: domain.PRIVATE,
		FileType: domain.MY_NUMBER_BACK,
		FileUrl:  "https://cdn.dev.energyshares.jp/user_file/2024-05-09/cou793qnulrn470ulh50.jpeg",
		Etc:      "2",
	}

	// TestPrivateUserUsecase_VerifyWithEkyc
	mockPrivateUserVerifyWithEkyc = domain.PrivateUser{
		User: domain.User{
			Id:       1,
			Name:     "直井 怜",
			Email:    "test@email.com",
			Uid:      "any uid",
			CommonId: "any common id",
			Password: "any password",
			Salt:     "any salt",

			// TODO: マイナンバー復活後に元に戻す
			// マイナンバー登録ありの場合、規約同意→マイナンバー登録→eKYCの順に登録する
			// マイナンバー登録なしの場合、規約同意→eKYCの順に登録する
			// SignupStatus:      domain.SIGNUP_MYNUMBER,
			SignupStatus: domain.SIGNUP_AGREE,

			Address:           "123 Main St",
			ZipCode:           "12345",
			Province:          "Any Province",
			Address2:          lo.ToPtr("   Apt 101 "),
			City:              "Any City",
			InvestmentPurpose: domain.INVESTMENT_PURPOSE_ETC,
		},
		PrivateProfile: domain.PrivateProfile{
			UserId:         1,
			FirstName:      "怜",
			FamilyName:     "直井",
			FirstNameKana:  "レイ",
			FamilyNameKana: "ナオイ",
			Birthday:       jsonutils.JSONTime(time.Date(1996, time.January, 1, 0, 0, 0, 0, lo.Must(time.LoadLocation("Asia/Tokyo")))),
			Gender:         domain.FEMALE,
			Nationality:    domain.JP,
			RequestId:      "any request id",
		},
	}

	mockPrivateUserVerifyWithEkyc2 = domain.PrivateUser{
		User: domain.User{
			Id:       1,
			Name:     "直井 怜",
			Email:    "test2@email.com",
			Uid:      "any uid 2",
			CommonId: "any common id 2",
			Password: "any password",
			Salt:     "any salt",

			// TODO: マイナンバー復活後に元に戻す
			// マイナンバー登録ありの場合、規約同意→マイナンバー登録→eKYCの順に登録する
			// マイナンバー登録なしの場合、規約同意→eKYCの順に登録する
			// SignupStatus:      domain.SIGNUP_MYNUMBER,
			SignupStatus: domain.SIGNUP_AGREE,

			Address:           "123 Main St",
			ZipCode:           "12345",
			Province:          "Any Province",
			Address2:          nil,
			City:              "Any City",
			InvestmentPurpose: domain.INVESTMENT_PURPOSE_ETC,
		},
		PrivateProfile: domain.PrivateProfile{
			UserId:         1,
			FirstName:      "怜",
			FamilyName:     "直井",
			FirstNameKana:  "レイ",
			FamilyNameKana: "ナオイ",
			Birthday:       jsonutils.JSONTime(time.Date(1996, time.January, 1, 0, 0, 0, 0, lo.Must(time.LoadLocation("Asia/Tokyo")))),
			Gender:         domain.FEMALE,
			Nationality:    domain.JP,
			RequestId:      "any request id",
		},
	}

	// TestPrivateUserUsecase_UpdateEkycStatus
	mockUserUpdateEkycStatus = domain.User{
		Id:    1,
		Email: "test@example.com",

		// TODO: マイナンバー復活後に元に戻す
		// マイナンバー登録ありの場合、規約同意→マイナンバー登録→eKYCの順に登録する
		// マイナンバー登録なしの場合、規約同意→eKYCの順に登録する
		// SignupStatus:      domain.SIGNUP_MYNUMBER,
		SignupStatus: domain.SIGNUP_AGREE,

		AuditStatus: domain.CHECK_ANTISOCIAL_AUDIT, // TODO: マイナンバー登録復活後にdomain.CHECK_MY_NUMBER_AUDITに戻す
	}

	mockEkycRequestLogUpdateEkycStatus = domain.EkycRequestLog{
		Id:          1,
		Email:       "test@example.com",
		UserId:      1,
		EkycId:      "any eKYC id",
		ApplicantId: int64(1),
		Status:      domain.EKYC_REQUEST_STATUS_NONE,
	}

	// Test_privateUserUsecase_GetSignUpStatus
	mockPrivateUserGetSignUpStatus = domain.PrivateUser{
		User: domain.User{
			Id:           1,
			Name:         "直井 怜",
			Email:        "test@email.com",
			Uid:          "any uid",
			CommonId:     "any common id",
			Password:     "any password",
			Salt:         "any salt",
			SignupStatus: domain.SIGNUP_PRIVATE_PROFILE_BASIC,
		},
		PrivateProfile: domain.PrivateProfile{
			UserId:         1,
			FirstName:      "怜",
			FamilyName:     "直井",
			FirstNameKana:  "レイ",
			FamilyNameKana: "ナオイ",
			Birthday:       jsonutils.JSONTime(time.Date(1996, time.January, 1, 0, 0, 0, 0, lo.Must(time.LoadLocation("Asia/Tokyo")))),
			Gender:         domain.FEMALE,
			Nationality:    domain.JP,
		},
	}

	// Test_privateUserUsecase_GetOne
	mockPrivateUserGetOne = domain.PrivateUser{
		User: domain.User{
			Id:                111,
			Name:              "直井 怜",
			Email:             "mockPrivateUserGetOne@email.com",
			Uid:               "any uid",
			CommonId:          "any common id",
			Password:          "any password",
			Salt:              "any salt",
			SignupStatus:      domain.SIGNUP_COMPLETE,
			Address:           "123 Main St",
			ZipCode:           "12345",
			Province:          "Any Province",
			Address2:          lo.ToPtr("Apt 101"),
			City:              "Any City",
			InvestmentPurpose: domain.INVESTMENT_PURPOSE_ETC,
			BankName:          "イルミ銀行",
			BranchName:        "イルミ支店",
			BankCode:          lo.ToPtr("123"),
			BranchCode:        lo.ToPtr("456"),
			DepositType:       domain.ORDINARY_ACCOUNT,
			BankAccount:       "1234567",
			Depositor:         "直井 怜",
			IsNewsletter:      true,
			Created:           jsonutils.JSONTime(time.Date(2024, time.January, 1, 0, 0, 0, 0, lo.Must(time.LoadLocation("Asia/Tokyo")))),
			AuditStatus:       domain.WAIT_FINAL_APPROVAL_AUDIT,
		},
		PrivateProfile: domain.PrivateProfile{
			UserId:         111,
			FirstName:      "怜",
			FamilyName:     "直井",
			FirstNameKana:  "レイ",
			FamilyNameKana: "ナオイ",
			Birthday:       jsonutils.JSONTime(time.Date(1996, time.January, 1, 0, 0, 0, 0, lo.Must(time.LoadLocation("Asia/Tokyo")))),
			Gender:         domain.FEMALE,
			Nationality:    domain.JP,
		},
	}
)
