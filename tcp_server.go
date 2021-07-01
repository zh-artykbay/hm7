package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"time"
)

const (
	PORT = ":8082"
)

func main() {
	ctx, cancelFunc := context.WithCancel(context.Background())

	go goroutineProcessSignals(cancelFunc)

	if err := listenSocket(ctx); err != nil {
		log.Fatal(err)
	}
}

func goroutineProcessSignals(cancelFunc context.CancelFunc) {
	signalChan := make(chan os.Signal)

	signal.Notify(signalChan, os.Interrupt)

	for {
		sig := <-signalChan
		switch sig {

		case os.Interrupt:
			log.Println("Signal SIGINT is received, probably due to `Ctrl-C`, exiting ...")
			cancelFunc()
			return
		}
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		message := scanner.Text()
		fmt.Println("Message Received:", message)
		num, _ := strconv.Atoi(message)
		newMessage := num * num
		conn.Write([]byte(strconv.Itoa(newMessage) + "\n"))
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("error:", err)
	}
}

func listenSocket(ctx context.Context) error {
	localAddr, err := net.ResolveTCPAddr("tcp", PORT)
	if err != nil {
		return err
	}

	l, err := net.ListenTCP("tcp", localAddr)
	if err != nil {
		return err
	}

	defer l.Close()
	log.Println("Start listening on the TCP socket", PORT, ".")

	for {
		select {
		case <-ctx.Done():
			log.Println("Stop listening on the TCP socket", PORT, ".")
			return nil

		default:
			if err := l.SetDeadline(time.Now().Add(time.Second)); err != nil {
				return err
			}

			conn, err := l.Accept()
			if err != nil {
				if os.IsTimeout(err) {
					continue
				}
				return err
			}
			go handleConnection(conn)

			log.Println("New connection to the listening TCP socket", PORT, ".")
		}
	}
}
