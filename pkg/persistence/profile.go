package persistence

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	sql "github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

//go:embed schema.sql
var sql_init string

//go:embed default_profile.png
var default_profile_image []byte

type Profile struct {
	ProfileDir string
	DB         *sql.DB
}

var ErrTargetAlreadyExists = fmt.Errorf("Target already exists")

// Create a new profile in the given location.
// Fails if target location already exists (i.e., is a file or directory).
//
// args:
// - target_dir: location to create the new profile directory
//
// returns:
// - the newly created Profile
func NewProfile(target_dir string) (Profile, error) {
	if file_exists(target_dir) {
		return Profile{}, fmt.Errorf("Could not create target %q:\n  %w", target_dir, ErrTargetAlreadyExists)
	}

	sqlite_file := filepath.Join(target_dir, "twitter.db")
	profile_images_dir := filepath.Join(target_dir, "profile_images")
	default_profile_image_file := filepath.Join(target_dir, "profile_images/default_profile.png")
	link_thumbnails_dir := filepath.Join(target_dir, "link_preview_images")
	images_dir := filepath.Join(target_dir, "images")
	videos_dir := filepath.Join(target_dir, "videos")
	video_thumbnails_dir := filepath.Join(target_dir, "video_thumbnails")

	// Create the directory
	fmt.Printf("Creating new profile: %s\n", target_dir)
	err := os.Mkdir(target_dir, os.FileMode(0755))
	if err != nil {
		return Profile{}, fmt.Errorf("Error creating directory %q:\n  %w", target_dir, err)
	}

	// Create `twitter.db`
	fmt.Printf("Creating............. %s\n", sqlite_file)
	db := sql.MustOpen("sqlite3", sqlite_file+"?_foreign_keys=on")
	db.MustExec(sql_init)

	// Create `profile_images`
	fmt.Printf("Creating............. %s/\n", profile_images_dir)
	err = os.Mkdir(profile_images_dir, os.FileMode(0755))
	if err != nil {
		return Profile{}, fmt.Errorf("Error creating %q:\n  %w", profile_images_dir, err)
	}
	// Put the default profile image in it
	fmt.Printf("Creating............. %s/\n", default_profile_image_file)
	err = os.WriteFile(default_profile_image_file, default_profile_image, os.FileMode(0644))
	if err != nil {
		return Profile{}, fmt.Errorf("Error creating default profile image file %q:\n  %w", default_profile_image, err)
	}

	// Create `link_thumbnail_images`
	fmt.Printf("Creating............. %s/\n", link_thumbnails_dir)
	err = os.Mkdir(link_thumbnails_dir, os.FileMode(0755))
	if err != nil {
		return Profile{}, fmt.Errorf("Error creating %q:\n  %w", link_thumbnails_dir, err)
	}

	// Create `images`
	fmt.Printf("Creating............. %s/\n", images_dir)
	err = os.Mkdir(images_dir, os.FileMode(0755))
	if err != nil {
		return Profile{}, fmt.Errorf("Error creating %q:\n  %w", images_dir, err)
	}

	// Create `videos`
	fmt.Printf("Creating............. %s/\n", videos_dir)
	err = os.Mkdir(videos_dir, os.FileMode(0755))
	if err != nil {
		return Profile{}, fmt.Errorf("Error creating %q:\n  %w", videos_dir, err)
	}

	// Create `video_thumbnails`
	fmt.Printf("Creating............. %s/\n", video_thumbnails_dir)
	err = os.Mkdir(video_thumbnails_dir, os.FileMode(0755))
	if err != nil {
		return Profile{}, fmt.Errorf("Error creating %q:\n  %w", video_thumbnails_dir, err)
	}

	return Profile{ProfileDir: target_dir, DB: db}, nil
}

// Loads the profile at the given location.  Fails if the given directory is not a Profile.
//
// args:
// - profile_dir: location to check for the profile
//
// returns:
// - the loaded Profile
func LoadProfile(profile_dir string) (Profile, error) {
	sqlite_file := filepath.Join(profile_dir, "twitter.db")
	if !file_exists(sqlite_file) {
		return Profile{}, fmt.Errorf("Invalid profile, could not find file: %s", sqlite_file)
	}

	db := sql.MustOpen("sqlite3", fmt.Sprintf("%s?_foreign_keys=on&_journal_mode=WAL", sqlite_file))

	ret := Profile{
		ProfileDir: profile_dir,
		DB:         db,
	}
	err := ret.check_and_update_version()
	return ret, err
}

func (p Profile) ListSessions() []scraper.UserHandle {
	result, err := filepath.Glob(filepath.Join(p.ProfileDir, "*.session"))
	if err != nil {
		panic(err)
	}
	ret := []scraper.UserHandle{}
	for _, filename := range result {
		ret = append(ret, scraper.UserHandle(filepath.Base(filename[:len(filename)-len(".session")])))
	}
	return ret
}
