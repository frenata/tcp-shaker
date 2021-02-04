package tcp

import (
	"log"
	"net"
	"runtime"

	"golang.org/x/sys/unix"
)

// parseSockAddr resolves given addr to unix.Sockaddr
func parseSockAddr(addr string) (unix.Sockaddr, bool, error) {

	ipv6Addr, err := net.ResolveTCPAddr("tcp6", addr)
	if err != nil {
		log.Print(err)
	}
	log.Println("ipv6: ", ipv6Addr)

	var addr16 [16]byte
	if ipv6Addr != nil {
		copy(addr16[:], ipv6Addr.IP.To16())
		log.Println(ipv6Addr.Port, addr16, ipv6Addr)
		return &unix.SockaddrInet6{Port: ipv6Addr.Port, Addr: addr16}, true, nil
	}

	tAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, false, err
	}
	log.Print("ipv4: ", tAddr)

	var addr4 [4]byte
	if tAddr.IP != nil {
		copy(addr4[:], tAddr.IP.To4()) // copy last 4 bytes of slice to array
	}
	return &unix.SockaddrInet4{Port: tAddr.Port, Addr: addr4}, false, nil
}

// connect calls the connect syscall with error handled.
func connect(fd int, addr unix.Sockaddr) (success bool, err error) {
	serr := unix.Connect(fd, addr)
	log.Print(fd, addr, serr)
	switch serr {
	case unix.EALREADY, unix.EINPROGRESS, unix.EINTR:
		// Connection could not be made immediately but asynchronously.
		success = false
		err = nil
	case nil, unix.EISCONN:
		// The specified socket is already connected.
		success = true
		err = nil
	case unix.EINVAL:
		// On Solaris we can see EINVAL if the socket has
		// already been accepted and closed by the server.
		// Treat this as a successful connection--writes to
		// the socket will see EOF.  For details and a test
		// case in C see https://golang.org/issue/6828.
		if runtime.GOOS == "solaris" {
			success = true
			err = nil
		} else {
			// error must be reported
			success = false
			err = serr
		}
	default:
		// Connect error.
		success = false
		err = serr
	}
	return success, err
}
