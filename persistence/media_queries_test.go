package persistence_test

import (
	"testing"
    "math/rand"
    "fmt"
    "time"

    "github.com/go-test/deep"

    "offline_twitter/scraper"
)


/**
 * Create an Image, save it, reload it, and make sure it comes back the same
 */
func TestSaveAndLoadImage(t *testing.T) {
    profile_path := "test_profiles/TestMediaQueries"
    profile := create_or_load_profile(profile_path)

    tweet := create_stable_tweet()

    // Create a fresh Image to test on
    rand.Seed(time.Now().UnixNano())
    filename := fmt.Sprint(rand.Int())
    img := scraper.Image{TweetID: tweet.ID, Filename: filename, IsDownloaded: false}

    // Save the Image
    result, err := profile.SaveImage(img)
    if err != nil {
        t.Fatalf("Failed to save the image: %s", err.Error())
    }
    last_insert, err := result.LastInsertId()
    if err != nil {
        t.Fatalf("last insert??? %s", err.Error())
    }
    img.ID = scraper.ImageID(last_insert)

    // Reload the Image
    imgs, err := profile.GetImagesForTweet(tweet)
    if err != nil {
        t.Fatalf("Could not load images: %s", err.Error())
    }

    var new_img scraper.Image
    for index := range imgs {
        if imgs[index].ID == img.ID {
            new_img = imgs[index]
        }
    }
    if new_img.ID != img.ID {
        t.Fatalf("Could not find image for some reason: %d, %d; %+v", new_img.ID, img.ID, imgs)
    }
    if diff := deep.Equal(img, new_img); diff != nil {
        t.Error(diff)
    }
}

/**
 * Change an Image, save the changes, reload it, and check if it comes back the same
 */
func TestModifyImage(t *testing.T) {
    profile_path := "test_profiles/TestMediaQueries"
    profile := create_or_load_profile(profile_path)

    tweet := create_stable_tweet()
    img := tweet.Images[0]

    if img.ID != 1 {
        t.Fatalf("Got the wrong image back: wanted ID %d, got %d", 1, img.ID)
    }

    img.Filename = "local/sdfjk.jpg"
    img.IsDownloaded = true

    // Save the changes
    result, err := profile.SaveImage(img)
    if err != nil {
        t.Error(err)
    }
    rows_affected, err := result.RowsAffected()
    if err != nil {
        t.Error(err)
    }
    if rows_affected != 1 {
        t.Errorf("Expected 1 row changed, but got %d", rows_affected)
    }

    // Reload it
    imgs, err := profile.GetImagesForTweet(tweet)
    if err != nil {
        t.Fatalf("Could not load images: %s", err.Error())
    }
    new_img := imgs[0]
    if new_img.ID != img.ID {
        t.Fatalf("Got the wrong image back: wanted ID %d, got %d", 1, new_img.ID)
    }

    if diff := deep.Equal(img, new_img); diff != nil {
        t.Error(diff)
    }
}


/**
 * Create an Video, save it, reload it, and make sure it comes back the same
 */
func TestSaveAndLoadVideo(t *testing.T) {
    profile_path := "test_profiles/TestMediaQueries"
    profile := create_or_load_profile(profile_path)

    tweet := create_stable_tweet()

    // Create a fresh Video to test on
    rand.Seed(time.Now().UnixNano())
    filename := fmt.Sprint(rand.Int())
    vid := scraper.Video{TweetID: tweet.ID, Filename: filename, IsDownloaded: false}

    // Save the Video
    result, err := profile.SaveVideo(vid)
    if err != nil {
        t.Fatalf("Failed to save the video: %s", err.Error())
    }
    last_insert, err := result.LastInsertId()
    if err != nil {
        t.Fatalf("last insert??? %s", err.Error())
    }
    vid.ID = scraper.VideoID(last_insert)

    // Reload the Video
    vids, err := profile.GetVideosForTweet(tweet)
    if err != nil {
        t.Fatalf("Could not load videos: %s", err.Error())
    }

    var new_vid scraper.Video
    for index := range vids {
        if vids[index].ID == vid.ID {
            new_vid = vids[index]
        }
    }
    if new_vid.ID != vid.ID {
        t.Fatalf("Could not find video for some reason: %d, %d; %+v", new_vid.ID, vid.ID, vids)
    }
    if diff := deep.Equal(vid, new_vid); diff != nil {
        t.Error(diff)
    }
}

/**
 * Change an Image, save the changes, reload it, and check if it comes back the same
 */
func TestModifyVideo(t *testing.T) {
    profile_path := "test_profiles/TestMediaQueries"
    profile := create_or_load_profile(profile_path)

    tweet := create_stable_tweet()
    vid := tweet.Videos[0]

    if vid.ID != 1 {
        t.Fatalf("Got the wrong video back: wanted ID %d, got %d", 1, vid.ID)
    }

    vid.Filename = "local/sdfjk.jpg"
    vid.IsDownloaded = true

    // Save the changes
    result, err := profile.SaveVideo(vid)
    if err != nil {
        t.Error(err)
    }
    rows_affected, err := result.RowsAffected()
    if err != nil {
        t.Error(err)
    }
    if rows_affected != 1 {
        t.Errorf("Expected 1 row changed, but got %d", rows_affected)
    }

    // Reload it
    vids, err := profile.GetVideosForTweet(tweet)
    if err != nil {
        t.Fatalf("Could not load videos: %s", err.Error())
    }
    new_vid := vids[0]
    if new_vid.ID != vid.ID {
        t.Fatalf("Got the wrong video back: wanted ID %d, got %d", 1, new_vid.ID)
    }

    if diff := deep.Equal(vid, new_vid); diff != nil {
        t.Error(diff)
    }
}
