package persistence

import (
	"encoding/json"
	"os"

	log "github.com/sirupsen/logrus"
)

func (p Profile) SaveSession(userhandle UserHandle, data []byte) {
	log.Debug("Profile Dir: " + p.ProfileDir)
	err := os.WriteFile(p.ProfileDir+"/"+string(userhandle+".session"), data, os.FileMode(0644))
	if err != nil {
		panic(err)
	}
}

func (p Profile) LoadSession(userhandle UserHandle, result json.Unmarshaler) {
	data, err := os.ReadFile(p.ProfileDir + "/" + string(userhandle+".session"))
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(data, result)
	if err != nil {
		panic(err)
	}
}
