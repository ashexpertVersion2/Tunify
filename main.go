package main

import (
	// "fmt"
	"log"
	"os"

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
	link, err := netlink.LinkByName(linkName)
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

	err = net.CreateRtable(*ns, subnet.IP, link)
	if err != nil {
		log.Fatalf("could not create routing table: %v\n", err)
	}

	err = net.AddMasqurade(*subnet, linkName)
	if err != nil {
		log.Fatalf("could not add iptable nat rule: %v\n", err)
	}
	// port mapping outside ns
	udp53Process, err := proc.ExecSC(53, "", "UDP", "UNIX")
	if err != nil {
		log.Fatalf("could not run socat: %v\n", err)
	}

	log.Default().Printf("executable: %s\n", os.Args[2])
	net.EnterNetworkNs(*ns)

	// port mapping outside ns
	unix53Process, err := proc.ExecSC(53, "127.0.0.53", "UNIX", "UDP")
	if err != nil {
		log.Fatalf("could not run socat: %v\n", err)
	}

	// forking and waiting
	process, err := proc.Exec(os.Args[2], []string{})
	if err != nil {
		log.Fatalf("could not execute the process: %s\n", err)
	}
	process.Wait()
	udp53Process.Kill()
	unix53Process.Kill()
	log.Default().Println("Done!")
}
