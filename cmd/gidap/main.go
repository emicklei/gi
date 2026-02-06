package main

import (
	"log"

	"github.com/emicklei/gi/pkg/dap"
)

// This program accepts the Delve (dlv) flags and args because that is hardcoded in the vscode-go plugin. Example is:
//
//	dap --listen=127.0.0.1:52950 --log-dest=3 --log
func main() {
	addr := flagValueString(getListenFlag())
	srv := dap.Server{Addr: addr}
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
}
