package net

import (
	"fmt"
	"net"
	"os/exec"

	"github.com/vishvananda/netlink"
)

// isSubnetFree checks if a given subnet is free in the current routing table.
func isSubnetFree(subnet *net.IPNet, routes []netlink.Route) bool {
	for _, route := range routes {
		if route.Dst != nil && (subnet.Contains(route.Dst.IP) || route.Dst.Contains(subnet.IP)) {
			return false
		}
	}
	return true
}

// findFreeSubnet finds a free /31 subnet based on the current routing table.
func FindFreeSubnet() (*net.IPNet, error) {
	// Get all existing routes.
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return nil, fmt.Errorf("failed to list routes: %w", err)
	}

	// Iterate through possible /31 subnets and check if they are free.
	for i := 0; i < 254; i++ {
		subnet := &net.IPNet{
			IP:   net.IPv4(10, byte(i), 0, 0),
			Mask: net.CIDRMask(31, 32),
		}
		if isSubnetFree(subnet, routes) {
			return subnet, nil
		}
	}

	return nil, fmt.Errorf("no free /31 subnet found in 10.0.0.0/8")
}

// developers heavily advocate using iptables utlity for programmatic manipulation.
func AddMasqurade(subnet net.IPNet, device string) error {
	cmd := exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-o", device, "-s", subnet.String(), "-j", "MASQUERADE")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error executing iptables command: %w , output was: %s", err, output)
	}
	return nil
}

func CleanUpMasqurade(subnet net.IPNet, device string) error {
	cmd := exec.Command("iptables", "-t", "nat", "-D", "POSTROUTING", "-o", device, "-s", subnet.String(), "-j", "MASQUERADE")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error executing iptables command: %w , output was: %s", err, output)
	}
	return nil
}
