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
	"github.com/MatrixAINetwork/go-matrix/event"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/p2p/discover"
	"github.com/MatrixAINetwork/go-matrix/p2p/nat"
	"github.com/MatrixAINetwork/go-matrix/p2p/netutil"
	"strconv"
	"path/filepath"
	"os"
	"github.com/MatrixAINetwork/go-matrix/base58"
	"encoding/json"
	"io/ioutil"
	"github.com/MatrixAINetwork/go-matrix/crypto"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/mc"
)

const (
	defaultDialTimeout = 15 * time.Second
	datadirPrivateKey      = "nodekey.json"   // Path within the datadir to the node's private key
	datadirManSignature    = "signature" // Path within the datadir to the signature
	datadirManAddress      = "address"

	// Connectivity defaults.
	maxActiveDialTasks     = 100
	defaultMaxPendingPeers = 10
	defaultDialRatio       = 3

	// Maximum time allowed for reading a complete message.
	// This is effectively the amount of time a connection can be idle.
	frameReadTimeout = 30 * time.Second

	// Maximum amount of time allowed for writing a complete message.
	frameWriteTimeout = 20 * time.Second

	numSubserver = 1
)

var errServerStopped = errors.New("server stopped")

// Config holds Server options.
type Config struct {

	//subServerConfig
	SubServers []SubConfig

	// MaxPeers is the maximum number of peers that can be
	// connected. It must be greater than zero.
	MaxPeers int

	// MaxPendingPeers is the maximum number of peers that can be pending in the
	// handshake phase, counted separately for inbound and outbound connections.
	// Zero defaults to preset values.
	MaxPendingPeers int `toml:",omitempty"`

	// DialRatio controls the ratio of inbound to dialed connections.
	// Example: a DialRatio of 2 allows 1/2 of connections to be dialed.
	// Setting DialRatio to zero defaults it to 3.
	DialRatio int `toml:",omitempty"`

	// NoDiscovery can be used to disable the peer discovery mechanism.
	// Disabling is useful for protocol debugging (manual topology).
	NoDiscovery bool

	// If NoDial is true, the server will not dial any peers.
	NoDial bool `toml:",omitempty"`

	// DiscoveryV5 specifies whether the the new topic-discovery based V5 discovery
	// protocol should be started or not.
	//DiscoveryV5 bool `toml:",omitempty"`

	// Name sets the node name of this server.
	// Use common.MakeName to create a name that follows existing conventions.
	Name string `toml:"-"`

	// BootstrapNodes are used to establish connectivity
	// with the rest of the network.
	BootstrapNodes []*discover.Node

	// BootstrapNodesV5 are used to establish connectivity
	// with the rest of the network using the V5 discovery
	// protocol.
	//BootstrapNodesV5 []*discv5.Node `toml:",omitempty"`

	// Static nodes are used as pre-configured connections which are always
	// maintained and re-connected on disconnects.
	StaticNodes []*discover.Node

	// Trusted nodes are used as pre-configured connections which are always
	// allowed to connect, even above the peer limit.
	TrustedNodes []*discover.Node

	// Connectivity can be restricted to certain IP networks.
	// If this option is set to a non-nil value, only hosts which match one of the
	// IP networks contained in the list are considered.
	NetRestrict *netutil.Netlist `toml:",omitempty"`

	// NodeDatabase is the path to the database containing the previously seen
	// live nodes in the network.
	NodeDatabase string `toml:",omitempty"`

	// Protocols should contain the protocols supported
	// by the server. Matching protocols are launched for
	// each peer.
	Protocols []Protocol `toml:"-"`


	// If EnableMsgEvents is set then the server will emit PeerEvents
	// whenever a message is sent to or received from a peer
	EnableMsgEvents bool

	// Logger is a custom logger to use with the p2p.Server.
	Logger log.Logger `toml:",omitempty"`

	// NetWorkId
	NetWorkId uint64
}

type SubPrivateInfo struct {
	PrvKey string       `json:"private"`
	ManAddr string      `json:"address"`
	ManAddr0 string      `json:"address0"`
	Signature string    `json:"signature"`
	Time   time.Time       `json:"time"`
}
func (cfg *Config)ReadSubConfig(path string)  {
	subInfo := cfg.parsePersistentNodes(filepath.Join(path,datadirPrivateKey))
	if subInfo != nil{
		cfg.SubServers = make([]SubConfig,len(subInfo))
		for i:=0;i<len(subInfo);i++{
			if len(subInfo[i].PrvKey)>20{
				cfg.SubServers[i].PrivateKey,_ = crypto.ToECDSA(common.FromHex(subInfo[i].PrvKey))
			}else{
				cfg.SubServers[i].PrivateKey,_ = crypto.GenerateKey()
			}
			cfg.SubServers[i].ManAddress,_ = base58.Base58DecodeToAddress(subInfo[i].ManAddr)
			cfg.SubServers[i].ManAddress0,_ = base58.Base58DecodeToAddress(subInfo[i].ManAddr0)
			copy(cfg.SubServers[i].Signature[:],common.FromHex(subInfo[i].Signature))
			cfg.SubServers[i].SignTime = subInfo[i].Time
		}
	}
}
func(cfg *Config)WriteSubConfig(path string)  {
	subInfo := make([]SubPrivateInfo,len(cfg.SubServers))
	for i:=0;i<len(subInfo);i++{
		subInfo[i].PrvKey = common.ToHex(cfg.SubServers[i].PrivateKey.D.Bytes())
		subInfo[i].ManAddr = base58.Base58EncodeToString("MAN",cfg.SubServers[i].ManAddress)
		subInfo[i].ManAddr0 = base58.Base58EncodeToString("MAN",cfg.SubServers[i].ManAddress0)
		subInfo[i].Signature = common.ToHex(cfg.SubServers[i].Signature[:])
		subInfo[i].Time = cfg.SubServers[i].SignTime
	}
	out, _ := json.MarshalIndent(subInfo, "", "  ")
	if err := ioutil.WriteFile(filepath.Join(path,datadirPrivateKey), out, 0644); err != nil {
		fmt.Errorf("Failed to save Node file", "err=%v", err)
	}
}
func (c *Config) parsePersistentNodes(path string) []SubPrivateInfo {
	// Short circuit if no node config is present
	if path == "" {
		return nil
	}
	if _, err := os.Stat(path); err != nil {
		return nil
	}
	// Load the nodes from the config file.
	var nodelist []SubPrivateInfo
	if err := common.LoadJSON(path, &nodelist); err != nil {
		log.Error(fmt.Sprintf("Can't load node file %s: %v", path, err))
		return nil
	}

	return nodelist
}

// Server manages all peer connections.
type Server struct {
	// Config fields may not be modified while the server is running.
	Config

	// Hooks for testing. These are useful because we can inhibit
	// the whole protocol stack.
	subServers []*subServer
	friends     map[discover.NodeID]*Peer
	newTransport func(net.Conn) transport
	newPeerHook  func(*Peer)

	lock    sync.Mutex // protects running
	running bool

	ntab         discoverTable
	listener     net.Listener
	ourHandshake *protoHandshake
	lastLookup   time.Time

	recentLeader *RecentLeaderSet
	//DiscV5       *discv5.Network

	// These are for Peers, PeerCount (and nothing else).
	peerOp     chan peerOpFunc
	peerOpDone chan struct{}

	quit          chan struct{}
	addstatic     chan common.Address
	removestatic  chan common.Address
	headerChan	  chan *types.Header
	miningRequestCh       chan *mc.HD_MiningReqMsg
	addpeer       chan *conn
	delpeer       chan peerDrop
	loopWG        sync.WaitGroup // loop, listenLoop
	peerFeed      event.Feed
	log           log.Logger
	tasks         map[common.Address]int
	staticMap	  map[discover.NodeID]int
	taskLock      sync.RWMutex
	lastReqTime   uint64
}

var ServerP2p = &Server{}

type peerOpFunc func(map[discover.NodeID]*Peer)

type peerDrop struct {
	*Peer
	err       error
	requested bool // true if signaled by the peer
}

type connFlag int

const (
	dynDialedConn connFlag = 1 << iota
	staticDialedConn
	inboundConn
	trustedConn
)

// conn wraps a network connection with information gathered
// during the two handshakes.
type conn struct {
	fd net.Conn
	transport
	flags connFlag
	cont  chan error      // The run loop uses cont to signal errors to SetupConn.
	id    discover.NodeID // valid after the encryption handshake
	caps  []Cap           // valid after the protocol handshake
	name  string          // valid after the protocol handshake
}

type transport interface {
	// The two handshakes.
	doEncHandshake(prv *ecdsa.PrivateKey, dialDest *discover.Node) (discover.NodeID, error)
	doProtoHandshake(our *protoHandshake) (*protoHandshake, error)
	// The MsgReadWriter can only be used after the encryption
	// handshake has completed. The code uses conn.id to track this
	// by setting it to a non-nil value after the encryption handshake.
	MsgReadWriter
	// transports must provide Close because we use MsgPipe in some of
	// the tests. Closing the actual network connection doesn't do
	// anything in those tests because NsgPipe doesn't use it.
	close(err error)
}

func (c *conn) String() string {
	s := c.flags.String()
	if (c.id != discover.NodeID{}) {
		s += " " + c.id.String()
	}
	s += " " + c.fd.RemoteAddr().String()
	return s
}

func (f connFlag) String() string {
	s := ""
	if f&trustedConn != 0 {
		s += "-trusted"
	}
	if f&dynDialedConn != 0 {
		s += "-dyndial"
	}
	if f&staticDialedConn != 0 {
		s += "-staticdial"
	}
	if f&inboundConn != 0 {
		s += "-inbound"
	}
	if s != "" {
		s = s[1:]
	}
	return s
}

func (c *conn) is(f connFlag) bool {
	return c.flags&f != 0
}
func (srv *Server) ContainAddr(addr common.Address) bool {
	for _,sub := range srv.subServers {
		if sub.ManAddress == addr{
			return true
		}
	}
	return false
}

// Peers returns all connected peers.
func (srv *Server) Peers(addr0 common.Address) []*Peer {
	for _,sub := range srv.subServers {
		if sub.ManAddress0 == addr0{
			return sub.Peers()
		}
	}
	return nil
}

func (srv *Server) AddressTable() map[common.Address]*discover.Node {
	return srv.ntab.GetAllAddress()
}

func (srv *Server) ConvertAddressToId(addr common.Address) discover.NodeID {
	node := srv.ntab.ResolveNode(addr, EmptyNodeId)
	if node != nil {
		return node.ID
	}
	return EmptyNodeId
}

func (srv *Server) ConvertIdToAddress(id discover.NodeID) common.Address {
	node := srv.ntab.ResolveNode(EmptyAddress, id)
	if node != nil {
		return node.Address
	}
	return EmptyAddress
}

// PeerCount returns the number of connected peers.
func (srv *Server) PeerCount() int {
	var count int
	select {
	case srv.peerOp <- func(ps map[discover.NodeID]*Peer) { count = len(ps) }:
		<-srv.peerOpDone
	case <-srv.quit:
	}
	return count
}

// AddPeer connects to the given node and maintains the connection until the
// server is shut down. If the connection fails for any reason, the server will
// attempt to reconnect the peer.


func (srv *Server) AddPeerByAddress(addr common.Address) {
	if srv.ContainAddr(addr) || len(srv.subServers) == 0 || srv.subServers[0].ntab == nil{
		return
	}
	node := srv.subServers[0].ntab.GetNodeByAddress(addr)
	if node == nil {
		return
	}
	srv.recentLeader.findNode(addr)
	if _,exist := srv.staticMap[node.ID];!exist {
		for _,sub := range srv.subServers{
			sub.addstatic <- node
		}
	}
	return
}

func (srv *Server) AddPeerTask(addr common.Address) {
	srv.AddTasks(addr)
}

// RemovePeer disconnects from the given node
func (srv *Server) RemovePeer(node *discover.Node) {
}

func (srv *Server) RemovePeerByAddress(addr common.Address) {

	node := srv.subServers[0].ntab.GetNodeByAddress(addr)
	if node != nil {
		if _,exist := srv.staticMap[node.ID];!exist {
			for _, sub := range srv.subServers {
				sub.removestatic <- node
			}
		}
		return
	}
	srv.log.Info("can not found node info and remove from table", "addr", addr)
}

// SubscribePeers subscribes the given channel to peer events
func (srv *Server) SubscribeEvents(ch chan *PeerEvent) event.Subscription {
	return srv.peerFeed.Subscribe(ch)
}
/*
// Self returns the local node's endpoint information.
func (srv *Server) Self() *discover.Node {
	srv.lock.Lock()
	defer srv.lock.Unlock()

	if !srv.running {
		return &discover.Node{IP: net.ParseIP("0.0.0.0")}
	}
	return srv.makeSelf(srv.listener, srv.ntab)
}

func (srv *Server) makeSelf(listener net.Listener, ntab discoverTable) *discover.Node {
	// If the server's not running, return an empty node.
	// If the node is running but discovery is off, manually assemble the node infos.
	if ntab == nil {
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
	// Otherwise return the discovery node.
	return ntab.Self()
}
*/
// Stop terminates the server and all active peer connections.
// It blocks until all active connections have been closed.
func (srv *Server) Stop() {
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
	srv.loopWG.Wait()
}

// sharedUDPConn implements a shared connection. Write sends messages to the underlying connection while read returns
// messages that were found unprocessable and sent to the unhandled channel by the primary listener.
type sharedUDPConn struct {
	*net.UDPConn
	unhandled chan discover.ReadPacket
}

// ReadFromUDP implements discv5.conn
func (s *sharedUDPConn) ReadFromUDP(b []byte) (n int, addr *net.UDPAddr, err error) {
	packet, ok := <-s.unhandled
	if !ok {
		return 0, nil, fmt.Errorf("Connection was closed")
	}
	l := len(packet.Data)
	if l > len(b) {
		l = len(b)
	}
	copy(b[:l], packet.Data[:l])
	return l, packet.Addr, nil
}

// Close implements discv5.conn
//func (s *sharedUDPConn) Close() error {
//	return nil
//}

// Start starts running the server.
// Servers can not be re-used after stopping.
func (srv *Server) Start() (err error) {

	srv.lock.Lock()
	defer srv.lock.Unlock()
	if srv.running {
		return errors.New("server already running")
	}
	srv.running = true
	srv.log = srv.Config.Logger
	if srv.log == nil {
		srv.log = log.New()
	}
	srv.log.Info("Starting P2P networking")

	srv.subServers = make([]*subServer,len(srv.Config.SubServers))
	// static fields
	srv.quit = make(chan struct{})
	srv.addpeer = make(chan *conn)
	srv.delpeer = make(chan peerDrop)
	srv.addstatic = make(chan common.Address,5)
	srv.removestatic = make(chan common.Address,5)
	srv.peerOp = make(chan peerOpFunc)
	srv.peerOpDone = make(chan struct{})
	srv.tasks = make(map[common.Address]int)
	srv.recentLeader = newRecentLeaderSet(600,srv)
	srv.headerChan = make(chan *types.Header,100)
	srv.miningRequestCh = make(chan *mc.HD_MiningReqMsg, 100)
	srv.staticMap = make(map[discover.NodeID]int,len(srv.Config.StaticNodes))
	srv.friends  = make(map[discover.NodeID]*Peer)
	for _,node := range srv.Config.StaticNodes  {
		srv.staticMap[node.ID] = 0
	}
	for _,node := range srv.Config.TrustedNodes {
		srv.friends[node.ID] = nil
	}
	mc.SubscribeEvent(mc.NewBlockMessage, srv.headerChan)
	mc.SubscribeEvent(mc.HD_V2_MiningReq, srv.miningRequestCh)
	port := 50505
	for i:=0;i<len(srv.subServers);i++ {
		sub := new(subServer)
		srv.subServers[i] = sub
		sub.SubConfig = srv.Config.SubServers[i]
		sub.SubConfig.ListenAddr = ":"+strconv.Itoa(port)
		sub.NAT = nat.Any()
		sub.owner = srv
		port++
		sub.Start()
	}
	go srv.run()


	//if !srv.NoDiscovery && srv.DiscoveryV5 {
	//	unhandled = make(chan discover.ReadPacket, 100)
	//	sconn = &sharedUDPConn{conn, unhandled}
	//}

//	srv.log.Info("server start info", "man address", srv.ManAddress, "signature", srv.Signature, "time", srv.SignTime)



	//add by zw
//	go Receiveudp()
//	go CustSend()
	//add by zw
	Custsrv = srv
	srv.running = true

//	go Buckets.Start()
	go Link.Start()
//	go UdpStart()

	return nil
}


type dialer interface {
	newTasks(running int, peers map[discover.NodeID]*Peer, now time.Time) []task
	taskDone(task, time.Time)
	addStatic(*discover.Node)
	removeStatic(*discover.Node)
}

func (srv *Server) run() {
	defer srv.loopWG.Done()

running:
	for {

		select {
		case <-srv.quit:
			// The server was stopped. Run the cleanup logic.
			for _, sub := range srv.subServers {
				close(sub.quit)
			}
			break running
		case op := <-srv.peerOp:
			// This channel is used by Peers and PeerCount.
			for _, sub := range srv.subServers {
				sub.peerOp <- op
			}
			srv.peerOpDone <- struct{}{}
		case header := <-srv.headerChan:
			go srv.recentLeader.insertLeader(header)
		case addr := <- srv.addstatic:
			go srv.AddPeerByAddress(addr)
		case addr := <- srv.removestatic:
			go srv.RemovePeerByAddress(addr)
		case req := <-srv.miningRequestCh:
			curtime := req.Header.Time.Uint64()
			if curtime > srv.lastReqTime{
				srv.lastReqTime = curtime
				go srv.recentLeader.insertLeader(req.Header)
				go srv.sendMinerReqTofriends(req)
			}
		}

	}

	srv.log.Trace("P2P networking is spinning down")

	// Terminate discovery. If there is a running lookup it will terminate soon.
	if srv.ntab != nil {
		srv.ntab.Close()
	}
	Buckets.Stop()
	Link.Stop()
	// Wait for peers to shut down. Pending connections and tasks are
	// not handled here and will terminate soon-ish because srv.quit
	// is closed.
}

func(srv *Server) sendMinerReqTofriends(req *mc.HD_MiningReqMsg){
	if !req.IsRemote {
		for _,peer := range srv.friends {
			if peer != nil{
				Send(peer.rw,common.FriendMinerMsg,req)
			}
		}
	}
}

func (srv *Server) maxInboundConns() int {
	return srv.MaxPeers - srv.maxDialedConns()
}

func (srv *Server) maxDialedConns() int {
	if srv.NoDiscovery || srv.NoDial {
		return 0
	}
	r := srv.DialRatio
	if r == 0 {
		r = defaultDialRatio
	}
	return srv.MaxPeers / r
}

type tempError interface {
	Temporary() bool
}

// listenLoop runs in its own goroutine and accepts
// inbound connections.
func (srv *Server) listenLoop() {
	defer srv.loopWG.Done()
//	srv.log.Info("RLPx listener up", "self", srv.makeSelf(srv.listener, srv.ntab))

	tokens := defaultMaxPendingPeers
	if srv.MaxPendingPeers > 0 {
		tokens = srv.MaxPendingPeers
	}
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

		// Reject connections that do not match NetRestrict.
		if srv.NetRestrict != nil {
			if tcp, ok := fd.RemoteAddr().(*net.TCPAddr); ok && !srv.NetRestrict.Contains(tcp.IP) {
				srv.log.Debug("Rejected conn (not whitelisted in NetRestrict)", "addr", fd.RemoteAddr())
				fd.Close()
				slots <- struct{}{}
				continue
			}
		}

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
func (srv *Server) SetupConn(fd net.Conn, flags connFlag, dialDest *discover.Node) error {
	return nil
	/*
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
	*/
}

func truncateName(s string) string {
	if len(s) > 20 {
		return s[:20] + "..."
	}
	return s
}


// NodeInfo represents a short summary of the information known about the host.
type NodeInfo struct {
	ID    string `json:"id"`    // Unique node identifier (also the encryption key)
	Name  string `json:"name"`  // Name of the node, including client type, version, OS, custom data
	Enode string `json:"enode"` // Enode URL for adding this peer from remote peers
	IP    string `json:"ip"`    // IP address of the node
	Ports struct {
		Discovery int `json:"discovery"` // UDP listening port for discovery protocol
		Listener  int `json:"listener"`  // TCP listening port for RLPx
	} `json:"ports"`
	ListenAddr string                 `json:"listenAddr"`
	Protocols  map[string]interface{} `json:"protocols"`
}

// NodeInfo gathers and returns a collection of metadata known about the host.
func (srv *Server) NodeInfo() *NodeInfo {
	return nil
	/*
	node := srv.Self()

	// Gather and assemble the generic node infos
	info := &NodeInfo{
		Name:       srv.Name,
		Enode:      node.String(),
		ID:         node.ID.String(),
		IP:         node.IP.String(),
		ListenAddr: srv.ListenAddr,
		Protocols:  make(map[string]interface{}),
	}
	info.Ports.Discovery = int(node.UDP)
	info.Ports.Listener = int(node.TCP)

	// Gather all the running protocol infos (only once per protocol type)
	for _, proto := range srv.Protocols {
		if _, ok := info.Protocols[proto.Name]; !ok {
			nodeInfo := interface{}("unknown")
			if query := proto.NodeInfo; query != nil {
				nodeInfo = proto.NodeInfo()
			}
			info.Protocols[proto.Name] = nodeInfo
		}
	}
	return info
	*/
}

// PeersInfo returns an array of metadata objects describing connected peers.
func (srv *Server) PeersInfo() []*PeerInfo {
	// Gather all the generic and sub-protocol specific infos
	infos := make([]*PeerInfo, 0, srv.PeerCount())
	for _, peer := range srv.Peers(common.Address{}) {
		if peer != nil {
			infos = append(infos, peer.Info())
		}
	}
	// Sort the result array alphabetically by node identifier
	for i := 0; i < len(infos); i++ {
		for j := i + 1; j < len(infos); j++ {
			if infos[i].ID > infos[j].ID {
				infos[i], infos[j] = infos[j], infos[i]
			}
		}
	}
	return infos
}

func (srv *Server) runTask() {
	tk := time.NewTicker(time.Second * 3)
	defer tk.Stop()

	for {
		select {
		case <-tk.C:
			srv.taskLock.Lock()
			for a := range srv.tasks {
				go srv.AddPeerByAddress(a)
			}
			srv.taskLock.Unlock()
		case <-srv.quit:
			return
		}
	}
}

func (srv *Server) AddTasks(addr common.Address) {
	srv.taskLock.Lock()
	srv.tasks[addr] = 0
	srv.taskLock.Unlock()
}

func (srv *Server) DelTasks(addr common.Address) {
	srv.taskLock.Lock()
	delete(srv.tasks, addr)
	srv.taskLock.Unlock()
}

func (srv *Server) CouTask(addr common.Address) {
	srv.taskLock.Lock()
	srv.tasks[addr] = srv.tasks[addr] + 1
	if srv.tasks[addr] > 30 {
		delete(srv.tasks, addr)
	}
	srv.taskLock.Unlock()
}
