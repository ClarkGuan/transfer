package server

import (
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"sync/atomic"
	"syscall"

	"github.com/ClarkGuan/transfer/model"
	"google.golang.org/protobuf/proto"
)

var (
	ErrFileNotExist = errors.New("file name is empty")
)

func Main(args []string) error {
	flagSet := flag.NewFlagSet("transfer server", flag.ExitOnError)
	addr := flagSet.String("addr", ":54321", "address which to listen")
	output := flag.String("out", ".", "output directory")
	if err := flagSet.Parse(args); err != nil {
		return err
	}

	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	localAddr := listener.Addr()
	fmt.Printf("listener: %s", localAddr)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGSTOP)
	flag := int32(0)

	go func() {
		_ = <-signals
		atomic.StoreInt32(&flag, 1)
		listener.Close()
	}()

	for atomic.LoadInt32(&flag) == 0 {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s", err)
			continue
		}

		go handleConn(conn, *output)
	}

	fmt.Println("Server stopping.")

	return nil
}

func handleConn(conn net.Conn, dir string) {
	defer conn.Close()

	lenBuf := make([]byte, 4)
	_, err := io.ReadFull(conn, lenBuf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v", conn.RemoteAddr(), err)
		return
	}
	len := binary.BigEndian.Uint32(lenBuf)

	buf := make([]byte, len)
	_, err = io.ReadFull(conn, buf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v", conn.RemoteAddr(), err)
		return
	}

	files := new(model.Files)
	err = proto.Unmarshal(buf, files)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v", conn.RemoteAddr(), err)
		return
	}

	for _, file := range files.GetFiles() {
		if newPath, err := saveFile(file, dir); err == nil {
			fmt.Printf("File saved successful: remote(%s) -> local(%s)\n", file.GetName(), newPath)
		} else {
			fmt.Fprintf(os.Stderr, "File saved failed: %s", err)
		}
	}
}

func saveFile(file *model.File, dir string) (string, error) {
	if file == nil {
		return "", nil
	}

	if file.Name == nil {
		return "", ErrFileNotExist
	}

	name := filepath.Join(dir, filepath.Base(file.GetName()))
	if info, _ := os.Stat(name); info != nil {
		return "", fmt.Errorf("%s already exist", name)
	}

	if err := os.WriteFile(name, file.Content, 0666); err != nil {
		return "", err
	}

	abs, err := filepath.Abs(name)
	if err != nil {
		return "", err
	}
	return abs, nil
}
