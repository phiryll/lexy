package lexy_test

import (
	"testing"

	"github.com/phiryll/lexy"
)

func TestMapUnderlyingType(t *testing.T) {
	t.Parallel()
	type mStringInt map[string]int32
	testBasicMapWithPrefix(t, pNilLast, lexy.NilsLast(lexy.CastMapOf[mStringInt](lexy.String(), lexy.Int32())))
}
