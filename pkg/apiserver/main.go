package apiserver

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"

	"github.com/ds-test-framework/scheduler/pkg/log"
	transport "github.com/ds-test-framework/scheduler/pkg/transports/http"
	"github.com/ds-test-framework/scheduler/pkg/types"
)

type APIServer struct {
	transport *transport.HttpTransport
	ctx       *types.Context

	counter     int
	counterLock *sync.Mutex

	logger *log.Logger
}

type state struct {
	State string `json:"state"`
}

type logM struct {
	Params map[string]interface{} `json:"params"`
}

func NewAPIServer(ctx *types.Context) *APIServer {
	t := transport.NewHttpTransport(ctx.Config("transport"))
	server := &APIServer{
		transport: t,
		ctx:       ctx,

		counter:     0,
		counterLock: new(sync.Mutex),
		logger: ctx.Logger.With(map[string]string{
			"service": "APIServer",
		}),
	}

	t.AddHandler("/message", server.HandleMessage)
	t.AddHandler("/timeout", server.HandleTimeout)
	t.AddHandler("/replica", server.HandleReplica)
	t.AddHandler("/state", server.HandleStateUpdate)
	t.AddHandler("/log", server.HandleLog)

	return server
}

func (srv *APIServer) nextID() int {
	srv.counterLock.Lock()
	srv.counter = srv.counter + 1
	counter := srv.counter
	srv.counterLock.Unlock()
	return counter
}

func (srv *APIServer) respond(w http.ResponseWriter, r *transport.Response) {
	w.Header().Add("Content-Type", "application/json")
	respB, err := json.Marshal(r)
	if err != nil {
		respB, _ = json.Marshal(transport.InternalError)
	}
	w.Write(respB)
}

func (srv *APIServer) readRequest(w http.ResponseWriter, r *http.Request) ([]byte, bool) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		srv.respond(w, &transport.MethodNotAllowed)
		return nil, false
	}
	bodyBytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		srv.respond(w, &transport.InternalError)
		return nil, false
	}
	return bodyBytes, true
}

func (srv *APIServer) HandleMessage(w http.ResponseWriter, r *http.Request) {
	body, ok := srv.readRequest(w, r)
	if !ok {
		return
	}

	var msg types.Message
	err := json.Unmarshal(body, &msg)
	if err != nil {
		srv.respond(w, &transport.InternalError)
	}
	if msg.ID == "" {
		msg.ID = strconv.Itoa(srv.nextID())
	}

	srv.ctx.NewMessage(types.NewMessage(
		msg.Type,
		msg.ID,
		msg.From,
		msg.To,
		0,
		false,
		[]byte(msg.Msg),
		msg.Intercept,
	))
	srv.respond(w, &transport.AllOk)
}

func (srv *APIServer) HandleTimeout(w http.ResponseWriter, r *http.Request) {
	body, ok := srv.readRequest(w, r)
	if !ok {
		return
	}

	var msg types.Timeout
	err := json.Unmarshal(body, &msg)
	if err != nil {
		srv.respond(w, &transport.InternalError)
	}

	srv.ctx.NewMessage(types.NewMessage(
		msg.Type,
		strconv.Itoa(srv.nextID()),
		msg.Replica,
		msg.Replica,
		msg.Duration,
		true,
		nil,
		true,
	))
	srv.respond(w, &transport.AllOk)
}

func (srv *APIServer) HandleReplica(w http.ResponseWriter, r *http.Request) {
	body, ok := srv.readRequest(w, r)
	if !ok {
		return
	}

	var replica types.Replica
	err := json.Unmarshal(body, &replica)
	if err != nil {
		srv.respond(w, &transport.InternalError)
	}

	srv.ctx.NewReplica(&replica)
	srv.respond(w, &transport.AllOk)
}

func (srv *APIServer) HandleStateUpdate(w http.ResponseWriter, r *http.Request) {
	body, ok := srv.readRequest(w, r)
	if !ok {
		return
	}

	var s state
	err := json.Unmarshal(body, &s)
	if err != nil {
		srv.respond(w, &transport.InternalError)
	}

	srv.ctx.StateUpdates.AddUpdate(s.State)
	srv.respond(w, &transport.AllOk)
}

func (srv *APIServer) HandleLog(w http.ResponseWriter, r *http.Request) {
	body, ok := srv.readRequest(w, r)
	if !ok {
		return
	}

	var l logM
	err := json.Unmarshal(body, &l)
	if err != nil {
		srv.respond(w, &transport.InternalError)
	}

	srv.ctx.Logs.AddUpdate(l.Params)
	srv.respond(w, &transport.AllOk)
}

func (srv *APIServer) Start() {
	srv.logger.With(map[string]string{
		"addr": srv.transport.Addr(),
	}).Info("Starting API server")
	go srv.transport.Run()
}

func (srv *APIServer) Stop() {
	srv.transport.Stop()
}
