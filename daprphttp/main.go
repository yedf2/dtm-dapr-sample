package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dtm-labs/client/dtmcli"
	daprdriver "github.com/dtm-labs/dtmdriver-dapr"

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
	finishRequest()
}

// QsStartSvr quick start: start server
func startSvr() {
	app := gin.New()
	qsAddRoute(app)
	log.Printf("quick start examples listening at %d", qsBusiPort)
	go func() {
		_ = app.Run(fmt.Sprintf(":%d", qsBusiPort))
	}()
	time.Sleep(100 * time.Millisecond)
}

func qsAddRoute(app *gin.Engine) {
	app.POST(qsBusiAPI+"/TransIn", func(c *gin.Context) {
		log.Printf("TransIn")
		c.JSON(200, "")
		// c.JSON(409, "") // Status 409 for Failure. Won't be retried
	})
	app.POST(qsBusiAPI+"/TransInCompensate", func(c *gin.Context) {
		log.Printf("TransInCompensate")
		c.JSON(200, "")
	})
	app.POST(qsBusiAPI+"/TransOut", func(c *gin.Context) {
		log.Printf("TransOut")
		c.JSON(200, "")
	})
	app.POST(qsBusiAPI+"/TransOutCompensate", func(c *gin.Context) {
		log.Printf("TransOutCompensate")
		c.JSON(200, "")
	})
}

var dtmServer = daprdriver.AddrForProxiedHTTP("dtm", "/api/dtmsvr")

func finishRequest() string {
	req := &gin.H{"amount": 30} // load of micro-service
	// DtmServer is the url of dtm
	saga := dtmcli.NewSaga(dtmServer, shortuuid.New()).
		// add a TransOut sub-transaction，forward operation with url: qsBusi+"/TransOut", reverse compensation operation with url: qsBusi+"/TransOutCompensate"
		Add(qsBusi+"/TransOut", qsBusi+"/TransOutCompensate", req).
		// add a TransIn sub-transaction, forward operation with url: qsBusi+"/TransIn", reverse compensation operation with url: qsBusi+"/TransInCompensate"
		Add(qsBusi+"/TransIn", qsBusi+"/TransInCompensate", req)
	// submit the created saga transaction，dtm ensures all sub-transactions either complete or get revoked
	saga.WaitResult = true
	err := saga.Submit()

	if err != nil {
		panic(err)
	}
	return saga.Gid
}
