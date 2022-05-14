package scraper

type SpaceID string

type Space struct {
	ID      SpaceID
	TweetID TweetID

	Url string
}

func ParseAPISpace(apiCard APICard) Space {
	ret := Space{}
	ret.ID = SpaceID(apiCard.BindingValues.ID.StringValue)
	ret.Url = apiCard.ShortenedUrl

	return ret
}
