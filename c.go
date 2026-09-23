package main

import (
	// "strings"
	"io"
	"net"
	"os"
	"fmt"
)

func runclient() error {
	if _, err := os.Stat(ipcpath); err != nil {
		fmt.Fprintf(os.Stderr, "cannot stat %s: %v. is daemon running?\n", ipcpath, err)
		os.Exit(1)
	}
	conn, err := net.Dial("unix", ipcpath)
	if err != nil {
		return fmt.Errorf("cannot dial: %w", err)
	}
	defer conn.Close()
	communicate := func (request string) (string, error) {
		if _, err := conn.Write([]byte(request)); err != nil { return "", err }
		buf, err := io.ReadAll(conn)
		if err != nil { return "", fmt.Errorf("can't read: %w", err) }
		return string(buf), nil
	}
	switch {
	case quit:
		s, err := communicate("-Q")
		if err != nil { return err }
		if !quiet { fmt.Println(s) }
	case list:
		s, err := communicate("-l")
		if err != nil { return err }
		fmt.Println(s)
	case add != "":
		s, err := communicate(fmt.Sprintf("-a %s", add))
		if err != nil { return err }
		if !quiet { fmt.Println(s) }
	case cancel != "":
		s, err := communicate(fmt.Sprintf("-r %s", cancel))
		if err != nil { return err }
		if !quiet { fmt.Println(s) }
	default:
		return fmt.Errorf("command is unknown or unpecified")
	}
	return nil
}
