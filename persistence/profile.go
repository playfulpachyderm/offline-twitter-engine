package persistence

import (
	_ "embed"
	"fmt"
	"os"
	"path"
	sql "github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v2"
)

//go:embed schema.sql
var sql_init string

type Settings struct{}

type Profile struct {
	ProfileDir string
	Settings   Settings
	DB         *sql.DB
}

/**
 * Custom error
 */
type ErrTargetAlreadyExists struct {
	target string
}

func (err ErrTargetAlreadyExists) Error() string {
	return fmt.Sprintf("Target already exists: %s", err.target)
}

/**
 * Create a new profile in the given location.
 * Fails if target location already exists (i.e., is a file or directory).
 *
 * args:
 * - target_dir: location to create the new profile directory
 *
 * returns:
 * - the newly created Profile
 */
func NewProfile(target_dir string) (Profile, error) {
	if file_exists(target_dir) {
		return Profile{}, ErrTargetAlreadyExists{target_dir}
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
		return Profile{}, err
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
		return Profile{}, err
	}
	err = os.WriteFile(settings_file, data, os.FileMode(0644))
	if err != nil {
		return Profile{}, err
	}

	// Create `profile_images`
	fmt.Printf("Creating............. %s/\n", profile_images_dir)
	err = os.Mkdir(profile_images_dir, os.FileMode(0755))
	if err != nil {
		return Profile{}, err
	}

	// Create `link_thumbnail_images`
	fmt.Printf("Creating............. %s/\n", link_thumbnails_dir)
	err = os.Mkdir(link_thumbnails_dir, os.FileMode(0755))
	if err != nil {
		return Profile{}, err
	}

	// Create `images`
	fmt.Printf("Creating............. %s/\n", images_dir)
	err = os.Mkdir(images_dir, os.FileMode(0755))
	if err != nil {
		return Profile{}, err
	}

	// Create `videos`
	fmt.Printf("Creating............. %s/\n", videos_dir)
	err = os.Mkdir(videos_dir, os.FileMode(0755))
	if err != nil {
		return Profile{}, err
	}

	// Create `video_thumbnails`
	fmt.Printf("Creating............. %s/\n", video_thumbnails_dir)
	err = os.Mkdir(video_thumbnails_dir, os.FileMode(0755))
	if err != nil {
		return Profile{}, err
	}

	return Profile{target_dir, settings, db}, nil
}

/**
 * Loads the profile at the given location.  Fails if the given directory is not a Profile.
 *
 * args:
 * - profile_dir: location to check for the profile
 *
 * returns:
 * - the loaded Profile
 */
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
		return Profile{}, err
	}
	settings := Settings{}
	err = yaml.Unmarshal(settings_data, &settings)
	if err != nil {
		return Profile{}, err
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
