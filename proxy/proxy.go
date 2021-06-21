package proxy

import (
	"fmt"
	"strings"

	"github.com/hyprspace/relay/config"
	"golang.zx2c4.com/wireguard/tun/netstack"
)

func Init(tnet *netstack.Net, clients []config.Client) {
	for _, client := range clients {
		for _, port := range client.Ports {
			go RunTCPProxy(fmt.Sprintf(":%d", port),
				fmt.Sprintf("%s:%d", strings.Split(client.Address, "/")[0], port),
				tnet,
			)
		}
	}
}
