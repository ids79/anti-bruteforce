package internalgrpc

import (
	"context"
	"net"

	"github.com/ids79/anti-bruteforcer/internal/app"
	"github.com/ids79/anti-bruteforcer/internal/config"
	"github.com/ids79/anti-bruteforcer/internal/logger"
	"github.com/ids79/anti-bruteforcer/internal/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedCLiApiServer
	ls      net.Listener
	server  *grpc.Server
	backets app.WorkWithBackets
	iplist  app.WorkWithIPList
	logg    logger.Logg
	conf    *config.Config
}

func NewServer(backets *app.Backets, iplist *app.IPList, logg logger.Logg, config config.Config) *Server {
	return &Server{
		logg:    logg,
		backets: backets,
		iplist:  iplist,
		conf:    &config,
	}
}

func (s *Server) Start(ctx context.Context) error {
	var err error
	s.ls, err = net.Listen("tcp", s.conf.GRPCServer.Address)
	if err != nil {
		s.logg.Error(err)
		return nil
	}
	s.server = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			UnaryServerMiddleWareInterceptor(s.loggingReq)),
	)
	pb.RegisterCLiApiServer(s.server, s)
	s.logg.Info("starting grpc server on ", s.ls.Addr().String())
	if err := s.server.Serve(s.ls); err != nil {
		s.logg.Error(err)
		s.ls.Close()
	}
	return nil
}

func (s *Server) Close() {
	s.logg.Info("server grpc is stopping...")
	s.server.Stop()
	s.ls.Close()
}

func (s *Server) AddWhite(ctx context.Context, req *pb.IPMask) (*pb.Responce, error) {
	err := s.iplist.AddWhiteList(req.IP, req.Mask)
	switch {
	case err == nil:
		return &pb.Responce{Result: "IP was added in the white list"}, nil
	default:
		return &pb.Responce{Result: err.Error()}, status.Errorf(codes.InvalidArgument, "Invalid parameters")
	}
}

func (s *Server) DelWhite(ctx context.Context, IP *pb.IP) (*pb.Responce, error) {
	err := s.iplist.DelWhiteList(IP.IP)
	switch {
	case err == nil:
		return &pb.Responce{Result: "IP was deleted from the white list"}, nil
	default:
		return &pb.Responce{Result: err.Error()}, status.Errorf(codes.InvalidArgument, "Invalid parameters")
	}
}
