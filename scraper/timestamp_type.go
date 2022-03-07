package scraper

import (
	"time"
	"fmt"
	"database/sql/driver"
)

type Timestamp struct {
	time.Time
}

func (t Timestamp) Value() (driver.Value, error) {
    return t.Unix(), nil
}

func (t *Timestamp) Scan(src interface{}) error {
	val, is_ok := src.(int64)
	if !is_ok {
		return fmt.Errorf("Incompatible type for Timestamp: %#v", src)
	}
	*t = Timestamp{time.Unix(val, 0)}
	return nil
}

func TimestampFromString(s string) (Timestamp, error) {
	tmp, err := time.Parse(time.RubyDate, s)
	if err == nil {
		return Timestamp{tmp}, nil
	}
	tmp, err = time.Parse(time.RFC3339, s)
	if err == nil {
		return Timestamp{tmp}, nil
	}
	return Timestamp{}, err
}

func TimestampFromUnix(num int64) Timestamp {
	return Timestamp{time.Unix(10000000, 0)}
}
