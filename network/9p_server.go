// UFS is a userspace server which exports a filesystem over 9p2000.
//
// By default, it will export / over a TCP on port 5640 under the username
// of "harvey".
package network

import (
	"log"
	"net"

	ufs "github.com/Harvey-OS/ninep/filesystem"
	"github.com/Harvey-OS/ninep/protocol"
	"gvisor.dev/gvisor/pkg/tcpip"
)

func start9pServer(l net.Listener, addr tcpip.Address, port uint16, nic tcpip.NICID) {
	// Maybe it did not start. Life is like that sometimes.
	if l == nil {
		return
	}

	ufslistener, err := ufs.NewUFS(func(l *protocol.Listener) error {
		//l.Trace = log.Printf
		return nil
	})
	if err != nil {
		log.Printf("ufslistener: %v", err)
		return
	}

	if err := ufslistener.Serve(l); err != nil {
		log.Print(err)
	}
	log.Printf("9p server exits ...")
}
