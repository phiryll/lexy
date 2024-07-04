package internal_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/phiryll/lexy/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTime(t *testing.T) {
	codec := internal.TimeCodec

	locNYC, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	var zero time.Time
	// Before the epoch start on Jan 1, 1970
	past := time.Date(1900, 1, 2, 3, 4, 5, 600_000_000, time.UTC)
	local := time.Date(2000, 1, 2, 3, 4, 5, 6, time.Local)
	utc := time.Date(2000, 1, 2, 3, 4, 5, 600_000_000, time.UTC)
	nyc := time.Date(2000, 1, 2, 3, 4, 5, 999_999_999, locNYC)
	noZoneName, err := time.Parse(time.RFC3339Nano, "2000-01-02T03:04:05.6-05:00")
	require.NoError(t, err)

	for _, tt := range []struct {
		string
		time.Time
	}{
		{"zero", zero},
		{"past", past},
		{"utc", utc},
		{"local", local},
		{"nyc", nyc},
		{"no zone name", noZoneName},
	} {
		t.Run(tt.string, func(t *testing.T) {
			when := tt.Time
			_, expectedOffset := when.Zone()

			var b bytes.Buffer
			err := codec.Write(&b, when)
			require.NoError(t, err)

			got, err := codec.Read(bytes.NewReader(b.Bytes()))
			require.NoError(t, err)
			_, actualOffset := got.Zone()

			// Can't use == because it doesn't work for time.Time
			assert.True(t, when.Equal(got), "round trip")
			assert.Equal(t, expectedOffset, actualOffset, "offsets")
		})
	}

	testCodecFail[time.Time](t, codec, local)
}

func TestTimeOrder(t *testing.T) {
	codec := internal.TimeCodec

	// in order from west to east, expected sort order
	locFixed := time.FixedZone("fixed", -12*3600)
	locLA, err := time.LoadLocation("America/Los_Angeles")
	require.NoError(t, err)
	locNYC, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)

	// UTC times in order, pos/neg relative to epoch and 6/7 nanoseconds
	// test each of these in multiple timezones
	negUTC6 := time.Date(1900, 1, 2, 3, 4, 5, 6, time.UTC)
	negUTC7 := time.Date(1900, 1, 2, 3, 4, 5, 7, time.UTC)
	posUTC6 := time.Date(2000, 1, 2, 3, 4, 5, 6, time.UTC)
	posUTC7 := time.Date(2000, 1, 2, 3, 4, 5, 7, time.UTC)

	var prev []byte
	for i, tt := range []struct {
		string
		time.Time
	}{
		// Encodings should sort in this order.
		//
		// before and after the epoch start (Jan 1, 1970)
		// with different nanoseconds within the same second
		// with different timezones
		{"neg fixed 6", negUTC6.In(locFixed)},
		{"neg LA 6", negUTC6.In(locLA)},
		{"neg NYC 6", negUTC6.In(locNYC)},
		{"neg UTC 6", negUTC6},

		{"neg fixed 7", negUTC7.In(locFixed)},
		{"neg LA 7", negUTC7.In(locLA)},
		{"neg NYC 7", negUTC7.In(locNYC)},
		{"neg UTC 7", negUTC7},

		{"pos fixed 6", posUTC6.In(locFixed)},
		{"pos LA 6", posUTC6.In(locLA)},
		{"pos NYC 6", posUTC6.In(locNYC)},
		{"pos UTC 6", posUTC6},

		{"pos fixed 7", posUTC7.In(locFixed)},
		{"pos LA 7", posUTC7.In(locLA)},
		{"pos NYC 7", posUTC7.In(locNYC)},
		{"pos UTC 7", posUTC7},
	} {
		t.Run(tt.string, func(t *testing.T) {
			var b bytes.Buffer
			err := codec.Write(&b, tt.Time)
			require.NoError(t, err)
			current := b.Bytes()
			if i > 0 {
				assert.Less(t, prev, current)
			}
			prev = current
		})
	}
}
