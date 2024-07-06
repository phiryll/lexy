package internal

import (
	"fmt"
	"io"
	"time"
)

var (
	TimeCodec Codec[time.Time] = timeCodec{}
)

// timeCodec is the Codec for time.Time instances.
//
// Unlike most Codecs, timeCodec is lossy. It encodes the timezone's offset, but not its name.
// It will therefore lose information about Daylight Saving Time.
// Timezone names and DST behavior are defined outside go's control (as they must be),
// and time.Time.Zone can return names that will fail with time.Location.LoadLocation.
// The order of encoded instances is UTC time first, timezone offset second.
//
// A time.Time is encoded as the below values,
// using the appropriate uint/int Codecs so that the encoded sort order is correct.
//
//	int64 seconds since epoch (UTC)
//	uint32 nanoseconds with the second
//	int32 timezone offset in seconds east of UTC
type timeCodec struct{}

var formatCache = NewCache[int32, string](formatOffset)

func formatOffset(seconds int32) string {
	sign := '+'
	if seconds < 0 {
		sign = '-'
		seconds = -seconds
	}
	minutes := seconds / 60
	hours := minutes / 60
	return fmt.Sprintf("%c%02d:%02d:%02d", sign, hours, minutes%60, seconds%60)
}

func (c timeCodec) Read(r io.Reader) (time.Time, error) {
	var zero time.Time
	seconds, err := Int64Codec.Read(r)
	if err != nil {
		return zero, unexpectedIfEOF(err)
	}
	nanos, err := Uint32Codec.Read(r)
	if err != nil {
		return zero, unexpectedIfEOF(err)
	}
	offset, err := Int32Codec.Read(r)
	if err != nil && err != io.EOF {
		return zero, err
	}
	loc := time.FixedZone(formatCache.Get(offset), int(offset))
	return time.Unix(seconds, int64(nanos)).In(loc), nil
}

func (c timeCodec) Write(w io.Writer, value time.Time) error {
	utc := value.UTC()
	seconds := utc.Unix()     // int64 seconds since epoch
	nanos := utc.Nanosecond() // int nanoseconds within second (9 decimal digits, cast to int32)
	_, offset := value.Zone() // abbreviation (ignored), int seconds east of UTC (cast to int32)

	if err := Int64Codec.Write(w, seconds); err != nil {
		return err
	}
	if err := Uint32Codec.Write(w, uint32(nanos)); err != nil {
		return err
	}
	return Int32Codec.Write(w, int32(offset))
}
