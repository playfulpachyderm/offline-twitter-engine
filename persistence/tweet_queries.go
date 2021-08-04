package persistence

import (
    "fmt"
    "time"
    "strings"
    "database/sql"

    "offline_twitter/scraper"
)

func (p Profile) SaveTweet(t scraper.Tweet) error {
    db := p.DB

    tx, err := db.Begin()
    if err != nil {
        return err
    }
    _, err = db.Exec(`
        insert into tweets (id, user_id, text, posted_at, num_likes, num_retweets, num_replies, num_quote_tweets, in_reply_to, quoted_tweet, mentions, hashtags)
        values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            on conflict do update
           set num_likes=?,
               num_retweets=?,
               num_replies=?,
               num_quote_tweets=?
        `,
        t.ID, t.UserID, t.Text, t.PostedAt.Unix(), t.NumLikes, t.NumRetweets, t.NumReplies, t.NumQuoteTweets, t.InReplyTo, t.QuotedTweet, scraper.JoinArrayOfHandles(t.Mentions), strings.Join(t.Hashtags, ","),
        t.NumLikes, t.NumRetweets, t.NumReplies, t.NumQuoteTweets,
    )

    if err != nil {
        return err
    }
    for _, url := range t.Urls {
        _, err := db.Exec("insert into urls (tweet_id, text) values (?, ?) on conflict do nothing", t.ID, url)
        if err != nil {
            return err
        }
    }
    for _, image := range t.Images {
        err := p.SaveImage(image)
        if err != nil {
            return err
        }
    }
    for _, video := range t.Videos {
        _, err := p.SaveVideo(video)
        if err != nil {
            return err
        }
    }
    for _, hashtag := range t.Hashtags {
        _, err := db.Exec("insert into hashtags (tweet_id, text) values (?, ?) on conflict do nothing", t.ID, hashtag)
        if err != nil {
            return err
        }
    }

    err = tx.Commit()
    if err != nil {
        return err
    }
    return nil
}

func (p Profile) IsTweetInDatabase(id scraper.TweetID) bool {
    db := p.DB

    var dummy string
    err := db.QueryRow("select 1 from tweets where id = ?", id).Scan(&dummy)
    if err != nil {
        if err != sql.ErrNoRows {
            // A real error
            panic(err)
        }
        return false
    }
    return true
}

func (p Profile) attach_urls(t *scraper.Tweet) error {
    println("Attaching urls")
    stmt, err := p.DB.Prepare("select text from urls where tweet_id = ?")
    if err != nil {
        return err
    }
    defer stmt.Close()
    rows, err := stmt.Query(t.ID)
    if err != nil {
        return err
    }
    var url string
    for rows.Next() {
        err = rows.Scan(&url)
        if err != nil {
            return err
        }
        t.Urls = append(t.Urls, url)
        fmt.Printf("%v\n", t.Urls)
    }
    return nil
}

func (p Profile) GetTweetById(id scraper.TweetID) (scraper.Tweet, error) {
    db := p.DB

    stmt, err := db.Prepare(`
        select id, user_id, text, posted_at, num_likes, num_retweets, num_replies, num_quote_tweets, in_reply_to, quoted_tweet, mentions, hashtags
          from tweets
         where id = ?
    `)

    if err != nil {
        return scraper.Tweet{}, err
    }
    defer stmt.Close()

    var t scraper.Tweet
    var postedAt int
    var mentions string
    var hashtags string

    row := stmt.QueryRow(id)
    err = row.Scan(&t.ID, &t.UserID, &t.Text, &postedAt, &t.NumLikes, &t.NumRetweets, &t.NumReplies, &t.NumQuoteTweets, &t.InReplyTo, &t.QuotedTweet, &mentions, &hashtags)
    if err != nil {
        return t, err
    }

    t.PostedAt = time.Unix(int64(postedAt), 0)  // args are `seconds` and `nanoseconds`
    for _, m := range strings.Split(mentions, ",") {
        t.Mentions = append(t.Mentions, scraper.UserHandle(m))
    }
    t.Hashtags = strings.Split(hashtags, ",")

    imgs, err := p.GetImagesForTweet(t)
    if err != nil {
        return t, err
    }
    t.Images = imgs

    vids, err := p.GetVideosForTweet(t)
    if err != nil {
        return t, err
    }
    t.Videos = vids

    err = p.attach_urls(&t)
    return t, err
}


/**
 * Populate the `User` field on a tweet with an actual User
 */
func (p Profile) LoadUserFor(t *scraper.Tweet) error {
    if t.User != nil {
        // Already there, no need to load it
        return nil
    }

    user, err := p.GetUserByID(t.UserID)
    if err != nil {
        return err
    }
    t.User = &user
    return nil
}
