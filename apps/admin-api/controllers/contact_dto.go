package controllers

type ContactSearchInput struct {
	SearchInput
	Name      *string `query:"name"`
	Title     *string `query:"title"`
	Email     *string `query:"email"`
	StartDate string  `query:"start_date"`
	EndDate   string  `query:"end_date"`
}
