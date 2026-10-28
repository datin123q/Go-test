package mysql

import (
	"context"

	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

// db factory mock
type dbFactoryMock struct{}

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
		Id:            1001,
		UserId:        2000,
		FileName:      "c4b0ggd1u6bfi2k1njs0", // ユニークである必要がある
		OriginalName:  "dd.pdf",
		FileType:      domain.IDENTIFICATION1_BACK,
		FileSize:      317576,
		FileExtension: ".pdf",
		FilePath:      "/user_file/2021-06-24/",
		FileUrl:       "https://ilm-stg-eskk-file-storage.s3.ap-northeast-2.amazonaws.com/user_file/2021-06-24/c3a0ggd1u6bfi2k1njs0.pdf",
		OwnerId:       3000,
		UpdateId:      4000,
		UserType:      domain.PRIVATE,
		Etc:           "3",
	}

	mockUserFiles1 = domain.UserFiles{mockUserFile1, mockUserFile2}
	mockUserFiles2 = domain.UserFiles{mockUserFile2}

	mockUserFileCreate = domain.UserFile{
		Id:            1002,
		UserId:        2002,
		FileName:      "z4b0ggd1u6bfi2k1njs0", // ユニークである必要がある
		OriginalName:  "dd.pdf",
		FileType:      domain.IDENTIFICATION1_BACK,
		FileSize:      317576,
		FileExtension: ".pdf",
		FilePath:      "/user_file/2021-06-24/",
		FileUrl:       "https://ilm-stg-eskk-file-storage.s3.ap-northeast-2.amazonaws.com/user_file/2021-06-24/c3a0ggd1u6bfi2k1njs0.pdf",
		OwnerId:       3000,
		UpdateId:      4000,
		UserType:      domain.PRIVATE,
		Etc:           "4",
	}

	// 論理削除して使う。file_name は他と重複させない
	mockUserFileDeleted = domain.UserFile{
		Id:            1900,
		UserId:        2000,
		FileName:      "c5c0ggd1u6bfi2k1njs0",
		OriginalName:  "deleted.pdf",
		FileType:      domain.CERTIFIED_COPY,
		FileSize:      1000,
		FileExtension: ".pdf",
		FilePath:      "/user_file/2021-06-24/",
		FileUrl:       "https://ilm-stg-eskk-file-storage.s3.ap-northeast-2.amazonaws.com/user_file/2021-06-24/c5c0ggd1u6bfi2k1njs0.pdf",
		OwnerId:       3000,
		UpdateId:      4000,
		UserType:      domain.PRIVATE,
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.UserFile))
	xormEngine.ShowSQL(false)

	// テストデータ挿入
	xormEngine.Insert(mockUserFile1, mockUserFile2)

	// 論理削除済みの行。file_name の unique 制約は deleted を見ないため、
	// 削除済みでも同名は作れない。ExistsByFileName がこれを拾えるかを確かめる。
	xormEngine.Insert(mockUserFileDeleted)
	xormEngine.ID(mockUserFileDeleted.Id).Delete(new(domain.UserFile))

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
