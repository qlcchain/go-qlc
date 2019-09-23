// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package rpc

import (
	"net"
	"net/http"

	rpc "github.com/qlcchain/jsonrpc2"
)

// StartHTTPEndpoint starts the HTTP RPC endpoint, configured with cors/vhosts/modules
func (r *RPC) StartHTTPEndpoint(endpoint string, apis []rpc.API, modules []string, cors []string, vhosts []string, timeouts rpc.HTTPTimeouts) (net.Listener, *rpc.Server, error) {
	// Generate the whitelist based on the allowed modules
	whitelist := make(map[string]bool)
	for _, module := range modules {
		whitelist[module] = true
	}
	// Register all the APIs exposed by the services
	handler := rpc.NewServer()

	for _, api := range apis {
		if whitelist[api.Namespace] || (len(whitelist) == 0 && api.Public) {
			if err := handler.RegisterName(api.Namespace, api.Service); err != nil {
				return nil, nil, err
			}
			r.logger.Debug("HTTP registered ", "namespace ", api.Namespace)
		}
	}
	// All APIs registered, start the HTTP listener
	var (
		listener net.Listener
		err      error
	)
	network, address, err := scheme(endpoint)
	if err != nil {
		return nil, nil, err
	}
	if listener, err = net.Listen(network, address); err != nil {
		return nil, nil, err
	}

	hServer := new(http.Server)
	go func(hServer *http.Server) {
		hServer = rpc.NewHTTPServer(cors, vhosts, timeouts, handler)
		hServer.Serve(listener)
		select {
		case <-r.ctx.Done():
			hServer.Close()
		}
	}(hServer)

	return listener, handler, err
}

// StartWSEndpoint starts a websocket endpoint
func (r *RPC) StartWSEndpoint(endpoint string, apis []rpc.API, modules []string, wsOrigins []string, exposeAll bool) (net.Listener, *rpc.Server, error) {
	// Generate the whitelist based on the allowed modules
	whitelist := make(map[string]bool)
	for _, module := range modules {
		whitelist[module] = true
	}
	// Register all the APIs exposed by the services
	handler := rpc.NewServer()
	for _, api := range apis {
		if exposeAll || whitelist[api.Namespace] || (len(whitelist) == 0 && api.Public) {
			if err := handler.RegisterName(api.Namespace, api.Service); err != nil {
				return nil, nil, err
			}
			r.logger.Debug("WebSocket registered ", " service ", api.Service, " namespace ", api.Namespace)
		}
	}
	// All APIs registered, start the HTTP listener
	var (
		listener net.Listener
		err      error
	)
	network, address, err := scheme(endpoint)
	if err != nil {
		return nil, nil, err
	}

	if listener, err = net.Listen(network, address); err != nil {
		return nil, nil, err
	}

	//go rpc.NewWSServer(wsOrigins, handler).Serve(listener)
	hServer := new(http.Server)
	go func(hServer *http.Server) {
		hServer = rpc.NewWSServer(wsOrigins, handler)
		hServer.Serve(listener)
		select {
		case <-r.ctx.Done():
			hServer.Close()
		}
	}(hServer)

	return listener, handler, err
}
