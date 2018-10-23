// Copyright (c) 2008 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
// Copyright 2015 The go-matrix Authors
// This file is part of the go-matrix library.
//
// The go-matrix library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-matrix library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-matrix library. If not, see <http://www.gnu.org/licenses/>.

package man

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/matrix/go-matrix/man/downloader"
	"github.com/matrix/go-matrix/p2p"
	"github.com/matrix/go-matrix/p2p/discover"
)

// Tests that fast sync gets disabled as soon as a real block is successfully
// imported into the blockchain.
func TestFastSyncDisabling(t *testing.T) {
	// Create a pristine protocol manager, check that fast sync is left enabled
	pmEmpty, _ := newTestProtocolManagerMust(t, downloader.FastSync, 0, nil, nil)
	if atomic.LoadUint32(&pmEmpty.fastSync) == 0 {
		t.Fatalf("fast sync disabled on pristine blockchain")
	}
	// Create a full protocol manager, check that fast sync gets disabled
	pmFull, _ := newTestProtocolManagerMust(t, downloader.FastSync, 1024, nil, nil)
	if atomic.LoadUint32(&pmFull.fastSync) == 1 {
		t.Fatalf("fast sync not disabled on non-empty blockchain")
	}
	// Sync up the two peers
	io1, io2 := p2p.MsgPipe()

	go pmFull.handle(pmFull.newPeer(63, p2p.NewPeer(discover.NodeID{}, "empty", nil), io2))
	go pmEmpty.handle(pmEmpty.newPeer(63, p2p.NewPeer(discover.NodeID{}, "full", nil), io1))

	time.Sleep(250 * time.Millisecond)
	pmEmpty.synchronise(pmEmpty.peers.BestPeer())

	// Check that fast sync was disabled
	if atomic.LoadUint32(&pmEmpty.fastSync) == 1 {
		t.Fatalf("fast sync not disabled after successful synchronisation")
	}
}
