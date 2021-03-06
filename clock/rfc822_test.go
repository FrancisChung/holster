package clock

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Time RFC822Time `json:"ts"`
}

func TestRFC822New(t *testing.T) {
	stdTime, err := Parse(RFC3339, "2019-08-29T11:20:07.123456+03:00")
	assert.NoError(t, err)

	rfc822TimeFromTime := NewRFC822Time(stdTime)
	rfc822TimeFromUnix := NewRFC822TimeFromUnix(stdTime.Unix())
	assert.True(t, rfc822TimeFromTime.Equal(rfc822TimeFromUnix.Time),
		"want=%s, got=%s", rfc822TimeFromTime.Time, rfc822TimeFromUnix.Time)

	assert.Equal(t, "Thu, 29 Aug 2019 11:20:07 MSK", rfc822TimeFromTime.String())
	assert.Equal(t, "Thu, 29 Aug 2019 08:20:07 UTC", rfc822TimeFromUnix.String())
}

// NewRFC822Time truncates to second precision.
func TestRFC822SecondPrecision(t *testing.T) {
	stdTime1, err := Parse(RFC3339, "2019-08-29T11:20:07.111111+03:00")
	assert.NoError(t, err)
	stdTime2, err := Parse(RFC3339, "2019-08-29T11:20:07.999999+03:00")
	assert.NoError(t, err)
	assert.False(t, stdTime1.Equal(stdTime2))

	rfc822Time1 := NewRFC822Time(stdTime1)
	rfc822Time2 := NewRFC822Time(stdTime2)
	assert.True(t, rfc822Time1.Equal(rfc822Time2.Time),
		"want=%s, got=%s", rfc822Time1.Time, rfc822Time2.Time)
}

// Marshaled representation is truncated down to second precision.
func TestRFC822Marshaling(t *testing.T) {
	stdTime, err := Parse(RFC3339Nano, "2019-08-29T11:20:07.123456789+03:30")
	assert.NoError(t, err)

	ts := testStruct{Time: NewRFC822Time(stdTime)}
	encoded, err := json.Marshal(&ts)
	assert.NoError(t, err)
	assert.Equal(t, `{"ts":"Thu, 29 Aug 2019 11:20:07 +0330"}`, string(encoded))
}

func TestRFC822Unmarshaling(t *testing.T) {
	for i, tc := range []struct {
		inRFC822   string
		outRFC3339 string
		outRFC822  string
	}{{
		inRFC822:   "Thu, 29 Aug 2019 11:20:07 GMT",
		outRFC3339: "2019-08-29T11:20:07Z",
		outRFC822:  "Thu, 29 Aug 2019 11:20:07 GMT",
	}, {
		inRFC822:   "Thu, 29 Aug 2019 11:20:07 MSK",
		outRFC3339: "2019-08-29T11:20:07+03:00",
		outRFC822:  "Thu, 29 Aug 2019 11:20:07 MSK",
	}, {
		inRFC822:   "Thu, 29 Aug 2019 11:20:07 -0000",
		outRFC3339: "2019-08-29T11:20:07Z",
		outRFC822:  "Thu, 29 Aug 2019 11:20:07 -0000",
	}, {
		inRFC822:   "Thu, 29 Aug 2019 11:20:07 +0000",
		outRFC3339: "2019-08-29T11:20:07Z",
		outRFC822:  "Thu, 29 Aug 2019 11:20:07 +0000",
	}, {
		inRFC822:   "Thu, 29 Aug 2019 11:20:07 +0300",
		outRFC3339: "2019-08-29T11:20:07+03:00",
		outRFC822:  "Thu, 29 Aug 2019 11:20:07 MSK",
	}, {
		inRFC822:   "Thu, 29 Aug 2019 11:20:07 +0330",
		outRFC3339: "2019-08-29T11:20:07+03:30",
		outRFC822:  "Thu, 29 Aug 2019 11:20:07 +0330",
	}} {
		tcDesc := fmt.Sprintf("Test case #%d: %v", i, tc)
		var ts testStruct

		inEncoded := []byte(fmt.Sprintf(`{"ts":"%s"}`, tc.inRFC822))
		err := json.Unmarshal(inEncoded, &ts)
		assert.NoError(t, err, tcDesc)
		assert.Equal(t, tc.outRFC3339, ts.Time.Format(RFC3339), tcDesc)

		actualEncoded, err := json.Marshal(&ts)
		assert.NoError(t, err, tcDesc)
		outEncoded := fmt.Sprintf(`{"ts":"%s"}`, tc.outRFC822)
		assert.Equal(t, outEncoded, string(actualEncoded), tcDesc)
	}
}

func TestRFC822UnmarshalingError(t *testing.T) {
	for _, tc := range []struct {
		inEncoded string
		outError  string
	}{{
		inEncoded: `{"ts": "Thu, 29 Aug 2019 11:20:07"}`,
		outError:  `parsing time "Thu, 29 Aug 2019 11:20:07" as "Mon, 02 Jan 2006 15:04:05 -0700": cannot parse "" as "-0700"`,
	}, {
		inEncoded: `{"ts": "foo"}`,
		outError:  `parsing time "foo" as "Mon, 02 Jan 2006 15:04:05 MST": cannot parse "foo" as "Mon"`,
	}, {
		inEncoded: `{"ts": 42}`,
		outError:  "invalid syntax",
	}} {
		var ts testStruct
		err := json.Unmarshal([]byte(tc.inEncoded), &ts)
		assert.EqualError(t, err, tc.outError)
	}
}
