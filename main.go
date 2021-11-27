package main

import (
	"flag"
	"fmt"
	"github.com/eiannone/keyboard"
	"net"
	"os"
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
	messagePull := make(chan string)
	responsePull := make(chan string)

	clearConsole()
	fmt.Println("Trying " + host)

	go initConnection("tcp", host, port, messagePull, responsePull)

	// response printer
	go func() {
		for {
			response := <-responsePull
			fmt.Println(response)
		}
	}()

	startInputProcessing(messagePull)

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
		_, err := fmt.Fprintf(conn, message)
		if err == nil {
			// Listen response
			tmp := make([]byte, 4096)
			n, err := conn.Read(tmp)
			if err != nil {
				fmt.Println("Error, Connection closed")
				os.Exit(1)
			} else {
				responsePull <- string(tmp[0:n])
				//fmt.Println(string(tmp[0:n]))
			}
		} else {
			fmt.Println("Error, Connection closed")
			os.Exit(1)
		}
	}
}

// startInputProcessing runs keyboard listener, which collects user input in local buffer
// message goes into messagePull when user press Enter
// arrows up and down scrolls between written messages
func startInputProcessing(messagePull chan string) {

	keysEvents, err := keyboard.GetKeys(10)
	if err != nil {
		panic(err)
	}

	defer func() {
		_ = keyboard.Close()
	}()

	var history []string
	var buffer string
	historyIterator := 0

readLoop:
	for {
		event := <-keysEvents
		if event.Err != nil {
			fmt.Println(event.Err)
			continue // Just drop event if key don't recognized
			//panic(event.Err)
		}

		switch event.Key {
		case keyboard.KeyArrowUp:
			// History popup
			if historyIterator < len(history) {
				historyIterator++
				buffer = history[len(history)-historyIterator]
			}
			clearConsole()
			fmt.Print(buffer)
			break
		case keyboard.KeyArrowDown:
			// History popup
			if historyIterator > 1 {
				historyIterator--
				buffer = history[len(history)-historyIterator]
				clearConsole()
				fmt.Print(buffer)
			} else {
				clearConsole()
				buffer = ""
			}
			break
		case keyboard.KeyEsc:
			fallthrough
		case keyboard.KeyCtrlC:
			// Exit
			break readLoop
		case keyboard.KeyBackspace:
		case keyboard.KeyBackspace2:
			// remove last character
			if len(buffer) > 0 {
				buffer = buffer[:(len(buffer) - 1)]
				clearConsole()
				fmt.Print(buffer)
			}
			break
		case keyboard.KeyEnter:
			// Send message
			if len(buffer) > 0 {
				if len(history) == 0 || history[len(history)-1] != buffer {
					history = append(history, buffer)
				}
				buffer += "\n"
				messagePull <- buffer
			} else {
				messagePull <- "\n"
			}
			fmt.Print("\n")
			buffer = ""
			historyIterator = 0
			break
		case keyboard.KeySpace:
			// HACK github.com/eiannone/keyboard returns only Key value for KeySpace, Rune is empty
			clearConsole()
			buffer += " "
			fmt.Print(buffer)
			break
		default:
			clearConsole()
			buffer += string(event.Rune)
			fmt.Print(buffer)
			break
		}
	}
}

func clearConsole() {
	fmt.Print("\033[H\033[2J")
}
