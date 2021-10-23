package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/hyprspace/hyprspace/p2p"
	"github.com/hyprspace/relay/config"
	"github.com/hyprspace/relay/proxy"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"golang.org/x/net/ipv4"
	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/tun/netstack"
)

var (
	// tunDev is the tun device used to pass packets between
	// Hyprspace and the relay proxy.
	tunDev tun.Device
	// tnet is the virtualized network to communicate with the TUN device within
	// the proxy.
	tnet *netstack.Net
	// RevLookup allow quick lookups of an incoming stream
	// for security before accepting or responding to any data.
	RevLookup map[string]bool
)

func main() {
	fmt.Println(`    __  __                 _____                     
   / / / /_  ______  _____/ ___/____  ____ _________   
  / /_/ / / / / __ \/ ___/\__ \/ __ \/ __ '/ ___/ _ \  
 / __  / /_/ / /_/ / /   ___/ / /_/ / /_/ / /__/  __/  
/_/ /_/\__, / .___/_/   /____/ .___/\__,_/\___/\___/    
      /____/_/              /_/                      `)
	fmt.Println("Relay")
	fmt.Printf("Version: %s\n\n", config.Global.General.Version)

	// Setup Variables
	var err error

	// Create new virtualized TUN device.
	tunDev, tnet, err = netstack.CreateNetTUN(
		[]net.IP{net.ParseIP(strings.Split(config.Global.Relay.Address, "/")[0])},
		[]net.IP{net.ParseIP("1.1.1.1")},
		1420)
	if err != nil {
		log.Fatal(err)
	}

	// Setup System Context
	ctx := context.Background()

	fmt.Println("[+] Creating LibP2P Node")

	// Check that the listener port is available.
	var ln net.Listener
	port := config.Global.Relay.Port
	ln, err = net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(errors.New("could not create node, listen port already in use by something else"))
	}
	if ln != nil {
		ln.Close()
	}
	// Convert the listen port to an integer
	portInt, err := strconv.Atoi(port)
	if err != nil {
		log.Fatal(err)
	}
	privateKey, err := base64.StdEncoding.DecodeString(config.Global.Relay.PrivateKey)
	if err != nil {
		log.Fatal(err)
	}

	// Create P2P Node
	host, dht, err := p2p.CreateNode(ctx,
		string(privateKey),
		portInt,
		streamHandler)
	if err != nil {
		log.Fatal(err)
	}

	// Setup Peer Table for Quick Packet --> Dest ID lookup
	peerTable := make(map[string]peer.ID)
	for _, client := range config.Global.Clients {
		peerTable[client.Address], err = peer.Decode(client.ID)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Setup reverse lookup hash map for authentication.
	RevLookup = make(map[string]bool, len(config.Global.Clients))
	for _, client := range config.Global.Clients {
		RevLookup[client.ID] = true
	}

	fmt.Println("[+] Setting Up Node Discovery via DHT")
	// Setup P2P Discovery
	go p2p.Discover(ctx, host, dht, peerTable)

	proxy.Init(tnet, config.Global.Clients)

	fmt.Println("[+] Network Setup Complete...Waiting on Node Discovery")
	// Listen For New Packets on TUN Interface
	packet := make([]byte, 1420)
	var stream network.Stream
	var header *ipv4.Header
	var plen int
	for {
		plen, err = tunDev.Read(packet, 0)
		if err != nil {
			if err.Error() == "EOF" {
				continue
			}
			log.Fatal(err)
		}
		header, _ = ipv4.ParseHeader(packet)
		_, ok := peerTable[header.Dst.String()]
		if ok {
			stream, err = host.NewStream(ctx, peerTable[header.Dst.String()], p2p.Protocol)
			if err != nil {
				log.Println(err)
				continue
			}
			stream.Write(packet[:plen])
			stream.Close()
		}
	}

}

func streamHandler(stream network.Stream) {
	// If the remote node ID isn't in the list of known nodes don't respond.
	if _, ok := RevLookup[stream.Conn().RemotePeer().Pretty()]; !ok {
		stream.Reset()
	}
	buf := make([]byte, 1420)
	plen, _ := stream.Read(buf)
	tunDev.Write(buf[:plen], 0)
	stream.Close()
}
