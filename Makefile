.PHONY: protogen build buildrace buildcore buildcorerace buildclient
.DEFAULT_GOAL := build

protogen:
	@protoc \
		--go_out=raft/service/genproto \
		--go-grpc_out=raft/service/genproto \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative\
		--proto_path=raft/service/proto raft/service/proto/raft_service.proto;

	@protoc \
		--go_out=datastore/kv/service/genproto \
		--go-grpc_out=datastore/kv/service/genproto \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative\
		--proto_path=datastore/kv/service/proto datastore/kv/service/proto/kv_service.proto;

buildclient: protogen
	@go build -o bin/linux-amd64/ ./client/

buildcore: protogen
	@go build -o bin/linux-amd64/

buildcorerace: protogen
	@go build -o bin/linux-amd64/ -race 

build: protogen buildcore buildclient

buildrace: protogen buildcorerace buildclient
