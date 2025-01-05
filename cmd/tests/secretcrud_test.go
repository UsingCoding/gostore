package tests

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/UsingCoding/gostore/cmd/tests/api"
	"github.com/UsingCoding/gostore/internal/common/maybe"
)

func TestSecretCRUD(t *testing.T) {
	s, err := newSuite()
	require.NoError(t, err)
	t.Cleanup(func() {
		s.cleanup()
	})

	err = s.gostore().Init(api.InitRequest{
		ID:         "main",
		Recipients: nil,
		Remote:     maybe.Maybe[string]{},
	})
	require.NoError(t, err)

	const (
		path = "supa/secret"
		data = "data"
	)

	t.Run("add to store", func(t *testing.T) {
		err2 := s.gostore().Add(api.AddRequest{
			Path: path,
			Data: bytes.NewBufferString(data),
		})
		require.NoError(t, err2)

		resp, err2 := s.gostore().Get(api.ReadRequest{
			Path: path,
		})
		require.NoError(t, err2)

		require.Equal(t, data, string(resp.Data))

		err2 = s.gostore().Remove(api.RemoveRequest{
			Path: path,
		})
		require.NoError(t, err2)
	})

	t.Run("copy secret", func(t *testing.T) {
		dstP := "dst"
		t.Cleanup(func() {
			err2 := s.gostore().Remove(api.RemoveRequest{
				Path: path,
			})
			assert.NoError(t, err2)
			err2 = s.gostore().Remove(api.RemoveRequest{
				Path: dstP,
			})
			assert.NoError(t, err2)
		})

		err2 := s.gostore().Add(api.AddRequest{
			Path: path,
			Data: bytes.NewBufferString(data),
		})
		require.NoError(t, err2)

		err2 = s.gostore().Copy(api.CopyRequest{
			Src: path,
			Dst: dstP,
		})
		require.NoError(t, err2)

		resp1, err2 := s.gostore().Get(api.ReadRequest{
			Path: path,
		})
		require.NoError(t, err2)

		resp2, err2 := s.gostore().Get(api.ReadRequest{
			Path: dstP,
		})
		require.NoError(t, err2)

		require.Equal(t, resp1.Data, resp2.Data)
	})

	t.Run("move secret", func(t *testing.T) {
		dstP := "dst"
		t.Cleanup(func() {
			err2 := s.gostore().Remove(api.RemoveRequest{
				Path: dstP,
			})
			assert.NoError(t, err2)
		})

		err2 := s.gostore().Add(api.AddRequest{
			Path: path,
			Data: bytes.NewBufferString(data),
		})
		require.NoError(t, err2)

		err2 = s.gostore().Move(api.MoveRequest{
			Src: path,
			Dst: dstP,
		})
		require.NoError(t, err2)

		_, err2 = s.gostore().Get(api.ReadRequest{
			Path: path,
		})
		require.Error(t, err2)

		resp2, err2 := s.gostore().Get(api.ReadRequest{
			Path: dstP,
		})
		require.NoError(t, err2)

		require.Equal(t, data, string(resp2.Data))
	})
}
