package policy

import (
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

var (
	mockModifier = jwthelper.User{
		Id:    1,
		Email: "modifier@energyshares.jp",
		Name:  "modifier",
	}

	mockPolicy1 = domain.Policy{
		Id:             1,
		CommonId:       "eeae3a9d-1ca8-43c6-a804-a235983bd102",
		OwnerId:        1,
		ModifierId:     1,
		PolicyViewType: domain.POLICY_VIEW_JOIN,
		PolicyType:     domain.POLICY_PROTECTION,
		ContentHtml:    "content",
		ModifierEmail:  "admin@energyshares.jp",
		Attachment:     &mockAttachment1,
	}

	mockAttachment1 = domain.Attachment{
		Id:         1,
		ResourceId: "eeae3a9d-1ca8-43c6-a804-a235983bd102",
	}

	mockNewAttachment = domain.Attachment{
		FileName:      "original_file_name",
		OriginalName:  "original_file_name.jpg",
		FileSize:      1234,
		FileExtension: ".jpg",
		FilePath:      "/attachment_file/2024-09-09/",
		FileUrl:       "https://cdn.dev.energyshares.jp/attachment_file/2024-09-09/original_file_name.jpg",
		ResourceId:    mockPolicy1.CommonId,
	}

	mockPolicy2 = domain.Policy{
		Id:             2,
		CommonId:       "c2c546ab-0715-4527-a40f-16c126fb1c6b",
		OwnerId:        1,
		ModifierId:     1,
		PolicyViewType: domain.POLICY_VIEW_FOOTER,
		PolicyType:     domain.POLICY_CAUTION,
		ContentHtml:    "content",
		ModifierEmail:  "admin@energyshares.jp",
		Attachment:     &mockAttachment2,
	}

	mockAttachment2 = domain.Attachment{
		Id:         2,
		ResourceId: "c2c546ab-0715-4527-a40f-16c126fb1c6b",
	}

	mockPolicyHistory1 = domain.PolicyInfoHistory{
		Id:              1,
		PolicyId:        1,
		SubstanceType:   "substance_type",
		ContentType:     "content_type",
		Name:            mockModifier.Name,
		Email:           mockModifier.Email,
		IsPosted:        true,
		Label:           jsonutils.JSONTime(time.Date(2024, 2, 9, 0, 0, 0, 0, time.UTC)),
		PrevContent:     "prev_content_1",
		PrevContentHtml: "prev_content_html_1",
		PrevUrl:         "prev_url_1",
		CurrContent:     "curr_content_1",
		CurrContentHtml: "curr_content_html_1",
		CurrUrl:         "curr_url_1",
	}
	mockPolicyHistory2 = domain.PolicyInfoHistory{
		Id:              2,
		PolicyId:        1,
		SubstanceType:   "substance_type",
		ContentType:     "content_type",
		Name:            mockModifier.Name,
		Email:           mockModifier.Email,
		IsPosted:        false,
		Label:           jsonutils.JSONTime(time.Time{}),
		PrevContent:     "prev_content_2",
		PrevContentHtml: "prev_content_html_2",
		PrevUrl:         "prev_url_2",
		CurrContent:     "curr_content_2",
		CurrContentHtml: "curr_content_html_2",
		CurrUrl:         "curr_url_2",
	}
	mockPolicyHistory3 = domain.PolicyInfoHistory{
		Id:              3,
		PolicyId:        1,
		SubstanceType:   "substance_type",
		ContentType:     "content_type",
		Name:            mockModifier.Name,
		Email:           mockModifier.Email,
		IsPosted:        true,
		Label:           jsonutils.JSONTime(time.Date(2024, 3, 14, 0, 0, 0, 0, time.UTC)),
		PrevContent:     "prev_content_1",
		PrevContentHtml: "prev_content_html_1",
		PrevUrl:         "prev_url_1",
		CurrContent:     "curr_content_1",
		CurrContentHtml: "curr_content_html_1",
		CurrUrl:         "curr_url_1",
	}
)
