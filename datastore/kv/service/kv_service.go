package service

import (
	"context"
	"github.com/gr455/omnipresence/datastore/kv"
	pb "github.com/gr455/omnipresence/datastore/kv/service/genproto"
	"github.com/gr455/omnipresence/mq"
	"log"
)

type KeyValueServer struct {
	kv *kv.KeyValueStore

	pb.UnimplementedKeyValueServer
}

func NewKeyValueServer(mq *mq.MessageQueue) *KeyValueServer {
	k, err := kv.Initialize(mq)
	if err != nil {
		log.Fatalf("Fatal: Cannot create KV object: %v", err)
		return nil
	}

	return &KeyValueServer{kv: k}
}

func (s *KeyValueServer) Get(ctx context.Context, request *pb.GetRequest) (*pb.GetResponse, error) {
	value := s.kv.Get(request.Key)
	return &pb.GetResponse{Value: value}, nil
}

func (s *KeyValueServer) Put(ctx context.Context, request *pb.PutRequest) (*pb.PutResponse, error) {
	s.kv.Put(request.Key, request.Value)
	return &pb.PutResponse{}, nil

}

func (s *KeyValueServer) GetMap(ctx context.Context, request *pb.GetMapRequest) (*pb.GetMapResponse, error) {
	mp := s.kv.GetMap()
	return &pb.GetMapResponse{Mp: mp}, nil
}
