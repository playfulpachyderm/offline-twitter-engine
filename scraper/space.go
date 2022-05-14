package scraper

type SpaceID string

type Space struct {
	ID       SpaceID `db:"id"`
	ShortUrl string  `db:"short_url"`
}

func ParseAPISpace(apiCard APICard) Space {
	ret := Space{}
	ret.ID = SpaceID(apiCard.BindingValues.ID.StringValue)
	ret.ShortUrl = apiCard.ShortenedUrl

	return ret
}
