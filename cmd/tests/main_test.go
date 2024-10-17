package tests

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/UsingCoding/gostore/cmd/tests/api"
	"github.com/UsingCoding/gostore/internal/common/maybe"
)

func TestSecretCRUD(t *testing.T) {
	s, err := newSuite()
	assert.NoError(t, err)
	t.Cleanup(func() {
		s.cleanup()
	})

	err = s.gostore().Init(api.InitRequest{
		ID:         "main",
		Recipients: nil,
		Remote:     maybe.Maybe[string]{},
	})
	assert.NoError(t, err)

	const (
		path = "supa/secret"
		data = "data"
	)

	t.Run("add to store", func(t *testing.T) {
		err2 := s.gostore().Add(api.AddRequest{
			Path: path,
			Data: bytes.NewBufferString(data),
		})
		assert.NoError(t, err2)

		resp, err2 := s.gostore().Read(api.ReadRequest{
			Path: path,
		})
		assert.NoError(t, err2)

		assert.Equal(t, data, string(resp.Data))

		err2 = s.gostore().Remove(api.RemoveRequest{
			Path: path,
		})
		assert.NoError(t, err2)
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
		assert.NoError(t, err2)

		err2 = s.gostore().Copy(api.CopyRequest{
			Src: path,
			Dst: dstP,
		})
		assert.NoError(t, err2)

		resp1, err2 := s.gostore().Read(api.ReadRequest{
			Path: path,
		})
		assert.NoError(t, err2)

		resp2, err2 := s.gostore().Read(api.ReadRequest{
			Path: dstP,
		})
		assert.NoError(t, err2)

		assert.Equal(t, resp1.Data, resp2.Data)
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
		assert.NoError(t, err2)

		err2 = s.gostore().Move(api.MoveRequest{
			Src: path,
			Dst: dstP,
		})
		assert.NoError(t, err2)

		_, err2 = s.gostore().Read(api.ReadRequest{
			Path: path,
		})
		assert.Error(t, err2)

		resp2, err2 := s.gostore().Read(api.ReadRequest{
			Path: dstP,
		})
		assert.NoError(t, err2)

		assert.Equal(t, data, string(resp2.Data))
	})
}
