package internal

import (
	"bufio"

	"github.com/google/go-dap"
)

type dapAdapter struct {
	vm               *VM
	rw               bufio.ReadWriter
	seq              int
	activeStackFrame *dap.StackFrame
}

func (a *dapAdapter) send(msg dap.Message) error {
	return dap.WriteProtocolMessage(a.rw, msg)
}
func (a *dapAdapter) receive() (dap.Message, error) {
	msg, err := dap.ReadProtocolMessage(a.rw.Reader)
	if mseq := msg.GetSeq(); mseq > a.seq {
		a.seq = mseq
	}
	return msg, err
}
func (a *dapAdapter) newRequest(command string) dap.Request {
	a.seq++
	return dap.Request{
		ProtocolMessage: dap.ProtocolMessage{
			Seq:  a.seq,
			Type: "request",
		},
		Command: command,
	}
}

// talk to the debugger directly without stdio or tcp.
type dapDriver struct {
	adapter *dapAdapter
}

func (d *dapDriver) launch() error { return nil }

func newEvent(event string) *dap.Event {
	return &dap.Event{
		ProtocolMessage: dap.ProtocolMessage{
			Seq:  0,
			Type: "event",
		},
		Event: event,
	}
}
func newResponse(requestSeq int, command string) *dap.Response {
	return &dap.Response{
		ProtocolMessage: dap.ProtocolMessage{
			Seq:  0,
			Type: "response",
		},
		Command:    command,
		RequestSeq: requestSeq,
		Success:    true,
	}
}
func newErrorResponse(requestSeq int, command string, message string) *dap.ErrorResponse {
	er := &dap.ErrorResponse{}
	er.Response = *newResponse(requestSeq, command)
	er.Success = false
	er.Message = "unsupported"
	er.Body.Error.Format = message
	er.Body.Error.Id = 12345
	return er
}
