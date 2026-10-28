package look_up_my_number

import (
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

var (
	mockLookUpMyNumber1 = domain.LookUpMyNumber{
		Id:         1000,
		MyNumberId: 2000,
		MyNumber:   "bBBbNyUqDZTowZNAcdPDgzlj9GsnJhMR/sxC4HGT5EoLcK2omgsgOXWZLOkZPun5", // encrypted "410d3741-7180-4024-a969-648723a57bbe"
	}

	mockLookUpMyNumber2 = domain.LookUpMyNumber{
		Id:         1000,
		MyNumberId: 2000,
		MyNumber:   "410d3741-7180-4024-a969-648723a57bbe",
	}

	mockLookUpMyNumberDeleted1 = domain.LookUpMyNumber{
		Id:         1001,
		MyNumberId: 2001,
		MyNumber:   "bBBbNyUqDZTowZNAcdPDgzlj9GsnJhMR/sxC4HGT5EoLcK2omgsgOXWZLOkZPun5",
	}

	mockLookUpMyNumberDeleted2 = domain.LookUpMyNumber{
		Id:         1001,
		MyNumberId: 2001,
		MyNumber:   "410d3741-7180-4024-a969-648723a57bbe",
	}
)
