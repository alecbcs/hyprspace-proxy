package proxy

import (
	"io"
	"log"
	"net"

	"golang.zx2c4.com/wireguard/tun/netstack"
)

func RunTCPProxy(src string, dest string, tnet *netstack.Net) {
	listener, err := net.Listen("tcp", src)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("error accepting connection", err)
			continue
		}
		go func() {
			conn2, err := tnet.Dial("tcp", dest)
			if err != nil {
				log.Println("error dialing remote addr", err)
				return
			}
			go func() {
				_, err = io.Copy(conn2, conn)
				if err != nil {
					log.Println(err)
				}
				err = conn2.Close()
				if err != nil {
					log.Println(err)
				}
			}()
			_, err = io.Copy(conn, conn2)
			if err != nil {
				log.Println(err)
			}
			err = conn.Close()
			if err != nil {
				log.Println(err)
			}
		}()
	}
}
