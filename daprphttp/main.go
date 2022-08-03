package main

import (
	"fmt"
	"time"

	"github.com/dtm-labs/client/dtmcli"
	daprdriver "github.com/dtm-labs/dtmdriver-dapr"
	"github.com/dtm-labs/logger"

	"github.com/gin-gonic/gin"
	"github.com/lithammer/shortuuid/v3"
)

// busi address
const qsBusiAPI = "/api/busi_start"
const qsBusiPort = 8081

var qsBusi = daprdriver.AddrForProxiedHTTP("app-phttp", "/api/busi_start")

func main() {
	daprdriver.Use()
	startSvr()
	finishRequest("")
	finishRequest("FAILURE")
}

// QsStartSvr quick start: start server
func startSvr() {
	app := gin.New()
	qsAddRoute(app)
	logger.Infof("quick start examples listening at %d", qsBusiPort)
	go func() {
		err := app.Run(fmt.Sprintf(":%d", qsBusiPort))
		logger.FatalIfError(err)
	}()
	time.Sleep(500 * time.Millisecond)
}

func qsAddRoute(app *gin.Engine) {
	app.POST(qsBusiAPI+"/TransIn", func(c *gin.Context) {
		logger.Infof("TransIn")
		b, err := dtmcli.BarrierFromQuery(c.Request.URL.Query())
		logger.Infof("barrier info: %v err: %v", b, err)
		var req gin.H
		err = c.BindJSON(&req)
		logger.FatalIfError(err)
		if req["result"] == "FAILURE" {
			c.JSON(409, "user trigger a failure") // Status 409 for Failure. Won't be retried
		} else {
			c.JSON(200, "")
		}
	})
	app.POST(qsBusiAPI+"/TransInCompensate", func(c *gin.Context) {
		logger.Infof("TransInCompensate")
		c.JSON(200, "")
	})
	app.POST(qsBusiAPI+"/TransOut", func(c *gin.Context) {
		logger.Infof("TransOut")
		c.JSON(200, "")
	})
	app.POST(qsBusiAPI+"/TransOutCompensate", func(c *gin.Context) {
		logger.Infof("TransOutCompensate")
		c.JSON(200, "")
	})
}

var dtmServer = daprdriver.AddrForProxiedHTTP("dtm", "/api/dtmsvr")

func finishRequest(result string) {
	req := &gin.H{"amount": 30, "result": result} // load of micro-service
	// DtmServer is the url of dtm
	saga := dtmcli.NewSaga(dtmServer, shortuuid.New()).
		// add a TransOut sub-transaction，forward operation with url: qsBusi+"/TransOut", reverse compensation operation with url: qsBusi+"/TransOutCompensate"
		Add(qsBusi+"/TransOut", qsBusi+"/TransOutCompensate", req).
		// add a TransIn sub-transaction, forward operation with url: qsBusi+"/TransIn", reverse compensation operation with url: qsBusi+"/TransInCompensate"
		Add(qsBusi+"/TransIn", qsBusi+"/TransInCompensate", req)
	// submit the created saga transaction，dtm ensures all sub-transactions either complete or get revoked
	saga.WaitResult = true
	err := saga.Submit()
	logger.Infof("result is: %v", err)
}
