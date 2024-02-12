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

func (s *Server) AddWhite(ctx context.Context, req *pb.IPMask) (*pb.Responce, error) {
	_ = ctx
	err := s.iplist.AddWhiteList(req.IP, req.Mask)
	switch {
	case err == nil:
		return &pb.Responce{Result: "IP was added in the white list"}, nil
	default:
		return &pb.Responce{Result: err.Error()}, status.Errorf(codes.InvalidArgument, "Invalid parameters")
	}
}

func (s *Server) DelWhite(ctx context.Context, ip *pb.IP) (*pb.Responce, error) {
	_ = ctx
	err := s.iplist.DelWhiteList(ip.IP)
	switch {
	case err == nil:
		return &pb.Responce{Result: "IP was deleted from the white list"}, nil
	default:
		return &pb.Responce{Result: err.Error()}, status.Errorf(codes.InvalidArgument, "Invalid parameters")
	}
}

func (s *Server) AddBlack(ctx context.Context, req *pb.IPMask) (*pb.Responce, error) {
	_ = ctx
	err := s.iplist.AddBlackList(req.IP, req.Mask)
	switch {
	case err == nil:
		return &pb.Responce{Result: "IP was added in the black list"}, nil
	default:
		return &pb.Responce{Result: err.Error()}, status.Errorf(codes.InvalidArgument, "Invalid parameters")
	}
}

func (s *Server) DelBlack(ctx context.Context, ip *pb.IP) (*pb.Responce, error) {
	_ = ctx
	err := s.iplist.DelBlackList(ip.IP)
	switch {
	case err == nil:
		return &pb.Responce{Result: "IP was deleted from the black list"}, nil
	default:
		return &pb.Responce{Result: err.Error()}, status.Errorf(codes.InvalidArgument, "Invalid parameters")
	}
}

func (s *Server) ResetBacket(ctx context.Context, backet *pb.Backet) (*pb.Responce, error) {
	_ = ctx
	if app.BacketType(backet.Type) == app.IP {
		if _, err := app.IPtoInt(backet.Backet); err != nil {
			s.logg.Error(err)
			return &pb.Responce{Result: err.Error()}, status.Errorf(codes.InvalidArgument, "Invalid parameters")
		}
	}
	if errors.Is(s.backets.ResetBacket(backet.Backet, app.BacketType(backet.Type)), app.ErrBacketNotFound) {
		return &pb.Responce{Result: fmt.Sprintf("%s backet not found", backet.Type)}, nil
	}
	return &pb.Responce{Result: fmt.Sprintf("%s backet was reset", backet.Type)}, nil
}

func IPFormApptoPB(list []app.IPItem) *pb.List {
	l := make([]*pb.IPitem, len(list))
	for i, it := range list {
		l[i] = &pb.IPitem{
			IP:     it.IP,
			Mask:   it.Mask,
			IPfrom: it.IPfrom,
			IPto:   it.IPto,
		}
	}
	return &pb.List{Items: l}
}

func (s *Server) GetList(ctx context.Context, tipe *pb.TypeList) (*pb.List, error) {
	_ = ctx
	list := s.iplist.GetList(tipe.Type)
	return IPFormApptoPB(list), nil
}
