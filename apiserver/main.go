package apiserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/ds-test-framework/scheduler/log"
	transport "github.com/ds-test-framework/scheduler/transports/http"
	"github.com/ds-test-framework/scheduler/types"
	"github.com/ds-test-framework/scheduler/util"
)

type APIServer struct {
	transport *transport.HttpTransport
	ctx       *types.Context
	gen       *util.IDGenerator

	logger *log.Logger
}

func NewAPIServer(ctx *types.Context) *APIServer {
	t := transport.NewHttpTransport(ctx.Config("transport"))
	server := &APIServer{
		transport: t,
		ctx:       ctx,
		gen:       ctx.IDGen,
		logger: ctx.Logger.With(map[string]interface{}{
			"service": "APIServer",
		}),
	}

	t.AddHandler("/message", server.HandleMessage)
	t.AddHandler("/timeout", server.HandleTimeout)
	t.AddHandler("/replica", server.HandleReplica)
	t.AddHandler("/event", server.HandleEvent)
	t.AddHandler("/log", server.HandleLog)

	return server
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

type eventS struct {
	types.Event `json:",inline"`
	Params      map[string]string `json:"params"`
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
		msg.ID = strconv.Itoa(srv.gen.Next())
	}

	srv.ctx.Publish(
		types.InterceptedMessage,
		&types.MessageWrapper{
			Run: srv.ctx.GetRun(),
			Msg: types.NewMessage(
				msg.Type,
				msg.ID,
				msg.From,
				msg.To,
				0,
				false,
				msg.Msg,
				msg.Intercept,
			),
		},
	)
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

	srv.ctx.Publish(
		types.InterceptedMessage,
		&types.MessageWrapper{
			Run: srv.ctx.GetRun(),
			Msg: types.NewMessage(
				msg.Type,
				strconv.Itoa(srv.gen.Next()),
				msg.Replica,
				msg.Replica,
				msg.Duration,
				true,
				nil,
				true,
			),
		},
	)
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

	srv.logger.With(log.LogParams{
		"replica_id": replica.ID,
		"ready":      replica.Ready,
		"info":       replica.Info,
	}).Info("Received replica information")

	srv.ctx.Publish(types.ReplicaUpdate, &replica)
	srv.ctx.Replicas.UpdateReplica(&replica)
	srv.respond(w, &transport.AllOk)
}

func (srv *APIServer) HandleEvent(w http.ResponseWriter, r *http.Request) {
	body, ok := srv.readRequest(w, r)
	if !ok {
		return
	}
	var e eventS
	err := json.Unmarshal(body, &e)
	if err != nil {
		srv.respond(w, &transport.InternalError)
	}

	event := types.NewEvent(
		uint(srv.gen.Next()),
		e.Replica,
		types.NewReplicaEvent(e.TypeS, e.Params),
		e.Timestamp,
	)

	srv.logger.With(log.LogParams{
		"replica_id": event.Replica,
		"type":       event.TypeS,
	}).Debug("Received event")

	srv.ctx.EventGraph.AddEvent(event)
	srv.ctx.Publish(types.EventMessage, event)
	srv.respond(w, &transport.AllOk)
}

func (srv *APIServer) HandleLog(w http.ResponseWriter, r *http.Request) {
	body, ok := srv.readRequest(w, r)
	if !ok {
		return
	}

	var l types.ReplicaLog
	err := json.Unmarshal(body, &l)
	if err != nil {
		srv.respond(w, &transport.InternalError)
	}

	srv.logger.With(map[string]interface{}{
		"params":  fmt.Sprintf("%#v", l.Params),
		"replica": string(l.Replica),
	}).Debug("Received log")

	srv.ctx.Publish(types.LogMessage, &l)
	srv.ctx.Logs.AddUpdate(&l)
	srv.respond(w, &transport.AllOk)
}

func (srv *APIServer) Start() {
	srv.logger.With(map[string]interface{}{
		"addr": srv.transport.Addr(),
	}).Info("Starting API server")
	go srv.transport.Run()
}

func (srv *APIServer) Stop() {
	srv.logger.Debug("Stopping API server")
	srv.transport.Stop()
}
