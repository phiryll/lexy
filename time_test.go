package lexy_test

import (
	"testing"
	"time"

	"github.com/phiryll/lexy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func timeTestCases() []testCase[time.Time] {
	// West of UTC, negative timezone offset
	locNYC, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}
	// East of UTC, positive timezone offset
	locBerlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		panic(err)
	}
	var zero time.Time
	// Before the epoch start on Jan 1, 1970
	past := time.Date(1900, 1, 2, 3, 4, 5, 600_000_000, time.UTC)
	local := time.Date(2000, 1, 2, 3, 4, 5, 6, time.Local)
	utc := time.Date(2000, 1, 2, 3, 4, 5, 600_000_000, time.UTC)
	nyc := time.Date(2000, 1, 2, 3, 4, 5, 999_999_999, locNYC)
	berlin := time.Date(2000, 1, 2, 3, 4, 5, 999_999_999, locBerlin)
	noZoneName, err := time.Parse(time.RFC3339Nano, "2000-01-02T03:04:05.6-05:00")
	if err != nil {
		panic(err)
	}
	return []testCase[time.Time]{
		{"zero", zero, nil},
		{"past", past, nil},
		{"utc", utc, nil},
		{"local", local, nil},
		{"nyc", nyc, nil},
		{"berlin", berlin, nil},
		{"no zone name", noZoneName, nil},
	}
}

func TestTimeWithZoneNames(t *testing.T) {
	t.Parallel()
	codec := lexy.Time()
	assert.False(t, codec.RequiresTerminator())
	for _, tt := range timeTestCases() {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			when := tt.value
			_, expectedOffset := when.Zone()
			buf := codec.Append(nil, when)
			got, _ := codec.Get(buf)
			_, actualOffset := got.Zone()
			// Can't use == because it doesn't work for time.Time
			assert.True(t, when.Equal(got), "round trip")
			assert.Equal(t, expectedOffset, actualOffset, "offsets")
		})
	}
}

func TestTimeStripZoneName(t *testing.T) {
	t.Parallel()
	codec := lexy.Time()
	var testCases []testCase[time.Time]
	for _, tt := range timeTestCases() {
		_, offset := tt.value.Zone()
		testCases = append(testCases, testCase[time.Time]{
			tt.name,
			tt.value.In(time.FixedZone("", offset)),
			tt.data,
		})
	}
	testCodec(t, codec, fillTestData(codec, testCases))
}

func TestTimeOrdering(t *testing.T) {
	t.Parallel()
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

	testOrdering(t, lexy.Time(), []testCase[time.Time]{
		// Encodings should sort in this order.
		//
		// before and after the epoch start (Jan 1, 1970)
		// with different nanoseconds within the same second
		// with different timezones
		{"neg fixed 6", negUTC6.In(locFixed), nil},
		{"neg LA 6", negUTC6.In(locLA), nil},
		{"neg NYC 6", negUTC6.In(locNYC), nil},
		{"neg UTC 6", negUTC6, nil},
		{"neg Berlin 6", negUTC6.In(locBerlin), nil},

		{"neg fixed 7", negUTC7.In(locFixed), nil},
		{"neg LA 7", negUTC7.In(locLA), nil},
		{"neg NYC 7", negUTC7.In(locNYC), nil},
		{"neg UTC 7", negUTC7, nil},
		{"neg Berlin 7", negUTC7.In(locBerlin), nil},

		{"pos fixed 6", posUTC6.In(locFixed), nil},
		{"pos LA 6", posUTC6.In(locLA), nil},
		{"pos NYC 6", posUTC6.In(locNYC), nil},
		{"pos UTC 6", posUTC6, nil},
		{"pos Berlin 6", posUTC6.In(locBerlin), nil},

		{"pos fixed 7", posUTC7.In(locFixed), nil},
		{"pos LA 7", posUTC7.In(locLA), nil},
		{"pos NYC 7", posUTC7.In(locNYC), nil},
		{"pos UTC 7", posUTC7, nil},
		{"pos Berlin 7", posUTC7.In(locBerlin), nil},
	})
}
