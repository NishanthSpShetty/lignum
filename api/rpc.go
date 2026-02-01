package api

import (
	"context"
	"fmt"
	"net"

	interceptors "github.com/NishanthSpShetty/grpc-interceptors"
	proto "github.com/NishanthSpShetty/lignum/proto"
	"github.com/NishanthSpShetty/lignum/replication"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ proto.LignumServer = (*Server)(nil)

// Echo implements proto.LignumServer
func (s *Server) Echo(ctx context.Context, req *proto.EchoMessage) (*proto.EchoMessage, error) {
	return &proto.EchoMessage{
		Message: req.GetMessage(),
	}, nil
}

// Ping implements proto.LignumServer
func (s *Server) Ping(context.Context, *proto.PingRequest) (*proto.PingResponse, error) {
	return &proto.PingResponse{
		Message: "pong",
	}, nil
}

// CreateTopic implements proto.LignumServer
func (s *Server) CreateTopic(ctx context.Context, req *proto.Topic) (*proto.Ok, error) {
	name := req.GetName()

	if name == "" {
		return nil, status.Error(codes.InvalidArgument, "topic name is empty")
	}

	if s.message.TopicExist(name) {
		log.Error().Str("topic", name).Msg("topic exist with the given name")
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("topic %s exist", name))
	}

	bufferSize := s.message.BufferSize()

	liveReplication := req.GetEnableLiveReplication()

	s.message.CreateNewTopic(name, bufferSize, liveReplication)
	log.Debug().Str("topic", name).Bool("enable_live_replication", liveReplication).
		Uint64("quorum_count", req.GetQorumCount()).
		Msg("topic created")
	return &proto.Ok{}, nil
}

// ListTopic implements proto.LignumServer
func (s *Server) ListTopic(ctx context.Context, req *proto.ListTopicRequest) (*proto.ListTopicResponse, error) {
	topics := make([]*proto.Topic, 0)

	topix := s.message.GetTopics()

	for _, topic := range topix {
		topics = append(topics, &proto.Topic{
			Name:                  topic.GetName(),
			EnableLiveReplication: topic.LiveReplication(),
			Count:                 topic.GetCurrentOffset(),
			QorumCount:            0,
		})
	}
	return &proto.ListTopicResponse{
		Topic: topics,
	}, nil
}

// Read implements proto.LignumServer
func (s *Server) Read(ctx context.Context, req *proto.Query) (*proto.Messages, error) {
	resp := &proto.Messages{}

	from := req.GetFrom()
	to := req.GetTo()
	topic := req.GetTopic()

	if from < 0 || to <= from {
		return nil, status.Error(codes.InvalidArgument, "invalid range specified")
	}

	if topic == "" {
		return nil, status.Error(codes.InvalidArgument, "topic is empty")
	}

	if !s.message.TopicExist(topic) {
		log.Error().Str("topic", topic).Msg("topic does not exist")
		return nil, status.Error(codes.InvalidArgument, "topic does not exist")
	}

	messages := s.message.Get(topic, from, to)
	ms := make([]*proto.Message, 0, len(messages))
	for _, m := range messages {
		ms = append(ms, &proto.Message{
			Topic: topic,
			Data:  m.Data,
		})
	}

	resp.Message = ms
	return resp, nil
}

// Send implements proto.LignumServer
func (s *Server) Send(ctx context.Context, req *proto.Message) (*proto.Ok, error) {
	data := req.GetData()
	topic := req.GetTopic()

	if topic == "" {
		return nil, status.Error(codes.InvalidArgument, "topic is empty")
	}

	log.Debug().Bytes("data", data).Str("topic", topic).Msg("message received")
	mesg, liveReplication := s.message.Put(ctx, topic, data)
	if liveReplication {
		// write messages to replication queue
		payload := replication.Payload{
			Topic: topic,
			Id:    mesg.Id,
			Data:  mesg.Data,
		}
		s.replicationQueue <- payload
	}
	return &proto.Ok{}, nil
}

func (s *Server) setupGrpc() {
	interceptor := interceptors.NewInterceptor("lignum", log.Logger)
	opts := []grpc.ServerOption{interceptor.Get()}
	grpcServer := grpc.NewServer(opts...)
	proto.RegisterLignumServer(grpcServer, s)
	s.grpcServer = grpcServer
}

func (s *Server) startGrpc() error {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.config.Host, s.config.GrpcPort))
	if err != nil {
		return errors.Wrap(err, "Server.startGrpc")
	}
	log.Info().Str("host", s.config.Host).Int("port", s.config.GrpcPort).Msg("starting grpc server")
	return s.grpcServer.Serve(lis)
}
