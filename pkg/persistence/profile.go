package persistence

import (
	_ "embed"
	"fmt"
	"os"
	"path"
	"path/filepath"

	sql "github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v2"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

//go:embed schema.sql
var sql_init string

type Settings struct{}

type Profile struct {
	ProfileDir string
	Settings   Settings
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

	settings_file := path.Join(target_dir, "settings.yaml")
	sqlite_file := path.Join(target_dir, "twitter.db")
	profile_images_dir := path.Join(target_dir, "profile_images")
	link_thumbnails_dir := path.Join(target_dir, "link_preview_images")
	images_dir := path.Join(target_dir, "images")
	videos_dir := path.Join(target_dir, "videos")
	video_thumbnails_dir := path.Join(target_dir, "video_thumbnails")

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
	InitializeDatabaseVersion(db)
	db.Mapper = reflectx.NewMapperFunc("db", ToSnakeCase)

	// Create `settings.yaml`
	fmt.Printf("Creating............. %s\n", settings_file)
	settings := Settings{}
	data, err := yaml.Marshal(&settings)
	if err != nil {
		return Profile{}, fmt.Errorf("Error YAML-marshalling [empty!] settings file:\n  %w", err)
	}
	err = os.WriteFile(settings_file, data, os.FileMode(0644))
	if err != nil {
		return Profile{}, fmt.Errorf("Error creating settings file %q:\n  %w", settings_file, err)
	}

	// Create `profile_images`
	fmt.Printf("Creating............. %s/\n", profile_images_dir)
	err = os.Mkdir(profile_images_dir, os.FileMode(0755))
	if err != nil {
		return Profile{}, fmt.Errorf("Error creating %q:\n  %w", profile_images_dir, err)
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

	return Profile{target_dir, settings, db}, nil
}

// Loads the profile at the given location.  Fails if the given directory is not a Profile.
//
// args:
// - profile_dir: location to check for the profile
//
// returns:
// - the loaded Profile
func LoadProfile(profile_dir string) (Profile, error) {
	settings_file := path.Join(profile_dir, "settings.yaml")
	sqlite_file := path.Join(profile_dir, "twitter.db")

	for _, file := range []string{
		settings_file,
		sqlite_file,
	} {
		if !file_exists(file) {
			return Profile{}, fmt.Errorf("Invalid profile, could not find file: %s", file)
		}
	}

	settings_data, err := os.ReadFile(settings_file)
	if err != nil {
		return Profile{}, fmt.Errorf("Error reading %q:\n  %w", settings_file, err)
	}
	settings := Settings{}
	err = yaml.Unmarshal(settings_data, &settings)
	if err != nil {
		return Profile{}, fmt.Errorf("Error YAML-unmarshalling %q:\n  %w", settings_file, err)
	}

	db := sql.MustOpen("sqlite3", fmt.Sprintf("%s?_foreign_keys=on&_journal_mode=WAL", sqlite_file))
	db.Mapper = reflectx.NewMapperFunc("db", ToSnakeCase)

	ret := Profile{
		ProfileDir: profile_dir,
		Settings:   settings,
		DB:         db,
	}
	err = ret.check_and_update_version()

	return ret, err
}

func (p Profile) ListSessions() []scraper.UserHandle {
	result, err := filepath.Glob(path.Join(p.ProfileDir, "*.session"))
	if err != nil {
		panic(err)
	}
	ret := []scraper.UserHandle{}
	for _, filename := range result {
		ret = append(ret, scraper.UserHandle(path.Base(filename[:len(filename)-len(".session")])))
	}
	return ret
}
