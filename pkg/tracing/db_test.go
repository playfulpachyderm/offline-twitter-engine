package tracing_test

import (
	"fmt"
	"math/rand"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/tracing"
)

var test_db DB

func init() {
	i := rand.Uint32()
	var err error
	test_db_path := fmt.Sprintf("../../sample_data/profile/testtracing-%d.db", i)
	test_db, err = DBCreate(test_db_path)
	if err != nil {
		panic(err)
	}

	_, err = DBConnect(test_db_path)
	if err != nil {
		panic(err)
	}
}
