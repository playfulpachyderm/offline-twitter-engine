package persistence

import (
	_ "embed"
	"fmt"
	"os"
	"path"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v2"

	"offline_twitter/scraper"
)

//go:embed schema.sql
var sql_init string

type Settings struct {}

type Profile struct {
	ProfileDir string
	UsersList []scraper.UserHandle
	Settings Settings
	DB *sql.DB
}


// Create a new profile in the given location.
// `path` is a directory
func NewProfile(target_dir string) (Profile, error) {
	user_list_file := path.Join(target_dir, "users.txt")
	settings_file := path.Join(target_dir, "settings.yaml")
	sqlite_file := path.Join(target_dir, "twitter.db")
	profile_images_dir := path.Join(target_dir, "profile_images")
	images_dir := path.Join(target_dir, "images")
	videos_dir := path.Join(target_dir, "videos")


	for _, file := range []string{
			user_list_file,
			settings_file,
			sqlite_file,
			profile_images_dir,
			images_dir,
			videos_dir,
		} {
		if file_exists(file) {
			return Profile{}, fmt.Errorf("File already exists: %s", file)
		}
	}

	// Create `twitter.db`
	fmt.Printf("Creating %s\n", sqlite_file)
	db, err := sql.Open("sqlite3", sqlite_file)
	if err != nil {
		return Profile{}, err
	}
	_, err = db.Exec(sql_init)
	if err != nil {
		return Profile{}, err
	}

	// Create `users.txt`
	fmt.Printf("Creating %s\n", user_list_file)
	err = os.WriteFile(user_list_file, []byte{}, os.FileMode(0644))
	if err != nil {
		return Profile{}, err
	}

	// Create `settings.yaml`
	fmt.Printf("Creating %s\n", settings_file)
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
	fmt.Printf("Creating %s/\n", profile_images_dir)
	err = os.Mkdir(profile_images_dir, os.FileMode(0755))
	if err != nil {
		return Profile{}, err
	}

	// Create `images`
	fmt.Printf("Creating %s/\n", images_dir)
	err = os.Mkdir(images_dir, os.FileMode(0755))
	if err != nil {
		return Profile{}, err
	}

	// Create `videos`
	fmt.Printf("Creating %s/\n", videos_dir)
	err = os.Mkdir(videos_dir, os.FileMode(0755))
	if err != nil {
		return Profile{}, err
	}

	return Profile{target_dir, []scraper.UserHandle{}, settings, db}, nil
}


func LoadProfile(profile_dir string) (Profile, error) {
	user_list_file := path.Join(profile_dir, "users.txt")
	settings_file := path.Join(profile_dir, "settings.yaml")
	sqlite_file := path.Join(profile_dir, "twitter.db")

	for _, file := range []string{
			user_list_file,
			settings_file,
			sqlite_file,
		} {
		if !file_exists(file) {
			return Profile{}, fmt.Errorf("Invalid profile, could not find file: %s", file)
		}
	}

	users_data, err := os.ReadFile(user_list_file)
	if err != nil {
		return Profile{}, err
	}
	users_list := parse_users_file(users_data)

	settings_data, err := os.ReadFile(settings_file)
	if err != nil {
		return Profile{}, err
	}
	settings := Settings{}
	err = yaml.Unmarshal(settings_data, &settings)
	if err != nil {
		return Profile{}, err
	}
	db, err := sql.Open("sqlite3", sqlite_file)
	if err != nil {
		return Profile{}, err
	}

	return Profile{profile_dir, users_list, settings, db}, nil
}
