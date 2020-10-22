package main

import (
	"fmt"
	"net"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/iantal/lua/internal/files"
	"github.com/iantal/lua/internal/server"
	protos "github.com/iantal/lua/protos/lua"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // postgres
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	viper.AutomaticEnv()
	log := hclog.Default()

	// create a new gRPC server, use WithInsecure to allow http connections
	gs := grpc.NewServer()

	bp := fmt.Sprintf("%v", viper.Get("BASE_PATH"))
	rm := fmt.Sprintf("%v", viper.Get("RM_HOST"))
	ld := fmt.Sprintf("%v", viper.Get("LD_HOST"))

	stor, err := files.NewLocal(bp, 1024*1000*1000*5)
	if err != nil {
		log.Error("Unable to create storage", "error", err)
		os.Exit(1)
	}

	user := viper.Get("POSTGRES_USER")
	password := viper.Get("POSTGRES_PASSWORD")
	database := viper.Get("POSTGRES_DB")
	host := viper.Get("POSTGRES_HOST")
	port := viper.Get("POSTGRES_PORT")
	connection := fmt.Sprintf("host=%v port=%v user=%v dbname=%v password=%v sslmode=disable", host, port, user, database, password)

	db, err := gorm.Open("postgres", connection)
	defer db.Close()
	if err != nil {
		fmt.Println(err)
		panic("Failed to connect to database!")
	}

	err = db.DB().Ping()
	if err != nil {
		panic("Ping failed!")
	}

	// c := server.NewMCDownloader(log, stor, db)
	c := server.NewLibraryUsageAnalyser(log, stor, db, rm, ld)

	// register the currency server
	protos.RegisterAnalyzerServer(gs, c)

	// register the reflection service which allows clients to determine the methods
	// for this gRPC service
	reflection.Register(gs)

	// create a TCP socket for inbound server connections
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", 8008))
	if err != nil {
		log.Error("Unable to create listener", "error", err)
		os.Exit(1)
	}

	log.Info("Starting server", "bind_address", l.Addr().String())
	// listen for requests
	gs.Serve(l)
}