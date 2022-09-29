package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPatchResponse_ErrorsOnNil(t *testing.T) {
	p := &Patch{}
	resp := p.PatchResponse([]byte{}, "not json")
	assert.Equal(t, false, resp.Allowed)
}

func TestPatchResponse_OK(t *testing.T) {
	p := &Patch{}
	d := struct {
		Hello string
	}{
		Hello: "world",
	}
	resp := p.PatchResponse([]byte("{}"), &d)
	assert.Equal(t, true, resp.Allowed)
}
