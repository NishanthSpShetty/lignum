package wal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestisMarker(t *testing.T) {

	assert.True(t, isMarker(MARKER1), "marker1 is marker")
	assert.True(t, isMarker(MARKER2), "marker2 is marker")
	assert.True(t, isMarker(MARKER_META_END), "marker_meta_end is marker")

}
