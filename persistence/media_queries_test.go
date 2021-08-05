package persistence_test

import (
	"testing"
    "math/rand"
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
    img := create_image_from_id(rand.Int())
    img.TweetID = tweet.ID

    // Save the Image
    err := profile.SaveImage(img)
    if err != nil {
        t.Fatalf("Failed to save the image: %s", err.Error())
    }

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

    if img.ID != -1 {
        t.Fatalf("Got the wrong image back: wanted ID %d, got %d", -1, img.ID)
    }

    img.IsDownloaded = true

    // Save the changes
    err := profile.SaveImage(img)
    if err != nil {
        t.Error(err)
    }

    // Reload it
    imgs, err := profile.GetImagesForTweet(tweet)
    if err != nil {
        t.Fatalf("Could not load images: %s", err.Error())
    }
    new_img := imgs[0]
    if new_img.ID != img.ID {
        t.Fatalf("Got the wrong image back: wanted ID %d, got %d", -1, new_img.ID)
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
    vid := create_video_from_id(rand.Int())
    vid.TweetID = tweet.ID

    // Save the Video
    err := profile.SaveVideo(vid)
    if err != nil {
        t.Fatalf("Failed to save the video: %s", err.Error())
    }

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

    if vid.ID != -1 {
        t.Fatalf("Got the wrong video back: wanted ID %d, got %d", -1, vid.ID)
    }

    vid.IsDownloaded = true

    // Save the changes
    err := profile.SaveVideo(vid)
    if err != nil {
        t.Error(err)
    }

    // Reload it
    vids, err := profile.GetVideosForTweet(tweet)
    if err != nil {
        t.Fatalf("Could not load videos: %s", err.Error())
    }
    new_vid := vids[0]
    if new_vid.ID != vid.ID {
        t.Fatalf("Got the wrong video back: wanted ID %d, got %d", -1, new_vid.ID)
    }

    if diff := deep.Equal(vid, new_vid); diff != nil {
        t.Error(diff)
    }
}
