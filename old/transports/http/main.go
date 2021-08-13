package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	logger "github.com/ds-test-framework/scheduler/log"
	"github.com/ds-test-framework/scheduler/types"
	"github.com/spf13/viper"
)

const (
	ErrSendFailed       = "SEND_FAILED"
	ErrBadResponse      = "BAD_RESPONSE"
	ErrResponseReadFail = "RESPONSE_READ_FAILED"
)

var (
	InternalError    = Response{Err: "Internal server error"}
	MethodNotAllowed = Response{Err: "Method not allowed"}

	AllOk = Response{Status: "ok"}
)

// HttpTransport wrapps around net/http to start a server and to send http messages
// The transport sends the message in a channel which can be obtained by calling ReceiveChan
type HttpTransport struct {
	listenAddr string
	outChan    chan string
	server     *http.Server
	mux        *http.ServeMux
}

// Response message that is used to respond to incoming http messages
type Response struct {
	Status string `json:"status"`
	Err    string `json:"error"`
}

// NewHttpTransport returns an HttpTransport with the given options.
// Options:
// - addr: the address to listen for http connections for
func NewHttpTransport(options *viper.Viper) *HttpTransport {

	mux := http.NewServeMux()
	options.SetDefault("addr", "0.0.0.0:7074")
	t := &HttpTransport{
		listenAddr: options.GetString("addr"),
		outChan:    make(chan string, 10),
		mux:        mux,
	}

	t.server = &http.Server{
		Addr:    t.listenAddr,
		Handler: mux,
	}
	return t

}
func (t *HttpTransport) AddHandler(route string, handler func(http.ResponseWriter, *http.Request)) {
	t.mux.HandleFunc(route, handler)
}

func (t *HttpTransport) respond(w http.ResponseWriter, r *Response) {
	w.Header().Add("Content-Type", "application/json")
	respB, err := json.Marshal(r)
	if err != nil {
		respB, _ = json.Marshal(InternalError)
	}
	w.Write(respB)
}

// Handler function that receives all requests
func (t *HttpTransport) Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		t.respond(w, &MethodNotAllowed)
		return
	}
	bodyBytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err == nil {
		// logger.Debug(fmt.Sprintf("Transport: Received message, %s", string(bodyBytes)))
		go func(msg []byte) { t.outChan <- string(msg) }(bodyBytes)
	}
	t.respond(w, &AllOk)
}

// ReceiveChan returns the channel on which incoming messages are published
func (t *HttpTransport) ReceiveChan() chan string {
	return t.outChan
}

func (t *HttpTransport) Addr() string {
	return t.listenAddr
}

// Run start the http server
func (t *HttpTransport) Run() {
	logger.Debug(fmt.Sprintf("Starting server at: %s", t.listenAddr))
	if err := t.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal(fmt.Sprintf("Could not start server: %s", err.Error()))
	}
}

// Stop stops the http server
func (t *HttpTransport) Stop() {
	sCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer func() {
		cancel()
	}()

	if err := t.server.Shutdown(sCtx); err != nil {
		logger.Debug(
			fmt.Sprintf("Could not shutdown server: %s", err.Error()),
		)
	}
}

// RequestOption can be used to modify the request that is to be sent
type RequestOption func(*http.Request)

// JsonRequest sets the content type to application/json
func JsonRequest() RequestOption {
	return func(r *http.Request) {
		r.Header.Set("Content-Type", "application/json")
	}
}

// SendMsg sends a message to the specified address using the specified method
// Returns the response if successful or returns error if the response was not 2xx
func SendMsg(method, toAddr, msg string, options ...RequestOption) (string, *types.Error) {
	t := &HttpTransport{}
	return t.SendMsg(method, toAddr, msg, options...)
}

// SendMsg sends a message to the specified address using the specified method
// Returns the response if successful or returns error if the response was not 2xx
func (t *HttpTransport) SendMsg(method, toAddr, msg string, options ...RequestOption) (string, *types.Error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, "http://"+toAddr, bytes.NewBuffer([]byte(msg)))
	if err != nil {
		return "", types.NewError(
			ErrSendFailed,
			fmt.Sprintf("Could not create request object: %s", err.Error()),
		)
	}

	for _, o := range options {
		o(req)
	}

	// logger.Debug(fmt.Sprintf("Transport: Sending message: %s", msg))
	resp, err := client.Do(req)
	if err != nil {
		return "", types.NewError(
			ErrSendFailed,
			"Failed to send message",
		)
	}
	defer resp.Body.Close()
	statusOK := resp.StatusCode >= 200 && resp.StatusCode < 300
	if statusOK {
		bodyB, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", types.NewError(
				ErrResponseReadFail,
				"Failed to read response",
			)
		}
		return string(bodyB), nil
	}
	return "", types.NewError(
		ErrBadResponse,
		"Server did not return success response",
	)
}
