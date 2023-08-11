package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"path"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

// Save the given User to the database.
// If the User is already in the database, it will update most of its attributes (follower count, etc)
//
// args:
// - u: the User
func (p Profile) SaveUser(u *scraper.User) error {
	if u.IsNeedingFakeID {
		err := p.DB.QueryRow("select id from users where lower(handle) = lower(?)", u.Handle).Scan(&u.ID)
		if errors.Is(err, sql.ErrNoRows) {
			// We need to continue-- create a new fake user
			u.ID = p.NextFakeUserID()
		} else if err == nil {
			// We're done; a user exists with this handle already.  No need to fake anything, and we have no new data
			// to provide (since the ID is fake).
			// ID has already been scanned into the User, for use by the caller.
			return nil
		} else {
			// A real error occurred
			panic(fmt.Errorf("Error checking for existence of fake user with handle %q:\n  %w", u.Handle, err))
		}
	}

	_, err := p.DB.NamedExec(`
	    insert into users (id, display_name, handle, bio, following_count, followers_count, location, website, join_date, is_private,
	                       is_verified, is_banned, profile_image_url, profile_image_local_path, banner_image_url, banner_image_local_path,
	                       pinned_tweet_id, is_content_downloaded, is_id_fake)
	    values (:id, :display_name, :handle, :bio, :following_count, :followers_count, :location, :website, :join_date, :is_private,
	            :is_verified, :is_banned, :profile_image_url, :profile_image_local_path, :banner_image_url, :banner_image_local_path,
	            :pinned_tweet_id, :is_content_downloaded, :is_id_fake)
	        on conflict do update
	       set handle=:handle,
	           bio=:bio,
	           display_name=:display_name,
	           following_count=:following_count,
	           followers_count=:followers_count,
	           location=:location,
	           website=:website,
	           is_private=:is_private,
	           is_verified=:is_verified,
	           is_banned=:is_banned,
	           profile_image_url=:profile_image_url,
	           profile_image_local_path=:profile_image_local_path,
	           banner_image_url=:banner_image_url,
	           banner_image_local_path=:banner_image_local_path,
	           pinned_tweet_id=:pinned_tweet_id,
	           is_content_downloaded=(is_content_downloaded or :is_content_downloaded)
	    `,
		u,
	)
	if err != nil {
		return fmt.Errorf("Error executing SaveUser(%s):\n  %w", u.Handle, err)
	}

	return nil
}

// Check if the database has a User with the given user handle.
//
// args:
// - handle: the user handle to search for
//
// returns:
// - true if there is such a User in the database, false otherwise
func (p Profile) UserExists(handle scraper.UserHandle) bool {
	db := p.DB

	var dummy string
	err := db.QueryRow("select 1 from users where lower(handle) = lower(?)", handle).Scan(&dummy)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			// A real error
			panic(err)
		}
		return false
	}
	return true
}

// Retrieve a User from the database, by handle.
//
// args:
// - handle: the user handle to search for
//
// returns:
// - the User, if it exists
func (p Profile) GetUserByHandle(handle scraper.UserHandle) (scraper.User, error) {
	db := p.DB

	var ret scraper.User
	err := db.Get(&ret, `
	    select id, display_name, handle, bio, following_count, followers_count, location, website, join_date, is_private, is_verified,
	           is_banned, profile_image_url, profile_image_local_path, banner_image_url, banner_image_local_path, pinned_tweet_id,
	           is_content_downloaded, is_followed
	      from users
	     where lower(handle) = lower(?)
	`, handle)

	if errors.Is(err, sql.ErrNoRows) {
		return ret, ErrNotInDatabase{"User", handle}
	}
	return ret, nil
}

// Retrieve a User from the database, by user ID.
//
// args:
// - id: the user ID to search for
//
// returns:
// - the User, if it exists
func (p Profile) GetUserByID(id scraper.UserID) (scraper.User, error) {
	db := p.DB

	var ret scraper.User

	err := db.Get(&ret, `
	    select id, display_name, handle, bio, following_count, followers_count, location, website, join_date, is_private, is_verified,
	           is_banned, profile_image_url, profile_image_local_path, banner_image_url, banner_image_local_path, pinned_tweet_id,
	           is_content_downloaded, is_followed
	      from users
	     where id = ?
	`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ret, ErrNotInDatabase{"User", id}
	}
	if err != nil {
		panic(err)
	}
	return ret, nil
}

// Returns `true` if content download is needed, `false` otherwise
//
// If the user is banned, returns false because downloading will be impossible.
//
// If:
// - the user isn't in the DB at all (first time scraping), OR
// - `is_content_downloaded` is false in the DB, OR
// - the banner / profile image URL has changed from what the DB has
// then it needs to be downloaded.
//
// The `user` object will always have `is_content_downloaded` = false on every scrape.  This is
// why the No Worsening Principle is needed.
func (p Profile) CheckUserContentDownloadNeeded(user scraper.User) bool {
	row := p.DB.QueryRow(`select is_content_downloaded, profile_image_url, banner_image_url from users where id = ?`, user.ID)

	var is_content_downloaded bool
	var profile_image_url string
	var banner_image_url string
	err := row.Scan(&is_content_downloaded, &profile_image_url, &banner_image_url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return true
		} else {
			panic(err)
		}
	}

	if !is_content_downloaded {
		return true
	}

	banner_path := p.get_banner_image_output_path(user)
	if banner_path != "" && !file_exists(banner_path) {
		return true
	}
	profile_path := p.get_profile_image_output_path(user)
	return !file_exists(profile_path)
}

// Follow / unfollow a user.  Update the given User object's IsFollowed field.
func (p Profile) SetUserFollowed(user *scraper.User, is_followed bool) {
	result, err := p.DB.Exec("update users set is_followed = ? where id = ?", is_followed, user.ID)
	if err != nil {
		panic(fmt.Errorf("Error inserting user with handle %q:\n  %w", user.Handle, err))
	}
	count, err := result.RowsAffected()
	if err != nil {
		panic(fmt.Errorf("Unknown error retrieving row count:\n  %w", err))
	}
	if count != 1 {
		panic(fmt.Errorf("User with handle %q not found", user.Handle))
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

func (p Profile) IsFollowing(user scraper.User) bool {
	row := p.DB.QueryRow("select is_followed from users where id like ?", user.ID)
	var ret bool
	err := row.Scan(&ret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false
		}
		panic(err) // A real error
	}
	return ret
}

// Utility function to compute the path to save banner image to
func (p Profile) get_banner_image_output_path(u scraper.User) string {
	return path.Join(p.ProfileDir, "profile_images", u.BannerImageLocalPath)
}

// Utility function to compute the path to save profile image to
func (p Profile) get_profile_image_output_path(u scraper.User) string {
	if u.ProfileImageUrl == "" {
		return path.Join(p.ProfileDir, "profile_images", path.Base(scraper.DEFAULT_PROFILE_IMAGE_URL))
	}
	return path.Join(p.ProfileDir, "profile_images", u.ProfileImageLocalPath)
}

// Do a text search for users
func (p Profile) SearchUsers(s string) []scraper.User {
	var ret []scraper.User
	val := fmt.Sprintf("%%%s%%", s)
	err := p.DB.Select(&ret, `
	    select id, display_name, handle, bio, following_count, followers_count, location, website, join_date, is_private, is_verified,
	           is_banned, profile_image_url, profile_image_local_path, banner_image_url, banner_image_local_path, pinned_tweet_id,
	           is_content_downloaded, is_followed
	      from users
	     where handle like ?
	        or display_name like ?
	     order by followers_count desc
	`, val, val)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		panic(err)
	}
	return ret
}
