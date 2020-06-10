package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	"AliveVirtualGift_SessionService/src/database"
	"AliveVirtualGift_SessionService/src/proto"
	"AliveVirtualGift_SessionService/src/service"
	"AliveVirtualGift_SessionService/src/utils"
)

func init() {

	var err error
	// Load environment
	err = godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	utils.InitRedis()
}

func main() {

	listener, err := net.Listen("tcp", ":8000")
	if err != nil {
		panic(err)
	}

	var newdb database.DBInfo
	db, err := newdb.GetDB()
	if err != nil {
		fmt.Printf("failed to open database: %v", err)
		return
	}

	srv := grpc.NewServer()

	service := service.NewSessionServiceServer(db)

	proto.RegisterSessionServiceServer(srv, service)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	ctx := context.Background()
	go func() {
		for range c {
			// sig is a ^C, handle it
			log.Println("Shutting down gRPC Session service server...")

			srv.GracefulStop()

			<-ctx.Done()
		}
	}()

	if e := srv.Serve(listener); e != nil {
		panic(err)
	}
}
