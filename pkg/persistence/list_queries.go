package persistence

import (
	"database/sql"
	"errors"
	"fmt"
)

// Create an empty list, or rename an existing list
func (p Profile) SaveList(l *List) {
	// Since the unique column is managed by the database (auto-increment) due to the existence of
	// offline lists, we have to check for its existence first
	var rowid ListID
	if l.IsOnline {
		// Online list; look up its rowid by its online ID
		// TODO: maybe extract to a function
		err := p.DB.Get(&rowid, "select rowid from lists where is_online = 1 and online_list_id = ?", l.ID)
		if errors.Is(err, sql.ErrNoRows) {
			// Doesn't exist yet
			rowid = ListID(0)
		} else if err != nil {
			panic(err)
		}
	} else {
		// For offline lists, just use the rowid
		rowid = l.ID
	}

	// If `rowid` is 0, then it doesn't exist yet; create it.  Otherwise, update it
	if rowid == ListID(0) {
		result, err := p.DB.NamedExec(`
		    insert into lists (is_online, online_list_id, name)
		    values (:is_online, :online_list_id, :name)
		`, l)
		if err != nil {
			panic(err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			panic(err)
		}
		l.ID = ListID(id)
	} else {
		// Do update
		_, err := p.DB.NamedExec(`
			update lists set name = :name where rowid = :rowid
		`, l)
		if err != nil {
			panic(err)
		}
	}
}

func (p Profile) DeleteList(list_id ListID) {
	_, err := p.DB.Exec(`delete from lists where rowid = ?`, list_id)
	if err != nil {
		panic(fmt.Errorf("Error executing DeleteList(%d):\n  %w", list_id, err).Error())
	}
}

func (p Profile) SaveListUsers(list_id ListID, trove TweetTrove) {
	for user_id := range trove.Users {
		p.SaveListUser(list_id, user_id)
	}
}

func (p Profile) SaveListUser(list_id ListID, user_id UserID) {
	_, err := p.DB.Exec(`insert into list_users (list_id, user_id) values (?, ?) on conflict do nothing`, list_id, user_id)
	if err != nil {
		panic(fmt.Errorf("Error executing AddListUser(%d, %d):\n  %w", list_id, user_id, err).Error())
	}
}

func (p Profile) DeleteListUser(list_id ListID, user_id UserID) {
	_, err := p.DB.Exec(`delete from list_users where list_id = ? and user_id = ?`, list_id, user_id)
	if err != nil {
		panic(fmt.Errorf("Error executing DeleteListUser(%d, %d):\n  %w", list_id, user_id, err).Error())
	}
}

func (p Profile) GetListById(list_id ListID) (List, error) {
	var ret List
	err := p.DB.Get(&ret, `select rowid, is_online, online_list_id, name from lists where rowid = ?`, list_id)
	if errors.Is(err, sql.ErrNoRows) {
		return List{}, ErrNotInDatabase
	} else if err != nil {
		panic(err)
	}
	return ret, nil
}

func (p Profile) GetListUsers(list_id ListID) []User {
	var ret []User
	err := p.DB.Select(&ret, `
	    select `+USERS_ALL_SQL_FIELDS+`
	      from users
	     where id in (select user_id from list_users where list_id = ?)
	`, list_id)
	if err != nil {
		panic(err)
	}
	return ret
}

func (p Profile) GetAllLists() []List {
	var lists []List
	err := p.DB.Select(&lists, `select rowid, is_online, online_list_id, name from lists`)
	if err != nil {
		panic(err)
	}
	for i := range lists {
		lists[i].Users = p.GetListUsers(lists[i].ID)
	}
	return lists
}
