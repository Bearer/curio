package tree

import (
	"context"

	sitter "github.com/smacker/go-tree-sitter"
)

type Tree struct {
	input        []byte
	sitterTree   *sitter.Tree
	unifiedNodes map[NodeID][]*Node
}

func Parse(sitterLanguage *sitter.Language, input string) (*Tree, error) {
	inputBytes := []byte(input)

	parser := sitter.NewParser()
	defer parser.Close()

	parser.SetLanguage(sitterLanguage)

	sitterTree, err := parser.ParseCtx(context.Background(), nil, inputBytes)
	if err != nil {
		return nil, err
	}

	return &Tree{
		input:        inputBytes,
		sitterTree:   sitterTree,
		unifiedNodes: make(map[NodeID][]*Node),
	}, nil
}

func (tree *Tree) RootNode() *Node {
	return tree.wrap(tree.sitterTree.RootNode())
}

func (tree *Tree) Close() {
	tree.sitterTree.Close()
}

func (tree *Tree) wrap(sitterNode *sitter.Node) *Node {
	if sitterNode == nil {
		return nil
	}

	return &Node{tree: tree, sitterNode: sitterNode}
}

func (tree *Tree) unifyNodes(laterNode *Node, earlierNode *Node) {
	if laterNode.Equal(earlierNode) {
		return
	}

	existingUnifiedNodes := tree.unifiedNodes[laterNode.ID()]

	for _, other := range existingUnifiedNodes {
		if other.Equal(earlierNode) {
			// already unified
			return
		}
	}

	tree.unifiedNodes[laterNode.ID()] = append(existingUnifiedNodes, earlierNode)
}

func (tree *Tree) unifiedNodesFor(node *Node) []*Node {
	return tree.unifiedNodes[node.ID()]
}
