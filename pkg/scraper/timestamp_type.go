package scraper

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type Timestamp struct {
	time.Time
}

func (t Timestamp) Value() (driver.Value, error) {
	return t.UnixMilli(), nil
}

func (t *Timestamp) Scan(src interface{}) error {
	val, is_ok := src.(int64)
	if !is_ok {
		return fmt.Errorf("Incompatible type for Timestamp: %#v", src)
	}
	*t = Timestamp{time.UnixMilli(val)}
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
	return Timestamp{}, fmt.Errorf("Error parsing timestamp:\n  %w", err)
}

func TimestampFromUnix(num int64) Timestamp {
	return Timestamp{time.Unix(num, 0)}
}
func TimestampFromUnixMilli(num int64) Timestamp {
	return Timestamp{time.UnixMilli(num)}
}
