package user_antisocial_check

import (
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

var (
	mockUserAntisocialCheck1 = domain.UserAntisocialCheck{
		Id:            1000,
		CheckUserId:   2000,
		Checker:       "checker1",
		InvalidReason: "반사회 1",
	}

	mockAntisocialCheckFile1 = domain.AntisocialCheckFile{
		Id:            1000,
		CheckId:       1000,
		FileName:      "test",
		OriginalName:  "test",
		FileSize:      1000,
		FileExtension: "pdf",
		FilePath:      "test",
		FileUrl:       "",
		OwnerId:       1000,
		UpdateId:      1000,
	}
	mockAntisocialCheckFile2 = domain.AntisocialCheckFile{
		Id:            2000,
		CheckId:       1001,
		FileName:      "test",
		OriginalName:  "test",
		FileSize:      1000,
		FileExtension: "pdf",
		FilePath:      "test",
		FileUrl:       "",
		OwnerId:       1000,
		UpdateId:      1000,
	}
	mocAntisocialCheckFiles1 = domain.AntisocialCheckFiles{
		mockAntisocialCheckFile1,
		mockAntisocialCheckFile2,
	}

	mocAntisocialCheckFilesInUserAntisocialCheck = []*domain.AntisocialCheckFile{
		&mockAntisocialCheckFile1,
		&mockAntisocialCheckFile2,
	}

	mockUserAntisocialCheck2 = domain.UserAntisocialCheck{
		Id:            1000,
		CheckUserId:   2000,
		Checker:       "checker1",
		InvalidReason: "반사회 1",
		CheckFiles:    mocAntisocialCheckFilesInUserAntisocialCheck,
	}
)
