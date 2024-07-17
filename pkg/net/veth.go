package net

import (
	"fmt"
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

	err = netlink.LinkSetNsFd(peerLink, int(peerNs))
	if err != nil {
		return nil, nil, err
	}

	originalNs, err := EnterNetworkNs(peerNs)
	if err != nil {
		return nil, nil, err
	}

	broadcast := net.IP(make([]byte, 4))
	for i := range broadcast {
		tempByte := []byte(subnet.Mask)[i]
		broadcast[i] = subnet.IP.To4()[i] | ^tempByte
	}

	// assign the last ip in the subnet to peer
	ipNet := netlink.NewIPNet(broadcast.To4())
	ipNet.Mask = subnet.Mask
	err = netlink.AddrAdd(peerLink, &netlink.Addr{
		IPNet: ipNet,
	})
	if err != nil {
		return nil, nil, err
	}

	_, err = EnterNetworkNs(*originalNs)

	if err != nil {
		return nil, nil, err
	}

	vethLink, err := netlink.LinkByName(vethName)
	if err != nil {
		return nil, nil, err
	}

	// assign the first ip in subnet to veth
	ipNet = netlink.NewIPNet(subnet.IP.To4())
	ipNet.Mask = subnet.Mask
	err = netlink.AddrAdd(vethLink, &netlink.Addr{
		IPNet: ipNet,
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

func CreateRtable(peerNs netns.NsHandle, innerGateway net.IP, link netlink.Link, outerGateway net.IP) error {
	// going into namespace
	originalNs, err := EnterNetworkNs(peerNs)
	if err != nil {
		return err
	}
	peerLink, err := netlink.LinkByName(peerName)
	if err != nil {
		return fmt.Errorf("can not do ip link list:%v", err)
	}

	err = netlink.LinkSetUp(peerLink)
	if err != nil {
		return fmt.Errorf("can not do ip link set up:%v", err)
	}

	_, defaultNetMast, _ := net.ParseCIDR("0.0.0.0/0")
	route := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Dst:       defaultNetMast,
		Gw:        innerGateway.To4(),
	}
	err = netlink.RouteAdd(route)
	if err != nil {
		return fmt.Errorf("can not do ip r add inside ns:%v", err)
	}

	_, err = EnterNetworkNs(*originalNs)
	if err != nil {
		return fmt.Errorf("can not do ip net exec:%v", err)
	}

	route = &netlink.Route{
		LinkIndex: link.Attrs().Index,
		Dst:       defaultNetMast,
		Gw:        outerGateway.To4(),
		Table:     tableID,
	}
	err = netlink.RouteAdd(route)
	if err != nil {
		return fmt.Errorf("can not do ip r add outside ns:%v", err)
	}

	rule := netlink.NewRule()
	rule.IifName = vethName
	rule.Table = tableID
	err = netlink.RuleAdd(rule)
	if err != nil {
		return fmt.Errorf("can not do ip rule add:%v", err)
	}
	return nil
}

func SetLOUp() error {
	lo, err := netlink.LinkByName("lo")
	if err != nil {
		return fmt.Errorf("can not do ip link set up lo")
	}
	err = netlink.LinkSetUp(lo)
	if err != nil {
		return fmt.Errorf("can not do ip link set up lo")
	}
	return nil
}
