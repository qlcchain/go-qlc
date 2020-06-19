PROTOCHEAD=protoc -I/usr/local/include -I. -I${GOPATH}/src -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis

all: proto gw swagger

proto:
	@$(PROTOCHEAD) --go_out=plugins=grpc,paths=source_relative:..   ./types/*.proto
	@$(PROTOCHEAD) --go_out=plugins=grpc,paths=source_relative:..   ./*.proto

gw:
	@$(PROTOCHEAD) --grpc-gateway_out=logtostderr=true,paths=source_relative:..  ./types/*.proto
	@$(PROTOCHEAD) --grpc-gateway_out=logtostderr=true,paths=source_relative:..  ./*.proto

swagger:
	@$(PROTOCHEAD) --swagger_out=logtostderr=true:../swagger *.proto
