// Copyright 2014 struktur AG. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httputils

import (
	_ "crypto/sha512" // Make sure to link in SHA512 (see https://codereview.appspot.com/84700045/).
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// Provide a HTTP server implementation which can listen on TPC
// and Unix Domain sockets.
type Server struct {
	http.Server
	*log.Logger
	listener net.Listener
	closing  bool
	quit     chan struct{}
}

// Listen binds sockets according to the configuration of srv.
func (srv *Server) Listen() error {
	addr := srv.Addr
	if addr == "" {
		addr = ":http"
	}

	var err error
	srv.listener, err = srv.socketListen(addr)
	return err
}

// ListenAndServe binds sockets according to the configuration of srv and blocks
// until the socket closes or an exit signal is received.
func (srv *Server) ListenAndServe() error {
	if err := srv.Listen(); err != nil {
		return err
	}
	return srv.serveUntilSignalled()
}

// ListenTLS binds sockets according to the configuration of srv.
func (srv *Server) ListenTLS(certFile, keyFile string) error {
	config := &tls.Config{}
	if srv.TLSConfig != nil {
		*config = *srv.TLSConfig
	}

	var err error
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}

	return srv.ListenTLSWithConfig(config)
}

// ListenAndServeTLS binds sockets according to the configuration of srv and blocks
// until the socket closes or an exit signal is received.
func (srv *Server) ListenAndServeTLS(certFile, keyFile string) error {
	if err := srv.ListenTLS(certFile, keyFile); err != nil {
		return nil
	}
	return srv.serveUntilSignalled()

}

// ListenAndServeTLSWithConfig binds sockets according to the provided TLS
// config and blocks until the socket closes or an exit signal is received.
func (srv *Server) ListenTLSWithConfig(config *tls.Config) error {
	addr := srv.Addr
	if addr == "" {
		addr = ":https"
	}

	if config == nil {
		return errors.New("TLSConfig required")
	}
	if len(config.Certificates) == 0 {
		return errors.New("TLSConfig has no certificate")
	}

	if config.NextProtos == nil {
		config.NextProtos = []string{"http/1.1"}
	}

	if l, err := srv.socketListen(addr); err == nil {
		srv.listener = tls.NewListener(l, config)
	} else {
		return err
	}
	return nil
}

// ListenAndServeTLSWithConfig binds sockets according to the provided TLS
// config and blocks until the socket closes or an exit signal is received.
func (srv *Server) ListenAndServeTLSWithConfig(config *tls.Config) error {
	if err := srv.ListenTLSWithConfig(config); err != nil {
		return nil
	}
	return srv.serveUntilSignalled()
}

// Start runs a server whose sockets were previously bound by calling Listen,
// ListenTLS, or ListenTLSWithConfig and waits for its socket to close or
// Stop to be called.
//
// Note that signals are not handled by the server when started in this manner,
// the caller should do so as needed.
func (srv *Server) Start() error {
	if srv.listener == nil {
		return fmt.Errorf("Listen must be called before Start")
	}

	failed := make(chan error)
	go func() {
		failed <- srv.Serve(srv.listener)
	}()

	srv.quit = make(chan struct{})
	select {
	case <- srv.quit:
		return nil
	case err := <- failed:
		if srv.closing {
			err = nil
		}
		return err
	}
}

// Stop closes a previously started server.
func (srv *Server) Stop() error {
	if srv.quit == nil {
		return fmt.Errorf("Server was not started")
	}

	srv.closing = true
	err := srv.listener.Close()
	srv.quit <- struct{}{}	
	return err
}

func (srv *Server) serveUntilSignalled() error {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sig)

	go func() {
		s := <-sig
		msg := "Received exit signal %d - Closing ..."
		if srv.Logger != nil {
			srv.Logger.Printf(msg, s)
		} else {
			log.Printf(msg, s)
		}
		srv.Stop()
	}()

	return srv.Start()
}

func (srv *Server) socketListen(addr string) (net.Listener, error) {

	var err error
	var l net.Listener

	if strings.HasPrefix(addr, "/") {
		var laddr *net.UnixAddr
		if laddr, err = net.ResolveUnixAddr("unix", addr); err != nil {
			return nil, err
		}
		if l, err = createUnixSocket(laddr); err != nil {
			// Unix-domain-socket already exists, try to connect to it to
			// see if it still is usedb by another process
			if _, err = net.Dial("unix", addr); err != nil {
				if err = os.Remove(addr); err != nil {
					return nil, err
				}
				if l, err = createUnixSocket(laddr); err != nil {
					return nil, err
				}
			} else {
				return nil, fmt.Errorf("another process seems to be listening on %s already", addr)
			}
		}
	} else {
		var laddr *net.TCPAddr
		if laddr, err = net.ResolveTCPAddr("tcp", addr); err != nil {
			return nil, err
		}
		if l, err = net.ListenTCP("tcp", laddr); err != nil {
			return nil, err
		}
	}

	return l, nil

}

func createUnixSocket(addr *net.UnixAddr) (l net.Listener, err error) {
	l, err = net.ListenUnix("unix", addr)

	if err != nil {
		return
	}

	// TODO(lcooper): Not sure if this sequence is completely safe.
	// It would be better if we could get the underlying FD of the socket
	// and stat() + chmod() that instead.
	fi, err := os.Stat(addr.String())
	if err != nil {
		return
	}

	// NOTE(lcooper): This only ensures g+w other then on Linux,
	// BSD systems only use parent directory permissions.
	// See http://stackoverflow.com/questions/5977556 .
	err = os.Chmod(addr.String(), fi.Mode()|0060)
	return
}
