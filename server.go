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
}

// ListenAndServe binds sockets according to the configuration of srv and blocks
// until the socket closes or an exit signal is received.
func (srv *Server) ListenAndServe() error {

	addr := srv.Addr
	if addr == "" {
		addr = ":http"
	}

	var closing = false
	var err error
	var l net.Listener
	if l, err = srv.socketListen(addr); err != nil {
		return err
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		s := <-sig
		msg := "Received exit signal %d - Closing ..."
		if srv.Logger != nil {
			srv.Logger.Printf(msg, s)
		} else {
			log.Printf(msg, s)
		}
		closing = true
		l.Close()
	}()

	err = srv.Serve(l)
	if err != nil {
		if closing {
			return nil
		}
	}
	return err

}

// ListenAndServeTLS binds sockets according to the configuration of srv and blocks
// until the socket closes or an exit signal is received.
func (srv *Server) ListenAndServeTLS(certFile, keyFile string) error {

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

	return srv.ListenAndServeTLSWithConfig(config)

}

// ListenAndServeTLSWithConfig binds sockets according to the provided TLS
// config and blocks until the socket closes or an exit signal is received.
func (srv *Server) ListenAndServeTLSWithConfig(config *tls.Config) error {

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

	var err error
	var closing = false
	var l net.Listener
	if l, err = srv.socketListen(addr); err != nil {
		return err
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		s := <-sig
		msg := "Received exit signal %d - Closing ..."
		if srv.Logger != nil {
			srv.Logger.Printf(msg, s)
		} else {
			log.Printf(msg, s)
		}
		closing = true
		l.Close()
	}()

	tlsListener := tls.NewListener(l, config)

	err = srv.Serve(tlsListener)
	if err != nil {
		if closing {
			return nil
		}
	}
	return err

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
