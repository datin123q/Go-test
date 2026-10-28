package policy

import (
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

var (
	mockPolicy1 = domain.Policy{
		Id:             1,
		CommonId:       "eeae3a9d-1ca8-43c6-a804-a235983bd102",
		OwnerId:        1,
		ModifierId:     1,
		PolicyViewType: domain.POLICY_VIEW_FOOTER,
		PolicyType:     domain.POLICY_CAUTION,
		ContentHtml:    "content",
		ModifierEmail:  "admin@energyshares.jp",
		Attachment:     &mockAttachment1,
	}

	mockAttachment1 = domain.Attachment{
		Id:         1,
		ResourceId: "eeae3a9d-1ca8-43c6-a804-a235983bd102",
	}

	mockPolicy2 = domain.Policy{
		Id:             2,
		CommonId:       "rrrr0000-0000-aaaa-bbbb-a235983bd102",
		OwnerId:        1,
		ModifierId:     1,
		PolicyViewType: domain.POLICY_VIEW_FOOTER,
		PolicyType:     domain.POLICY_PROTECTION,
		ContentHtml:    "content",
		ModifierEmail:  "admin@energyshares.jp",
		Attachment:     nil,
	}

	mockPolicyHistory1 = domain.PolicyInfoHistory{
		Id:              1,
		PolicyId:        1,
		SubstanceType:   "substance_type",
		ContentType:     "content_type",
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
		IsPosted:        false,
		Label:           jsonutils.JSONTime(time.Time{}),
		PrevContent:     "prev_content_2",
		PrevContentHtml: "prev_content_html_2",
		PrevUrl:         "prev_url_2",
		CurrContent:     "curr_content_2",
		CurrContentHtml: "curr_content_html_2",
		CurrUrl:         "curr_url_2",
	}
)
