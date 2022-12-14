package main

import (
	"context"
	"fmt"
	"time"

	"github.com/dapr/go-sdk/service/common"
	daprd "github.com/dapr/go-sdk/service/grpc"
	"github.com/dtm-labs/client/dtmcli/logger"
	"github.com/dtm-labs/client/dtmgrpc"
	daprdriver "github.com/dtm-labs/dtmdriver-dapr"
	"github.com/lithammer/shortuuid/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	daprdriver.Use()
	logger.InitLog("debug")
	s, err := daprd.NewService(":8084")
	logger.FatalIfError(err)
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
		b, err := dtmgrpc.BarrierFromGrpc(ctx)
		logger.Infof("barrier is: %v err: %v", b, err)
		result := string(in.Data)
		logger.Debugf("data is: %s", result)
		if result == "FAILURE" {
			return nil, status.Error(codes.Aborted, "user return failed")
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

var dtmServer = fmt.Sprintf("%s://DAPR_ENV/%s", daprdriver.SchemaProxiedGrpc, "dtm")

const appid = "app-grpc"

func finishRequest(result string) {
	req := []byte(result) // load of micro-service
	saga := dtmgrpc.NewSagaGrpc(dtmServer, shortuuid.New()).
		Add(daprdriver.AddrForGrpc(appid, "TransOut"), daprdriver.AddrForGrpc(appid, "TransOutRevert"), daprdriver.PayloadForGrpc(req)).
		Add(daprdriver.AddrForGrpc(appid, "TransIn"), daprdriver.AddrForGrpc(appid, "TransInRevert"), daprdriver.PayloadForGrpc(req))
	saga.WaitResult = true
	err := saga.Submit()
	logger.Infof("submit return: %v", err)
}
