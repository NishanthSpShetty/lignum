package api

import (
	"context"
	"fmt"
	"net"

	interceptors "github.com/NishanthSpShetty/grpc-interceptors"
	proto "github.com/NishanthSpShetty/lignum/proto"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
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
func (s *Server) CreateTopic(context.Context, *proto.Topic) (*proto.Ok, error) {
	panic("unimplemented")
}

// Read implements proto.LignumServer
func (s *Server) Read(context.Context, *proto.Query) (*proto.Messages, error) {
	panic("unimplemented")
}

// Send implements proto.LignumServer
func (s *Server) Send(context.Context, *proto.Message) (*proto.Ok, error) {
	panic("unimplemented")
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
