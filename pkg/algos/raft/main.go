package raft

import (
	"fmt"
	"sync"
	"time"

	logger "github.com/ds-test-framework/scheduler/pkg/log"
	transport "github.com/ds-test-framework/scheduler/pkg/transports/http"
	"github.com/ds-test-framework/scheduler/pkg/types"
	"github.com/spf13/viper"
)

const (
	ErrFailedMasterConnect     = "MASTER_CONNECT_FAILED"
	ErrWorkloadInjectionFailed = "WORKLOAD_INJECTION_FAILED"
)

type RaftDriver struct {
	masterAddr        string
	run               int
	lock              *sync.Mutex
	transport         *transport.HttpTransport
	transportRecvChan chan string
	engineOut         chan *types.MessageWrapper
	engineIn          chan *types.MessageWrapper
	msgStore          map[string]string
	stopChan          chan bool
	runObj            *types.RunObj
	options           *viper.Viper
	intercepting      bool
}

func NewRaftDriver(options *viper.Viper) *RaftDriver {
	options.SetDefault("retries_limit", 5)
	options.SetDefault("retries_gap", 1)
	transport := transport.NewHttpTransport(options.Sub("transport"))
	d := &RaftDriver{
		masterAddr:   options.GetString("master_addr"),
		run:          0,
		lock:         new(sync.Mutex),
		transport:    transport,
		engineOut:    make(chan *types.MessageWrapper, 10),
		engineIn:     make(chan *types.MessageWrapper, 10),
		msgStore:     make(map[string]string),
		stopChan:     make(chan bool, 1),
		options:      options,
		intercepting: true,
	}
	d.transportRecvChan = transport.ReceiveChan()
	return d
}

func (d *RaftDriver) Init() {
	go d.transport.Run()
	i := 0
	retriesLimit := d.options.GetInt("retries_limit")
	retryGap := d.options.GetInt("retries_gap")
	for {
		err := d.connectToMaster()
		if err == nil {
			break
		}
		if i >= retriesLimit {
			d.transport.Stop()
			logger.Fatal(
				fmt.Sprintf("Could not connect to master within the retry limit: %s", err.Error()),
			)
		}
		i = i + 1
		time.Sleep(time.Duration(retryGap) * time.Second)
	}
	logger.Debug("Connected to master")
	go d.poll()
}

func (d *RaftDriver) Ready() (bool, *types.Error) {
	return true, nil
}

func (d *RaftDriver) InChan() chan *types.MessageWrapper {
	return d.engineIn
}

func (d *RaftDriver) OutChan() chan *types.MessageWrapper {
	return d.engineOut
}

func (d *RaftDriver) connectToMaster() *types.Error {
	return d.sendMasterMsg(NewControlMessage(Ready))
}

func (d *RaftDriver) handleControlMsg(msg *ControlMessage) {
	switch msg.Type {
	case "RunCompleted":
		d.lock.Lock()
		runObj := d.runObj
		d.lock.Unlock()
		runObj.Ch <- true
	default:
	}
}

func (d *RaftDriver) poll() {
	for {
		select {
		case m := <-d.transportRecvChan:
			if IsControlMessage(m) {
				cMsg, err := UnmarshalControlMessage(m)
				if err == nil {
					d.handleControlMsg(cMsg)
				}
			} else {
				msg, err := UnmarshalMessage(m)
				if err == nil {
					d.lock.Lock()
					run := d.run
					intercepting := d.intercepting
					d.msgStore[msg.ID] = m
					d.lock.Unlock()
					// logger.Debug(fmt.Sprintf("Driver: Received message from transport: %v, %s", intercepting, m))
					if intercepting {
						d.engineOut <- &types.MessageWrapper{
							Run: run,
							Msg: msg,
						}
					}
				}
			}
		case msgW := <-d.engineIn:
			// logger.Debug(fmt.Sprintf("Received message from Engine: %#v", msgW.Msg))
			d.lock.Lock()
			m, ok := d.msgStore[msgW.Msg.ID]
			d.lock.Unlock()
			if ok {
				go func() {
					d.transport.SendMsg("POST", d.masterAddr+"/controller", m, transport.JsonRequest())
				}()
			}
		case _ = <-d.stopChan:
			return
		}
	}
}

func (d *RaftDriver) masterReset() *types.Error {
	return d.sendMasterMsg(NewControlMessage(Reset))
}

func (d *RaftDriver) sendMasterMsg(m *ControlMessage) *types.Error {
	msgB, err := m.Marshall()
	if err != nil {
		return err
	}
	_, err = d.transport.SendMsg("POST", d.masterAddr+"/control", string(msgB), transport.JsonRequest())
	return err
}

func (d *RaftDriver) masterInjectWorkload() *types.Error {
	i := 0
	for {
		err := d.sendMasterMsg(NewControlMessage(ReadyToInject))
		if err == nil {
			break
		}
		if i > 10 {
			return types.NewError(
				ErrWorkloadInjectionFailed,
				"Could not inject worlkload as system is not ready",
			)
		}
		i = i + 1
		time.Sleep(500 * time.Millisecond)
	}
	return d.sendMasterMsg(NewControlMessage(InjectWorkload))
}

func (d *RaftDriver) StartRun(no int) (*types.RunObj, *types.Error) {
	d.masterReset()
	go d.masterInjectWorkload()
	d.lock.Lock()
	defer d.lock.Unlock()

	runObj := &types.RunObj{
		Ch: make(chan bool, 2),
	}

	d.run = no
	d.runObj = runObj
	return runObj, nil
}

func (d *RaftDriver) StopRun() {
	d.lock.Lock()
	d.intercepting = false
	d.lock.Unlock()

	d.masterReset()
	time.Sleep(2 * time.Second)

	d.lock.Lock()
	d.intercepting = true
	d.lock.Unlock()
}

func (d *RaftDriver) Destroy() {
	logger.Debug("Stopping driver")
	d.stopChan <- true
	d.transport.Stop()
}
