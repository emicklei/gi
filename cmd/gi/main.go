package main

import (
	"log"
	"os"

	"github.com/emicklei/gi"
	"github.com/emicklei/gi/pkg/dap"
)

// The gi tool has multiple modes of operation:
//
//  1. Start a DAP server for debugging
//     Is activated when the dap sub command is invoked.
//  2. Run a program from a directory of file
//     Is activated when the run sub command is invoked.
//  3. Step through a program from a directory of file
//     Is activated when the step sub command is invoked.
func main() {
	if hasDAPCommand() {
		startDAP()
		return
	}
	if hasStepCommand() {
		startStepper()
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

func runProgram() {
	if err := gi.Run("."); err != nil {
		print(err.Error())
		os.Exit(1)
	}
}
