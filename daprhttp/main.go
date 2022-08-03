package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/dapr/go-sdk/service/common"
	daprd "github.com/dapr/go-sdk/service/http"
	"github.com/dtm-labs/client/dtmcli/logger"
	"github.com/dtm-labs/dtm/client/dtmcli"
	daprdriver "github.com/dtm-labs/dtmdriver-dapr"
	"github.com/lithammer/shortuuid/v3"
)

func main() {
	daprdriver.Use()
	logger.InitLog("debug")
	s := daprd.NewService(":8082")
	addHandlers(s)

	go func() {
		err := s.Start()
		logger.FatalIfError(err)
	}()
	time.Sleep(2 * time.Second)
	finishRequest("success")
	finishRequest("FAILURE")
	select {}
}

func mustAddHandler(s common.Service, method string, fn func(ctx context.Context, in *common.InvocationEvent) (*common.Content, error)) {
	err := s.AddServiceInvocationHandler(method, fn)
	logger.FatalIfError(err)
}

func addHandlers(s common.Service) {
	mustAddHandler(s, "TransOut", func(ctx context.Context, in *common.InvocationEvent) (*common.Content, error) {
		logger.Infof("TransOut")
		return &common.Content{Data: []byte("")}, nil
	})
	mustAddHandler(s, "TransIn", func(ctx context.Context, in *common.InvocationEvent) (*common.Content, error) {
		logger.Infof("TransIn")
		values, err := url.ParseQuery(in.QueryString)
		logger.FatalIfError(err)
		b, err := dtmcli.BarrierFromQuery(values)
		logger.Infof("barrier is: %v err: %v", b, err)
		var result string
		err = json.Unmarshal(in.Data, &result)
		logger.FatalIfError(err)
		logger.Debugf("data is: %v", result)
		if result == "FAILURE" { // TODO there should be some way to return StatusCode 409
			return nil, fmt.Errorf("this error should be modified to return StatusCode 409, currently not supported")
		}
		return nil, nil
	})

	mustAddHandler(s, "TransOutRevert", func(ctx context.Context, in *common.InvocationEvent) (*common.Content, error) {
		logger.Infof("TransOutRevert")
		return nil, nil
	})
	mustAddHandler(s, "TransInRevert", func(ctx context.Context, in *common.InvocationEvent) (*common.Content, error) {
		logger.Infof("TransInRevert")
		return nil, nil
	})
}

var dtmServer = fmt.Sprintf("%s://DAPR_ENV/%s/api/dtmsvr", daprdriver.SchemaProxiedHTTP, "dtm")

const appid = "app-http"

func finishRequest(result string) {
	saga := dtmcli.NewSaga(dtmServer, shortuuid.New()).
		Add(daprdriver.AddrForHTTP(appid, "TransOut"), daprdriver.AddrForHTTP(appid, "TransOutRevert"), result).
		Add(daprdriver.AddrForHTTP(appid, "TransIn"), daprdriver.AddrForHTTP(appid, "TransInRevert"), result)
	saga.WaitResult = true
	err := saga.Submit()
	logger.Infof("submit return: %v", err)
}
