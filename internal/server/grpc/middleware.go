package internalgrpc

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type Middleware func(ctx context.Context) error

func UnaryServerMiddleWareInterceptor(mid Middleware) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		h grpc.UnaryHandler,
	) (interface{}, error) {
		if err := mid(ctx); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "%s is rejected by middleware. Error: %v", info.FullMethod, err)
		}
		return h(ctx, req)
	}
}

func (s *Server) loggingReq(ctx context.Context) error {
	if ctx == nil {
		s.logg.Error("context is empty, logging will not perform")
		return errors.New("context is empty, logging will not perform")
	}
	p, _ := peer.FromContext(ctx)
	ip := p.Addr.String()
	var sb strings.Builder
	sb.WriteString(ip)
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		ctype := md.Get("content-type")
		method := md.Get("method")
		if len(ctype) > 0 {
			sb.WriteString(" ")
			sb.WriteString(ctype[0])
		}
		if len(method) > 0 {
			sb.WriteString(" ")
			sb.WriteString(method[0])
		}
	}
	s.logg.Info(sb.String())
	return nil
}
