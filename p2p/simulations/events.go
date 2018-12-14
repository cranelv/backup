// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

package simulations

import (
	"fmt"
	"time"
)

// EventType is the type of event emitted by a simulation network
type EventType string

const (
	// EventTypeNode is the type of event emitted when a node is either
	// created, started or stopped
	EventTypeNode EventType = "node"

	// EventTypeConn is the type of event emitted when a connection is
	// is either established or dropped between two nodes
	EventTypeConn EventType = "conn"

	// EventTypeMsg is the type of event emitted when a p2p message it
	// sent between two nodes
	EventTypeMsg EventType = "msg"
)

// Event is an event emitted by a simulation network
type Event struct {
	// Type is the type of the event
	Type EventType `json:"type"`

	// Time is the time the event happened
	Time time.Time `json:"time"`

	// Control indicates whether the event is the result of a controlled
	// action in the network
	Control bool `json:"control"`

	// Node is set if the type is EventTypeNode
	Node *Node `json:"node,omitempty"`

	// Conn is set if the type is EventTypeConn
	Conn *Conn `json:"conn,omitempty"`

	// Msg is set if the type is EventTypeMsg
	Msg *Msg `json:"msg,omitempty"`
}

// NewEvent creates a new event for the given object which should be either a
// Node, Conn or Msg.
//
// The object is copied so that the event represents the state of the object
// when NewEvent is called.
func NewEvent(v interface{}) *Event {
	event := &Event{Time: time.Now()}
	switch v := v.(type) {
	case *Node:
		event.Type = EventTypeNode
		node := *v
		event.Node = &node
	case *Conn:
		event.Type = EventTypeConn
		conn := *v
		event.Conn = &conn
	case *Msg:
		event.Type = EventTypeMsg
		msg := *v
		event.Msg = &msg
	default:
		panic(fmt.Sprintf("invalid event type: %T", v))
	}
	return event
}

// ControlEvent creates a new control event
func ControlEvent(v interface{}) *Event {
	event := NewEvent(v)
	event.Control = true
	return event
}

// String returns the string representation of the event
func (e *Event) String() string {
	switch e.Type {
	case EventTypeNode:
		return fmt.Sprintf("<node-event> id: %s up: %t", e.Node.ID().TerminalString(), e.Node.Up)
	case EventTypeConn:
		return fmt.Sprintf("<conn-event> nodes: %s->%s up: %t", e.Conn.One.TerminalString(), e.Conn.Other.TerminalString(), e.Conn.Up)
	case EventTypeMsg:
		return fmt.Sprintf("<msg-event> nodes: %s->%s proto: %s, code: %d, received: %t", e.Msg.One.TerminalString(), e.Msg.Other.TerminalString(), e.Msg.Protocol, e.Msg.Code, e.Msg.Received)
	default:
		return ""
	}
}
