package api

import (
	"io"

	"github.com/UsingCoding/gostore/internal/common/maybe"
)

type InitRequest struct {
	ID         string
	Recipients []string
	Remote     maybe.Maybe[string]
}

type AddRequest struct {
	Path string
	Key  maybe.Maybe[string]

	Data io.Reader
}

type ReadRequest struct {
	Path string
	Key  maybe.Maybe[string]
}

type ReadResponse struct {
	Data []byte
}

type RemoveRequest struct {
	Path string
	Key  maybe.Maybe[string]
}

type ListRequest struct {
	Path maybe.Maybe[string]
}

type ListResponse struct {
	ListNode
}

type ListNode struct {
	Name  string
	Nodes []ListNode
}

type MoveRequest struct {
	Src, Dst string
}

type CopyRequest struct {
	Src, Dst string
}
