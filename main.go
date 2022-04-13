package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	var host string
	var port string

	_ = flag.String("host", "", "Host address")
	_ = flag.String("port", "", "Port")

	flag.Parse()
	if flag.NArg() == 2 {
		host = flag.Arg(0)
		port = flag.Arg(1)
	} else {
		flag.Usage()
		return
	}

	// Channel to connect user input and goroutine
	messagePool := make(chan string)
	responsePool := make(chan string)

	fmt.Println("Trying " + host)
	go initConnection("tcp", host, port, messagePool, responsePool)

	p := tea.NewProgram(initialModel(messagePool, responsePool))
	if err := p.Start(); err != nil {
		fmt.Printf("Error occured: %v", err)
		os.Exit(1)
	}
}

// initConnection connects to remove host via protocol provided in network string
// awaits for data from messagePull and writes responses into responsePull
func initConnection(network string, host string, port string, messagePull chan string, responsePull chan string) {

	conn, err := net.Dial(network, host+":"+port)
	if err != nil {
		panic(err)
	}

	fmt.Println("Connected to " + host)
	fmt.Println("To exit press ESC or Ctrl+C")

	defer func() {
		_ = conn.Close()
	}()

	for {
		message := <-messagePull
		// Send to socket
		_, err := fmt.Fprint(conn, message)
		if err == nil {
			// Listen response
			tmp := make([]byte, 4096)
			n, err := conn.Read(tmp)
			if err != nil {
				fmt.Println("Error, Connection closed")
				os.Exit(1)
			} else {
				responsePull <- string(tmp[0:n])
				fmt.Println(string(tmp[0:n]))
			}
		} else {
			fmt.Println("Error, Connection closed")
			os.Exit(1)
		}
	}
}

type model struct {
	index        int
	history      []string
	buffer       string
	response     string
	messagePool  chan string
	responsePool chan string
}

func initialModel(messagePool chan string, responsePool chan string) model {
	return model{
		index:        -1,
		buffer:       "",
		response:     "",
		messagePool:  messagePool,
		responsePool: responsePool,
	}
}

func (m model) Init() tea.Cmd {
	go func() {
		for {
			m.response = <-m.responsePool
		}
	}()
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:

		switch msg.String() {
		case "ctrl+q":
			m.buffer = "Wake up Neo...\nThe Matrix has you...\n" +
				"Follow the white rabbit.\n\n\nKnock, knock, Neo.\n"
			fallthrough
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.index < len(m.history) {
				m.index++
			}
			if m.index < len(m.history) {
				m.buffer = m.history[len(m.history)-1-m.index]
			} else {
				m.buffer = ""
			}
		case "down", "j":
			if m.index > -1 {
				m.index--
			}
			if m.index >= 0 && len(m.history) > 0 {
				m.buffer = m.history[len(m.history)-1-m.index]
			} else {
				m.buffer = ""
			}
		case "backspace":
			m.buffer = m.buffer[:(len(m.buffer) - 1)]
		case "enter":
			m.messagePool <- m.buffer + "\n"
			if m.buffer != "" &&
				(len(m.history) == 0 || m.history[len(m.history)-1] != m.buffer) {
				m.history = append(m.history, m.buffer)
			}
			m.buffer = ""
			m.index = -1
		default:
			if len(msg.String()) == 1 {
				m.buffer += msg.String()
			}
		}
	}

	return m, nil
}

func (m model) View() (s string) {
	if len(m.buffer) == 0 {
		s = m.response
	} else if len(m.response) == 0 {
		s = m.buffer
	}
	return s
}
