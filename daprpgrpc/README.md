# sample for dtm to call normal grpc service deployed by dapr

## start dtm
``` bash
git clone github.com/dtm-labs/dtm && cd dtm
MICRO_SERVICE_DRIVER=dtm-driver-dapr dapr run --app-id dtm --app-protocol grpc --app-port 36790 -- go run main.go -d -r
```

## run sample
``` bash
dapr run --app-id app-pgrpc --app-protocol grpc --app-port 8083 go run main.go
```