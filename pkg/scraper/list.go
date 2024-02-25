package scraper

type ListID int64
type OnlineListID int64

type List struct {
	ID       ListID       `db:"rowid"`
	IsOnline bool         `db:"is_online"`
	OnlineID OnlineListID `db:"online_list_id"`
	Name     string       `db:"name"`

	UserIDs []UserID
	Users   []*User
}
