package lottery

import "github.com/illuminarean/energy-shares-backend-repo/domain"

var (
	mockLotteryForDownload = domain.LotteryApplication{
		Id: "app-001",
	}
	mockLotteriesForDownload = []domain.LotteryApplication{
		mockLotteryForDownload,
	}

	mockLotteryApplied = domain.LotteryApplication{
		Id:     "app-applied",
		Status: domain.LOTTERY_STATUS_APPLIED,
	}
	mockLotteryLost = domain.LotteryApplication{
		Id:     "app-lost",
		Status: domain.LOTTERY_STATUS_LOST,
	}
	mockLotteryPartialWinning = domain.LotteryApplication{
		Id:     "app-partial-winning",
		Status: domain.LOTTERY_STATUS_PARTIAL_WINNING,
	}
	mockLotteryWinning = domain.LotteryApplication{
		Id:     "app-winning",
		Status: domain.LOTTERY_STATUS_WINNING,
	}
	mockLotteryCanceled = domain.LotteryApplication{
		Id:     "app-canceled",
		Status: domain.LOTTERY_STATUS_CANCELED,
	}
	mockLotteriesForBulkLottery = []domain.LotteryApplication{
		mockLotteryApplied,
		mockLotteryLost,
		mockLotteryPartialWinning,
		mockLotteryWinning,
		mockLotteryCanceled,
	}
	mockFilteredLotteriesForBulkLottery = []domain.LotteryApplication{
		mockLotteryApplied,
		mockLotteryLost,
		mockLotteryPartialWinning,
	}
)
