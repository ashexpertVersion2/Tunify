package main

import (
	// "fmt"
	"log"
	"net"
	"os"

	tunet "tunify/pkg/net"
	"tunify/pkg/proc"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

// create new namespace
// create veth pair in default and new namespace
// create rtable
// create nat rules
func main() {
	if len(os.Args) != 3 && len(os.Args) != 4 {
		log.Fatalln("Usage: tunify <link_name> <executable> <gateway>")
	}
	var gateway net.IP
	if len(os.Args) == 4 {
		gateway = net.ParseIP(os.Args[3])
		if gateway == nil {
			log.Fatalln("inavlid gateway")
		}
	}
	linkName := os.Args[1]
	link, err := netlink.LinkByName(linkName)
	if err != nil {
		log.Fatalf("%s is not a valid link name: %s\n", linkName, err)
	}

	subnet, err := tunet.FindFreeSubnet()
	if err != nil {
		log.Fatalf("could allocate subnet: %v\n", err)
	}
	ns, err := tunet.CreateNetworkNs()
	if err != nil {
		log.Fatalf("could not create namespace: %v\n", err)
	}

	mainNs, err := netns.Get()
	if err != nil {
		log.Fatalf("could not open namespace file: %v\n", err)
	}

	_, _, err = tunet.CreateVethPair(*ns, *subnet)
	if err != nil {
		log.Fatalf("could not create veth: %v\n", err)
	}

	err = tunet.CreateRtable(*ns, subnet.IP, link, gateway)
	if err != nil {
		log.Fatalf("could not create routing table: %v\n", err)
	}

	err = tunet.AddMasqurade(*subnet, linkName)
	if err != nil {
		log.Fatalf("could not add iptable nat rule: %v\n", err)
	}
	// port mapping outside ns
	udp53Process, err := proc.ExecSC(53, "127.0.0.53", "UNIX", "UDP")
	if err != nil {
		log.Fatalf("could not run socat: %v\n", err)
	}

	log.Default().Printf("executable: %s\n", os.Args[2])
	tunet.EnterNetworkNs(*ns)

	err = tunet.SetLOUp()
	if err != nil {
		log.Fatalf("could not set loopback on: %v\n", err)
	}
	// port mapping inside ns
	unix53Process, err := proc.ExecSC(53, "127.0.0.53", "UDP", "UNIX")
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
	netns.Set(mainNs)
	tunet.CleanUpMasqurade(*subnet, linkName)
	tunet.CleanUpRule(link)
	log.Default().Println("Done!")
}
