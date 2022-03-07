package scraper

import (
	"fmt"
)

var (
    END_OF_FEED  = fmt.Errorf("End of feed")
    DOESNT_EXIST = fmt.Errorf("Doesn't exist")
    EXTERNAL_API_ERROR = fmt.Errorf("Unexpected result from external API")
    API_PARSE_ERROR = fmt.Errorf("Couldn't parse the result returned from the API")
)
