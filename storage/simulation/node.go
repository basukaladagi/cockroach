// Copyright 2015 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License. See the AUTHORS file
// for names of contributors.
//
// Author: Bram Gruneir (bram+code@cockroachlabs.com)

package main

import (
	"bytes"
	"fmt"

	"github.com/cockroachdb/cockroach/gossip"
	"github.com/cockroachdb/cockroach/proto"
)

// Node is a simulated cockroach node.
type Node struct {
	desc   proto.NodeDescriptor
	stores map[proto.StoreID]*Store
	gossip *gossip.Gossip
}

// newNode creates a new node with no stores.
func newNode(nodeID proto.NodeID, gossip *gossip.Gossip) *Node {
	node := &Node{
		desc: proto.NodeDescriptor{
			NodeID: nodeID,
		},
		stores: make(map[proto.StoreID]*Store),
		gossip: gossip,
	}
	return node
}

// getStoreIDs returns the list of storeIDs from the stores contained on the
// node.
func (n *Node) getStoreIDs() []proto.StoreID {
	var storeIDs []proto.StoreID
	for storeID := range n.stores {
		storeIDs = append(storeIDs, storeID)
	}
	return storeIDs
}

// getNextStoreID gets the store ID that should be used when adding a new store
// to the node.
func (n *Node) getNextStoreID() proto.StoreID {
	return proto.StoreID((int(n.desc.NodeID) * 1000) + len(n.stores))
}

// addNewStore creates a new store and adds it to the node.
func (n *Node) addNewStore() *Store {
	newStoreID := n.getNextStoreID()
	newStore := newStore(newStoreID, n.desc, n.gossip)
	n.stores[newStoreID] = newStore
	return newStore
}

// String returns the current status of the node for human readable printing.
func (n *Node) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Node %d - Stores:[", n.desc.NodeID)
	first := true
	for storeID := range n.stores {
		if first {
			first = false
		} else {
			buf.WriteString(",")
		}
		buf.WriteString(storeID.String())
	}
	buf.WriteString("]")
	return buf.String()
}
