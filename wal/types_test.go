package wal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsMarker(t *testing.T) {

	assert.True(t, IsMarker(MARKER1), "marker1 is marker")
	assert.True(t, IsMarker(MARKER2), "marker2 is marker")
	assert.True(t, IsMarker(MARKER_META_END), "marker_meta_end is marker")

}
