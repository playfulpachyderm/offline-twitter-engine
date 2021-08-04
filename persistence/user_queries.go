package persistence

import (
    "fmt"
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

    tx, err := db.Begin()
    if err != nil {
        return err
    }
    _, err = db.Exec(`
        insert into users (id, display_name, handle, bio, following_count, followers_count, location, website, join_date, is_private, is_verified, profile_image_url, banner_image_url, pinned_tweet_id)
        values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            on conflict do update
           set bio=?,
                  following_count=?,
                  followers_count=?,
                  location=?,
                  website=?,
                  is_private=?,
                  is_verified=?,
                  profile_image_url=?,
                  banner_image_url=?,
                  pinned_tweet_id=?
        `,
        u.ID, u.DisplayName, u.Handle, u.Bio, u.FollowingCount, u.FollowersCount, u.Location, u.Website, u.JoinDate.Unix(), u.IsPrivate, u.IsVerified, u.ProfileImageUrl, u.BannerImageUrl, u.PinnedTweetID, u.Bio, u.FollowingCount, u.FollowersCount, u.Location, u.Website, u.IsPrivate, u.IsVerified, u.ProfileImageUrl, u.BannerImageUrl, u.PinnedTweetID,
    )
    if err != nil {
        return err
    }

    err = tx.Commit()
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
    var pinned_tweet_id int64

    err := row.Scan(&u.ID, &u.DisplayName, &u.Handle, &u.Bio, &u.FollowingCount, &u.FollowersCount, &u.Location, &u.Website, &joinDate, &u.IsPrivate, &u.IsVerified, &u.ProfileImageUrl, &u.BannerImageUrl, &pinned_tweet_id)
    if err != nil {
        return u, err
    }

    u.JoinDate = time.Unix(joinDate, 0)
    u.PinnedTweetID = scraper.TweetID(fmt.Sprint(pinned_tweet_id))

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
        select id, display_name, handle, bio, following_count, followers_count, location, website, join_date, is_private, is_verified, profile_image_url, banner_image_url, pinned_tweet_id
          from users
         where handle = ?
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
        select id, display_name, handle, bio, following_count, followers_count, location, website, join_date, is_private, is_verified, profile_image_url, banner_image_url, pinned_tweet_id
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
