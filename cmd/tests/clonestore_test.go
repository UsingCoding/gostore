package tests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/UsingCoding/gostore/cmd/tests/api"
	"github.com/UsingCoding/gostore/internal/common/maybe"
)

func TestCloneStore(t *testing.T) {
	var (
		remote = maybe.MapZero("")
	)

	//nolint:staticcheck
	r, ok := maybe.JustValid(remote)
	if !ok {
		t.Skipf("remote is not defined")
	}

	s, err := newSuite()
	require.NoError(t, err)
	t.Cleanup(func() {
		s.cleanup()
	})
	t.Skip()

	err = s.gostore().Clone(api.CloneRequest{
		ID:     "local",
		Remote: r,
	})
	require.NoError(t, err)
}
