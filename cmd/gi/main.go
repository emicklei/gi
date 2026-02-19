package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/emicklei/gi"
	"github.com/emicklei/gi/pkg"
	"github.com/emicklei/gi/pkg/dap"
)

// The gi tool has three modes of operation:
//
//  1. Start a DAP server for debugging
//     Mode 1 is activated when the dap sub command is invoked.
//  2. Start a REPL for interactive terminal-based debugging
//     Mode 2 is activated when the --repl flag is provided.
//  3. Run a program from a directory of file
//     Mode 3 is activated when neither dap nor --repl are provided.
func main() {
	if hasDAPCommand() {
		startDAP()
	} else if getREPLFlag() {
		startREPL()
	} else {
		runProgram()
	}
}

// This accepts the Delve (dlv) flags and args because that is hardcoded in the vscode-go plugin. Example is:
//
// gi --listen=127.0.0.1:52950 --log-dest=3 --log
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
	runner := pkg.NewVM(ipkg)
	runner.Launch("main", nil)
	var b []byte = make([]byte, 1)
	fmt.Println("Press any key to step, 'q' to quit")
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
		fmt.Println(runner.Location())
	}

}
func runProgram() {
	if err := gi.Run("."); err != nil {
		print(err.Error())
		os.Exit(1)
	}
}
