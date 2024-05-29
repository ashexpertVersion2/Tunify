package net

import (
	"github.com/vishvananda/netns"
)

func CreateNetworkNs() (*netns.NsHandle, error) {
	ns, err := netns.Get()
	if err != nil {
		return nil, err
	}
	newNs, err := netns.New()
	if err != nil {
		return nil, err
	}
	err = netns.Set(ns)
	if err != nil {
		return nil, err
	}
	return &newNs, nil
}

func EnterNetworkNs(newNs netns.NsHandle) (*netns.NsHandle, error) {
	oldns, err := netns.Get()
	if err != nil {
		return nil, err
	}
	err = netns.Set(newNs)
	if err != nil {
		return nil, err
	}
	return &oldns, nil

}
