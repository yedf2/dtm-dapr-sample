# sample for dtm to call normal http service deployed by dapr

## start dtm
``` bash
git clone github.com/dtm-labs/dtm && cd dtm
MICRO_SERVICE_DRIVER=dtm-driver-dapr dapr run --app-id dtm --app-protocol grpc --app-port 36790 --dapr-grpc-port 30004 -- go run main.go -d
```

## run sample
``` bash
dapr run --app-id app-grpc --app-protocol grpc --app-port 8084 --dapr-grpc-port 40004 go run main.go
```