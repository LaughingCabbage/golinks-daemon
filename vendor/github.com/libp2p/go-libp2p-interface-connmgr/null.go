package ifconnmgr

import (
	"context"

	inet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
)

type NullConnMgr struct{}

var _ ConnManager = (*NullConnMgr)(nil)

func (_ NullConnMgr) TagPeer(peer.ID, string, int)             {}
func (_ NullConnMgr) UntagPeer(peer.ID, string)                {}
func (_ NullConnMgr) UpsertTag(peer.ID, string, func(int) int) {}
func (_ NullConnMgr) GetTagInfo(peer.ID) *TagInfo              { return &TagInfo{} }
func (_ NullConnMgr) TrimOpenConns(context.Context)            {}
func (_ NullConnMgr) Notifee() inet.Notifiee                   { return &cmNotifee{} }
func (_ NullConnMgr) Protect(peer.ID, string)                  {}
func (_ NullConnMgr) Unprotect(peer.ID, string) bool           { return false }

type cmNotifee struct{}

func (nn *cmNotifee) Connected(n inet.Network, c inet.Conn)         {}
func (nn *cmNotifee) Disconnected(n inet.Network, c inet.Conn)      {}
func (nn *cmNotifee) Listen(n inet.Network, addr ma.Multiaddr)      {}
func (nn *cmNotifee) ListenClose(n inet.Network, addr ma.Multiaddr) {}
func (nn *cmNotifee) OpenedStream(inet.Network, inet.Stream)        {}
func (nn *cmNotifee) ClosedStream(inet.Network, inet.Stream)        {}
