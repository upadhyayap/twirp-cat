package main

import (
	"context"
	"log"
	"net/http"

	"github.com/twitchtv/twirp"
	pb "github.com/upadhyayap/twirp-cat/twirp/service"
)

type HelloWorldServer struct{}

func (s *HelloWorldServer) Hello(ctx context.Context, req *pb.HelloReq) (*pb.HelloResp, error) {
	if req.Subject == "" {
		return nil, twirp.NewError(twirp.InvalidArgument, "subject is required")
	}
	subject := "there"
	if req.Subject != "" {
		subject = req.Subject
	}

	return &pb.HelloResp{Text: "Hello " + subject}, nil
}

// Run the implementation in a local server
func main() {
	// default path prefix is /twirp
	//twirpHandler := pb.NewHelloWorldServer(&HelloWorldServer{})
	// to use custom path prefix
	 //twirpHandler := pb.NewHelloWorldServer(&HelloWorldServer{}, twirp.WithServerPathPrefix("custom"))

	// Adding interceptors and server hooks
	twirpHandler := pb.NewHelloWorldServer(&HelloWorldServer{}, twirp.WithServerInterceptors(NewInterceptorHijackRequest()), twirp.WithServerHooks(NewLoggingServerHooks()))

	// You can use any mux you like - NewHelloWorldServer gives you an http.Handler.
	mux := http.NewServeMux()
	// The generated code includes a method, PathPrefix(), which
	// can be used to mount your service on a mux.
	mux.Handle(twirpHandler.PathPrefix(), twirpHandler)
	http.ListenAndServe(":8080", mux)
}

// Server hook for logging
func NewLoggingServerHooks() *twirp.ServerHooks {
	return &twirp.ServerHooks{
		RequestRouted: func(ctx context.Context) (context.Context, error) {
			method, _ := twirp.MethodName(ctx)
			log.Println("Method: " + method)
			return ctx, nil
		},
		Error: func(ctx context.Context, twerr twirp.Error) context.Context {
			log.Println("Error: " + string(twerr.Code()))
			return ctx
		},
		ResponseSent: func(ctx context.Context) {
			log.Println("Response Sent (error or success)")
		},
	}
}

func NewInterceptorHijackRequest() twirp.Interceptor {
	return func(next twirp.Method) twirp.Method {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			methodBeingHandled, _ := twirp.MethodName(ctx)
			if methodBeingHandled == "Hello" {
				return next(ctx, &pb.HelloReq{Subject: "hijacked"})
			}
			return next(ctx, req)
		}
	}
}
