package main

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	ipc "github.com/james-barrow/golang-ipc"
	"golang.design/x/clipboard"
)

type Syncer struct {
	mtx    sync.Mutex
	socket commSocket

	clipboardContent []byte

	wg         sync.WaitGroup
	ctx        context.Context
	cancelFunc context.CancelFunc
}

type commSocket interface {
	Read() (*ipc.Message, error)
	Write(msgType int, message []byte) error
	StatusCode() ipc.Status
}

func (s *Syncer) Start(ctx context.Context, socket commSocket) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.socket = socket
	s.clipboardContent = clipboard.Read(clipboard.FmtText)
	s.wg.Add(2)

	s.ctx, s.cancelFunc = context.WithCancel(ctx)
	go func() {
		<-ctx.Done()
		s.Stop()
	}()

	go func() {
		defer s.wg.Done()
		defer fmt.Println("socket listener stopped")

		for {
			msg, err := socket.Read()
			if err != nil {
				fmt.Printf("got error on read: %v\n", err)
				break
			}
			if msg.MsgType == 1 {
				s.mtx.Lock()
				clipboard.Write(clipboard.FmtText, msg.Data)
				s.clipboardContent = msg.Data
				s.mtx.Unlock()
			}
		}
	}()

	initChan := make(chan struct{})

	go func() {
		defer s.wg.Done()

		for {
			if socket.StatusCode() == ipc.Connected {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		close(initChan)

		defer fmt.Println("clipboard listener stopped")
		clipboardChanges := clipboard.Watch(s.ctx, clipboard.FmtText)
		for newContent := range clipboardChanges {
			s.mtx.Lock()
			if !bytes.Equal(newContent, s.clipboardContent) {
				_ = socket.Write(1, newContent)
				s.clipboardContent = newContent
			}
			s.mtx.Unlock()
		}
	}()

	<-initChan
}

func (s *Syncer) Stop() {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.ctx != nil {
		s.cancelFunc()
	}

	s.wg.Wait()
	s.ctx = nil
}

func (s *Syncer) ForceSync() {
	_ = s.socket.Write(1, s.clipboardContent)
}
