package testlib

import (
	"github.com/ds-test-framework/scheduler/apiserver"
	"github.com/ds-test-framework/scheduler/config"
	"github.com/ds-test-framework/scheduler/context"
	"github.com/ds-test-framework/scheduler/dispatcher"
	"github.com/ds-test-framework/scheduler/log"
	"github.com/ds-test-framework/scheduler/types"
)

type TestingServer struct {
	apiserver  *apiserver.APIServer
	dispatcher *dispatcher.Dispatcher
	ctx        *context.RootContext
	messagesCh chan *types.Message
	*types.BaseService
}

func NewTestingServer(config *config.Config, testcases []interface{}) (*TestingServer, error) {
	log.Init(config.LogConfig)
	ctx := context.NewRootContext(config, log.DefaultLogger)
	ch, err := ctx.MessageQueue.Subscribe("testingserver")
	if err != nil {
		return nil, err
	}

	return &TestingServer{
		apiserver:   apiserver.NewAPIServer(ctx),
		dispatcher:  dispatcher.NewDispatcher(ctx),
		ctx:         ctx,
		messagesCh:  ch,
		BaseService: types.NewBaseService("TestingServer", log.DefaultLogger),
	}, nil
}

func (srv *TestingServer) Run() {
	srv.StartRunning()
	srv.apiserver.Start()
	srv.ctx.Start()
	srv.poll()
}

// dummy, for now just dispatch the messages that come through
func (srv *TestingServer) poll() {
	for {
		select {
		case m := <-srv.messagesCh:
			err := srv.dispatcher.DispatchMessage(m)
			if err != nil {
				srv.Logger.With(log.LogParams{
					"message_id": m.ID,
					"err":        err.Error(),
				}).Error("error dispatching message")
			}
		case <-srv.QuitCh():
			return
		}
	}
}

func (srv *TestingServer) Stop() {
	srv.StopRunning()
	srv.apiserver.Stop()
	srv.ctx.Stop()
}
