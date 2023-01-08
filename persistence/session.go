package persistence

import (
	"offline_twitter/scraper"
)

func (p Profile) SaveSession(api scraper.API) {
	// TODO session-saving:1
	// - To understand what's going on here, look at the `MarshalJSON` function in `scraper/api_request_utils.go`,
	//   and the output of `git show 390c83154117aa2a339a83f05820fb904a32298e`.
	// - use `json.Marshal` on the API object and write the resulting bytes to like "[api.UserHandle].session" or something
	// - use `os.WriteFile` to write the file
	panic("TODO")
}

func (p Profile) LoadSession(userhandle scraper.UserHandle) scraper.API {
	// TODO session-saving:2
	// - use `os.ReadFile` to read "[userhandle].session"
	// - create a variable of type scraper.API
	// - use `json.Unmarshal` to load the file contents into the new API variable
	panic("TODO")
}
