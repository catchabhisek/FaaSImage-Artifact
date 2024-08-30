package main

import "C"

import (
	"fmt"
	"net"

	"golang.org/x/sys/unix"
)

const (
	netlinkPort  = 30  // Adjust as needed, matching the kernel module port
	messageGroup = unix.NETLINK_KOBJECT_UEVENT // Group for netlink communication
)

func createNetlinkSocket() (*net.Conn, error) {
	conn, err := net.Dial("unixgram", fmt.Sprintf("unix:/dev/netlink:%d", netlinkPort))
	if err != nil {
		return nil, err
	}
	// Set socket groups to receive messages from the kernel module
	nl := conn.(*net.UnixConn)
	err = nl.SetGroups(messageGroup)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

func sendMessage(conn *net.Conn, message string) error {
	_, err := conn.Write([]byte(message))
	return err
}

func receiveMessage(conn *net.Conn) (string, error) {
	data := make([]byte, 1024)
	n, err := conn.Read(data)
	if err != nil {
		return "", err
	}
	return string(data[:n]), nil
}

func main() {
	conn, err := createNetlinkSocket()
	if err != nil {
		fmt.Printf("Error creating netlink socket: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Println("Netlink socket created and bound to port", netlinkPort)

	var choice string
	for {
		fmt.Println("\nAvailable options:")
		fmt.Println("  1. Send message")
		fmt.Println("  2. Receive message")
		fmt.Println("  q. Quit")

		fmt.Scanf("%s", &choice)

		if choice == "q" || choice == "Q" {
			break
		} else if choice == "1" {
			message := ""
			fmt.Println("Enter message to send: ")
			fmt.Scanf("%s", &message)
			err := sendMessage(conn, message)
			if err != nil {
				fmt.Printf("Error sending message: %v\n", err)
			} else {
				fmt.Println("Message sent to kernel module.")
			}
		} else if choice == "2" {
			message, err := receiveMessage(conn)
			if err != nil {
				fmt.Printf("Error receiving message: %v\n", err)
			} else if message != "" {
				fmt.Println("Received message from kernel module:", message)
			} else {
				fmt.Println("No message received from kernel module.")
			}
		} else {
			fmt.Println("Invalid choice.")
		}
	}

	fmt.Println("Exiting program...")
}
