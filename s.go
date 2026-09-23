package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Queue struct {
	// qid string
	tasks []Task
}
type ExecResult struct {
	statuscode int
	// stdout string
	// stderr string
}
type Task struct {
	id string
	// restartpolicy int
	completed bool
	result ExecResult
	exec string
}

type Request struct { 
	cmd string
	client chan string 
}


func taskexec(taskDoneCh chan ExecResult, task Task, taskCancelCh chan struct{}) {
	// fmt.Fprintf(os.Stderr, "%s #%s\n", task.exec, task.id)
	fmt.Fprintf(os.Stderr, "# start %s\n%s\n", task.id, task.exec)
	// var stdout, stderr bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		select {
		case <-taskCancelCh:
			cancel()
		case <-ctx.Done():
		}
	}()

	cmd := exec.CommandContext(ctx, "sh", "-c", task.exec)
	// xxx: maybe just stream it to stdout/stderr?
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	// cmd.Stdout = &stdout
	// cmd.Stderr = &stderr
    err := cmd.Run()
	result := ExecResult{0}
    if err != nil {
		if ctx.Err() == context.Canceled {
			result.statuscode = -2
		} else if exitErr, ok := err.(*exec.ExitError); ok {
			result.statuscode = exitErr.ExitCode()
		} else {
			result.statuscode = -1
		}
    }
	taskDoneCh <- result
	// fmt.Fprintf(os.Stderr, "%s: error:\n%s", task.iid, err.Error())

	// fmt.Fprintf(os.Stderr, "%s: stdout:\n%s", task.iid, string(st))
}

func runserver() error {
    //  defer func() {
    //     if r := recover(); r != nil {
    //         err = fmt.Errorf("panic: %v", r)
    //     }
    // }()
	var queue Queue
	currentTaskIdx := 0

	if conn, err := net.Dial("unix", ipcpath); err == nil {
		conn.Close()
		return fmt.Errorf("server already running on %s", ipcpath)
	}
	if err := os.Remove(ipcpath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("rm socket %s: %w", ipcpath, err)
	}
	ln, err := net.Listen("unix", ipcpath)
	if err != nil {
		return fmt.Errorf("cannot listen: %v", err)
	}
	defer os.Remove(ipcpath)
	defer ln.Close()

	taskDoneCh := make(chan ExecResult, 1)
	taskCancelCh := make(chan struct{}, 1)
	execrunning := false

	errCh := make(chan error, 1) // fatal errors
	requestchan := make(chan Request, 1)
	go func() { 
		for {
			conn, err := ln.Accept()
			if err != nil {
				errCh <- fmt.Errorf("cannot accept: %v", err)
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				buf := make([]byte, 1024)
				n, _ := c.Read(buf)
				cmd := string(buf[:n])
				client := make(chan string, 1)
				requestchan<-Request{cmd, client}
				for s := range client { // iterate channel until closed
					_, err := c.Write([]byte(s))
					if err != nil {

					}
				}
			}(conn)
		}
	}()
	mainloop: for {
		select {
		case err := <-errCh:
			return err
		case r := <-requestchan:
			// defer close(r.client) ?
			switch r.cmd[:2] {
				case"-Q":
				r.client <- "0"
				close(r.client)
				break mainloop
			case "-l":
				var s strings.Builder
				for _, task := range queue.tasks {
					c := ""; if task.completed { c = fmt.Sprintf("+%d", task.result.statuscode) }
					fmt.Fprintf(&s, "%s #%s%s\n", task.exec, task.id, c)
				}
				r.client <- s.String()
				close(r.client)
			case "-a":
				var task Task
				task.id = strconv.FormatInt(time.Now().UnixMicro(), 10)
				task.exec = r.cmd[3:]
				queue.tasks = append(queue.tasks, task)
				if !execrunning {
					execrunning = true
					go taskexec(taskDoneCh, queue.tasks[currentTaskIdx], taskCancelCh)
				}
				r.client <- task.id
				close(r.client)
			case "-r":
				id := r.cmd[3:]
				for i, task := range queue.tasks {
					if task.id == id {
						if task.completed {
							r.client <- "cannot cancel completed task"
							close(r.client)
							continue mainloop
						}
						if i == 0 || queue.tasks[i-1].completed {
							taskCancelCh <- struct{}{}
							r.client <- "ok" // TODO?
							close(r.client)
							continue mainloop
						}
						queue.tasks = append(queue.tasks[:i], queue.tasks[i+1:]...)
						r.client <- "ok"
						close(r.client)
						continue mainloop
					}
				}
				r.client <- "not found"
				close(r.client)
			default:
			}
		case result := <-taskDoneCh:
			task := &queue.tasks[currentTaskIdx]
			task.completed = true
			task.result = result
			// c := ""; if task.completed { c = fmt.Sprintf("+%d", task.result.statuscode) }
			fmt.Fprintf(os.Stderr, "# completed with status %d\n", task.result.statuscode)
			currentTaskIdx++
			execrunning = false
			if currentTaskIdx <= len(queue.tasks)-1 { // task exists
				execrunning = true
				go taskexec(taskDoneCh, queue.tasks[currentTaskIdx], taskCancelCh)
			}
		}
	}
	return nil
}
