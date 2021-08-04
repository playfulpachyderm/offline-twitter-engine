package persistence

import (
    "database/sql"

    "offline_twitter/scraper"
)

/**
 * Save an Image
 *
 * args:
 * - img: the Image to save
 */
func (p Profile) SaveImage(img scraper.Image) error {
    _, err := p.DB.Exec(`
        insert into images (id, tweet_id, filename, is_downloaded)
                    values (?, ?, ?, ?)
               on conflict do update
                       set is_downloaded=?
        `,
        img.ID, img.TweetID, img.Filename, img.IsDownloaded,
        img.IsDownloaded,
    )
    return err
}

/**
 * Save a Video.  If it's a new Video (no rowid), does an insert; otherwise, does an update.
 *
 * args:
 * - img: the Video to save
 *
 * returns:
 * - the rowid
 */
func (p Profile) SaveVideo(vid scraper.Video) (sql.Result, error) {
    if vid.ID == 0 {
        // New image
        return p.DB.Exec("insert into videos (tweet_id, filename) values (?, ?) on conflict do nothing", vid.TweetID, vid.Filename)
    } else {
        // Updating an existing image
        return p.DB.Exec("update videos set filename=?, is_downloaded=? where rowid=?", vid.Filename, vid.IsDownloaded, vid.ID)
    }
}

/**
 * Get the list of images for a tweet
 */
func (p Profile) GetImagesForTweet(t scraper.Tweet) (imgs []scraper.Image, err error) {
    stmt, err := p.DB.Prepare("select id, filename, is_downloaded from images where tweet_id=?")
    if err != nil {
        return
    }
    defer stmt.Close()
    rows, err := stmt.Query(t.ID)
    if err != nil {
        return
    }
    var img scraper.Image

    for rows.Next() {
        err = rows.Scan(&img.ID, &img.Filename, &img.IsDownloaded)
        if err != nil {
            return
        }
        img.TweetID = t.ID
        imgs = append(imgs, img)
    }
    return
}


/**
 * Get the list of videos for a tweet
 */
func (p Profile) GetVideosForTweet(t scraper.Tweet) (vids []scraper.Video, err error) {
    stmt, err := p.DB.Prepare("select rowid, filename, is_downloaded from videos where tweet_id=?")
    if err != nil {
        return
    }
    defer stmt.Close()
    rows, err := stmt.Query(t.ID)
    if err != nil {
        return
    }
    var vid scraper.Video
    for rows.Next() {
        err = rows.Scan(&vid.ID, &vid.Filename, &vid.IsDownloaded)
        if err != nil {
            return
        }
        vid.TweetID = t.ID
        vids = append(vids, vid)
    }
    return
}
