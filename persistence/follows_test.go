package persistence_test

import (
	"testing"

	. "offline_twitter/persistence"
)


func TestIsFollowing(t *testing.T) {
	p := Profile{}
	p.UsersList = []Follow{
		Follow{Handle: "guy1"},
		Follow{Handle: "guy2"},
	}

	if p.IsFollowing("guy3") {
		t.Errorf("Should not be following guy3")
	}
	if !p.IsFollowing("guy2") {
		t.Errorf("Should be following guy2")
	}
}
