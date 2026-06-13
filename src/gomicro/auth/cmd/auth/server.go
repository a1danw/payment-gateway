package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"

	"github.com/a1danw/gomicro/internal/implementation/auth"
	pb "github.com/a1danw/gomicro/proto"
	"google.golang.org/grpc"
)

// verify user credentials the gateway sends us (login request) via mysql
const (
	dbDriver   = "mysql"
	dbUser     = "root"
	dbPassword = "Admin123"
	dbName     = "User"
)

var db *sql.DB

func main() {
	var err error

	// (data source name) opens a db connection - "username:password@tcp(localhost:3306)/database_name"
	dsn := fmt.Sprintf("%s:%s@tcp(localhost:3306)/%s", dbUser, dbPassword, dbName)

	db, err = sql.Open(dbDriver, dsn)
	if err != nil {
		log.Fatal(err)
	} // exit application after logging the error

	// defer the closure of the db, in the event the application is exited/shutdown, make sure the db is closed/shutdown in a clean way
	// defer function essentially will run before the application closes
	defer func() {
		if db.Close(); err != nil {
			log.Printf("Error closing db: %s", err)
		}
	}()

	// Ping the db to ensure the connection is valid otherwise close db
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()                                     // creating empty grpc server
	authServerImplementation := auth.NewAuthImplementation(db)         // contains all the methods in internal/implementation/auth/auth.go so the gateway can call GetToken for login and ValidateToken for other routes
	pb.RegisterAuthServiceServer(grpcServer, authServerImplementation) // registers the auth service in order to use the grpc server - from generated auth_svc_grpc.pb.go - grpc passed and function implementations

	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("failed to listen on port 9000: %v", err)
	}
	log.Printf("server listening at %v", listener.Addr())
	// incoming grpc requests
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
