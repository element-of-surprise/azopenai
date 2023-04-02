package custom

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// UnixTime is a wrapper around time.Time that allows for marshaling and unmarshaling of
// unix timestamps.
type UnixTime struct {
	// Time is the underlying time.Time value.
	Time time.Time
}

// UnmarshalJSON unmarshals a unix timestamp into a time.Time.
func (u *UnixTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		u.Time = time.Time{}
		return nil
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("expected epoch time, got %q: %w", s, err)
	}
	u.Time = time.Unix(int64(i), 0)
	return nil
}

// MarshalJSON marshals a time.Time into a unix timestamp.
func (u *UnixTime) MarshalJSON() ([]byte, error) {
	if u.Time.IsZero() {
		return nil, nil
	}
	return []byte(fmt.Sprintf("%d", u.Time.Unix())), nil
}
