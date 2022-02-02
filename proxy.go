package main

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/hyprspace/proxy/config"
	"github.com/hyprspace/proxy/p2p"
	"github.com/hyprspace/proxy/proxy"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/tun/netstack"
)

var (
	// tunDev is the tun device used to pass packets between
	// Hyprspace and the proxy proxy.
	tunDev tun.Device
	// tnet is the virtualized network to communicate with the TUN device within
	// the proxy.
	tnet *netstack.Net
	// RevLookup allow quick lookups of an incoming stream
	// for security before accepting or responding to any data.
	RevLookup map[string]string
	// activeStreams is a map of active streams to a peer
	activeStreams map[string]network.Stream
)

func main() {
	fmt.Println(`    __  __                 _____                     
   / / / /_  ______  _____/ ___/____  ____ _________   
  / /_/ / / / / __ \/ ___/\__ \/ __ \/ __ '/ ___/ _ \  
 / __  / /_/ / /_/ / /   ___/ / /_/ / /_/ / /__/  __/  
/_/ /_/\__, / .___/_/   /____/ .___/\__,_/\___/\___/    
      /____/_/              /_/                      `)
	fmt.Println("Proxy")
	fmt.Printf("Version: %s\n\n", config.Global.General.Version)

	// Setup Variables
	var err error

	// Create new virtualized TUN device.
	tunDev, tnet, err = netstack.CreateNetTUN(
		[]net.IP{net.ParseIP(strings.Split(config.Global.Proxy.Address, "/")[0])},
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
	port := config.Global.Proxy.Port
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
	privateKey, err := base64.StdEncoding.DecodeString(config.Global.Proxy.PrivateKey)
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
		peerTable[net.ParseIP(client.Address).String()], err = peer.Decode(client.ID)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Setup reverse lookup hash map for authentication.
	RevLookup = make(map[string]string, len(config.Global.Clients))
	for _, client := range config.Global.Clients {
		RevLookup[client.ID] = client.Address
	}

	fmt.Println("[+] Setting Up Node Discovery via DHT")
	// Setup P2P Discovery
	go p2p.Discover(ctx, host, dht, peerTable)

	proxy.Init(tnet, config.Global.Clients)

	fmt.Println("[+] Network Setup Complete...Waiting on Node Discovery")
	// Listen For New Packets on TUN Interface
	activeStreams = make(map[string]network.Stream)
	packet := make([]byte, 1420)
	for {
		plen, err := tunDev.Read(packet, 0)
		if err != nil {
			log.Println(err)
			continue
		}
		dst := net.IPv4(packet[16], packet[17], packet[18], packet[19]).String()
		stream, ok := activeStreams[dst]
		if ok {
			err = binary.Write(stream, binary.LittleEndian, uint16(plen))
			if err == nil {
				_, err = stream.Write(packet[:plen])
				if err == nil {
					continue
				}
			}
			stream.Close()
			delete(activeStreams, dst)
			ok = false
		}
		if peer, ok := peerTable[dst]; ok {
			stream, err = host.NewStream(ctx, peer, p2p.Protocol)
			if err != nil {
				log.Println(err)
				continue
			}
			err = binary.Write(stream, binary.LittleEndian, uint16(plen))
			if err != nil {
				stream.Close()
				continue
			}
			_, err = stream.Write(packet[:plen])
			if err != nil {
				stream.Close()
				continue
			}
			activeStreams[dst] = stream
		}
	}

}

func streamHandler(stream network.Stream) {
	// If the remote node ID isn't in the list of known nodes don't respond.
	if _, ok := RevLookup[stream.Conn().RemotePeer().Pretty()]; !ok {
		stream.Reset()
		return
	}
	var packet = make([]byte, 1420)
	var packetSize = make([]byte, 2)
	for {
		_, err := stream.Read(packetSize)
		if err != nil {
			stream.Close()
			return
		}
		size := binary.LittleEndian.Uint16(packetSize)
		var plen uint16 = 0
		for plen < size {
			tmp, err := stream.Read(packet[plen:size])
			plen += uint16(tmp)
			if err != nil {
				stream.Close()
				return
			}
		}
		tunDev.Write(packet[:size], 0)
	}
}
