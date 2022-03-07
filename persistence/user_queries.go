package persistence

import (
	"fmt"
	"database/sql"
	"offline_twitter/scraper"
)

/**
 * Save the given User to the database.
 * If the User is already in the database, it will update most of its attributes (follower count, etc)
 *
 * args:
 * - u: the User
 */
func (p Profile) SaveUser(u *scraper.User) error {
	if u.IsNeedingFakeID {
		err := p.DB.QueryRow("select id from users where lower(handle) = lower(?)", u.Handle).Scan(&u.ID)
		if err == sql.ErrNoRows {
			// We need to continue-- create a new fake user
			u.ID = p.NextFakeUserID()
		} else if err == nil {
			// We're done; everything is fine (ID has already been scanned into the User)
			return nil
		} else {
			// A real error occurred
			panic(fmt.Sprintf("Error checking for existence of fake user with handle %q: %s", u.Handle, err.Error()))
		}
	}

	_, err := p.DB.Exec(`
        insert into users (id, display_name, handle, bio, following_count, followers_count, location, website, join_date, is_private,
                           is_verified, is_banned, profile_image_url, profile_image_local_path, banner_image_url, banner_image_local_path,
                           pinned_tweet_id, is_content_downloaded, is_id_fake)
        values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            on conflict do update
           set bio=?,
               display_name=?,
               following_count=?,
               followers_count=?,
               location=?,
               website=?,
               is_private=?,
               is_verified=?,
               is_banned=?,
               profile_image_url=?,
               profile_image_local_path=?,
               banner_image_url=?,
               banner_image_local_path=?,
               pinned_tweet_id=?,
               is_content_downloaded=(is_content_downloaded or ?)
        `,
		u.ID, u.DisplayName, u.Handle, u.Bio, u.FollowingCount, u.FollowersCount, u.Location, u.Website, u.JoinDate, u.IsPrivate,
		u.IsVerified, u.IsBanned, u.ProfileImageUrl, u.ProfileImageLocalPath, u.BannerImageUrl, u.BannerImageLocalPath, u.PinnedTweetID,
		u.IsContentDownloaded, u.IsIdFake,

		u.Bio, u.DisplayName, u.FollowingCount, u.FollowersCount, u.Location, u.Website, u.IsPrivate, u.IsVerified, u.IsBanned,
		u.ProfileImageUrl, u.ProfileImageLocalPath, u.BannerImageUrl, u.BannerImageLocalPath, u.PinnedTweetID, u.IsContentDownloaded,
	)
	if err != nil {
		return err
	}

	return nil
}

/**
 * Check if the database has a User with the given user handle.
 *
 * args:
 * - handle: the user handle to search for
 *
 * returns:
 * - true if there is such a User in the database, false otherwise
 */
func (p Profile) UserExists(handle scraper.UserHandle) bool {
    db := p.DB

    var dummy string
    err := db.QueryRow("select 1 from users where lower(handle) = lower(?)", handle).Scan(&dummy)
    if err != nil {
        if err != sql.ErrNoRows {
            // A real error
            panic(err)
        }
        return false
    }
    return true
}

/**
 * Helper function.  Create a User from a Row.
 */
func parse_user_from_row(row *sql.Row) (scraper.User, error) {
    var u scraper.User

    err := row.Scan(&u.ID, &u.DisplayName, &u.Handle, &u.Bio, &u.FollowingCount, &u.FollowersCount, &u.Location, &u.Website, &u.JoinDate,
                    &u.IsPrivate, &u.IsVerified, &u.IsBanned, &u.ProfileImageUrl, &u.ProfileImageLocalPath, &u.BannerImageUrl,
                    &u.BannerImageLocalPath, &u.PinnedTweetID, &u.IsContentDownloaded, &u.IsFollowed)
    return u, err
}


/**
 * Retrieve a User from the database, by handle.
 *
 * args:
 * - handle: the user handle to search for
 *
 * returns:
 * - the User, if it exists
 */
func (p Profile) GetUserByHandle(handle scraper.UserHandle) (scraper.User, error) {
    db := p.DB

    stmt, err := db.Prepare(`
        select id, display_name, handle, bio, following_count, followers_count, location, website, join_date, is_private, is_verified,
               is_banned, profile_image_url, profile_image_local_path, banner_image_url, banner_image_local_path, pinned_tweet_id,
               is_content_downloaded, is_followed
          from users
         where lower(handle) = lower(?)
    `)
    if err != nil {
        return scraper.User{}, err
    }
    defer stmt.Close()

    row := stmt.QueryRow(handle)
    ret, err := parse_user_from_row(row)
    if err == sql.ErrNoRows {
        return ret, ErrNotInDatabase{"User", handle}
    }
    return ret, nil
}


/**
 * Retrieve a User from the database, by user ID.
 *
 * args:
 * - id: the user ID to search for
 *
 * returns:
 * - the User, if it exists
 */
func (p Profile) GetUserByID(id scraper.UserID) (scraper.User, error) {
    db := p.DB

    stmt, err := db.Prepare(`
        select id, display_name, handle, bio, following_count, followers_count, location, website, join_date, is_private, is_verified,
               is_banned, profile_image_url, profile_image_local_path, banner_image_url, banner_image_local_path, pinned_tweet_id,
               is_content_downloaded, is_followed
          from users
         where id = ?
    `)
    if err != nil {
        return scraper.User{}, err
    }
    defer stmt.Close()

    row := stmt.QueryRow(id)
    ret, err := parse_user_from_row(row)
    if err == sql.ErrNoRows {
        return ret, ErrNotInDatabase{"User", id}
    }
    return ret, err
}

/**
 * Returns `true` if content download is needed, `false` otherwise
 *
 * If the user is banned, returns false because downloading will be impossible.
 *
 * If:
 * - the user isn't in the DB at all (first time scraping), OR
 * - `is_content_downloaded` is false in the DB, OR
 * - the banner / profile image URL has changed from what the DB has
 * then it needs to be downloaded.
 *
 * The `user` object will always have `is_content_downloaded` = false on every scrape.  This is
 * why the No Worsening Principle is needed.
 */
func (p Profile) CheckUserContentDownloadNeeded(user scraper.User) bool {
	row := p.DB.QueryRow(`select is_content_downloaded, profile_image_url, banner_image_url from users where id = ?`, user.ID)

	var is_content_downloaded bool
	var profile_image_url string
	var banner_image_url string
	err := row.Scan(&is_content_downloaded, &profile_image_url, &banner_image_url)
	if err != nil {
		if err == sql.ErrNoRows {
			return true
		} else {
			panic(err)
		}
	}

	if !is_content_downloaded {
		return true
	}
	if banner_image_url != user.BannerImageUrl {
		return true
	}
	if profile_image_url != user.ProfileImageUrl {
		return true
	}
	return false
}

/**
 * Follow / unfollow a user.  Update the given User object's IsFollowed field.
 */
func (p Profile) SetUserFollowed(user *scraper.User, is_followed bool) {
	result, err := p.DB.Exec("update users set is_followed = ? where id = ?", is_followed, user.ID)
	if err != nil {
		panic(fmt.Sprintf("Error inserting user with handle %q: %s", user.Handle, err.Error()))
	}
	count, err := result.RowsAffected()
	if err != nil {
		panic("Unknown error: " + err.Error())
	}
	if count != 1 {
		panic(fmt.Sprintf("User with handle %q not found", user.Handle))
	}
	user.IsFollowed = is_followed
}

func (p Profile) NextFakeUserID() scraper.UserID {
	_, err := p.DB.Exec("update fake_user_sequence set latest_fake_id = latest_fake_id + 1")
	if err != nil {
		panic(err)
	}
	var ret scraper.UserID
	err = p.DB.QueryRow("select latest_fake_id from fake_user_sequence").Scan(&ret)
	if err != nil {
		panic(err)
	}
	return ret
}

func (p Profile) GetAllFollowedUsers() []scraper.UserHandle {
	rows, err := p.DB.Query("select handle from users where is_followed = 1")
	if err != nil {
		panic(err)
	}

	ret := []scraper.UserHandle{}

	var tmp scraper.UserHandle

	for rows.Next() {
		err = rows.Scan(&tmp)
		if err != nil {
			panic(err)
		}
		ret = append(ret, tmp)
	}

	return ret
}

func (p Profile) IsFollowing(handle scraper.UserHandle) bool {
	for _, follow := range p.GetAllFollowedUsers() {
		if follow == handle {
			return true
		}
	}
	return false
}
