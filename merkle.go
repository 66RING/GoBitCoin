package main

import "crypto/sha256"

type MerkleTree struct {
	RootNode *MerkleTreeNode
}

type MerkleTreeNode struct {
	Left  *MerkleTreeNode
	Right *MerkleTreeNode
	data  []byte
}

func NewMerkleTree(data [][]byte) *MerkleTree {
	var nodes []MerkleTreeNode
	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}
	for _, datum := range data {
		node := NewMerkleTreeNode(nil, nil, datum)
		nodes = append(nodes, *node)
	}
	for i := 0; i < len(data)/2; i++ {
		var newlevel []MerkleTreeNode
		for j := 0; j < len(nodes); j += 2 {
			node := NewMerkleTreeNode(&nodes[j], &nodes[j+1], nil)
			newlevel = append(newlevel, *node)
		}
		nodes = newlevel
	}
	mTree := MerkleTree{&nodes[0]}
	return &mTree
}

func NewMerkleTreeNode(left, right *MerkleTreeNode, data []byte) *MerkleTreeNode {
	mNode := MerkleTreeNode{}
	if left == nil && right == nil {
		hash := sha256.Sum256(data)
		mNode.data = hash[:]
	} else {
		preHashes := append(left.data, right.data...)
		hash := sha256.Sum256(preHashes)
		mNode.data = hash[:]
	}
	mNode.Left = left
	mNode.Right = right
	return &mNode
}
