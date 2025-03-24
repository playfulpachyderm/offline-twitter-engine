package persistence

func (p Profile) SaveFollow(follower_id UserID, followee_id UserID) {
	_, err := p.DB.Exec(`
		insert into follows (follower_id, followee_id)
		     values (?, ?)
		on conflict do nothing
	`, follower_id, followee_id)
	if err != nil {
		panic(err)
	}
}
func (p Profile) DeleteFollow(follower_id UserID, followee_id UserID) {
	_, err := p.DB.Exec(`delete from follows where follower_id = ? and followee_id = ?`, follower_id, followee_id)
	if err != nil {
		panic(err)
	}
}

func (p Profile) SaveAsFollowersList(followee_id UserID, trove TweetTrove) {
	for follower_id := range trove.Users {
		p.SaveFollow(follower_id, followee_id)
	}
}

func (p Profile) SaveAsFolloweesList(follower_id UserID, trove TweetTrove) {
	for followee_id := range trove.Users {
		p.SaveFollow(follower_id, followee_id)
	}
}

// Returns true if the first user follows the second user, false otherwise
func (p Profile) IsXFollowingY(follower_id UserID, followee_id UserID) bool {
	rows, err := p.DB.Query(`select 1 from follows where follower_id = ? and followee_id = ?`, follower_id, followee_id)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	return rows.Next() // true if there is a row, false otherwise
}

func (p Profile) GetFollowers(followee_id UserID) []User {
	var ret []User
	err := p.DB.Select(&ret, `
	    select `+USERS_ALL_SQL_FIELDS+`
	      from users
	     where id in (select follower_id from follows where followee_id = ?)
	`, followee_id)
	if err != nil {
		panic(err)
	}
	return ret
}

func (p Profile) GetFollowees(follower_id UserID) []User {
	var ret []User
	err := p.DB.Select(&ret, `
	    select `+USERS_ALL_SQL_FIELDS+`
	      from users
	     where id in (select followee_id from follows where follower_id = ?)
	`, follower_id)
	if err != nil {
		panic(err)
	}
	return ret
}

func (p Profile) GetFollowersYouKnow(your_id UserID, their_id UserID) []User {
	var ret []User
	err := p.DB.Select(&ret, `
	    select `+USERS_ALL_SQL_FIELDS+`
	      from users
	     where id in (select followee_id from follows where follower_id = ?) -- you follow them
	       and id in (select follower_id from follows where followee_id = ?) -- they follow the other guy
	`, your_id, their_id)
	if err != nil {
		panic(err)
	}
	return ret
}

func (p Profile) GetFolloweesYouKnow(your_id UserID, their_id UserID) []User {
	var ret []User
	err := p.DB.Select(&ret, `
	    select `+USERS_ALL_SQL_FIELDS+`
	      from users
	     where id in (select followee_id from follows where follower_id = ?) -- you follow them
	       and id in (select followee_id from follows where follower_id = ?) -- the other guy also follows them
	`, your_id, their_id)
	if err != nil {
		panic(err)
	}
	return ret
}
