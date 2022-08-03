package main

import (
	"context"
	"fmt"
	"log"
	"time"

	dapr "github.com/dapr/go-sdk/client"
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
	if err != nil {
		log.Fatalf("failed to start the server: %v", err)
	}

	MustAddHandler(s, "TransOut", func(ctx context.Context, in *common.InvocationEvent) (*common.Content, error) {
		logger.Infof("TransOut")
		return nil, nil
	})
	MustAddHandler(s, "TransIn", func(ctx context.Context, in *common.InvocationEvent) (*common.Content, error) {
		logger.Infof("TransIn")
		return nil, nil
	})
	MustAddHandler(s, "TransInFailed", func(ctx context.Context, in *common.InvocationEvent) (*common.Content, error) {
		logger.Infof("TransIn")
		return nil, status.Error(codes.Aborted, "busi failed")
	})
	MustAddHandler(s, "TransOutRevert", func(ctx context.Context, in *common.InvocationEvent) (*common.Content, error) {
		logger.Infof("TransOutRevert")
		return nil, nil
	})
	MustAddHandler(s, "TransInRevert", func(ctx context.Context, in *common.InvocationEvent) (*common.Content, error) {
		logger.Infof("TransInRevert")
		return nil, nil
	})
	go func() {
		err := s.Start()
		logger.FatalIfError(err)
	}()
	finishRequest()
}

func MustAddHandler(s common.Service, method string, fn func(ctx context.Context, in *common.InvocationEvent) (*common.Content, error)) {
	err := s.AddServiceInvocationHandler(method, fn)
	logger.FatalIfError(err)
}

// var dtmServer = fmt.Sprintf("localhost:40004")

// var dtmServer = fmt.Sprintf("%s://localhost:30004/%s", daprdriver.SchemaProxiedGrpc, "dtm")

var dtmServer = fmt.Sprintf("%s://DAPR_ENV/%s", daprdriver.SchemaProxiedGrpc, "dtm")

const appid = "app-grpc"

func finishRequest() {
	logger.Debugf("sleeping to wait local service ready")
	time.Sleep(3 * time.Second)
	var err error
	client, err := dapr.NewClient()
	logger.FatalIfError(err)
	content := &dapr.DataContent{
		ContentType: "text/plain",
		Data:        []byte("hellow"),
	}
	logger.Debugf("calling to TransIn")
	_, err = client.InvokeMethodWithContent(context.Background(), appid, "TransIn", "post", content)
	logger.FatalIfError(err)

	req := []byte("amount: 30") // load of micro-service
	// DtmServer is the url of dtm
	saga := dtmgrpc.NewSagaGrpc(dtmServer, shortuuid.New()).
		Add(daprdriver.AddrForGrpc(appid, "TransOut"), daprdriver.AddrForGrpc(appid, "TransOutRevert"), daprdriver.PayloadForGrpc(req)).
		Add(daprdriver.AddrForGrpc(appid, "TransIn"), daprdriver.AddrForGrpc(appid, "TransOutRevert"), daprdriver.PayloadForGrpc(req))
	// submit the created saga transaction，dtm ensures all sub-transactions either complete or get revoked
	saga.WaitResult = true
	// saga.Context = metadata.AppendToOutgoingContext(saga.Context, "dapr-app-id", "dtm")
	err = saga.Submit()
	logger.Debugf("submit return: %v", err)

	saga2 := dtmgrpc.NewSagaGrpc(dtmServer, shortuuid.New()).
		Add(daprdriver.AddrForGrpc(appid, "TransOut"), daprdriver.AddrForGrpc(appid, "TransOutRevert"), daprdriver.PayloadForGrpc(req)).
		Add(daprdriver.AddrForGrpc(appid, "TransInFailed"), daprdriver.AddrForGrpc(appid, "TransOutRevert"), daprdriver.PayloadForGrpc(req))
	// submit the created saga transaction，dtm ensures all sub-transactions either complete or get revoked
	saga2.WaitResult = true
	// saga.Context = metadata.AppendToOutgoingContext(saga.Context, "dapr-app-id", "dtm")
	err = saga2.Submit()
	logger.Debugf("submit return: %v", err)

	select {}
}
