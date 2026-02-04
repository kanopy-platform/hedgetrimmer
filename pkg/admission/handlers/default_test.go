package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPatchResponse_ErrorsOnNil(t *testing.T) {
	resp := PatchResponse([]byte{}, "not json")
	assert.Equal(t, false, resp.Allowed)
}

func TestPatchResponse_OK(t *testing.T) {
	d := struct {
		Hello string
	}{
		Hello: "world",
	}
	resp := PatchResponse([]byte("{}"), &d)
	assert.Equal(t, true, resp.Allowed)
}
