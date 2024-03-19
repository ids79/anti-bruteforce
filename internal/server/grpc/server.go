package internalgrpc

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/ids79/anti-bruteforce/internal/app"
	"github.com/ids79/anti-bruteforce/internal/config"
	"github.com/ids79/anti-bruteforce/internal/logger"
	"github.com/ids79/anti-bruteforce/internal/pb"
	"github.com/ids79/anti-bruteforce/internal/storage"
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

func NewServer(backets app.WorkWithBackets, iplist *app.IPList, logg logger.Logg, config *config.Config) *Server {
	return &Server{
		logg:    logg,
		backets: backets,
		iplist:  iplist,
		conf:    config,
	}
}

func (s *Server) Start() error {
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

func (s *Server) AddWhite(ctx context.Context, req *pb.IP) (*pb.Responce, error) {
	err := s.iplist.AddWhiteList(ctx, req.IP)
	if err != nil {
		s.logg.Error("grpc, AddWhite: ", err)
		if errors.Is(err, storage.ErrIPAlreadyExistInBlackRange) {
			return &pb.Responce{Result: err.Error()}, nil
		}
		return nil, status.Errorf(codes.InvalidArgument, "Invalid parameters")
	}
	return &pb.Responce{Result: "IP was added in the white list"}, nil
}

func (s *Server) DelWhite(ctx context.Context, req *pb.IP) (*pb.Responce, error) {
	err := s.iplist.DelWhiteList(ctx, req.IP)
	if err != nil {
		s.logg.Error("grpc, DelWhite: ", err)
		return &pb.Responce{Result: err.Error()}, status.Errorf(codes.InvalidArgument, "Invalid parameters")
	}
	return &pb.Responce{Result: "IP was deleted from the white list"}, nil
}

func (s *Server) AddBlack(ctx context.Context, req *pb.IP) (*pb.Responce, error) {
	err := s.iplist.AddBlackList(ctx, req.IP)
	if err != nil {
		s.logg.Error("grpc, AddBlack: ", err)
		if errors.Is(err, storage.ErrIPAlreadyExistInWhiteRange) {
			return &pb.Responce{Result: err.Error()}, nil
		}
		return nil, status.Errorf(codes.InvalidArgument, "Invalid parameters")
	}
	return &pb.Responce{Result: "IP was added in the black list"}, nil
}

func (s *Server) DelBlack(ctx context.Context, req *pb.IP) (*pb.Responce, error) {
	err := s.iplist.DelBlackList(ctx, req.IP)
	if err != nil {
		s.logg.Error("grpc, DelBlack: ", err)
		return &pb.Responce{Result: err.Error()}, status.Errorf(codes.InvalidArgument, "Invalid parameters")
	}
	return &pb.Responce{Result: "IP was deleted from the black list"}, nil
}

func (s *Server) ResetBacket(ctx context.Context, backet *pb.Backet) (*pb.Responce, error) {
	if app.BacketType(backet.Type) == app.IP {
		if ip := net.ParseIP(backet.Backet); ip == nil {
			s.logg.Error("resetBucket: ", backet.Backet, " - parse IP error")
			return &pb.Responce{Result: "parse IP error"}, status.Errorf(codes.InvalidArgument, "Invalid parameters")
		}
	}
	err := s.backets.ResetBacket(ctx, backet.Backet, app.BacketType(backet.Type))
	if errors.Is(err, app.ErrBacketNotFound) {
		s.logg.Error("grpc, ResetBacket: ", err)
		return &pb.Responce{Result: fmt.Sprintf("%s backet not found", backet.Type)}, nil
	} else if errors.Is(err, app.ErrContextWasExpire) {
		s.logg.Error("grpc, ResetBacket: ", err)
		return &pb.Responce{Result: fmt.Sprintf("%s context was expire", backet.Type)}, nil
	}
	return &pb.Responce{Result: fmt.Sprintf("%s backet was reset", backet.Type)}, nil
}

func (s *Server) GetList(ctx context.Context, tipe *pb.TypeList) (*pb.List, error) {
	list := s.iplist.GetList(ctx, tipe.Type)
	return &pb.List{Items: list}, nil
}
