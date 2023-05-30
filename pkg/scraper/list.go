package scraper

type ListID int

type List struct {
	ID   ListID `db:"rowid"`
	Type string `db:"type"`
	Name string `db:"name"`

	UserIDs []UserID
	Users   []*User
}
