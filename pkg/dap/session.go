// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Copyright (c) 2026 Ernest Micklei
// This file contains modified source from https://github.com/google/go-dap

package dap

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/emicklei/gi/pkg"
	"github.com/google/go-dap"
)

// The debugging session will keep track of how many breakpoints
// have been set. Once start-up is done (i.e. configurationDone
// request is processed), it will "stop" at each breakpoint one by
// one, and once there are no more, it will trigger a terminated event.
type session struct {
	// rw is used to read requests and write events/responses
	rw *bufio.ReadWriter

	// sendQueue is used to capture messages from multiple request
	// processing goroutines while writing them to the client connection
	// from a single goroutine via sendFromQueue. We must keep track of
	// the multiple channel senders with a wait group to make sure we do
	// not close this channel prematurely. Closing this channel will signal
	// the sendFromQueue goroutine that it can exit.
	sendQueue chan dap.Message
	sendWg    sync.WaitGroup

	// stopStepping is used to notify long-running handlers to stop processing.
	stopStepping chan struct{}

	// bpSet is a counter of the remaining breakpoints that the debug
	// session is yet to stop at before the program terminates.
	bpSet    int
	bpSetMux sync.Mutex

	// vma represents program being debugged
	vma *pkg.DAPAccess
	// not sure if this is the right place
	dir string
}

func (ds *session) handleRequest() error {
	log.Println("Reading request...")
	request, err := dap.ReadProtocolMessage(ds.rw.Reader)
	if err != nil {
		return err
	}
	log.Printf("Received request\n\t%T\n", request)
	ds.sendWg.Go(func() {
		ds.dispatchRequest(request)
	})
	return nil
}

// send lets the sender goroutine know via a channel that there is
// a message to be sent to client. This is called by per-request
// goroutines to send events and responses for each request and
// to notify of events triggered by the fake debugger.
func (ds *session) send(message dap.Message) {
	ds.sendQueue <- message
}

// sendFromQueue is to be run in a separate goroutine to listen on a
// channel for messages to send back to the client. It will
// return once the channel is closed.
func (ds *session) sendFromQueue() {
	for message := range ds.sendQueue {
		dap.WriteProtocolMessage(ds.rw.Writer, message)
		log.Printf("Message sent\n\t%#v\n", message)
		ds.rw.Flush()
	}
}

// TODO make it work
func (ds *session) doContinue() {
	var e dap.Message
	ds.bpSetMux.Lock()
	if ds.bpSet == 0 {
		// Pretend that the program is running.
		// The delay will allow for all in-flight responses
		// to be sent before termination.
		time.Sleep(1000 * time.Millisecond)
		e = &dap.TerminatedEvent{
			Event: *newEvent("terminated"),
		}
	} else {
		e = &dap.StoppedEvent{
			Event: *newEvent("stopped"),
			Body:  dap.StoppedEventBody{Reason: "breakpoint", ThreadId: 1, AllThreadsStopped: true},
		}
		ds.bpSet--
	}
	ds.bpSetMux.Unlock()
	ds.send(e)
}

func (ds *session) onInitializeRequest(request *dap.InitializeRequest) {
	response := &dap.InitializeResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	response.Body.SupportsConfigurationDoneRequest = true
	response.Body.SupportsFunctionBreakpoints = false
	response.Body.SupportsConditionalBreakpoints = false
	response.Body.SupportsHitConditionalBreakpoints = false
	response.Body.SupportsEvaluateForHovers = false
	response.Body.ExceptionBreakpointFilters = []dap.ExceptionBreakpointsFilter{}
	response.Body.SupportsStepBack = false
	response.Body.SupportsSetVariable = false
	response.Body.SupportsRestartFrame = false
	response.Body.SupportsGotoTargetsRequest = false
	response.Body.SupportsStepInTargetsRequest = false
	response.Body.SupportsCompletionsRequest = false
	response.Body.CompletionTriggerCharacters = []string{}
	response.Body.SupportsModulesRequest = false
	response.Body.AdditionalModuleColumns = []dap.ColumnDescriptor{}
	response.Body.SupportedChecksumAlgorithms = []dap.ChecksumAlgorithm{}
	response.Body.SupportsRestartRequest = false
	response.Body.SupportsExceptionOptions = false
	response.Body.SupportsValueFormattingOptions = false
	response.Body.SupportsExceptionInfoRequest = false
	response.Body.SupportTerminateDebuggee = false
	response.Body.SupportsDelayedStackTraceLoading = false
	response.Body.SupportsLoadedSourcesRequest = false
	response.Body.SupportsLogPoints = false
	response.Body.SupportsTerminateThreadsRequest = false
	response.Body.SupportsSetExpression = false
	response.Body.SupportsTerminateRequest = false
	response.Body.SupportsDataBreakpoints = false
	response.Body.SupportsReadMemoryRequest = false
	response.Body.SupportsDisassembleRequest = false
	response.Body.SupportsCancelRequest = false
	response.Body.SupportsBreakpointLocationsRequest = false
	// Notify the client with an 'initialized' event. The client will end
	// the configuration sequence with 'configurationDone' request.
	e := &dap.InitializedEvent{Event: *newEvent("initialized")}
	ds.send(e)
	ds.send(response)
}

func (ds *session) onLaunchRequest(request *dap.LaunchRequest) {
	resp := new(dap.LaunchResponse)
	resp.Response = *newResponse(request.Seq, request.Command)
	resp.Success = true

	cwd, _ := os.Getwd()
	log.Println("starting program in", cwd)

	gopkg, err := pkg.LoadPackage(cwd, nil)
	if err != nil {
		log.Println("load package failed", err)
		resp.Message = "failed to load package: " + err.Error()
		resp.Success = false
		ds.send(resp)
		return
	}
	p, err := pkg.BuildPackage(gopkg)
	if err != nil {
		log.Println("build package failed", err)
		resp.Message = "failed to build package: " + err.Error()
		resp.Success = false
		ds.send(resp)
		return
	}
	ds.dir = cwd
	ds.vma = pkg.NewDAPAccess(pkg.NewVM(p))
	ds.vma.Launch("main", nil)
	ds.send(resp)

	// TODO simulate hit breakpoint
	{
		resp := &dap.StoppedEvent{}
		resp.Body.Reason = "breakpoint"
		resp.Body.ThreadId = 1
		resp.Body.AllThreadsStopped = true
		resp.Event = *newEvent("stopped")
		ds.send(resp)
	}

}

func (ds *session) onAttachRequest(request *dap.AttachRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "AttachRequest is not yet supported"))
}

func (ds *session) onDisconnectRequest(request *dap.DisconnectRequest) {
	// brutal
	ds.vma = nil
	resp := new(dap.DisconnectResponse)
	resp.Response = *newResponse(request.Seq, request.Command)
	resp.Success = true
	ds.send(resp)
}

func (ds *session) onTerminateRequest(request *dap.TerminateRequest) {
	// brutal
	ds.vma = nil
	resp := new(dap.TerminateResponse)
	resp.Response = *newResponse(request.Seq, request.Command)
	resp.Success = true
	ds.send(resp)
}

func (ds *session) onRestartRequest(request *dap.RestartRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "RestartRequest is not yet supported"))
}

func (ds *session) onSetBreakpointsRequest(request *dap.SetBreakpointsRequest) {
	log.Println("breakpoint args")
	for _, arg := range request.Arguments.Breakpoints {
		log.Printf("\t%#v\n", arg)
	}
	resp := new(dap.SetBreakpointsResponse)
	ds.send(resp)
}

func (ds *session) onSetFunctionBreakpointsRequest(request *dap.SetFunctionBreakpointsRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "SetFunctionBreakpointsRequest is not yet supported"))
}

func (ds *session) onSetExceptionBreakpointsRequest(request *dap.SetExceptionBreakpointsRequest) {
}

func (ds *session) onConfigurationDoneRequest(request *dap.ConfigurationDoneRequest) {}

func (ds *session) onContinueRequest(request *dap.ContinueRequest) {
}

func (ds *session) onNextRequest(request *dap.NextRequest) {
	if ds.vma != nil {
		ds.vma.Next()
	}
	resp := new(dap.NextResponse)
	resp.Response = *newResponse(request.Seq, request.Command)
	resp.Success = ds.vma != nil
	ds.send(resp)
}

func (ds *session) onStepInRequest(request *dap.StepInRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "StepInRequest is not yet supported"))
}

func (ds *session) onStepOutRequest(request *dap.StepOutRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "StepOutRequest is not yet supported"))
}

func (ds *session) onStepBackRequest(request *dap.StepBackRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "StepBackRequest is not yet supported"))
}

func (ds *session) onReverseContinueRequest(request *dap.ReverseContinueRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "ReverseContinueRequest is not yet supported"))
}

func (ds *session) onRestartFrameRequest(request *dap.RestartFrameRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "RestartFrameRequest is not yet supported"))
}

func (ds *session) onGotoRequest(request *dap.GotoRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "GotoRequest is not yet supported"))
}

func (ds *session) onPauseRequest(request *dap.PauseRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "PauseRequest is not yet supported"))
}

func (ds *session) onStackTraceRequest(request *dap.StackTraceRequest) {
	resp := new(dap.StackTraceResponse)
	resp.Response = *newResponse(request.Seq, request.Command)
	if ds.vma != nil {
		resp.Body.StackFrames = []dap.StackFrame{
			{Id: 1, Name: "main.main", Source: &dap.Source{Name: "main.go", Path: filepath.Join(ds.dir, "main.go")}, Line: 12},
		}
	}
	ds.send(resp)
}

// https://microsoft.github.io/debug-adapter-protocol//specification.html#Requests_Scopes
func (ds *session) onScopesRequest(request *dap.ScopesRequest) {
	resp := new(dap.ScopesResponse)
	resp.Response = *newResponse(request.Seq, request.Command)
	if ds.vma != nil {
		// https://microsoft.github.io/debug-adapter-protocol//specification.html#Types_Scope
		resp.Body.Scopes = []dap.Scope{
			{Name: "Local", VariablesReference: 1},
		}
	} else {
		resp.Success = false
	}
	ds.send(resp)
}

// https://microsoft.github.io/debug-adapter-protocol//specification.html#Requests_Variables
func (ds *session) onVariablesRequest(request *dap.VariablesRequest) {
	resp := new(dap.VariablesResponse)
	resp.Response = *newResponse(request.Seq, request.Command)
	resp.Body.Variables = []dap.Variable{
		{Name: "x", Value: "1", Type: "int"},
		{Name: "y", Value: "2", Type: "int"},
	}
	ds.send(resp)
}

// https://microsoft.github.io/debug-adapter-protocol//specification.html#Requests_SetVariable
func (ds *session) onSetVariableRequest(request *dap.SetVariableRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "setVariableRequest is not yet supported"))
}

// https://microsoft.github.io/debug-adapter-protocol//specification.html#Requests_SetExpression
func (ds *session) onSetExpressionRequest(request *dap.SetExpressionRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "SetExpressionRequest is not yet supported"))
}

// https://microsoft.github.io/debug-adapter-protocol//specification.html#Requests_Source
func (ds *session) onSourceRequest(request *dap.SourceRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "SourceRequest is not yet supported"))
}

// https://microsoft.github.io/debug-adapter-protocol//specification.html#Requests_Threads
func (ds *session) onThreadsRequest(request *dap.ThreadsRequest) {
	resp := new(dap.ThreadsResponse)
	resp.Response = *newResponse(request.Seq, request.Command)
	// check launched
	if ds.vma != nil {
		resp.Body.Threads = []dap.Thread{
			{Id: 1, Name: "main"},
		}
	}
	ds.send(resp)
}

// https://microsoft.github.io/debug-adapter-protocol//specification.html#Requests_TerminateThreads
func (ds *session) onTerminateThreadsRequest(request *dap.TerminateThreadsRequest) {
	// brutal
	ds.vma = nil
	resp := new(dap.TerminateThreadsResponse)
	resp.Response = *newResponse(request.Seq, request.Command)
	resp.Success = true
	ds.send(resp)
}

func (ds *session) onEvaluateRequest(request *dap.EvaluateRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "EvaluateRequest is not yet supported"))
}

func (ds *session) onStepInTargetsRequest(request *dap.StepInTargetsRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "StepInTargetRequest is not yet supported"))
}

func (ds *session) onGotoTargetsRequest(request *dap.GotoTargetsRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "GotoTargetRequest is not yet supported"))
}

func (ds *session) onCompletionsRequest(request *dap.CompletionsRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "CompletionRequest is not yet supported"))
}

func (ds *session) onExceptionInfoRequest(request *dap.ExceptionInfoRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "ExceptionRequest is not yet supported"))
}

func (ds *session) onLoadedSourcesRequest(request *dap.LoadedSourcesRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "LoadedRequest is not yet supported"))
}

func (ds *session) onDataBreakpointInfoRequest(request *dap.DataBreakpointInfoRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "DataBreakpointInfoRequest is not yet supported"))
}

func (ds *session) onSetDataBreakpointsRequest(request *dap.SetDataBreakpointsRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "SetDataBreakpointsRequest is not yet supported"))
}

func (ds *session) onReadMemoryRequest(request *dap.ReadMemoryRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "ReadMemoryRequest is not yet supported"))
}

func (ds *session) onDisassembleRequest(request *dap.DisassembleRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "DisassembleRequest is not yet supported"))
}

func (ds *session) onCancelRequest(request *dap.CancelRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "CancelRequest is not yet supported"))
}

func (ds *session) onBreakpointLocationsRequest(request *dap.BreakpointLocationsRequest) {
	ds.send(newErrorResponse(request.Seq, request.Command, "BreakpointLocationsRequest is not yet supported"))
}

func (ds *session) dispatchRequest(request dap.Message) {
	switch request := request.(type) {
	case *dap.InitializeRequest:
		ds.onInitializeRequest(request)
	case *dap.LaunchRequest:
		ds.onLaunchRequest(request)
	case *dap.AttachRequest:
		ds.onAttachRequest(request)
	case *dap.DisconnectRequest:
		ds.onDisconnectRequest(request)
	case *dap.TerminateRequest:
		ds.onTerminateRequest(request)
	case *dap.RestartRequest:
		ds.onRestartRequest(request)
	case *dap.SetBreakpointsRequest:
		ds.onSetBreakpointsRequest(request)
	case *dap.SetFunctionBreakpointsRequest:
		ds.onSetFunctionBreakpointsRequest(request)
	case *dap.SetExceptionBreakpointsRequest:
		ds.onSetExceptionBreakpointsRequest(request)
	case *dap.ConfigurationDoneRequest:
		ds.onConfigurationDoneRequest(request)
	case *dap.ContinueRequest:
		ds.onContinueRequest(request)
	case *dap.NextRequest:
		ds.onNextRequest(request)
	case *dap.StepInRequest:
		ds.onStepInRequest(request)
	case *dap.StepOutRequest:
		ds.onStepOutRequest(request)
	case *dap.StepBackRequest:
		ds.onStepBackRequest(request)
	case *dap.ReverseContinueRequest:
		ds.onReverseContinueRequest(request)
	case *dap.RestartFrameRequest:
		ds.onRestartFrameRequest(request)
	case *dap.GotoRequest:
		ds.onGotoRequest(request)
	case *dap.PauseRequest:
		ds.onPauseRequest(request)
	case *dap.StackTraceRequest:
		ds.onStackTraceRequest(request)
	case *dap.ScopesRequest:
		ds.onScopesRequest(request)
	case *dap.VariablesRequest:
		ds.onVariablesRequest(request)
	case *dap.SetVariableRequest:
		ds.onSetVariableRequest(request)
	case *dap.SetExpressionRequest:
		ds.onSetExpressionRequest(request)
	case *dap.SourceRequest:
		ds.onSourceRequest(request)
	case *dap.ThreadsRequest:
		ds.onThreadsRequest(request)
	case *dap.TerminateThreadsRequest:
		ds.onTerminateThreadsRequest(request)
	case *dap.EvaluateRequest:
		ds.onEvaluateRequest(request)
	case *dap.StepInTargetsRequest:
		ds.onStepInTargetsRequest(request)
	case *dap.GotoTargetsRequest:
		ds.onGotoTargetsRequest(request)
	case *dap.CompletionsRequest:
		ds.onCompletionsRequest(request)
	case *dap.ExceptionInfoRequest:
		ds.onExceptionInfoRequest(request)
	case *dap.LoadedSourcesRequest:
		ds.onLoadedSourcesRequest(request)
	case *dap.DataBreakpointInfoRequest:
		ds.onDataBreakpointInfoRequest(request)
	case *dap.SetDataBreakpointsRequest:
		ds.onSetDataBreakpointsRequest(request)
	case *dap.ReadMemoryRequest:
		ds.onReadMemoryRequest(request)
	case *dap.DisassembleRequest:
		ds.onDisassembleRequest(request)
	case *dap.CancelRequest:
		ds.onCancelRequest(request)
	case *dap.BreakpointLocationsRequest:
		ds.onBreakpointLocationsRequest(request)
	default:
		log.Fatalf("Unable to process %#v", request)
	}
}
