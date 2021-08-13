package testing

import (
	"errors"
	"sync"
	"time"

	"github.com/ds-test-framework/scheduler/apiserver"
	"github.com/ds-test-framework/scheduler/log"
	"github.com/ds-test-framework/scheduler/types"
	"github.com/spf13/viper"
)

type ServerConfig struct {
	Addr      string
	LogPath   string
	LogLevel  string
	LogFormat string
	Replicas  int
}

func (c *ServerConfig) setDefaults() {
	if c.Addr == "" {
		c.Addr = defaultServerAddr
	}
	if c.LogPath == "" {
		c.LogPath = defaultLogPath
	}
	if c.LogLevel == "" {
		c.LogLevel = defaultLogLevel
	}
	if c.LogFormat == "" {
		c.LogFormat = defaultLogFormat
	}
	if c.Replicas == 0 {
		c.Replicas = defaultReplicaCount
	}
}

func (c *ServerConfig) viperConfig() *viper.Viper {
	config := viper.New()

	config.Set("transport.addr", c.Addr)
	config.Set("num_replicas", c.Replicas)
	config.Set("log.format", c.LogFormat)
	if c.LogPath != "" {
		config.Set("log.path", c.LogPath)
	}

	return config
}

const (
	defaultServerAddr   = "0.0.0.0:87074"
	defaultLogPath      = ""
	defaultLogLevel     = "info"
	defaultReplicaCount = 4
	defaultLogFormat    = "plain"
)

type Server struct {
	TestCases []*TestCase
	driver    *testDriver
	apiServer *apiserver.APIServer
	logger    *log.Logger
	ctx       *types.Context

	resultMap *resultMap

	stopCh chan bool
	stopO  *sync.Once
}

func NewTestServer(config ServerConfig, testcases []*TestCase) (*Server, error) {
	config.setDefaults()
	ctxConfig := config.viperConfig()

	log.Init(ctxConfig.Sub("log"), config.LogLevel)

	if len(testcases) == 0 {
		log.DefaultLogger.With(map[string]interface{}{"service": "test-server"}).Info("No testcases to run, returning nil")
		return nil, errors.New("no test cases given")
	}

	context := types.NewContext(ctxConfig, log.DefaultLogger)

	server := &Server{
		TestCases: testcases,
		driver:    newTestDriver(context),
		apiServer: apiserver.NewAPIServer(context),
		logger:    log.DefaultLogger.With(map[string]interface{}{"service": "test-server"}),
		ctx:       context,

		resultMap: newResultMap(),

		stopCh: make(chan bool),
		stopO:  new(sync.Once),
	}

	go server.driver.Run()
	go server.apiServer.Start()

	return server, nil
}

func (s *Server) Run() {
	s.logger.Info("Starting server main loop")
	for i, testCase := range s.TestCases {

		logger := s.logger.With(map[string]interface{}{"testcase": testCase.Name})

		s.ctx.SetRun(&types.Run{
			Id:    i,
			Label: testCase.Name,
		})
		logger.Info("Starting test case")
		s.driver.StartRun(testCase)
		ok := s.driver.Ready()
		if !ok {
			logger.Info("Replicas failed to intialize, test case did not start")
			continue
		}
		testCase = testCase.WithContext(s.ctx)
		ctx := testCase.Run()

		select {
		case <-ctx.Done:
			logger.Debug("Testcase indicated done")
		case <-time.After(ctx.Timeout):
			logger.Debug("Testcase timeout reached")
		case <-s.stopCh:
			s.logger.Debug("Stopping server main loop prematurely")
			return
		}

		ok = testCase.Assert()
		s.resultMap.add(testCase.Name, ok)
		if !ok {
			logger.Info("Testcase failed")
			continue
		}
		logger.Info("Testcase succeded")

		select {
		case <-s.stopCh:
			s.logger.Debug("Stopping server main loop prematurely")
			return
		default:
		}
	}
	s.logger.Info("Completed all test cases")
	s.logger.Info("Summary: \n" + s.resultMap.summary())
	s.Stop()
}

func (s *Server) Stop() {
	s.logger.Info("Received Stop signal")
	s.stopO.Do(func() {
		close(s.stopCh)
		s.driver.Stop()
		s.apiServer.Stop()
		log.Destroy()
	})
}
