package scraper_test

import (
    "testing"
    "io/ioutil"
    "encoding/json"

    "offline_twitter/scraper"
)

func TestParseAPIUrlCard(t *testing.T) {
    data, err := ioutil.ReadFile("test_responses/url_card.json")
    if err != nil {
        panic(err)
    }
    var apiCard scraper.APICard
    err = json.Unmarshal(data, &apiCard)
    if err != nil {
        t.Fatal(err.Error())
    }
    url := scraper.ParseAPIUrlCard(apiCard)

    expected_domain := "reason.com"
    if url.Domain != expected_domain {
        t.Errorf("Expected %q, got %q", expected_domain, url.Domain)
    }
    expected_title := "L.A. Teachers Union Leader: 'There's No Such Thing As Learning Loss'"
    if url.Title != expected_title {
        t.Errorf("Expected %q, got %q", expected_title, url.Title)
    }
    expected_description := "\"It’s OK that our babies may not have learned all their times tables,\" says Cecily Myart-Cruz. \"They learned resilience.\""
    if url.Description != expected_description {
        t.Errorf("Expected %q, got %q", expected_description, url.Description)
    }
    expected_remote_url := "https://pbs.twimg.com/card_img/1434998862305968129/odDi9EqO?format=jpg&name=600x600"
    if url.ThumbnailRemoteUrl != expected_remote_url {
        t.Errorf("Expected %q, got %q", expected_remote_url, url.ThumbnailRemoteUrl)
    }
    expected_local_filename := "odDi9EqO_600x600.jpg"
    if url.ThumbnailLocalPath != expected_local_filename {
        t.Errorf("Expected %q, got %q", expected_local_filename, url.ThumbnailLocalPath)
    }
    expected_creator_id := scraper.UserID(155581583)
    if url.CreatorID != expected_creator_id {
        t.Errorf("Expected %d, got %d", expected_creator_id, url.CreatorID)
    }
    expected_site_id := scraper.UserID(16467567)
    if url.SiteID != expected_site_id {
        t.Errorf("Expected %d, got %d", expected_site_id, url.SiteID)
    }
    if url.IsContentDownloaded {
        t.Errorf("Expected it not to be downloaded, but it was")
    }
}

func TestParseAPIUrlCardWithPlayer(t *testing.T) {
    data, err := ioutil.ReadFile("test_responses/url_card_with_player.json")
    if err != nil {
        panic(err)
    }
    var apiCard scraper.APICard
    err = json.Unmarshal(data, &apiCard)
    if err != nil {
        t.Fatal(err.Error())
    }
    url := scraper.ParseAPIUrlCard(apiCard)

    expected_domain := "www.youtube.com"
    if url.Domain != expected_domain {
        t.Errorf("Expected %q, got %q", expected_domain, url.Domain)
    }
    expected_title := "The Politically Incorrect Guide to the Constitution (Starring Tom..."
    if url.Title != expected_title {
        t.Errorf("Expected %q, got %q", expected_title, url.Title)
    }
    expected_description := "Watch this episode on LBRY/Odysee: https://odysee.com/@capitalresearch:5/the-politically-incorrect-guide-to-the:8Watch this episode on Rumble: https://rumble..."
    if url.Description != expected_description {
        t.Errorf("Expected %q, got %q", expected_description, url.Description)
    }
    expected_remote_url := "https://pbs.twimg.com/card_img/1437849456423194639/_1t0btyt?format=jpg&name=800x320_1"
    if url.ThumbnailRemoteUrl != expected_remote_url {
        t.Errorf("Expected %q, got %q", expected_remote_url, url.ThumbnailRemoteUrl)
    }
    expected_local_filename := "_1t0btyt_800x320_1.jpg"
    if url.ThumbnailLocalPath != expected_local_filename {
        t.Errorf("Expected %q, got %q", expected_local_filename, url.ThumbnailLocalPath)
    }
    expected_site_id := scraper.UserID(10228272)
    if url.SiteID != expected_site_id {
        t.Errorf("Expected %d, got %d", expected_site_id, url.SiteID)
    }
    if url.IsContentDownloaded {
        t.Errorf("Expected it not to be downloaded, but it was")
    }
}
