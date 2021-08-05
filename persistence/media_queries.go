package persistence

import (
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
 * Save a Video
 *
 * args:
 * - img: the Video to save
 */
func (p Profile) SaveVideo(vid scraper.Video) error {
    _, err := p.DB.Exec(`
        insert into videos (id, tweet_id, remote_url, local_filename, is_downloaded)
                    values (?, ?, ?, ?, ?)
               on conflict do update
                       set is_downloaded=?
        `,
        vid.ID, vid.TweetID, vid.RemoteURL, vid.LocalFilename, vid.IsDownloaded,
        vid.IsDownloaded,
    )
    return err
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
    stmt, err := p.DB.Prepare("select id, remote_url, local_filename, is_downloaded from videos where tweet_id=?")
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
        err = rows.Scan(&vid.ID, &vid.RemoteURL, &vid.LocalFilename, &vid.IsDownloaded)
        if err != nil {
            return
        }
        vid.TweetID = t.ID
        vids = append(vids, vid)
    }
    return
}
