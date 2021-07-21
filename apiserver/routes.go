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

func (srv *APIServer) HandleReplica(c *gin.Context) {
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

}
