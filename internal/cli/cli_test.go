package cli

import (
	"testing"

	"github.com/kanopy-platform/hedgetrimmer/pkg/mutators"
	"github.com/stretchr/testify/assert"
)

func TestGetHandlers(t *testing.T) {
	t.Parallel()
	mutator := mutators.NewPodTemplateSpec()

	tests := []struct {
		msg       string
		resources []string
		wantLen   int
		wantError bool
	}{
		{
			msg:       "Full list of resources",
			resources: all_resources,
			wantLen:   8,
			wantError: false,
		},
		{
			msg:       "Unexpected resource",
			resources: []string{cronjobs, "unexpected"},
			wantLen:   0,
			wantError: true,
		},
		{
			msg:       "Handle empty input",
			resources: []string{},
			wantLen:   0,
			wantError: false,
		},
		{
			msg:       "Handle leading and trailing spaces",
			resources: []string{" daemonsets ", "  replicasets", "replicationcontrollers  "},
			wantLen:   3,
			wantError: false,
		},
		{
			msg:       "Handle duplicated resources",
			resources: []string{daemonsets, replicasets, replicasets, daemonsets, replicasets},
			wantLen:   2,
			wantError: false,
		},
	}

	for _, test := range tests {
		handlers, err := getHandlers(test.resources, mutator)
		assert.Len(t, handlers, test.wantLen, test.msg)
		assert.Equal(t, test.wantError, err != nil, test.msg)
	}
}
