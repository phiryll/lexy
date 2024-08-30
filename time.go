package lexy

import (
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

func splitTime(value time.Time) (int64, uint32, int32) {
	utc := value.UTC()
	seconds := utc.Unix()     // int64 seconds since epoch
	nanos := utc.Nanosecond() // int nanoseconds within second (9 decimal digits, cast to uint32)
	_, offset := value.Zone() // abbreviation (ignored), int seconds east of UTC (cast to int32)
	return seconds, uint32(nanos), int32(offset)
}

func (timeCodec) Append(buf []byte, value time.Time) []byte {
	seconds, nanos, offset := splitTime(value)
	//nolint:mnd
	buf = stdInt64.Append(extend(buf, 16), seconds)
	buf = stdUint32.Append(buf, nanos)
	return stdInt32.Append(buf, offset)
}

func (timeCodec) Put(buf []byte, value time.Time) []byte {
	seconds, nanos, offset := splitTime(value)
	buf = stdInt64.Put(buf, seconds)
	buf = stdUint32.Put(buf, nanos)
	return stdInt32.Put(buf, offset)
}

func (timeCodec) Get(buf []byte) (time.Time, []byte) {
	seconds, buf := stdInt64.Get(buf)
	nanos, buf := stdUint32.Get(buf)
	offset, buf := stdInt32.Get(buf)
	return time.Unix(seconds, int64(nanos)).In(time.FixedZone("", int(offset))), buf
}

func (timeCodec) RequiresTerminator() bool {
	return false
}
