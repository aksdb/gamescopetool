package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"sync"
	"time"

	ipc "github.com/james-barrow/golang-ipc"
	"golang.design/x/clipboard"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	if err := clipboard.Init(); err != nil {
		log.Fatalf("cannot initialize clipboard: %v", err)
	}

	args := ParseArgs(os.Args)

	if args.ClientSocket != "" {
		runClient(args)
	} else {
		runServer(args)
	}
}

func runClient(args Args) {
	client, err := ipc.StartClient(args.ClientSocket, nil)
	if err != nil {
		log.Fatalf("cannot initialize client IPC: %v", err)
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

	gamescopeCmd := exec.Command(args.GameAndArgs[0], args.GameAndArgs[1:]...)
	gamescopeCmd.Run()
}

func runServer(args Args) {
	socketName := randomName()
	if len(args.GameAndArgs) > 0 {
		server, err := ipc.StartServer(socketName, nil)
		if err != nil {
			log.Fatalf("cannot initialize server IPC: %v", err)
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
	}

	var cmdArgs []string
	cmdArgs = append(cmdArgs, args.GamescopeArgs...)
	if len(args.GameAndArgs) > 0 {
		cmdArgs = append(cmdArgs, "--")
		cmdArgs = append(cmdArgs, os.Args[0], "-client", socketName)
		cmdArgs = append(cmdArgs, args.GameAndArgs...)
	}

	clientCommand := exec.Command("gamescope", cmdArgs...)
	clientCommand.Stdout = os.Stdout
	clientCommand.Stderr = os.Stderr
	if err := clientCommand.Run(); err != nil {
		log.Printf("gamescope finished with error: %v", err)
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
