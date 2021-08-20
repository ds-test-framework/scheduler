package apiserver

import (
	"net/http"

	"github.com/ds-test-framework/scheduler/log"
	"github.com/ds-test-framework/scheduler/types"
	"github.com/gin-gonic/gin"
)

var (
	messageSendEventType    = "MessageSend"
	messageReceiveEventType = "MessageReceive"
)

func (srv *APIServer) HandleMessage(c *gin.Context) {
	srv.Logger.Info("Handling message")
	var msg types.Message
	if err := c.ShouldBindJSON(&msg); err != nil {
		srv.Logger.With(log.LogParams{"error": err}).Info("Bad message")
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to unmarshal request"})
		return
	}
	srv.ctx.MessageStore.Add(&msg)
	// TODO: logic for dispatching the message if it should not be intercepted
	srv.ctx.MessageQueue.Add(&msg)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (srv *APIServer) HandleReplicaPost(c *gin.Context) {
	var replica types.Replica
	if err := c.ShouldBindJSON(&replica); err != nil {
		srv.Logger.With(log.LogParams{"error": err}).Info("Bad replica request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to unmarshal request"})
		return
	}

	srv.Logger.With(log.LogParams{
		"replica_id": replica.ID,
		"ready":      replica.Ready,
		"info":       replica.Info,
	}).Info("Received replica information")

	srv.ctx.Replicas.Add(&replica)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type eventS struct {
	types.Event `json:",inline"`
	Params      map[string]string `json:"params"`
}

func (srv *APIServer) HandleEvent(c *gin.Context) {
	var e eventS
	if err := c.ShouldBindJSON(&e); err != nil {
		srv.Logger.With(log.LogParams{"error": err}).Info("Bad event request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to unmarshal request"})
		return
	}
	srv.Logger.With(log.LogParams{
		"replica": e.Replica,
		"type":    e.TypeS,
		"params":  e.Params,
	}).Debug("Received event")

	var eventType types.EventType
	switch e.TypeS {
	case messageSendEventType:
		messageID := e.Params["message_id"]
		eventType = types.NewMessageReceiveEventType(messageID)
	case messageReceiveEventType:
		messageID := e.Params["message_id"]
		eventType = types.NewMessageReceiveEventType(messageID)
	default:
		eventType = types.NewGenericEventType(e.Params, e.TypeS)
	}

	srv.ctx.EventQueue.Add(&types.Event{
		Replica:   e.Replica,
		Type:      eventType,
		TypeS:     eventType.String(),
		ID:        uint64(srv.gen.Next()),
		Timestamp: e.Timestamp,
	})
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
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
	replica, ok := srv.ctx.Replicas.Get(types.ReplicaID(replicaID))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "replica id does not exist"})
		return
	}
	c.JSON(http.StatusOK, replica)
}

func (srv *APIServer) HandleLog(c *gin.Context) {
	var l types.ReplicaLog
	if err := c.ShouldBindJSON(&l); err != nil {
		srv.Logger.With(log.LogParams{"error": err}).Debug("Bad replica log request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to unmarshall request"})
		return
	}

	srv.Logger.With(log.LogParams{
		"replica": l.Replica,
		"message": l.Message,
	}).Debug("Received replica log")

	srv.ctx.LogStore.Add(&l)
	srv.ctx.LogQueue.Add(&l)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
