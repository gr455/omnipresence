protogen:
	@protoc \
		--go_out=raft/service/genproto \
		--go-grpc_out=raft/service/genproto \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative\
		--proto_path=raft/service/proto raft/service/proto/raft_service.proto
