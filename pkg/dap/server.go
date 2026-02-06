// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This file contains modified source from https://github.com/google/go-dap

package dap

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/google/go-dap"
)

type Server struct {
	Addr string
}

// Starts a server that listens on a specified port
// and blocks indefinitely. This server can accept multiple
// client connections at the same time.
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	// Line must start with "DAP server listening at:"
	// see https://github.com/golang/vscode-go/blob/f907536117c3e9fc731be9277e992b8cc7cd74f1/extension/src/goDebugFactory.ts#L558
	fmt.Println("DAP server listening at:", s.Addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Connection failed:", err)
			continue
		}
		log.Println("Accepted connection from", conn.RemoteAddr())
		// Handle multiple client connections concurrently
		go s.handleConnection(conn)
	}
}

// handleConnection handles a connection from a single client.
// It reads and decodes the incoming data and dispatches it
// to per-request processing goroutines. It also launches the
// sender goroutine to send resulting messages over the connection
// back to the client.
func (s *Server) handleConnection(conn net.Conn) {
	session := &dapSession{
		rw:           bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
		sendQueue:    make(chan dap.Message),
		stopStepping: make(chan struct{}),
	}
	go session.sendFromQueue()

	for {
		err := session.handleRequest()
		if err != nil {
			if err == io.EOF {
				log.Println("No more data to read:", err)
				break
			}
			// There maybe more messages to process, but
			// we will start with the strict behavior of only accepting
			// expected inputs.
			log.Fatal("Server error: ", err)
		}
	}

	log.Println("Closing connection from", conn.RemoteAddr())
	close(session.stopStepping)
	session.sendWg.Wait()
	close(session.sendQueue)
	conn.Close()
}
