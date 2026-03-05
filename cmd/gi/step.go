package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/emicklei/gi/pkg"
	"github.com/emicklei/structexplorer"
	"github.com/google/go-dap"
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
	fmt.Println("Press return to step, 'v' for variables, 'q' to quit")
	for {
		os.Stdin.Read(b)
		if b[0] == 'q' {
			os.Exit(0)
		}
		if b[0] == 'v' {
			for _, thread := range runner.Threads() {
				args := godap.StackTraceArguments{ThreadId: thread.Id, StartFrame: 0, Levels: 10}
				frames := runner.StackFrames(args)
				frameId := frames[len(frames)-1].Id // last
				scopes := runner.Scopes(dap.ScopesArguments{FrameId: frameId})
				for _, scope := range scopes {
					fmt.Printf("%s %s :: %s\n", ansiGray("[scope"), scope.Name, ansiGray("]"))
					vars := runner.Variables(dap.VariablesArguments{VariablesReference: scope.VariablesReference})
					for _, v := range vars {
						fmt.Printf("  %s %s = %v\n", ansiYellow(v.Type), v.Name, v.Value)
					}
				}
				fmt.Println()
			}
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
			for i, frame := range frames {
				fmt.Printf("%s %d%s %d%s %s :: %s @ %s:%d ",
					ansiGray("[goroutine"),
					thread.Id,
					ansiGray("] [frame"),
					frame.Id,
					ansiGray("]"),
					frame.Name,
					ansiYellow(frame.InstructionPointerReference),
					ansiGray(frame.Source.Path), frame.Line)
				if i < len(frames)-1 { // last
					fmt.Println()
				}
			}
		}
	}
}

func ansiGray(s string) string {
	return "\033[90m" + s + "\033[0m"
}
func ansiYellow(s string) string {
	return "\033[93m" + s + "\033[0m"
}
