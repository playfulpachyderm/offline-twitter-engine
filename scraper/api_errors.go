package scraper

import (
	"fmt"
)

var END_OF_FEED  = fmt.Errorf("End of feed")
var DOESNT_EXIST = fmt.Errorf("Doesn't exist")
