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
	"fmt"

	"github.com/cockroachdb/cockroach/gossip"
	"github.com/cockroachdb/cockroach/proto"
)

const (
	// TODO(bram): Do we still need these? The default zone config might be
	// enough.
	bytesPerRange    = 64 << 20 // 64 MiB
	capacityPerStore = 1 << 40  // 1 TiB - 32768 ranges per store
)

// Store is a simulated cockroach store. To access the replicas in a store, use
// the ranges directly instead.
type Store struct {
	desc   proto.StoreDescriptor
	gossip *gossip.Gossip
}

// newStore returns a new store with using the passed in ID and node
// descriptor.
func newStore(storeID proto.StoreID, nodeDesc proto.NodeDescriptor, gossip *gossip.Gossip) *Store {
	return &Store{
		desc: proto.StoreDescriptor{
			StoreID: storeID,
			Node:    nodeDesc,
		},
		gossip: gossip,
	}
}

// getIDs returns the store's ID and its node's IDs.
func (s *Store) getIDs() (proto.StoreID, proto.NodeID) {
	return s.desc.StoreID, s.desc.Node.NodeID
}

// getDesc returns the store descriptor. The rangeCount is required to
// determine the current capacity.
func (s *Store) getDesc(rangeCount int) proto.StoreDescriptor {
	desc := s.desc
	desc.Capacity = s.getCapacity(rangeCount)
	return desc
}

// getCapacity returns the store capacity based on the numbers of ranges
// located in the store.
// TODO(bram): Change this to take the actual ranges for real counts.
func (s *Store) getCapacity(rangeCount int) proto.StoreCapacity {
	return proto.StoreCapacity{
		Capacity:   capacityPerStore,
		Available:  capacityPerStore - int64(rangeCount)*bytesPerRange,
		RangeCount: int32(rangeCount),
	}
}

// String returns the current status of the store in human readable format.
// Like the getDesc and getCapacity, it requires the number of ranges currently
// housed in the store.
func (s *Store) String(rangeCount int) string {
	desc := s.getDesc(rangeCount)
	return fmt.Sprintf("Store %d - Node:%d, Replicas:%d, AvailableReplicas:%d, Capacity:%d, Available:%d",
		desc.StoreID, desc.Node.NodeID, desc.Capacity.RangeCount, desc.Capacity.Available/bytesPerRange,
		desc.Capacity.Capacity, desc.Capacity.Available)
}

// GossipStore broadcasts the store on the gossip network.
func (s *Store) gossipStore(rangeCount int) error {
	desc := s.getDesc(rangeCount)
	// Unique gossip key per store.
	gossipKey := gossip.MakeStoreKey(desc.StoreID)
	// Gossip store descriptor.
	return s.gossip.AddInfoProto(gossipKey, &desc, 0)
}
