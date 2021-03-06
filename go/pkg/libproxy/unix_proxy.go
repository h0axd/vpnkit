package libproxy

import (
	"fmt"
	"log"
	"net"
)

// UnixProxy is a proxy for Unix connections. It implements the Proxy interface to
// handle Unix traffic forwarding between the frontend and backend addresses.
type UnixProxy struct {
	listener     net.Listener
	frontendAddr net.Addr
	backendAddr  *net.UnixAddr
}

// NewUnixProxy creates a new UnixProxy.
func NewUnixProxy(listener net.Listener, backendAddr *net.UnixAddr) (*UnixProxy, error) {
	log.Printf("NewUnixProxy from %s -> %s\n", listener.Addr().String(), backendAddr.String())
	return &UnixProxy{
		listener:     listener,
		frontendAddr: listener.Addr(),
		backendAddr:  backendAddr,
	}, nil
}

// HandleUnixConnection forwards the Unix traffic to a specified backend address
func HandleUnixConnection(client Conn, backendAddr *net.UnixAddr, quit chan struct{}) error {
	backend, err := net.DialUnix("unix", nil, backendAddr)
	if err != nil {
		return fmt.Errorf("can't forward traffic to backend unix/%v: %s", backendAddr, err)
	}
	return proxy(client, backend, quit)
}

// Run starts forwarding the traffic using Unix.
func (proxy *UnixProxy) Run() {
	quit := make(chan struct{})
	defer close(quit)
	for {
		client, err := proxy.listener.Accept()
		if err != nil {
			log.Printf("Stopping proxy on unix/%v for unix/%v (%s)", proxy.frontendAddr, proxy.backendAddr, err)
			return
		}
		go HandleUnixConnection(client.(Conn), proxy.backendAddr, quit)
	}
}

// Close stops forwarding the traffic.
func (proxy *UnixProxy) Close() { proxy.listener.Close() }

// FrontendAddr returns the Unix address on which the proxy is listening.
func (proxy *UnixProxy) FrontendAddr() net.Addr { return proxy.frontendAddr }

// BackendAddr returns the Unix proxied address.
func (proxy *UnixProxy) BackendAddr() net.Addr { return proxy.backendAddr }
