package ghost

type Post struct {
	Slug                string      `json:"slug"`
	Id                  string      `json:"id"`
	UUID                string      `json:"uuid"`
	Title               string      `json:"title"`
	HTML                string      `json:"html"`
	CommentID           string      `json:"comment_id"`
	FeatureImage        string      `json:"feature_image"`
	FeatureImageAlt     string      `json:"feature_image_alt"`
	FeatureImageCaption string      `json:"feature_image_caption"`
	Featured            bool        `json:"featured"`
	Visibility          string      `json:"visibility"`
	CreatedAt           string      `json:"created_at"`
	UpdatedAt           string      `json:"updated_at"`
	PublishedAt         string      `json:"published_at"`
	URL                 string      `json:"url"`
	Excerpt             string      `json:"excerpt"`
	ReadingTime         int         `json:"reading_time"`
	Access              bool        `json:"access"`
	OgImage             interface{} `json:"og_image"`
	OgTitle             interface{} `json:"og_title"`
	OgDescription       interface{} `json:"og_description"`
	TwitterImage        interface{} `json:"twitter_image"`
	TwitterTitle        interface{} `json:"twitter_title"`
	TwitterDescription  interface{} `json:"twitter_description"`
	MetaTitle           interface{} `json:"meta_title"`
	MetaDescription     interface{} `json:"meta_description"`
	EmailSubject        interface{} `json:"email_subject"`
}

type Posts struct {
	Posts []Post `json:"posts"`
}
