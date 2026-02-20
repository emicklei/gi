package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/emicklei/gi/pkg"
	"github.com/emicklei/structexplorer"
	godap "github.com/google/go-dap"
)

func startStepper() {
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

	// serve explorer
	go structexplorer.NewService("vm", runner).Start()

	// loop
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
		for _, thread := range runner.Threads() {
			args := godap.StackTraceArguments{ThreadId: thread.Id, StartFrame: 0, Levels: 10}
			frames := runner.StackFrames(args)
			for _, frame := range frames {
				fmt.Printf("[thread %d] %s @ %s:%d\n", thread.Id, frame.Name, frame.Source.Path, frame.Line)
			}
		}
	}
}
