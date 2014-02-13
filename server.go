// Copyright 2014 struktur AG. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httputils

import (
    "net/http"
    "strings"
    "net"
    "os"
    "os/signal"
    "fmt"
    "log"
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
    go func(){
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
	err = os.Chmod(addr.String(), fi.Mode() | 0060)
	return
}
