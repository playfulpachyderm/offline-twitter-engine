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
        insert into images (id, tweet_id, width, height, remote_url, local_filename, is_downloaded)
                    values (?, ?, ?, ?, ?, ?, ?)
               on conflict do update
                       set is_downloaded=(is_downloaded or ?)
        `,
        img.ID, img.TweetID, img.Width, img.Height, img.RemoteURL, img.LocalFilename, img.IsDownloaded,
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
        insert into videos (id, tweet_id, width, height, remote_url, local_filename, is_downloaded, is_gif)
                    values (?, ?, ?, ?, ?, ?, ?, ?)
               on conflict do update
                       set is_downloaded=(is_downloaded or ?)
        `,
        vid.ID, vid.TweetID, vid.Width, vid.Height, vid.RemoteURL, vid.LocalFilename, vid.IsDownloaded, vid.IsGif,
        vid.IsDownloaded,
    )
    return err
}

/**
 * Save an Url
 */
func (p Profile) SaveUrl(url scraper.Url) error {
    _, err := p.DB.Exec(`
        insert into urls (tweet_id, domain, text, title, description, creator_id, site_id, thumbnail_width, thumbnail_height, thumbnail_remote_url, thumbnail_local_path, has_card, has_thumbnail, is_content_downloaded)
                  values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
             on conflict do update
                     set is_content_downloaded=(is_content_downloaded or ?)
        `,
        url.TweetID, url.Domain, url.Text, url.Title, url.Description,  url.CreatorID, url.SiteID, url.ThumbnailWidth, url.ThumbnailHeight, url.ThumbnailRemoteUrl, url.ThumbnailLocalPath, url.HasCard, url.HasThumbnail, url.IsContentDownloaded,
        url.IsContentDownloaded,
    )
    return err
}

/**
 * Get the list of images for a tweet
 */
func (p Profile) GetImagesForTweet(t scraper.Tweet) (imgs []scraper.Image, err error) {
    stmt, err := p.DB.Prepare("select id, width, height, remote_url, local_filename, is_downloaded from images where tweet_id=?")
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
        err = rows.Scan(&img.ID, &img.Width, &img.Height, &img.RemoteURL, &img.LocalFilename, &img.IsDownloaded)
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
    stmt, err := p.DB.Prepare("select id, width, height, remote_url, local_filename, is_downloaded, is_gif from videos where tweet_id=?")
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
        err = rows.Scan(&vid.ID, &vid.Width, &vid.Height, &vid.RemoteURL, &vid.LocalFilename, &vid.IsDownloaded, &vid.IsGif)
        if err != nil {
            return
        }
        vid.TweetID = t.ID
        vids = append(vids, vid)
    }
    return
}

/**
 * Get the list of Urls for a Tweet
 */
func (p Profile) GetUrlsForTweet(t scraper.Tweet) (urls []scraper.Url, err error) {
    stmt, err := p.DB.Prepare("select domain, text, title, description, creator_id, site_id, thumbnail_width, thumbnail_height, thumbnail_remote_url, thumbnail_local_path, has_card, has_thumbnail, is_content_downloaded from urls where tweet_id=? order by rowid")
    if err != nil {
        return
    }
    defer stmt.Close()
    rows, err := stmt.Query(t.ID)
    if err != nil {
        return
    }
    var url scraper.Url
    for rows.Next() {
        err = rows.Scan(&url.Domain, &url.Text, &url.Title, &url.Description, &url.CreatorID, &url.SiteID, &url.ThumbnailWidth, &url.ThumbnailHeight, &url.ThumbnailRemoteUrl, &url.ThumbnailLocalPath, &url.HasCard, &url.HasThumbnail, &url.IsContentDownloaded)
        if err != nil {
            return
        }
        url.TweetID = t.ID
        urls = append(urls, url)
    }
    return
}
