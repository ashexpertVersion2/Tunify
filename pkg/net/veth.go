package net

import (
	"net"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

const tableID = 219
const vethName = "tunifyveth0"
const peerName = "tunifyveth1"

func CreateVethPair(peerNs netns.NsHandle, subnet net.IPNet) (netlink.Link, netlink.Link, error) {
	la := netlink.NewLinkAttrs()
	la.Name = vethName
	vethSpec := &netlink.Veth{LinkAttrs: la, PeerName: peerName}
	err := netlink.LinkAdd(vethSpec)
	if err != nil {
		return nil, nil, err
	}
	peerLink, err := netlink.LinkByName(peerName)
	if err != nil {
		return nil, nil, err
	}

	err = netlink.LinkSetUp(peerLink)
	if err != nil {
		return nil, nil, err
	}

	// assign the last ip in the subnet to peer
	err = netlink.AddrAdd(peerLink, &netlink.Addr{
		IPNet: netlink.NewIPNet(net.IP(subnet.Mask)),
	})
	if err != nil {
		return nil, nil, err
	}

	err = netlink.LinkSetNsFd(peerLink, int(peerNs))
	if err != nil {
		return nil, nil, err
	}

	vethLink, err := netlink.LinkByName(vethName)
	if err != nil {
		return nil, nil, err
	}

	// assign the first ip in subnet to veth
	err = netlink.AddrAdd(vethLink, &netlink.Addr{
		IPNet: netlink.NewIPNet(net.IP(subnet.IP)),
	})
	if err != nil {
		return nil, nil, err
	}

	err = netlink.LinkSetUp(vethLink)
	if err != nil {
		return nil, nil, err
	}

	return vethLink, peerLink, nil
}

func CreateRtable(peerNs netns.NsHandle) error {
	// going into namespace
	originalNs, err := EnterNetworkNs(peerNs)
	if err != nil {
		return err
	}
	peerLink, err := netlink.LinkByName(peerName)
	if err != nil {
		return err
	}

	err = netlink.LinkSetUp(peerLink)
	if err != nil {
		return err
	}

	_, defaultNetMast, _ := net.ParseCIDR("0.0.0.0/0")
	route := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Dst:       defaultNetMast,
		Table:     tableID,
	}
	err = netlink.RouteAdd(route)
	if err != nil {
		return err
	}

	_, err = EnterNetworkNs(*originalNs)
	if err != nil {
		return err
	}

	rule := netlink.NewRule()
	rule.IifName = vethName
	rule.Table = tableID
	err = netlink.RuleAdd(rule)
	if err != nil {
		return err
	}
	return nil
}
