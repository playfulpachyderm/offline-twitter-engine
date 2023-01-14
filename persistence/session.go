package persistence

import (
	"encoding/json"
	"offline_twitter/scraper"
	"os"
)

func (p Profile) SaveSession(api scraper.API) {
	data, err := json.Marshal(api)
	if err != nil {
		panic(err)
	}

	os.WriteFile(p.ProfileDir+"/"+string(api.UserHandle+".session"), data, os.FileMode(0644))
	if err != nil {
		panic(err)
	}
}

func (p Profile) LoadSession(userhandle scraper.UserHandle) scraper.API {
	data, err := os.ReadFile(p.ProfileDir + "/" + string(userhandle+".session"))
	if err != nil {
		panic(err)
	}

	var result scraper.API
	err = json.Unmarshal(data, &result)
	if err != nil {
		panic(err)
	}

	return result
}
