package main

import (
	"time"

	"github.com/dtm-labs/client/dtmcli/logger"
	"github.com/dtm-labs/client/dtmgrpc"
	daprdriver "github.com/dtm-labs/dtmdriver-dapr"
	"github.com/lithammer/shortuuid/v3"
	"github.com/yedf2/dtm-dapr-sample/daprpgrpc/busi"
)

func main() {
	logger.InitLog("debug")
	daprdriver.Use()

	s := busi.GrpcNewServer()
	busi.GrpcStartup(s)
	logger.Infof("grpc simple transaction begin")
	time.Sleep(2 * time.Second)
	finishRequest(false)
	finishRequest(true)
	select {}
}

func finishRequest(failed bool) {
	req := &busi.BusiReq{Amount: 30}
	if failed {
		req.TransInResult = "FAILURE"
	}
	// req := &busi.BusiReq{Amount: 30, TransInResult: "FAILURE"}
	saga := dtmgrpc.NewSagaGrpc(busi.DtmGrpcServer, shortuuid.New()).
		Add(busi.BusiGrpc+"/busi.Busi/TransOut", busi.BusiGrpc+"/busi.Busi/TransOutRevert", req).
		Add(busi.BusiGrpc+"/busi.Busi/TransIn", busi.BusiGrpc+"/busi.Busi/TransInRevert", req)
	saga.WaitResult = true
	err := saga.Submit()
	logger.Infof("result is: %v", err)
}
