package tests

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/UsingCoding/gostore/cmd/tests/api"
	"github.com/UsingCoding/gostore/internal/common/maybe"
)

func TestTreeCopy(t *testing.T) {
	s, err := newSuite()
	assert.NoError(t, err)
	t.Cleanup(func() {
		s.cleanup()
	})

	err = s.gostore().Init(api.InitRequest{
		ID: "main",
	})
	assert.NoError(t, err)

	t.Run("nested tree", func(t *testing.T) {
		err2 := createSubTree(
			"",
			s,
			api.ListNode{
				Name: "node1",
				Nodes: []api.ListNode{
					{
						Name: "node2",
					},
					{
						Name: "node3",
						Nodes: []api.ListNode{
							{
								Name: "node4",
							},
							{
								Name: "node5",
							},
							{
								Name: "node6",
							},
						},
					},
				},
			},
		)
		assert.NoError(t, err2)
		t.Cleanup(func() {
			err2 = clearSubtree("", s)
			assert.NoError(t, err2)
		})

		// copy node3 subtree under node1 as node7
		err2 = s.gostore().Copy(api.CopyRequest{
			Src: "node1/node3",
			Dst: "node1/node7",
		})
		assert.NoError(t, err2)

		list1, err2 := s.gostore().List(api.ListRequest{
			Path: maybe.NewJust("node1/node3"),
		})
		assert.NoError(t, err2)

		list2, err2 := s.gostore().List(api.ListRequest{
			Path: maybe.NewJust("node1/node7"),
		})
		assert.NoError(t, err2)

		assert.Equal(
			t,
			list1.Nodes,
			list2.Nodes,
		)
	})

	t.Run("on same position", func(t *testing.T) {
		err2 := createSubTree(
			"",
			s,
			api.ListNode{
				Name: "node1",
				Nodes: []api.ListNode{
					{
						Name: "node2",
					},
					{
						Name: "node3",
						Nodes: []api.ListNode{
							{
								Name: "node4",
							},
							{
								Name: "node5",
							},
							{
								Name: "node6",
							},
						},
					},
				},
			},
		)
		assert.NoError(t, err2)
		t.Cleanup(func() {
			err2 = clearSubtree("", s)
			assert.NoError(t, err2)
		})

		// get root
		root, err2 := s.gostore().List(api.ListRequest{})
		assert.NoError(t, err2)

		// copy node3 subtree under node3
		err2 = s.gostore().Copy(api.CopyRequest{
			Src: "node1/node3",
			Dst: "node1/node3",
		})
		assert.NoError(t, err2)

		newRoot, err2 := s.gostore().List(api.ListRequest{})
		assert.NoError(t, err2)

		// tree should not change
		assert.Equal(t, root, newRoot)
	})

	t.Run("merge trees", func(t *testing.T) {
		err2 := createSubTree(
			"",
			s,
			api.ListNode{
				Name: "node1",
				Nodes: []api.ListNode{
					{
						Name: "node2",
					},
					{
						Name: "node3",
						Nodes: []api.ListNode{
							{
								Name: "node4",
							},
							{
								Name: "node5",
							},
						},
					},
					{
						Name: "node6",
						Nodes: []api.ListNode{
							{
								Name: "node7",
							},
							{
								Name: "node8",
							},
						},
					},
				},
			},
		)
		assert.NoError(t, err2)
		t.Cleanup(func() {
			err2 = clearSubtree("", s)
			assert.NoError(t, err2)
		})

		// copy node1 under node6 as node9
		err2 = s.gostore().Copy(api.CopyRequest{
			Src: "node1/node3",
			Dst: "node1/node6/node9",
		})
		assert.NoError(t, err2)

		list, err2 := s.gostore().List(api.ListRequest{})
		assert.NoError(t, err2)

		assert.Equal(t,
			api.ListNode{
				Name: "node1",
				Nodes: []api.ListNode{
					{
						Name:  "node2",
						Nodes: []api.ListNode{},
					},
					{
						Name: "node3",
						Nodes: []api.ListNode{
							{
								Name:  "node4",
								Nodes: []api.ListNode{},
							},
							{
								Name:  "node5",
								Nodes: []api.ListNode{},
							},
						},
					},
					{
						Name: "node6",
						Nodes: []api.ListNode{
							{
								Name:  "node7",
								Nodes: []api.ListNode{},
							},
							{
								Name:  "node8",
								Nodes: []api.ListNode{},
							},
							{
								Name: "node9",
								Nodes: []api.ListNode{
									{
										Name:  "node4",
										Nodes: []api.ListNode{},
									},
									{
										Name:  "node5",
										Nodes: []api.ListNode{},
									},
								},
							},
						},
					},
				},
			},
			list.Nodes[0],
		)
	})

}

func createSubTree(basePath string, s suite, tree api.ListNode) error {
	p := path.Join(basePath, tree.Name)
	if len(tree.Nodes) == 0 {
		return s.gostore().Add(api.AddRequest{
			Path: p,
			Data: generateData(),
		})

	}

	for _, node := range tree.Nodes {
		err := createSubTree(
			p,
			s,
			node,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func clearSubtree(p string, s suite) error {
	list, err := s.gostore().List(api.ListRequest{
		Path: maybe.NewJust(p),
	})
	if err != nil {
		return err
	}

	nodes := list.Nodes

	var recursiveRM func(basePath string, nodes []api.ListNode) error
	recursiveRM = func(basePath string, nodes []api.ListNode) error {
		for _, node := range nodes {
			if len(node.Nodes) == 0 {
				err2 := s.gostore().Remove(api.RemoveRequest{
					Path: path.Join(basePath, node.Name),
				})
				if err2 != nil {
					return err2
				}
				continue
			}

			err2 := recursiveRM(
				path.Join(basePath, node.Name),
				node.Nodes,
			)
			if err2 != nil {
				return err2
			}
		}
		return nil
	}

	return recursiveRM(p, nodes)
}
