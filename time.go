package lexy

import (
	"fmt"
	"io"
	"time"
)

// timeCodec is the Codec for time.Time instances.
//
// Unlike most Codecs, timeCodec is lossy. It encodes the timezone's offset, but not its name.
// It will therefore lose information about Daylight Saving Time.
// Timezone names and DST behavior are defined outside Go's control (as they must be),
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

const timeSize = uint64Size + uint32Size + uint32Size

var formatCache = makeCache(formatOffset)

//nolint:mnd
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

func splitTime(value time.Time) (int64, uint32, int32) {
	utc := value.UTC()
	seconds := utc.Unix()     // int64 seconds since epoch
	nanos := utc.Nanosecond() // int nanoseconds within second (9 decimal digits, cast to uint32)
	_, offset := value.Zone() // abbreviation (ignored), int seconds east of UTC (cast to int32)
	return seconds, uint32(nanos), int32(offset)
}

func buildTime(seconds int64, nanos uint32, offset int32) time.Time {
	loc := time.FixedZone(formatCache.Get(offset), int(offset))
	return time.Unix(seconds, int64(nanos)).In(loc)
}

func (timeCodec) Append(buf []byte, value time.Time) []byte {
	seconds, nanos, offset := splitTime(value)
	buf = stdInt64.Append(buf, seconds)
	buf = stdUint32.Append(buf, nanos)
	return stdInt32.Append(buf, offset)
}

func (timeCodec) Put(buf []byte, value time.Time) int {
	seconds, nanos, offset := splitTime(value)
	n := stdInt64.Put(buf, seconds)
	n += stdUint32.Put(buf, nanos)
	return n + stdInt32.Put(buf, offset)
}

func (timeCodec) Write(w io.Writer, value time.Time) error {
	seconds, nanos, offset := splitTime(value)
	if err := stdInt64.Write(w, seconds); err != nil {
		return err
	}
	if err := stdUint32.Write(w, nanos); err != nil {
		return err
	}
	return stdInt32.Write(w, offset)
}

func (timeCodec) Get(buf []byte) (time.Time, int) {
	seconds, _ := stdInt64.Get(buf)
	nanos, _ := stdUint32.Get(buf[uint64Size:])
	offset, _ := stdInt32.Get(buf[uint64Size+uint32Size:])
	return buildTime(seconds, nanos, offset), timeSize
}

func (timeCodec) Read(r io.Reader) (time.Time, error) {
	var zero time.Time
	seconds, err := stdInt64.Read(r)
	if err != nil {
		return zero, err
	}
	nanos, err := stdUint32.Read(r)
	if err != nil {
		return zero, UnexpectedIfEOF(err)
	}
	offset, err := stdInt32.Read(r)
	if err != nil {
		return zero, UnexpectedIfEOF(err)
	}
	return buildTime(seconds, nanos, offset), nil
}

func (timeCodec) MaxSize() int {
	return timeSize
}

func (timeCodec) RequiresTerminator() bool {
	return false
}
