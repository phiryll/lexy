package lexy_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/phiryll/lexy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTime(t *testing.T) {
	codec := lexy.Time()

	// West of UTC, negative timezone offset
	locNYC, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	// East of UTC, positive timezone offset
	locBerlin, err := time.LoadLocation("Europe/Berlin")
	require.NoError(t, err)
	var zero time.Time
	// Before the epoch start on Jan 1, 1970
	past := time.Date(1900, 1, 2, 3, 4, 5, 600_000_000, time.UTC)
	local := time.Date(2000, 1, 2, 3, 4, 5, 6, time.Local)
	utc := time.Date(2000, 1, 2, 3, 4, 5, 600_000_000, time.UTC)
	nyc := time.Date(2000, 1, 2, 3, 4, 5, 999_999_999, locNYC)
	berlin := time.Date(2000, 1, 2, 3, 4, 5, 999_999_999, locBerlin)
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
		{"berlin", berlin},
		{"no zone name", noZoneName},
	} {
		t.Run(tt.string, func(t *testing.T) {
			when := tt.Time
			_, expectedOffset := when.Zone()

			buf := bytes.NewBuffer([]byte{})
			err := codec.Write(buf, when)
			require.NoError(t, err)

			got, err := codec.Read(bytes.NewReader(buf.Bytes()))
			require.NoError(t, err)
			_, actualOffset := got.Zone()

			// Can't use == because it doesn't work for time.Time
			assert.True(t, when.Equal(got), "round trip")
			assert.Equal(t, expectedOffset, actualOffset, "offsets")
		})
	}

	testCodecFail(t, codec, local)
}

func TestTimeOrder(t *testing.T) {
	// in order from west to east, expected sort order,
	// UTC is between NYC and Berlin.
	locFixed := time.FixedZone("fixed", -12*3600)
	locLA, err := time.LoadLocation("America/Los_Angeles")
	require.NoError(t, err)
	locNYC, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	locBerlin, err := time.LoadLocation("Europe/Berlin")
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
		{"neg Berlin 6", negUTC6.In(locBerlin)},

		{"neg fixed 7", negUTC7.In(locFixed)},
		{"neg LA 7", negUTC7.In(locLA)},
		{"neg NYC 7", negUTC7.In(locNYC)},
		{"neg UTC 7", negUTC7},
		{"neg Berlin 7", negUTC7.In(locBerlin)},

		{"pos fixed 6", posUTC6.In(locFixed)},
		{"pos LA 6", posUTC6.In(locLA)},
		{"pos NYC 6", posUTC6.In(locNYC)},
		{"pos UTC 6", posUTC6},
		{"pos Berlin 6", posUTC6.In(locBerlin)},

		{"pos fixed 7", posUTC7.In(locFixed)},
		{"pos LA 7", posUTC7.In(locLA)},
		{"pos NYC 7", posUTC7.In(locNYC)},
		{"pos UTC 7", posUTC7},
		{"pos Berlin 7", posUTC7.In(locBerlin)},
	} {
		t.Run(tt.string, func(t *testing.T) {
			buf := bytes.NewBuffer([]byte{})
			err := lexy.Time().Write(buf, tt.Time)
			require.NoError(t, err)
			current := buf.Bytes()
			if i > 0 {
				assert.Less(t, prev, current)
			}
			prev = current
		})
	}
}
