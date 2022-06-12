package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"sync"
	"time"

	ipc "github.com/james-barrow/golang-ipc"
	"golang.design/x/clipboard"
)

var clientSocket = flag.String("client", "", "Name of the socket to be the client for.")

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	if err := clipboard.Init(); err != nil {
		panic(err)
	}

	if *clientSocket != "" {
		runClient()
	} else {
		runServer()
	}
}

func runClient() {
	fmt.Printf("client for %q\n", *clientSocket)
	client, err := ipc.StartClient(*clientSocket, nil)
	if err != nil {
		panic("##1 " + err.Error())
	}
	defer client.Close()

	var mtx sync.Mutex
	var clipboardContent = clipboard.Read(clipboard.FmtText)
	var watchCtx, cancelWatch = context.WithCancel(context.Background())
	defer cancelWatch()

	go func() {
		for {
			msg, err := client.Read()
			if err != nil {
				fmt.Printf("got error on client read: %v\n", err)
				break
			}
			if msg.MsgType == 1 {
				mtx.Lock()
				clipboard.Write(clipboard.FmtText, msg.Data)
				clipboardContent = msg.Data
				mtx.Unlock()
			}
		}
	}()

	for {
		if client.StatusCode() == ipc.Connected {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	go func() {
		clipboardChanges := clipboard.Watch(watchCtx, clipboard.FmtText)
		for newContent := range clipboardChanges {
			mtx.Lock()
			if !bytes.Equal(newContent, clipboardContent) {
				_ = client.Write(1, newContent)
				clipboardContent = newContent
			}
			mtx.Unlock()
		}
	}()

	args := flag.Args()
	gamescopeCmd := exec.Command(args[0], args[1:]...)
	gamescopeCmd.Run()
}

func runServer() {
	socketName := randomName()
	fmt.Printf("server for %q\n", socketName)
	server, err := ipc.StartServer(socketName, nil)
	if err != nil {
		panic(err)
	}
	defer server.Close()

	var mtx sync.Mutex
	var clipboardContent = clipboard.Read(clipboard.FmtText)
	var watchCtx, cancelWatch = context.WithCancel(context.Background())
	defer cancelWatch()

	go func() {
		for {
			if server.StatusCode() == ipc.Connected {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

		_ = server.Write(1, clipboardContent)

		clipboardChanges := clipboard.Watch(watchCtx, clipboard.FmtText)
		for newContent := range clipboardChanges {
			mtx.Lock()
			if !bytes.Equal(newContent, clipboardContent) {
				_ = server.Write(1, newContent)
				clipboardContent = newContent
			}
			mtx.Unlock()
		}
	}()

	go func() {
		for {
			msg, err := server.Read()
			if err != nil {
				fmt.Printf("got error on server read: %v\n", err)
				break
			}
			if msg.MsgType == 1 {
				mtx.Lock()
				clipboard.Write(clipboard.FmtText, msg.Data)
				clipboardContent = msg.Data
				mtx.Unlock()
			}
		}
	}()

	allArgs := flag.Args()
	var gamescopeArgs []string
	var gameArgs = allArgs
	for i := range allArgs {
		if allArgs[i] == "--" {
			gamescopeArgs = allArgs[:i]
			gameArgs = allArgs[i+1:]
			break
		}
	}

	var args []string
	args = append(args, gamescopeArgs...)
	args = append(args, "--")
	args = append(args, os.Args[0], "-client", socketName)
	args = append(args, gameArgs...)

	clientCommand := exec.Command("gamescope", args...)
	clientCommand.Stdout = os.Stdout
	clientCommand.Stderr = os.Stderr
	if err := clientCommand.Run(); err != nil {
		panic(err)
	}
}

var randomChars = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

func randomName() string {
	var result string
	for i := 0; i < 16; i++ {
		result = result + string(randomChars[rand.Intn(len(randomChars))])
	}
	return result
}
