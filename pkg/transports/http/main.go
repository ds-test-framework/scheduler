package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ds-test-framework/model-checker/pkg/logger"
	"github.com/ds-test-framework/model-checker/pkg/types"
	"github.com/spf13/viper"
)

const (
	ErrSendFailed       = "SEND_FAILED"
	ErrBadResponse      = "BAD_RESPONSE"
	ErrResponseReadFail = "RESPONSE_READ_FAILED"
)

var (
	internalError    = Response{Err: "Internal server error"}
	methodNotAllowed = Response{Err: "Method not allowed"}

	allOk = Response{Status: "ok"}
)

type HttpTransport struct {
	listenAddr string
	outChan    chan string
	server     *http.Server
}

type Response struct {
	Status string `json:"status"`
	Err    string `json:"error"`
}

func NewHttpTransport(options *viper.Viper) *HttpTransport {
	t := &HttpTransport{
		listenAddr: options.GetString("addr"),
		outChan:    make(chan string, 10),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", t.Handler)

	t.server = &http.Server{
		Addr:    t.listenAddr,
		Handler: mux,
	}
	return t

}

func (t *HttpTransport) respond(w http.ResponseWriter, r *Response) {
	w.Header().Add("Content-Type", "application/json")
	respB, err := json.Marshal(r)
	if err != nil {
		respB, _ = json.Marshal(internalError)
	}
	w.Write(respB)
}

func (t *HttpTransport) Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		t.respond(w, &methodNotAllowed)
		return
	}
	bodyBytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err == nil {
		// logger.Debug(fmt.Sprintf("Transport: Received message, %s", string(bodyBytes)))
		t.outChan <- string(bodyBytes)
	}
	t.respond(w, &allOk)
}

func (t *HttpTransport) ReceiveChan() chan string {
	return t.outChan
}

func (t *HttpTransport) Run() {
	logger.Debug(fmt.Sprintf("Starting server at: %s", t.listenAddr))
	if err := t.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Could not start server")
	}
}

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

type RequestOption func(*http.Request)

func JsonRequest() RequestOption {
	return func(r *http.Request) {
		r.Header.Set("Content-Type", "application/json")
	}
}

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
	if resp.StatusCode == http.StatusOK {
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
