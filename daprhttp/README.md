# sample for dtm to call normal http service deployed by dapr

## start dtm
``` bash
git clone github.com/dtm-labs/dtm && cd dtm
MICRO_SERVICE_DRIVER=dtm-driver-dapr dapr run --app-id dtm --app-protocol http --app-port 36789 -- go run main.go -d -r
```

## run sample
``` bash
dapr run --app-id app-http --app-protocol http --app-port 8082 go run main.go
```