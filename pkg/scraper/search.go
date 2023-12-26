package scraper

func TimestampToDateString(timestamp int) string {
	panic("???") // TODO
}

/**
 * TODO: Search modes:
 * - regular ("top")
 * - latest / "live"
 * - search for users
 * - photos
 * - videos
 */
func Search(query string, min_results int) (trove TweetTrove, err error) {
	return the_api.GetPaginatedQuery(PaginatedSearch{query}, min_results)
}
