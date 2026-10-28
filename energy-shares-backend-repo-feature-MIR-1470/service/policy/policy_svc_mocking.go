package policy

import "github.com/illuminarean/energy-shares-backend-repo/domain"

var (
	mockBeforePolicy = domain.Policy{
		Id:             1,
		CommonId:       "eeae3a9d-1ca8-43c6-a804-a235983bd102",
		OwnerId:        1,
		ModifierId:     1,
		PolicyViewType: domain.POLICY_VIEW_JOIN,
		PolicyType:     domain.POLICY_PROTECTION,
		ContentHtml:    "content",
		ModifierEmail:  "admin@energyshares.jp",
		Attachment: &domain.Attachment{
			OriginalName: "test.txt",
			FileUrl:      "http://test.com/test.txt",
		},
	}

	mockAfterPolicy = domain.Policy{
		Id:             1,
		CommonId:       "eeae3a9d-1ca8-43c6-a804-a235983bd102",
		OwnerId:        1,
		ModifierId:     2,
		PolicyViewType: domain.POLICY_VIEW_JOIN,
		PolicyType:     domain.POLICY_PROTECTION,
		ContentHtml:    "content-changed",
		ModifierEmail:  "test@example.com",
		Attachment: &domain.Attachment{
			OriginalName: "test-changed.txt",
			FileUrl:      "http://test.com/test-changed.txt",
		},
	}

	mockAdminUser = domain.AdminUser{
		Id:    2,
		Name:  "Admin2",
		Email: "test@example.com",
	}
)
