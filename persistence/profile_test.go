package persistence_test

import (
	"testing"
	"os"
	"path"
	"errors"

	"offline_twitter/persistence"
)

// DUPE 1
func file_exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		panic(err)
	}
}

func isdir_map(is_dir bool) string {
	if is_dir {
		return "directory"
	}
	return "file"
}


func TestNewProfile(t *testing.T) {
	profile_path := "test_profiles/TestNewProfile"
	if !file_exists(profile_path) {
		err := os.Mkdir(profile_path, 0755)
		if err != nil {
			panic(err)
		}
	}

	contents, err := os.ReadDir(profile_path)
	if err != nil {
		panic(err)
	}
	if len(contents) != 0 {
		t.Fatalf("test_profile not empty at start of test!")
	}

	profile, err := persistence.NewProfile(profile_path)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if profile.ProfileDir != profile_path {
		t.Errorf("ProfileDir should be %s, but it is %s", profile_path, profile.ProfileDir)
	}
	if len(profile.UsersList) != 0 {
		t.Errorf("Expected empty users list, got %v instead", profile.UsersList)
	}

	// Check files were created
	contents, err = os.ReadDir(profile_path)
	if err != nil {
		panic(err)
	}
	if len(contents) != 6 {
		t.Fatalf("Expected 6 contents, got %d instead", len(contents))
	}

	expected_files := []struct {
		filename string
		isDir bool
	} {
		{"images", true},
		{"profile_images", true},
		{"settings.yaml", false},
		{"twitter.db", false},
		{"users.txt", false},
		{"videos", true},
	}

	for i, v := range expected_files {
		if contents[i].Name() != v.filename || contents[i].IsDir() != v.isDir {
			t.Fatalf("Expected `%s` to be a %s, but got %s [%s]", v.filename, isdir_map(v.isDir), contents[i].Name(), isdir_map(contents[i].IsDir()))
		}
	}
}

func TestLoadProfile(t *testing.T) {
	profile_path := "test_profiles/TestLoadProfile"
	if !file_exists(profile_path) {
		err := os.Mkdir(profile_path, 0755)
		if err != nil {
			panic(err)
		}
	}

	contents, err := os.ReadDir(profile_path)
	if err != nil {
		panic(err)
	}
	if len(contents) != 0 {
		t.Fatalf("test_profile not empty at start of test!")
	}

	_, err = persistence.NewProfile(profile_path)
	if err != nil {
		t.Fatalf(err.Error())
	}

	// Create some users
	err = os.WriteFile(path.Join(profile_path, "users.txt"), []byte("user1\nuser2\n"), 0644)
	if err != nil {
		t.Fatalf(err.Error())
	}

	profile, err := persistence.LoadProfile(profile_path)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if profile.ProfileDir != profile_path {
		t.Errorf("Expected profile path to be %q, but got %q", profile_path, profile.ProfileDir)
	}

	if len(profile.UsersList) != 2 {
		t.Errorf("Expected 2 users, got %v", profile.UsersList)
	}

}
