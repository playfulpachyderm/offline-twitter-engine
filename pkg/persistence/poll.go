package persistence

import (
	"time"
)

type PollID int64

type Poll struct {
	ID         PollID  `db:"id"`
	TweetID    TweetID `db:"tweet_id"`
	NumChoices int     `db:"num_choices"`

	Choice1       string `db:"choice1"`
	Choice1_Votes int    `db:"choice1_votes"`
	Choice2       string `db:"choice2"`
	Choice2_Votes int    `db:"choice2_votes"`
	Choice3       string `db:"choice3"`
	Choice3_Votes int    `db:"choice3_votes"`
	Choice4       string `db:"choice4"`
	Choice4_Votes int    `db:"choice4_votes"`

	VotingDuration int       `db:"voting_duration"` // In seconds
	VotingEndsAt   Timestamp `db:"voting_ends_at"`

	LastUpdatedAt Timestamp `db:"last_scraped_at"`
}

// TODO: view-layer
// - view helpers should go in a view layer

func (p Poll) TotalVotes() int {
	return p.Choice1_Votes + p.Choice2_Votes + p.Choice3_Votes + p.Choice4_Votes
}
func (p Poll) VotePercentage(n int) float64 {
	return 100.0 * float64(n) / float64(p.TotalVotes())
}
func (p Poll) IsOpen() bool {
	return time.Now().Unix() < p.VotingEndsAt.Unix()
}
func (p Poll) FormatEndsAt() string {
	return p.VotingEndsAt.Format("Jan 2, 2006 3:04 pm")
}
func (p Poll) IsWinner(votes int) bool {
	if p.IsOpen() {
		// There's no winner if the poll is still open
		return false
	}
	return votes >= p.Choice1_Votes && votes >= p.Choice2_Votes && votes >= p.Choice3_Votes && votes >= p.Choice4_Votes
}
