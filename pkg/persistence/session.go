package persistence

import (
	"encoding/json"

	"os"

	log "github.com/sirupsen/logrus"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

func (p Profile) SaveSession(api API) {
	data, err := json.Marshal(api)
	if err != nil {
		panic(err)
	}

	log.Debug("Profile Dir: " + p.ProfileDir)
	err = os.WriteFile(p.ProfileDir+"/"+string(api.UserHandle+".session"), data, os.FileMode(0644))
	if err != nil {
		panic(err)
	}
}

func (p Profile) LoadSession(userhandle UserHandle) API {
	data, err := os.ReadFile(p.ProfileDir + "/" + string(userhandle+".session"))
	if err != nil {
		panic(err)
	}

	var result API
	err = json.Unmarshal(data, &result)
	if err != nil {
		panic(err)
	}

	return result
}
