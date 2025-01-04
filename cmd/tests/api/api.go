package api

import (
	"encoding/json"
	"io"

	"github.com/UsingCoding/fpgo/pkg/slices"
	"github.com/pkg/errors"

	"github.com/UsingCoding/gostore/internal/common/maybe"
)

const (
	gostorePath = "bin/gostore"
)

// API defines api for gostore cli
type API interface {
	Init(req InitRequest) error

	Add(req AddRequest) error
	Get(req ReadRequest) (ReadResponse, error)
	Remove(req RemoveRequest) error
	List(req ListRequest) (ListResponse, error)

	Move(req MoveRequest) error
	Copy(req CopyRequest) error
}

func New(basePath string) API {
	return api{
		basePath: basePath,
		storeID:  maybe.Maybe[string]{},
	}
}

type api struct {
	basePath string
	storeID  maybe.Maybe[string]
}

func (a api) Init(req InitRequest) error {
	args := []string{
		"init",
		"--id",
		req.ID,
	}

	if len(req.Recipients) > 0 {
		args = append(args, "-r")
		args = append(args, req.Recipients...)
	}

	if r, ok := maybe.JustValid(req.Remote); ok {
		args = append(args, "--remote", r)
	}

	_, err := a.gostore(input{
		args:  args,
		stdin: nil,
	})

	return err
}

func (a api) Add(req AddRequest) error {
	args := []string{
		"add",
		req.Path,
	}

	if k, ok := maybe.JustValid(req.Key); ok {
		args = append(args, k)
	}

	_, err := a.gostore(input{
		args:  args,
		stdin: req.Data,
	})

	return err
}

func (a api) Get(req ReadRequest) (ReadResponse, error) {
	args := []string{
		"cat",
		req.Path,
	}

	if k, ok := maybe.JustValid(req.Key); ok {
		args = append(args, k)
	}

	o, err := a.gostore(input{
		args: args,
	})
	if err != nil {
		return ReadResponse{}, err
	}

	data, err := io.ReadAll(o.stdout)
	if err != nil {
		return ReadResponse{}, errors.Wrap(err, "failed to read response")
	}

	return ReadResponse{
		Data: data,
	}, nil
}

func (a api) Remove(req RemoveRequest) error {
	args := []string{
		"rm",
		req.Path,
	}

	if k, ok := maybe.JustValid(req.Key); ok {
		args = append(args, k)
	}

	_, err := a.gostore(input{
		args: args,
	})
	return err
}

func (a api) List(req ListRequest) (ListResponse, error) {
	args := []string{
		"-o", "json",
		"ls",
	}

	if p, ok := maybe.JustValid(req.Path); ok {
		args = append(args, p)
	}

	o, err := a.gostore(input{
		args: args,
	})
	if err != nil {
		return ListResponse{}, err
	}

	type jsonTreeNode struct {
		Name  string         `json:"name"`
		Elems []jsonTreeNode `json:"children,omitempty"`
	}

	var res jsonTreeNode

	err = json.Unmarshal(o.stdout.Bytes(), &res)
	if err != nil {
		return ListResponse{}, errors.Wrap(err, "failed to unmarshal response")
	}

	var mapNode func(node jsonTreeNode) ListNode
	mapNode = func(node jsonTreeNode) ListNode {
		return ListNode{
			Name:  node.Name,
			Nodes: slices.Map(node.Elems, mapNode),
		}
	}

	return ListResponse{
		mapNode(res),
	}, nil
}

func (a api) Move(req MoveRequest) error {
	args := []string{
		"mv",
		req.Src,
		req.Dst,
	}

	_, err := a.gostore(input{args: args})
	return err
}

func (a api) Copy(req CopyRequest) error {
	args := []string{
		"cp",
		req.Src,
		req.Dst,
	}

	_, err := a.gostore(input{args: args})
	return err
}
