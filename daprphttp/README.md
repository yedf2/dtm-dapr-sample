# sample for dtm to call normal http service deployed by dapr

## start dtm
``` bash
git clone github.com/dtm-labs/dtm && cd dtm
MICRO_SERVICE_DRIVER=dtm-driver-dapr dapr run --app-id dtm --app-protocol http --app-port 36789 -- go run main.go -d
```

## run sample
``` bash
dapr run --app-id app-phttp --app-protocol http --app-port 8081 go run main.go
```