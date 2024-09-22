package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"path"

	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

type ErrConflictingUserHandle struct {
	ConflictingUserID scraper.UserID
}

func (e ErrConflictingUserHandle) Error() string {
	return fmt.Sprintf("active user with given handle already exists (id: %d)", e.ConflictingUserID)
}

const USERS_ALL_SQL_FIELDS = `
		id, display_name, handle, bio, following_count, followers_count, location, website, join_date, is_private, is_verified,
		is_banned, is_deleted, profile_image_url, profile_image_local_path, banner_image_url, banner_image_local_path,
		pinned_tweet_id, is_content_downloaded, is_followed`

// User save strategy:
//
//  1. Check if the user needs a fake ID; if so, assign one
//  2. Try to execute an update
//     2a. if the user is banned or deleted, don't overwrite other fields, blanking them
//     2b. if the user exists but `handle` conflicts with an active user, do conflict handling
//  3. If the user doesn't already exist, execute an insert.  Do conflict handling if applicable
//
// Conflict handling:
//
//  1. Look up the ID of the user with conflicting handle
//  2. TODO: handle case where the previous user has a fake ID
//     May have to rescrape that user's tweets to figure out if they're the same user or not.
//     Strategy 1: assume they're the same users
//     - Execute a full update on the old user, including their ID (we have a real ID for them now)
//     - Update all the other tables (tweets, follows, lists, etc) with the new ID
//     Strategy 2: assume they're different users
//     - Mark the old user as deactivated and be done with it
//  3. Mark the old user as deactivated, eliminating the conflict
//  4. Re-save the new user
//  5. Return an ErrConflictingUserHandle, notifying the caller of the conflict
func (p Profile) SaveUser(u *scraper.User) error {
	// First, check if the user needs a fake ID, and generate one if needed
	if u.IsNeedingFakeID {
		// User is fake; check if we already have them, in order to proceed
		err := p.DB.QueryRow("select id from users_by_handle where lower(handle) = lower(?)", u.Handle).Scan(&u.ID)
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

	// Handler function to deal with UNIQUE constraint violations on `handle`.
	//
	// We know the UNIQUE violation must be on `handle`, because we checked for users with this ID
	// above (`update` query).
	handle_conflict := func() error {
		var old_user scraper.User
		err := p.DB.Get(&old_user,
			`select id, is_id_fake from users where handle = ? and is_banned = 0 and is_deleted = 0`,
			u.Handle,
		)
		if err != nil {
			panic(err)
		}
		if old_user.IsIdFake {
			panic("TODO: user with fake ID")
		} else {
			// 1. The being-saved user ID doesn't exist yet (or was previously inactive)
			// 2. There's another user with the same handle who's currently considered active
			// 3. Their ID is not fake.
			// 4. The being-saved user is also understood to be active (otherwise a UNIQUE handle
			//    conflict wouldn't have occurred)
			//
			// Since we're saving an active user, the old user is presumably no longer active.
			// They will need to be rescraped when posssible, to find out what's going on.  For
			// now, we will just mark them as deleted.
			_, err := p.DB.Exec(`update users set is_deleted=1 where id = ?`, old_user.ID)
			if err != nil {
				panic(err)
			}
			// Now we can save our new user.  Should succeed since the conflict is cleared:
			err = p.SaveUser(u)
			if err != nil {
				panic(err)
			}
			// Notify caller of the duplicate for rescraping
			return ErrConflictingUserHandle{ConflictingUserID: old_user.ID}
		}
	}

	// Try to treat it like an `update` and see if it works
	var result sql.Result
	var err error
	if u.IsBanned || u.IsDeleted {
		// If user is banned or deleted, it's a stub, so don't update other fields
		result, err = p.DB.NamedExec(`update users set is_deleted=:is_deleted, is_banned=:is_banned where id = :id`, u)
	} else {
		// This could be re-activating a previously deleted / banned user
		result, err = p.DB.NamedExec(`
		    update users
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
		           is_deleted=:is_deleted,
		           profile_image_url=:profile_image_url,
		           profile_image_local_path=:profile_image_local_path,
		           banner_image_url=:banner_image_url,
		           banner_image_local_path=:banner_image_local_path,
		           pinned_tweet_id=:pinned_tweet_id,
		           is_content_downloaded=(is_content_downloaded or :is_content_downloaded)
		     where id = :id
		`, u)
	}
	if err != nil {
		// Check for UNIQUE constraint violation on `handle` field
		var sqliteErr sqlite3.Error
		is_ok := errors.As(err, &sqliteErr)
		if is_ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return handle_conflict()
		} else {
			// Unexpected error
			return fmt.Errorf("Error executing SaveUser(%s):\n  %w", u.Handle, err)
		}
	}
	// If a row was updated, then the User already exists and was updated successfully; we're done
	rows_affected, err := result.RowsAffected()
	if err != nil {
		panic(err)
	}
	if rows_affected > 0 {
		return nil
	}

	// It's a new user.  Try to insert it:
	_, err = p.DB.NamedExec(`
	    insert into users (id, display_name, handle, bio, following_count, followers_count, location, website, join_date, is_private,
	                       is_verified, is_banned, is_deleted, profile_image_url, profile_image_local_path, banner_image_url,
	                       banner_image_local_path, pinned_tweet_id, is_content_downloaded, is_id_fake)
	    values (:id, :display_name, :handle, :bio, :following_count, :followers_count, :location, :website, :join_date, :is_private,
	            :is_verified, :is_banned, :is_deleted, :profile_image_url, :profile_image_local_path, :banner_image_url,
	            :banner_image_local_path, :pinned_tweet_id, :is_content_downloaded, :is_id_fake)
	    `,
		u,
	)
	if err == nil {
		// It worked; user is inserted, we're done
		return nil
	}

	// If execution reaches this point, then an error has occurred; err is not nil.
	// Check if it's a UNIQUE CONSTRAINT FAILED:
	var sqliteErr sqlite3.Error
	is_ok := errors.As(err, &sqliteErr)
	if is_ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique { // Conflict detected
		return handle_conflict()
	} else {
		// Some other error
		return fmt.Errorf("Error executing SaveUser(%s):\n  %w", u.Handle, err)
	}
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
	    select `+USERS_ALL_SQL_FIELDS+`
	      from users_by_handle
	     where lower(handle) = lower(?)
	`, handle)

	if errors.Is(err, sql.ErrNoRows) {
		return ret, ErrNotInDatabase
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
	    select `+USERS_ALL_SQL_FIELDS+`
	      from users
	     where id = ?
	`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ret, ErrNotInDatabase
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

// TODO: This is only used in checking whether the media downloader should get the big or small version of
// a profile image.  That should be rewritten
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
	q, args, err := sqlx.Named(`
		select `+USERS_ALL_SQL_FIELDS+`
	      from users
	     where handle like :val
	        or display_name like :val
	     order by handle like :val or display_name like :val desc,
	              followers_count desc
	     `,
		struct {
			Val string `db:"val"`
		}{fmt.Sprintf("%%%s%%", s)},
	)
	if err != nil {
		panic(err)
	}
	err = p.DB.Select(&ret, q, args...)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		panic(err)
	}
	return ret
}
