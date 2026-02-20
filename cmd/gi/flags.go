package main

import (
	"os"
	"strconv"
	"strings"
)

func flagValueString(f string) string {
	if !strings.Contains(f, "=") {
		return ""
	}
	return strings.Split(f, "=")[1]
}

func flagValueInt(f string) int {
	if !strings.Contains(f, "=") {
		return 0
	}
	is := strings.Split(f, "=")[1]
	i, err := strconv.Atoi(is)
	if err != nil {
		return 0
	}
	return i
}

func getListenFlag() string {
	for _, each := range os.Args {
		if strings.HasPrefix(each, "--listen") {
			return each
		}
	}
	return ""
}

func hasDAPCommand() bool {
	return len(os.Args) > 1 && os.Args[1] == "dap"
}

func hasRunCommand() bool {
	return len(os.Args) > 1 && os.Args[1] == "run"
}

func hasReplCommand() bool {
	return len(os.Args) > 1 && os.Args[1] == "repl"
}

func hasStepCommand() bool {
	return len(os.Args) > 1 && os.Args[1] == "step"
}

func getLogDestFlag() string {
	for _, each := range os.Args {
		if strings.HasPrefix(each, "--log-dest") {
			return each
		}
	}
	return ""
}
func getLogFlag() bool {
	for _, each := range os.Args {
		if each == "--log" {
			return true
		}
	}
	return false
}
