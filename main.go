package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
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

	var syncer Syncer
	syncer.Start(context.Background(), client)
	defer func() {
		client.Close()
		syncer.Stop()
	}()

	runGame := func() {
		gamescopeCmd := exec.Command(args.GameAndArgs[0], args.GameAndArgs[1:]...)
		gamescopeCmd.Run()
	}

	if args.ShowDummyWindow {
		a := app.New()
		w := a.NewWindow("gamescopetool dummy windows")
		w.SetContent(widget.NewLabel("This is just a window to keep gamescope happy."))

		go func() {
			runGame()
			w.Close()
		}()

		w.ShowAndRun()
		return
	}

	runGame()
}

func runServer(args Args) {
	socketName := randomName()
	if len(args.GameAndArgs) > 0 {
		server, err := ipc.StartServer(socketName, nil)
		if err != nil {
			log.Fatalf("cannot initialize server IPC: %v", err)
		}

		var syncer Syncer
		initChan := make(chan struct{})

		defer func() {
			<-initChan
			server.Close()
			// Ignore the syncer cleanup. The server doesn't properly close the socket
			// locking the syncer in its worker loop. Until fixed, just let the OS
			// cleanup after the process exits.
			//syncer.Stop()
		}()

		go func() {
			defer close(initChan)

			syncer.Start(context.Background(), server)
			syncer.ForceSync()
		}()
	}

	var cmdArgs []string
	cmdArgs = append(cmdArgs, args.GamescopeArgs...)
	if len(args.GameAndArgs) > 0 {
		cmdArgs = append(cmdArgs, "--")
		cmdArgs = append(cmdArgs, os.Args[0], "--client", socketName)
		if args.ShowDummyWindow {
			cmdArgs = append(cmdArgs, "--dummy-window")
		}
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
