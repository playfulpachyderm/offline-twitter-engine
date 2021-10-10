package persistence

import (
    "database/sql"
    "time"
    "offline_twitter/scraper"
)

/**
 * Save the given User to the database.
 * If the User is already in the database, it will update most of its attributes (follower count, etc)
 *
 * args:
 * - u: the User
 */
func (p Profile) SaveUser(u scraper.User) error {
    db := p.DB

    _, err := db.Exec(`
        insert into users (id, display_name, handle, bio, following_count, followers_count, location, website, join_date, is_private, is_verified, profile_image_url, profile_image_local_path, banner_image_url, banner_image_local_path, pinned_tweet_id, is_content_downloaded)
        values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            on conflict do update
           set bio=?,
                  following_count=?,
                  followers_count=?,
                  location=?,
                  website=?,
                  is_private=?,
                  is_verified=?,
                  profile_image_url=?,
                  profile_image_local_path=?,
                  banner_image_url=?,
                  banner_image_local_path=?,
                  pinned_tweet_id=?,
                  is_content_downloaded=(is_content_downloaded or ?)
        `,
        u.ID, u.DisplayName, u.Handle, u.Bio, u.FollowingCount, u.FollowersCount, u.Location, u.Website, u.JoinDate.Unix(), u.IsPrivate, u.IsVerified, u.ProfileImageUrl, u.ProfileImageLocalPath, u.BannerImageUrl, u.BannerImageLocalPath, u.PinnedTweetID, u.IsContentDownloaded,
        u.Bio, u.FollowingCount, u.FollowersCount, u.Location, u.Website, u.IsPrivate, u.IsVerified, u.ProfileImageUrl, u.ProfileImageLocalPath, u.BannerImageUrl, u.BannerImageLocalPath, u.PinnedTweetID, u.IsContentDownloaded,
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
    var joinDate int64

    err := row.Scan(&u.ID, &u.DisplayName, &u.Handle, &u.Bio, &u.FollowingCount, &u.FollowersCount, &u.Location, &u.Website, &joinDate, &u.IsPrivate, &u.IsVerified, &u.ProfileImageUrl, &u.ProfileImageLocalPath, &u.BannerImageUrl, &u.BannerImageLocalPath, &u.PinnedTweetID, &u.IsContentDownloaded)
    if err != nil {
        return u, err
    }
    u.JoinDate = time.Unix(joinDate, 0)

    return u, nil
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
        select id, display_name, handle, bio, following_count, followers_count, location, website, join_date, is_private, is_verified, profile_image_url, profile_image_local_path, banner_image_url, banner_image_local_path, pinned_tweet_id, is_content_downloaded
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
        select id, display_name, handle, bio, following_count, followers_count, location, website, join_date, is_private, is_verified, profile_image_url, profile_image_local_path, banner_image_url, banner_image_local_path, pinned_tweet_id, is_content_downloaded
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
