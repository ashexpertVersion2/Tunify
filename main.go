package main

import (
	// "fmt"
	"log"
	"os"

	"os/exec"

	"tunify/pkg/net"
	"tunify/pkg/proc"

	"github.com/vishvananda/netlink"
)

// create new namespace
// create veth pair in default and new namespace
// create rtable
// create nat rules
func main() {
	if len(os.Args) != 3 {
		log.Fatalln("Usage: tunify <link_name> <executable>")
	}
	linkName := os.Args[1]
	_, err := netlink.LinkByName(linkName)
	if err != nil {
		log.Fatalf("%s is not a valid link name: %s\n", linkName, err)
	}

	subnet, err := net.FindFreeSubnet()
	if err != nil {
		log.Fatalf("could allocate subnet: %v\n", err)
	}

	ns, err := net.CreateNetworkNs()
	if err != nil {
		log.Fatalf("could not create namespace: %v\n", err)
	}

	_, _, err = net.CreateVethPair(*ns, *subnet)
	if err != nil {
		log.Fatalf("could not create veth: %v\n", err)
	}

	err = net.CreateRtable(*ns)
	if err != nil {
		log.Fatalf("could not create routing table: %v\n", err)
	}

	err = net.AddMasqurade(*subnet, linkName)
	if err != nil {
		log.Fatalf("could not add nat rule: %v\n", err)
	}

	// forking and waiting
	log.Default().Printf("executable: %s\n", os.Args[2])
	pwd, _ := os.Getwd()
	executablePath, err := exec.LookPath(os.Args[2])
	if err != nil {
		log.Fatalf("could not get executable from path: %s\n", err)
	}

	net.EnterNetworkNs(*ns)
	process, err := proc.Exec(executablePath, pwd)
	if err != nil {
		log.Fatalf("could not execute the process: %s\n", err)
	}
	process.Wait()
	log.Default().Println("Done!")
}
