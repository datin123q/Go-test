package domain

import (
	"context"
)

type Dashboard struct {
	CorporateUsers                    []CorporateUser                  `json:"corporate_users" xorm:"-" swagger:"desc(검토대기 법인회원 리스트 )"`
	CorporateUserCount                int64                            `json:"corporate_user_count" swagger:"desc(검토대기 법인회원 건수)"`
	PrivateUsers                      []PrivateUser                    `json:"private_users" xorm:"-" swagger:"desc(검토대기 개인회원 리스트 )"`
	PrivateUserCount                  int64                            `json:"private_user_count" swagger:"desc(검토대기 개인회원 건수 )"`
	FundingRequests                   []FundingRequest                 `json:"funding_requests" xorm:"-" swagger:"desc(검토대기 대출의뢰 리스트 )"`
	FundingRequestCount               int64                            `json:"funding_request_count" swagger:"검토대기 대출의뢰 건수"`
	Investments                       []InvestmentDashboardResponseDto `json:"investments" xorm:"-" swagger:"desc(최근 거래목록  리스트 )"`
	InvestmentCount                   int64                            `json:"investment_count" swagger:"desc(최근 거래 목록 건수)"`
	Projects                          []Project                        `json:"projects" xorm:"-" swagger:"desc(모집 시작일이 지난 미게시 상품리스트)"`
	ProjectCount                      int64                            `json:"project_count" swagger:"모집시작일이 지난 미게시 상품 건수"`
	AntisocialCheckPrivateUsers       []PrivateUser                    `json:"antisocial_check_private_user"  swagger:"desc(반사회 확인중 개인회원 목록)"`
	AntisocialCheckPrivateUserCount   int64                            `json:"antisocial_check_private_user_count" swagger:"desc(반사회 확인중 개인회원 건수)"`
	AntisocialCheckCorporateUsers     []CorporateUser                  `json:"antisocial_check_corporate_user" swagger:"desc(반사회 확인중 법인회원 목록)"`
	AntisocialCheckCorporateUserCount int64                            `json:"antisocial_check_corporate_user_count" swagger:"desc(반사회 확인중 법인회원 건수)"`
	QnaList                           []Qna                            `json:"qna_list" swagger:"desc(미확인 문의 목록)"`
	QnaCount                          int64                            `json:"qna_count" swagger:"desc(미확인 문의 목록 건수)"`
	WaitMyNumberCount                 int64                            `json:"wait_my_number_count" swagger:"desc(신청 대기중인 마이넘버 등록 요청 건수 )"`
	WithdrawalList                    []Withdrawal                     `json:"withdrawal_list" swagger:"desc(미승인 출금 의뢰 목록)"`
	WithdrawalCount                   int64                            `json:"withdrawal_count" swagger:"desc(미승인 출금 의뢰 목록 건수)"`
	EquityVaCount                     int64                            `json:"equity_va_count" swagger:"desc(사용 가능한 사업형 가상계좌 건수 )"`
}

func (Dashboard) GetDashboard(ctx context.Context, offset, limit int) (*Dashboard, error) {
	var dashboard Dashboard
	var err error

	dashboard.CorporateUserCount, dashboard.CorporateUsers, err = CorporateUser{}.GetDashboardList(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	dashboard.PrivateUserCount, dashboard.PrivateUsers, err = PrivateUser{}.GetDashboardList(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	dashboard.ProjectCount, dashboard.Projects, err = Project{}.GetDashboardList(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	dashboard.FundingRequestCount, dashboard.FundingRequests, err = FundingRequest{}.GetDashboardList(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	dashboard.InvestmentCount, dashboard.Investments, err = Investment{}.GetDashboardList(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	dashboard.QnaCount, dashboard.QnaList, err = Qna{}.GetNoAnswerQna(ctx, "", "", offset, limit)
	if err != nil {
		return nil, err
	}

	dashboard.AntisocialCheckPrivateUserCount, dashboard.AntisocialCheckPrivateUsers, err = PrivateUser{}.GetAntisocialCheckUsers(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	dashboard.AntisocialCheckCorporateUserCount, dashboard.AntisocialCheckCorporateUsers, err = CorporateUser{}.GetAntisocialCheckUsers(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	dashboard.WaitMyNumberCount, err = MyNumber{}.CountWaitApprove(ctx)
	if err != nil {
		return nil, err
	}

	dashboard.EquityVaCount, err = UserVirtualAccount{}.GetVaByUnLinkedAccountCount(ctx, LOAN)
	if err != nil {
		return nil, err
	}

	dashboard.WithdrawalList, dashboard.WithdrawalCount, err = Withdrawal{}.GetIncompleteWithdrawal(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	return &dashboard, nil
}
