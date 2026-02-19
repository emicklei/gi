package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/emicklei/gi"
	"github.com/emicklei/gi/pkg"
	"github.com/emicklei/gi/pkg/dap"
	godap "github.com/google/go-dap"
)

// The gi tool has three modes of operation:
//
//  1. Start a DAP server for debugging
//     Mode 1 is activated when the dap sub command is invoked.
//  2. Start a REPL for interactive terminal-based debugging
//     Mode 2 is activated when the repl sub command is invoked.
//  3. Run a program from a directory of file
//     Mode 3 is activated when the run sub command is invoked.
func main() {
	if hasDAPCommand() {
		startDAP()
		return
	}
	if hasReplCommand() {
		startREPL()
		return
	}
	if hasRunCommand() {
		runProgram()
		return
	}
	log.Println("[gi] unknown command")
}

// This accepts the Delve (dlv) flags and args because that is hardcoded in the vscode-go plugin. Example is:
//
// gi dap --listen=127.0.0.1:52950 --log-dest=3 --log
func startDAP() {
	addr := flagValueString(getListenFlag())
	srv := dap.Server{Addr: addr}
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
}

func startREPL() {
	gopkg, err := pkg.LoadPackage(".", nil)
	if err != nil {
		log.Fatal(err)
	}
	ipkg, err := pkg.BuildPackage(gopkg)
	if err != nil {
		log.Fatal(err)
	}
	runner := pkg.NewDAPAccess(pkg.NewVM(ipkg))
	runner.Launch("main", nil)

	// loop
	var b []byte = make([]byte, 1)
	fmt.Println("press any key to step, 'q' to quit")
	for {
		os.Stdin.Read(b)
		if b[0] == 'q' {
			os.Exit(0)
		}
		if err := runner.Next(); err != nil {
			if err == io.EOF {
				return
			}
			log.Fatal(err)
		}
		for _, thread := range runner.Threads() {
			args := godap.StackTraceArguments{ThreadId: thread.Id, StartFrame: 0, Levels: 10}
			frames := runner.StackFrames(args)
			for _, frame := range frames {
				fmt.Printf("[thread %d] %s @ %s:%d\n", thread.Id, frame.Name, frame.Source.Path, frame.Line)
			}
		}
	}
}
func runProgram() {
	if err := gi.Run("."); err != nil {
		print(err.Error())
		os.Exit(1)
	}
}
