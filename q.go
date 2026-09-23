package main

import (
	"flag"
	"os"
	"fmt"
)


var ipcpath string 
var server bool
var quit bool
// var qid string
var list bool
var add string
var cancel string
var quiet bool
// var args []string

func main() {
	flag.BoolVar(&server, "s", false, "start server")

	flag.StringVar(&ipcpath, "i", "/run/user/1000/q.sock", "ipc path")

	flag.BoolVar(&quit, "Q", false, "quit server")
	// flag.StringVar(&qid, "q", "0", "queue id")
	flag.BoolVar(&list, "l", false, "list tasks")
	flag.StringVar(&add, "a", "", "enqueue task")
	flag.StringVar(&cancel, "c", "", "cancel task")

	flag.BoolVar(&quiet, "q", false, "don't print useless stuff")

	flag.Parse()
	// args = flag.Args()

	if server {
		err := runserver()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	} else {
		err := runclient()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	}
}
