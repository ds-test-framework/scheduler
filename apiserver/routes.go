package apiserver

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ds-test-framework/scheduler/log"
	"github.com/ds-test-framework/scheduler/types"
	"github.com/gin-gonic/gin"
)

func (srv *APIServer) HandleMessage(c *gin.Context) {
	var msg types.Message
	if err := c.ShouldBindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to unmarshal request"})
		return
	}
	srv.ctx.Publish(
		types.InterceptedMessage,
		&types.MessageWrapper{
			Run: srv.ctx.GetCurRun(),
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
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (srv *APIServer) HandleTimeout(c *gin.Context) {
	var timeout types.Timeout
	if err := c.ShouldBindJSON(&timeout); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to unmarshal request"})
		return
	}
	srv.ctx.Publish(
		types.InterceptedMessage,
		&types.MessageWrapper{
			Run: srv.ctx.GetCurRun(),
			Msg: types.NewMessage(
				timeout.Type,
				strconv.Itoa(srv.gen.Next()),
				timeout.Replica,
				timeout.Replica,
				timeout.Duration,
				true,
				nil,
				true,
			),
		},
	)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (srv *APIServer) HandleReplicaPost(c *gin.Context) {
	var replica types.Replica
	if err := c.ShouldBindJSON(&replica); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to unmarshal request"})
		return
	}

	srv.logger.With(log.LogParams{
		"replica_id": replica.ID,
		"ready":      replica.Ready,
		"info":       replica.Info,
	}).Info("Received replica information")

	srv.ctx.Publish(types.ReplicaUpdate, &replica)
	srv.ctx.Replicas.UpdateReplica(&replica)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type eventS struct {
	types.Event `json:",inline"`
	Params      map[string]string `json:"params"`
}

func (srv *APIServer) HandleEvent(c *gin.Context) {
	var e eventS
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to unmarshal request"})
		return
	}
	srv.logger.With(log.LogParams{
		"replica": e.Replica,
		"type":    e.TypeS,
		"params":  e.Params,
	}).Debug("Received event")
	srv.ctx.Publish(types.EventMessage, types.NewEvent(
		uint(srv.gen.Next()),
		e.Replica,
		types.NewReplicaEventType(e.TypeS, e.Params),
		time.Now().Unix(),
	))
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (srv *APIServer) HandleLog(c *gin.Context) {
	var l types.ReplicaLog
	if err := c.ShouldBindJSON(&l); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to unmarshal request"})
		return
	}
	srv.logger.With(log.LogParams{
		"params":  fmt.Sprintf("%#v", l.Params),
		"replica": string(l.Replica),
	}).Debug("Received log")

	srv.ctx.Publish(types.LogMessage, &l)
	srv.ctx.Logs.AddUpdate(&l)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (srv *APIServer) handleRun(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"runs": srv.ctx.GetAllRuns(),
	})
}

func (srv *APIServer) handleRunLog(c *gin.Context) {
	run := c.Param("run")
	if run == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing run param"})
		return
	}
	runI, err := strconv.Atoi(run)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid run param"})
		return
	}
	replica := c.Param("replica")
	if replica == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing replica param"})
		return
	}

	logs, err := srv.ctx.Logs.GetLogs(runI, types.ReplicaID(replica))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

func (srv *APIServer) handleReplicas(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"replicas": srv.ctx.Replicas.Iter(),
	})
}

func (srv *APIServer) handleReplicaGet(c *gin.Context) {
	replicaID, ok := c.Params.Get("replica")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing replica param"})
		return
	}
	replica, ok := srv.ctx.Replicas.GetReplica(types.ReplicaID(replicaID))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "replica id does not exist"})
		return
	}
	c.JSON(http.StatusOK, replica)
}

func (srv *APIServer) handleRunGraph(c *gin.Context) {
	runS, ok := c.Params.Get("run")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing run param"})
		return
	}
	runI, err := strconv.Atoi(runS)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid run param"})
		return
	}

	graph, ok := srv.ctx.EventGraph.GetGraph(runI)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "run does not exist"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"graph": graph})
}
