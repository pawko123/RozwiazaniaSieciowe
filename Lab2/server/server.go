package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

type Server struct {
	addr     string
	handler  *Handler
	listener net.Listener
	quit     chan struct{}
	wg       sync.WaitGroup
}

func NewServer(addr string, handler *Handler) *Server {
	return &Server{
		addr:    addr,
		handler: handler,
		quit:    make(chan struct{}),
	}
}

func (server *Server) Start() error {
	listener, err := net.Listen("tcp", server.addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", server.addr, err)
	}
	server.listener = listener
	log.Printf("[SERWER] Nasłuchiwanie na TCP %s...", server.addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-server.quit:
				return nil
			default:
				log.Printf("[SERWER] Błąd Accept: %v", err)
				return err
			}
		}

		server.wg.Add(1)
		go server.handleConn(conn)
	}
}

func (server *Server) handleConn(conn net.Conn) {
	defer server.wg.Done()
	defer conn.Close()
	remoteAddr := conn.RemoteAddr().String()
	log.Printf("[POŁĄCZENIE] Nowy klient połączony: %s", remoteAddr)

	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString(';')
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Printf("[POŁĄCZENIE] Klient %s rozłączył się (EOF)", remoteAddr)
			} else {
				select {
				case <-server.quit:
				default:
					log.Printf("[POŁĄCZENIE] Błąd odczytu z %s: %v", remoteAddr, err)
				}
			}
			return
		}

		trimmed := strings.TrimSpace(msg)
		if trimmed == "" {
			continue
		}

		log.Printf("[ODEBRANO] Od %s: %s", remoteAddr, trimmed)
		resp := server.handler.Handle(trimmed)
		log.Printf("[WYSŁANO] Do %s: %s", remoteAddr, resp)

		_, writeErr := fmt.Fprintf(conn, "%s\n", resp)
		if writeErr != nil {
			log.Printf("[POŁĄCZENIE] Błąd zapisu do %s: %v", remoteAddr, writeErr)
			return
		}
	}
}

func (server *Server) Close() error {
	close(server.quit)
	var err error
	if server.listener != nil {
		err = server.listener.Close()
	}
	server.wg.Wait()
	return err
}
