package clock

import (
	"strconv"
)

// Allows seamless JSON encoding/decoding of rfc822 formatted timestamps.
// https://www.ietf.org/rfc/rfc822.txt section 5.
type RFC822Time struct {
	Time
}

// NewRFC822Time creates RFC822Time from a standard Time. The created value is
// truncated down to second precision because RFC822 does not allow for better.
func NewRFC822Time(t Time) RFC822Time {
	return RFC822Time{Time: t.Truncate(Second)}
}

// NewRFC822Time creates RFC822Time from a Unix timestamp (seconds from Epoch).
func NewRFC822TimeFromUnix(timestamp int64) RFC822Time {
	return RFC822Time{Time: Unix(timestamp, 0).UTC()}
}

func (t RFC822Time) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(t.Format(RFC1123))), nil
}

func (t *RFC822Time) UnmarshalJSON(s []byte) error {
	q, err := strconv.Unquote(string(s))
	if err != nil {
		return err
	}
	if t.Time, err = Parse(RFC1123, q); err == nil {
		return nil
	}
	if err, ok := err.(*ParseError); !ok || err.LayoutElem != "MST" {
		return err
	}
	if t.Time, err = Parse(RFC1123Z, q); err != nil {
		return err
	}
	return nil
}

func (t RFC822Time) String() string {
	return t.Format(RFC1123)
}
