package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"log-service/logs"
	data "log-service/models"
	"net"
)

type LogServer struct {
	logs.UnimplementedLogServiceServer
	Models data.MongoRepository
}

func (l *LogServer) WriteLog(ctx context.Context, req *logs.LogRequest) (*logs.LogResponse, error) {
	input := req.GetLogEntry()

	logEntry := data.LogEntry{
		Name: input.Name,
		Data: input.Data,
	}

	err := l.Models.Insert(logEntry)
	if err != nil {
		fmt.Println("Error in log-service/grpc, 28")
		return &logs.LogResponse{Result: "failed"}, err
	}

	return &logs.LogResponse{Result: "Logged!"}, nil
}

func (app *Config) gRPCListen() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", gRpcPort))
	if err != nil {
		fmt.Println("Error in log-service/grpc, 38")
		log.Fatalf("Failed to listen gRPC: %v", err)
	}

	s := grpc.NewServer()
	logs.RegisterLogServiceServer(s, &LogServer{Models: data.MongoRepository{}})

	log.Printf("gRPC server started on port %s", gRpcPort)
	if err := s.Serve(lis); err != nil {
		fmt.Println("Error in log-service/grpc, 47")
		log.Fatalf("Failed to listen gRPC: %v", err)
	}
}
