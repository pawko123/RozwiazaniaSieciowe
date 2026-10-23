package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"Lab2/shared/protocol"
)

var ErrServerError = errors.New("serwer zgłosił błąd (error)")

type Client struct {
	addr    string
	timeout time.Duration
}

func NewClient(addr string) *Client {
	if addr == "" {
		addr = fmt.Sprintf("127.0.0.1:%d", protocol.DefaultPort)
	}
	return &Client{
		addr:    addr,
		timeout: 5 * time.Second,
	}
}

func (client *Client) CheckKey(id string) (bool, error) {
	req := &protocol.Request{
		Command: protocol.CmdGet,
		NrID:    id,
	}
	resp, err := client.Send(req.String())
	if err != nil {
		return false, err
	}
	switch resp.Contents {
	case protocol.ContentYes:
		return true, nil
	case protocol.ContentNot:
		return false, nil
	case protocol.ContentError:
		return false, ErrServerError
	default:
		return false, fmt.Errorf("nieznana odpowiedź serwera: %s", resp.Contents)
	}
}

func (client *Client) AddKey(password, id string) (bool, error) {
	req := &protocol.Request{
		Command:  protocol.CmdSet,
		Password: password,
		NrID:     id,
	}
	resp, err := client.Send(req.String())
	if err != nil {
		return false, err
	}
	switch resp.Contents {
	case protocol.ContentYes:
		return true, nil
	case protocol.ContentNot:
		return false, nil
	case protocol.ContentError:
		return false, ErrServerError
	default:
		return false, fmt.Errorf("nieznana odpowiedź serwera: %s", resp.Contents)
	}
}

func (client *Client) Send(rawMessage string) (*protocol.Response, error) {
	conn, err := net.DialTimeout("tcp", client.addr, client.timeout)
	if err != nil {
		return nil, fmt.Errorf("nie udało się połączyć z serwerem (%s): %w", client.addr, err)
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(client.timeout)); err != nil {
		return nil, fmt.Errorf("błąd limitu czasu: %w", err)
	}

	if _, err := fmt.Fprintf(conn, "%s\n", strings.TrimSpace(rawMessage)); err != nil {
		return nil, fmt.Errorf("błąd wysyłania danych: %w", err)
	}

	reader := bufio.NewReader(conn)
	rawResp, err := reader.ReadString(';')
	if err != nil {
		return nil, fmt.Errorf("błąd odczytu odpowiedzi: %w", err)
	}

	return protocol.ParseResponse(rawResp)
}
