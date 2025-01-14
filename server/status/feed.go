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
// Author: Matt Tracy (matt@cockroachlabs.com)

package status

import (
	"github.com/cockroachdb/cockroach/proto"
	"github.com/cockroachdb/cockroach/util"
	"github.com/cockroachdb/cockroach/util/tracer"
)

// StartNodeEvent is published when a node is started.
type StartNodeEvent struct {
	Desc      proto.NodeDescriptor
	StartedAt int64
}

// CallSuccessEvent is published when a call to a node completes without error.
type CallSuccessEvent struct {
	NodeID proto.NodeID
	Method proto.Method
}

// CallErrorEvent is published when a call to a node returns an error.
type CallErrorEvent struct {
	NodeID proto.NodeID
	Method proto.Method
}

// NodeEventFeed is a helper structure which publishes node-specific events to a
// util.Feed. If the target feed is nil, event methods become no-ops.
type NodeEventFeed struct {
	id proto.NodeID
	f  *util.Feed
}

// NewNodeEventFeed creates a new NodeEventFeed which publishes events for a
// specific node to the supplied feed.
func NewNodeEventFeed(id proto.NodeID, feed *util.Feed) NodeEventFeed {
	return NodeEventFeed{
		id: id,
		f:  feed,
	}
}

// StartNode is called by a node when it has started.
func (nef NodeEventFeed) StartNode(desc proto.NodeDescriptor, startedAt int64) {
	nef.f.Publish(&StartNodeEvent{
		Desc:      desc,
		StartedAt: startedAt,
	})
}

// CallComplete is called by a node whenever it completes a request. This will
// publish an appropriate event to the feed based on the results of the call.
// TODO(tschottdorf): move to batch, account for multiple methods per batch.
// In particular, on error want an error position to identify the failed
// request.
func (nef NodeEventFeed) CallComplete(args proto.Request, reply proto.Response) {
	method := args.Method()
	if ba, ok := args.(*proto.BatchRequest); ok && len(ba.Requests) > 0 {
		method = ba.Requests[0].GetInner().Method()
	}
	if err := reply.Header().Error; err != nil &&
		err.TransactionRestart == proto.TransactionRestart_ABORT {
		nef.f.Publish(&CallErrorEvent{
			NodeID: nef.id,
			Method: method,
		})
	} else {
		nef.f.Publish(&CallSuccessEvent{
			NodeID: nef.id,
			Method: method,
		})
	}
}

// NodeEventListener is an interface that can be implemented by objects which
// listen for events published by nodes.
type NodeEventListener interface {
	OnStartNode(event *StartNodeEvent)
	OnCallSuccess(event *CallSuccessEvent)
	OnCallError(event *CallErrorEvent)
	// TODO(tschottdorf): break this out into a TraceEventListener.
	OnTrace(event *tracer.Trace)
}

// ProcessNodeEvent dispatches an event on the NodeEventListener.
func ProcessNodeEvent(l NodeEventListener, event interface{}) {
	switch specificEvent := event.(type) {
	case *StartNodeEvent:
		l.OnStartNode(specificEvent)
	case *tracer.Trace:
		l.OnTrace(specificEvent)
	case *CallSuccessEvent:
		l.OnCallSuccess(specificEvent)
	case *CallErrorEvent:
		l.OnCallError(specificEvent)
	}
}
