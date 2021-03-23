package main

import (
	"github.com/spf13/viper"
	"github.com/zeu5/model-checker/pkg/transports/http"
)

func main() {
	// engine := timeout.NewTimeoutEngine(viper.GetViper())

	// inChan := make(chan *types.MessageWrapper, 10)
	// outChan := make(chan *types.MessageWrapper, 10)

	// engine.SetChannels(inChan, outChan)

	// inChan <- &types.MessageWrapper{
	// 	Run: 0,
	// 	Msg: types.NewMessage("SomeType", "1", 0, 0, 1000, true),
	// }
	// engine.Run()

	viper.SetDefault("addr", "127.0.0.1:7074")
	transport := http.NewHttpTransport(viper.GetViper())

	transport.Run()
}
