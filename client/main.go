package client

import (
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/ClarkGuan/transfer/model"
	"google.golang.org/protobuf/proto"
)

var (
	ErrNothing = errors.New("nothing be sent")
)

func Main(arguments []string) error {
	flagSet := flag.NewFlagSet("transfer client", flag.ExitOnError)
	addr := flagSet.String("addr", ":54321", "remote address which socket connect to")
	if err := flagSet.Parse(arguments); err != nil {
		return err
	}

	filePaths := flagSet.Args()
	if len(filePaths) == 0 {
		return ErrNothing
	}

	var files []*model.File
	for _, path := range filePaths {
		if file, err := collectFile(path); err != nil {
			fmt.Fprintf(os.Stderr, "can't handle file: %s, error: %s\n", path, err)
		} else {
			files = append(files, file)
		}
	}

	if len(files) > 0 {
		var pbFiles = model.Files{Files: files}
		if err := socketSending(*addr, &pbFiles); err != nil {
			return err
		}
	} else {
		return ErrNothing
	}

	return nil
}

func collectFile(path string) (*model.File, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	pbFile := model.File{Name: proto.String(path), Content: content}
	return &pbFile, nil
}

func socketSending(addr string, files *model.Files) error {
	buf, err := proto.Marshal(files)
	if err != nil {
		return err
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(buf)))

	if err := writeAll(conn, lenBuf, buf); err != nil {
		return err
	}

	return nil
}

func writeAll(w io.Writer, bufs ...[]byte) error {
	for _, buf := range bufs {
		if _, err := w.Write(buf); err != nil {
			return err
		}
	}

	return nil
}
