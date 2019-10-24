// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

// Package p2p implements the Matrix p2p network protocols.
package p2p

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/common/mclock"
	"github.com/MatrixAINetwork/go-matrix/event"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/p2p/discover"
	"github.com/MatrixAINetwork/go-matrix/p2p/nat"
	"strings"
)


// Config holds Server options.
type SubConfig struct {
	// This field must be set to a valid secp256k1 private key.
	PrivateKey *ecdsa.PrivateKey `toml:"-"`

	// If ListenAddr is set to a non-nil address, the server
	// will listen for incoming connections.
	//
	// If the port is zero, the operating system will pick a port. The
	// ListenAddr field will be updated with the actual address when
	// the server is started.
	ListenAddr string

	// If set to a non-nil value, the given NAT port mapper
	// is used to make the listening port available to the
	// Internet.
	NAT nat.Interface `toml:",omitempty"`

	// If Dialer is set to a non-nil value, the given Dialer
	// is used to dial outbound peer connections.
	Dialer NodeDialer `toml:"-"`

	// If NoDial is true, the server will not dial any peers.
	NoDial bool `toml:",omitempty"`

	// ManAddress
	ManAddress common.Address
	ManAddress0 common.Address
	Signature  common.Signature
	SignTime   time.Time
}

// Server manages all peer connections.
type subServer struct {
	// Config fields may not be modified while the server is running.
	SubConfig

	owner *Server
	// Hooks for testing. These are useful because we can inhibit
	// the whole protocol stack.
	newTransport func(net.Conn) transport
	newPeerHook  func(*Peer)
	peerOp     chan peerOpFunc
	peerOpDone chan struct{}

	lock    sync.Mutex // protects running
	running bool

	listener     net.Listener
	ourHandshake *protoHandshake
	lastLookup   time.Time
	ntab         discoverTable
	//DiscV5       *discv5.Network

	quit          chan struct{}
	addstatic     chan *discover.Node
	removestatic  chan *discover.Node
	posthandshake chan *conn
	addpeer       chan *conn
	delpeer       chan peerDrop
	peerFeed      event.Feed
	log           log.Logger
	tasks         map[common.Address]int
	taskLock      sync.RWMutex
}

// Self returns the local node's endpoint information.
func (srv *subServer) Self() *discover.Node {

	if !srv.running {
		return &discover.Node{IP: net.ParseIP("0.0.0.0")}
	}
	return srv.makeSelf(srv.listener)
}
func (srv *subServer) SelfAddr() common.Address {
	return srv.SubConfig.ManAddress0
}

func (srv *subServer) makeSelf(listener net.Listener) *discover.Node {
	// Inbound connections disabled, use zero address.
	if listener == nil {
		return &discover.Node{IP: net.ParseIP("0.0.0.0"), ID: discover.PubkeyID(&srv.PrivateKey.PublicKey)}
	}
	// Otherwise inject the listener address too
	addr := listener.Addr().(*net.TCPAddr)
	return &discover.Node{
		ID:  discover.PubkeyID(&srv.PrivateKey.PublicKey),
		IP:  addr.IP,
		TCP: uint16(addr.Port),
	}
}

// Stop terminates the server and all active peer connections.
// It blocks until all active connections have been closed.
func (srv *subServer) Stop() {
	srv.lock.Lock()
	defer srv.lock.Unlock()
	if !srv.running {
		return
	}
	srv.running = false
	if srv.listener != nil {
		// this unblocks listener Accept
		srv.listener.Close()
	}
	close(srv.quit)
}

func (srv *subServer) Start() (err error) {

	srv.lock.Lock()
	defer srv.lock.Unlock()
	if srv.running {
		return errors.New("server already running")
	}
	srv.running = true
	srv.log = srv.owner.Config.Logger
	if srv.log == nil {
		srv.log = log.New()
	}
	srv.log.Info("Starting P2P networking")

	// static fields
	if srv.PrivateKey == nil {
		return fmt.Errorf("Server.PrivateKey must be set to a non-nil key")
	}
	if srv.newTransport == nil {
		srv.newTransport = newRLPX
	}
	if srv.Dialer == nil {
		srv.Dialer = TCPDialer{&net.Dialer{Timeout: defaultDialTimeout}}
	}
	srv.quit = make(chan struct{})
	srv.addpeer = make(chan *conn)
	srv.delpeer = make(chan peerDrop)
	srv.posthandshake = make(chan *conn)
	srv.addstatic = make(chan *discover.Node,5)
	srv.removestatic = make(chan *discover.Node,5)
	srv.tasks = make(map[common.Address]int)
	srv.peerOp = make(chan peerOpFunc)
	srv.peerOpDone = make(chan struct{})


	var (
		conn *net.UDPConn
		//sconn     *sharedUDPConn
		realaddr  *net.UDPAddr
		unhandled chan discover.ReadPacket
	)

	if !srv.owner.NoDiscovery /*|| srv.DiscoveryV5 */ {
		addr, err := net.ResolveUDPAddr("udp", srv.ListenAddr)
		if err != nil {
			return err
		}
		conn, err = net.ListenUDP("udp", addr)
		if err != nil {
			return err
		}
		realaddr = conn.LocalAddr().(*net.UDPAddr)
		if srv.NAT != nil {
			if !realaddr.IP.IsLoopback() {
				go nat.Map(srv.NAT, srv.quit, "udp", realaddr.Port, realaddr.Port, "matrix discovery")
			}
			// TODO: react to external IP changes over time.
			if ext, err := srv.NAT.ExternalIP(); err == nil {
				realaddr = &net.UDPAddr{IP: ext, Port: realaddr.Port}
			}
		}
	}

	//if !srv.NoDiscovery && srv.DiscoveryV5 {
	//	unhandled = make(chan discover.ReadPacket, 100)
	//	sconn = &sharedUDPConn{conn, unhandled}
	//}

	srv.log.Info("server start info", "man address", srv.ManAddress, "signature", srv.Signature, "time", srv.SignTime)
	cfg := discover.Config{
		PrivateKey:   srv.PrivateKey,
		AnnounceAddr: realaddr,
//		NodeDBPath:   srv.NodeDatabase,
//		NetRestrict:  srv.NetRestrict,
		//Bootnodes:    srv.BootstrapNodes,
		TrustNodes:   srv.owner.TrustedNodes,
		Unhandled:    unhandled,
		NetWorkId:    srv.owner.NetWorkId,
		Address:      srv.ManAddress,
		Signature:    srv.Signature,
		SignTime:     uint64(srv.SignTime.Unix()),
	}
	cfg.Bootnodes = append(cfg.Bootnodes,srv.owner.BootstrapNodes...)
	cfg.Bootnodes = append(cfg.Bootnodes,srv.owner.StaticNodes...)
	ntab, err := discover.ListenUDP(conn, cfg)
	if err != nil {
		return err
	}
	srv.ntab = ntab

	//if srv.DiscoveryV5 {
	//	var (
	//		ntab *discv5.Network
	//		err  error
	//	)
	//	if sconn != nil {
	//		ntab, err = discv5.ListenUDP(srv.PrivateKey, sconn, realaddr, "", srv.NetRestrict) //srv.NodeDatabase)
	//	} else {
	//		ntab, err = discv5.ListenUDP(srv.PrivateKey, conn, realaddr, "", srv.NetRestrict) //srv.NodeDatabase)
	//	}
	//	if err != nil {
	//		return err
	//	}
	//	if err := ntab.SetFallbackNodes(srv.BootstrapNodesV5); err != nil {
	//		return err
	//	}
	//	srv.DiscV5 = ntab
	//}

	dynPeers := srv.owner.maxDialedConns()
	dialer := newDialState(srv.owner.StaticNodes, srv.owner.BootstrapNodes, srv.Self(), dynPeers, srv.owner.NetRestrict)
	if strings.Index(srv.SubConfig.ListenAddr,":50505")>=0 {
		for _,fnode := range srv.owner.Config.TrustedNodes {
			dialer.addStatic(fnode)
		}
	}

	// handshake
	srv.ourHandshake = &protoHandshake{Version: baseProtocolVersion, Name: srv.owner.Name, ID: discover.PubkeyID(&srv.PrivateKey.PublicKey)}
	for _, p := range srv.owner.Protocols {
		srv.ourHandshake.Caps = append(srv.ourHandshake.Caps, p.cap())
	}
	// listen/dial
	if srv.ListenAddr != "" {
		if err := srv.startListening(); err != nil {
			return err
		}
	}
	if srv.NoDial && srv.ListenAddr == "" {
		srv.log.Warn("P2P server will be useless, neither dialing nor listening")
	}

	srv.owner.loopWG.Add(1)
	//add by zw
//	go Receiveudp()
//	go CustSend()
	//add by zw
	go srv.run(dialer)
//	go srv.runTask()

	srv.running = true

//	go Buckets.Start()
	go Link.Start()
//	go UdpStart()

	return nil
}

func (srv *subServer) startListening() error {
	// Launch the TCP listener.
	listener, err := net.Listen("tcp", srv.ListenAddr)
	if err != nil {
		return err
	}
	laddr := listener.Addr().(*net.TCPAddr)
	srv.ListenAddr = laddr.String()
	srv.listener = listener
	srv.owner.loopWG.Add(1)
	go srv.listenLoop()
	// Map the TCP listening port if NAT is configured.
	if !laddr.IP.IsLoopback() && srv.NAT != nil {
		srv.owner.loopWG.Add(1)
		go func() {
			nat.Map(srv.NAT, srv.quit, "tcp", laddr.Port, laddr.Port, "matrix p2p")
			srv.owner.loopWG.Done()
		}()
	}
	return nil
}
// Peers returns all connected peers.
func (srv *subServer) Peers() []*Peer {
	var ps []*Peer
	select {
	// Note: We'd love to put this function into a variable but
	// that seems to cause a weird compiler error in some
	// environments.
	case srv.peerOp <- func(peers map[discover.NodeID]*Peer) {
		for _, p := range peers {
			ps = append(ps, p)
		}
	}:
		<-srv.peerOpDone
	case <-srv.quit:
	}
	return ps
}

func (srv *subServer) run(dialstate dialer) {
	defer srv.owner.loopWG.Done()
	var (
		peers        = make(map[discover.NodeID]*Peer)
		inboundCount = 0
		trusted      = make(map[discover.NodeID]bool, 0)
		taskdone     = make(chan task, maxActiveDialTasks)
		runningTasks []task
		queuedTasks  []task // tasks that can't run yet
	)
	// Put trusted nodes into a map to speed up checks.
	// Trusted peers are loaded on startup and cannot be
	// modified while the server is running.
//	for _, n := range srv.owner.TrustedNodes {
//		trusted[n.ID] = true
//	}
	// removes t from runningTasks
	delTask := func(t task) {
		for i := range runningTasks {
			if runningTasks[i] == t {
				runningTasks = append(runningTasks[:i], runningTasks[i+1:]...)
				break
			}
		}
	}
	// starts until max number of active tasks is satisfied
	startTasks := func(ts []task) (rest []task) {
		i := 0
		for ; len(runningTasks) < maxActiveDialTasks && i < len(ts); i++ {
			t := ts[i]
			srv.log.Trace("New dial task", "task", t)
			go func() { t.Do(srv); taskdone <- t }()
			runningTasks = append(runningTasks, t)
		}
		return ts[i:]
	}
	scheduleTasks := func() {
		// Start from queue first.
		queuedTasks = append(queuedTasks[:0], startTasks(queuedTasks)...)
		// Query dialer for new tasks and start as many as possible now.
		if len(runningTasks) < maxActiveDialTasks {
			nt := dialstate.newTasks(len(runningTasks)+len(queuedTasks), peers, time.Now())
			queuedTasks = append(queuedTasks, startTasks(nt)...)
		}
	}

	tk := time.NewTicker(time.Second * 10)
running:
	for {
		scheduleTasks()

		select {
		case <-tk.C:
			continue
		case <-srv.quit:
			// The server was stopped. Run the cleanup logic.
			break running
		case t := <-taskdone:
			// A task got done. Tell dialstate about it so it
			// can update its state and remove it from the active
			// tasks list.
			srv.log.Trace("Dial task done", "task", t)
			dialstate.taskDone(t, time.Now())
			delTask(t)
		case c := <-srv.posthandshake:
			// A connection has passed the encryption handshake so
			// the remote identity is known (but hasn't been verified yet).
			if trusted[c.id] {
				// Ensure that the trusted flag is set before checking against MaxPeers.
				c.flags |= trustedConn
			}
			// TODO: track in-progress inbound node IDs (pre-Peer) to avoid dialing them.
			select {
			case c.cont <- srv.encHandshakeChecks(peers, inboundCount, c):
			case <-srv.quit:
				break running
			}
		case c := <-srv.addpeer:
			// At this point the connection is past the protocol handshake.
			// Its capabilities are known and the remote identity is verified.
			err := srv.protoHandshakeChecks(peers, inboundCount, c)
			if err == nil {
				// The handshakes are done and it passed all checks.
				p := newPeer(c,srv.Self().ID,srv.SelfAddr(), srv.owner.Protocols)
				// If message events are enabled, pass the peerFeed
				// to the peer
				if srv.owner.EnableMsgEvents {
					p.events = &srv.peerFeed
				}
				name := truncateName(c.name)
				srv.log.Debug("Adding p2p peer", "name", name, "addr", c.fd.RemoteAddr(), "peers", len(peers)+1)
				go srv.runPeer(p)
				peers[c.id] = p
				if _,exist := srv.owner.friends[c.id];exist{
					srv.owner.friends[c.id] = p
				}
				if p.Inbound() {
					inboundCount++
				}
			}
			// The dialer logic relies on the assumption that
			// dial tasks complete after the peer has been added or
			// discarded. Unblock the task last.
			select {
			case c.cont <- err:
			case <-srv.quit:
				break running
			}
		case pd := <-srv.delpeer:
			// A peer disconnected.
			d := common.PrettyDuration(mclock.Now() - pd.created)
			pd.log.Debug("Removing p2p peer", "duration", d, "peers", len(peers)-1, "req", pd.requested, "err", pd.err)
			delete(peers, pd.ID())
			// delete each peers
			//dialstate.removeStatic(discover.NewNode(pd.ID(), net.IP{}, 0, 0))
			//if pd.Inbound() {
			//	inboundCount--
			//}
		case op := <-srv.peerOp:
			// This channel is used by Peers and PeerCount.
			op(peers)
			srv.peerOpDone <- struct{}{}
		case n := <-srv.addstatic:
			// This channel is used by AddPeer to add to the
			// ephemeral static peer list. Add it to the dialer,
			// it will keep the node connected.
			srv.log.Debug("Adding static node", "node", n)
			dialstate.addStatic(n)
		case n := <-srv.removestatic:
			// This channel is used by RemovePeer to send a
			// disconnect request to a peer and begin the
			// stop keeping the node connected
			srv.log.Debug("Removing static node", "node", n)
			dialstate.removeStatic(n)
			if p, ok := peers[n.ID]; ok {
				p.Disconnect(DiscRequested)
			}

		}
	}

	srv.log.Trace("P2P networking is spinning down")

	//if srv.DiscV5 != nil {
	//	srv.DiscV5.Close()
	//}
	// Disconnect all peers.
	for _, p := range peers {
		p.Disconnect(DiscQuitting)
	}
	// Wait for peers to shut down. Pending connections and tasks are
	// not handled here and will terminate soon-ish because srv.quit
	// is closed.
	for len(peers) > 0 {
		p := <-srv.delpeer
		p.log.Trace("<-delpeer (spindown)", "remainingTasks", len(runningTasks))
		delete(peers, p.ID())
	}
}

func (srv *subServer) protoHandshakeChecks(peers map[discover.NodeID]*Peer, inboundCount int, c *conn) error {
	// Drop connections with no matching protocols.
	if len(srv.owner.Protocols) > 0 && countMatchingProtocols(srv.owner.Protocols, c.caps) == 0 {
		return DiscUselessPeer
	}
	// Repeat the encryption handshake checks because the
	// peer set might have changed between the handshakes.
	return srv.encHandshakeChecks(peers, inboundCount, c)
}

func (srv *subServer) encHandshakeChecks(peers map[discover.NodeID]*Peer, inboundCount int, c *conn) error {
	switch {
	case !c.is(trustedConn|staticDialedConn) && len(peers) >= srv.owner.MaxPeers:
		return DiscTooManyPeers
	case !c.is(trustedConn) && c.is(inboundConn) && inboundCount >= srv.owner.maxInboundConns():
		return DiscTooManyPeers
	case peers[c.id] != nil:
		return DiscAlreadyConnected
	case c.id == srv.Self().ID:
		return DiscSelf
	default:
		return nil
	}
}

// listenLoop runs in its own goroutine and accepts
// inbound connections.
func (srv *subServer) listenLoop() {
	defer srv.owner.loopWG.Done()
	srv.log.Info("RLPx listener up", "self", srv.makeSelf(srv.listener))

	tokens := defaultMaxPendingPeers

	slots := make(chan struct{}, tokens)
	for i := 0; i < tokens; i++ {
		slots <- struct{}{}
	}

	for {
		// Wait for a handshake slot before accepting.
		<-slots

		var (
			fd  net.Conn
			err error
		)
		for {
			fd, err = srv.listener.Accept()
			if tempErr, ok := err.(tempError); ok && tempErr.Temporary() {
				srv.log.Debug("Temporary read error", "err", err)
				continue
			} else if err != nil {
				srv.log.Debug("Read error", "err", err)
				return
			}
			break
		}
		bHave := false
		if tcp, ok := fd.RemoteAddr().(*net.TCPAddr); ok{
			if strings.Index(srv.SubConfig.ListenAddr,":50505")>=0 {
				for _,fnode := range srv.owner.Config.TrustedNodes {
					if fnode.IP.String() == tcp.IP.String(){
						bHave = true
						break
					}
				}
			}
		}
		if !bHave {
			srv.log.Info("Rejected conn (not whitelisted in NetRestrict)", "addr", fd.RemoteAddr())
			fd.Close()
			slots <- struct{}{}
			continue
		}
		/*
		// Reject connections that do not match NetRestrict.
		if srv.owner.NetRestrict != nil {
			if tcp, ok := fd.RemoteAddr().(*net.TCPAddr); ok && !srv.owner.NetRestrict.Contains(tcp.IP) {
				srv.log.Debug("Rejected conn (not whitelisted in NetRestrict)", "addr", fd.RemoteAddr())
				fd.Close()
				slots <- struct{}{}
				continue
			}
		}
		*/

		fd = newMeteredConn(fd, true)
		srv.log.Trace("Accepted connection", "addr", fd.RemoteAddr())
		go func() {
			srv.SetupConn(fd, inboundConn, nil)
			slots <- struct{}{}
		}()
	}
}

// SetupConn runs the handshakes and attempts to add the connection
// as a peer. It returns when the connection has been added as a peer
// or the handshakes have failed.
func (srv *subServer) SetupConn(fd net.Conn, flags connFlag, dialDest *discover.Node) error {
	self := srv.Self()
	if self == nil {
		return errors.New("shutdown")
	}
	c := &conn{fd: fd, transport: srv.newTransport(fd), flags: flags, cont: make(chan error)}
	err := srv.setupConn(c, flags, dialDest)
	if err != nil {
		c.close(err)
		srv.log.Trace("Setting up connection failed", "id", c.id, "err", err)
	}
	return err
}

func (srv *subServer) setupConn(c *conn, flags connFlag, dialDest *discover.Node) error {
	// Prevent leftover pending conns from entering the handshake.
	srv.lock.Lock()
	running := srv.running
	srv.lock.Unlock()
	if !running {
		return errServerStopped
	}
	// Run the encryption handshake.
	var err error
	if c.id, err = c.doEncHandshake(srv.PrivateKey, dialDest); err != nil {
		srv.log.Trace("Failed RLPx handshake", "addr", c.fd.RemoteAddr(), "conn", c.flags, "err", err)
		return err
	}
	clog := srv.log.New("id", c.id, "addr", c.fd.RemoteAddr(), "conn", c.flags)
	// For dialed connections, check that the remote public key matches.
	if dialDest != nil && c.id != dialDest.ID {
		clog.Trace("Dialed identity mismatch", "want", c, dialDest.ID)
		return DiscUnexpectedIdentity
	}
	err = srv.checkpoint(c, srv.posthandshake)
	if err != nil {
		clog.Trace("Rejected peer before protocol handshake", "err", err)
		return err
	}
	// Run the protocol handshake
	phs, err := c.doProtoHandshake(srv.ourHandshake)
	if err != nil {
		clog.Trace("Failed proto handshake", "err", err)
		return err
	}
	if phs.ID != c.id {
		clog.Trace("Wrong devp2p handshake identity", "err", phs.ID)
		return DiscUnexpectedIdentity
	}
	c.caps, c.name = phs.Caps, phs.Name
	err = srv.checkpoint(c, srv.addpeer)
	if err != nil {
		clog.Trace("Rejected peer", "err", err)
		return err
	}
	// If the checks completed successfully, runPeer has now been
	// launched by run.
	clog.Trace("connection set up", "inbound", dialDest == nil)
	return nil
}


// checkpoint sends the conn to run, which performs the
// post-handshake checks for the stage (posthandshake, addpeer).
func (srv *subServer) checkpoint(c *conn, stage chan<- *conn) error {
	select {
	case stage <- c:
	case <-srv.quit:
		return errServerStopped
	}
	select {
	case err := <-c.cont:
		return err
	case <-srv.quit:
		return errServerStopped
	}
}

// runPeer runs in its own goroutine for each peer.
// it waits until the Peer logic returns and removes
// the peer.
func (srv *subServer) runPeer(p *Peer) {
	if srv.newPeerHook != nil {
		srv.newPeerHook(p)
	}

	// broadcast peer add
	srv.peerFeed.Send(&PeerEvent{
		Type: PeerEventTypeAdd,
		Peer: p.ID(),
	})

	// run the protocol
	remoteRequested, err := p.run()

	// broadcast peer drop
	srv.peerFeed.Send(&PeerEvent{
		Type:  PeerEventTypeDrop,
		Peer:  p.ID(),
		Error: err.Error(),
	})

	// Note: run waits for existing peers to be sent on srv.delpeer
	// before returning, so this send should not select on srv.quit.
	srv.delpeer <- peerDrop{p, err, remoteRequested}
}


// NodeInfo gathers and returns a collection of metadata known about the host.
func (srv *subServer) NodeInfo() *NodeInfo {
	node := srv.Self()

	// Gather and assemble the generic node infos
	info := &NodeInfo{
		Name:       srv.owner.Name,
		Enode:      node.String(),
		ID:         node.ID.String(),
		IP:         node.IP.String(),
		ListenAddr: srv.ListenAddr,
		Protocols:  make(map[string]interface{}),
	}
	info.Ports.Discovery = int(node.UDP)
	info.Ports.Listener = int(node.TCP)

	// Gather all the running protocol infos (only once per protocol type)
	for _, proto := range srv.owner.Protocols {
		if _, ok := info.Protocols[proto.Name]; !ok {
			nodeInfo := interface{}("unknown")
			if query := proto.NodeInfo; query != nil {
				nodeInfo = proto.NodeInfo()
			}
			info.Protocols[proto.Name] = nodeInfo
		}
	}
	return info
}
/*
func (srv *subServer) AddPeerByAddress(addr common.Address) {
	if addr == srv.ManAddress {
		return
	}
	srv.log.Info("add peer by address into task", "addr", addr.Hex())
	node := srv.owner.ntab.GetNodeByAddress(addr)
	if node == nil {
		srv.CouTask(addr)
		srv.log.Error("add peer by address failed, node info not found", "addr", addr.Hex())
		return
	}
	select {
	case srv.addstatic <- node:
	case <-srv.quit:
	}
	srv.DelTasks(addr)
	return
}
*/
func (srv *subServer) runTask() {
	tk := time.NewTicker(time.Second * 3)
	defer tk.Stop()

	for {
		select {
		case <-tk.C:
			srv.taskLock.Lock()
//			for a := range srv.tasks {
//				go srv.AddPeerByAddress(a)
//			}
			srv.taskLock.Unlock()
		case <-srv.quit:
			return
		}
	}
}

func (srv *subServer) AddTasks(addr common.Address) {
	srv.taskLock.Lock()
	srv.tasks[addr] = 0
	srv.taskLock.Unlock()
}

func (srv *subServer) DelTasks(addr common.Address) {
	srv.taskLock.Lock()
	delete(srv.tasks, addr)
	srv.taskLock.Unlock()
}

func (srv *subServer) CouTask(addr common.Address) {
	srv.taskLock.Lock()
	srv.tasks[addr] = srv.tasks[addr] + 1
	if srv.tasks[addr] > 30 {
		delete(srv.tasks, addr)
	}
	srv.taskLock.Unlock()
}
