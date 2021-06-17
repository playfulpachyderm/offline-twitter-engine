package main

import (
	"os"
	"fmt"
	"offline_twitter/scraper"
	"log"
)

const INCLUDE_REPLIES = true;

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Must provide tweet!")
	}

	user_handle := os.Args[1]

	user, err := scraper.GetUser(user_handle)
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Printf("%v\n", user)
}
