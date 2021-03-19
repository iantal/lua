package main

import (
	"fmt"
	"net"
	"os"
	"time"

	ldprotos "github.com/iantal/ld/protos/ld"
	"github.com/iantal/lua/internal/domain"
	"github.com/iantal/lua/internal/files"
	"github.com/iantal/lua/internal/server"
	"github.com/iantal/lua/internal/util"
	protos "github.com/iantal/lua/protos/lua"
	luaresult "github.com/iantal/lua/protos/luaresult"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // postgres
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func gRPCConnection(host string) *grpc.ClientConn {
	conn, err := grpc.Dial(
		host,
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1000*3000)),
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(600*time.Second),
	)
	if err != nil {
		panic(err)
	}
	return conn
}

func main() {
	viper.AutomaticEnv()
	log := util.NewLogger()

	// create a new gRPC server, use WithInsecure to allow http connections
	gs := grpc.NewServer()

	bp := "/opt/data"
	rm := fmt.Sprintf("%v", viper.Get("RM_HOST"))
	ld := fmt.Sprintf("%v", viper.Get("LD_HOST"))

	stor, err := files.NewLocal(bp, 1024*1000*1000*5)
	if err != nil {
		log.WithField("error", err).Error("Unable to create storage")
		os.Exit(1)
	}

	user := viper.Get("POSTGRES_USER")
	password := viper.Get("POSTGRES_PASSWORD")
	database := viper.Get("POSTGRES_DB")
	host := viper.Get("POSTGRES_HOST")
	port := viper.Get("POSTGRES_PORT")
	connection := fmt.Sprintf("host=%v port=%v user=%v dbname=%v password=%v sslmode=disable", host, port, user, database, password)

	var db *gorm.DB
	db, err = gorm.Open("postgres", connection)
	defer db.Close()
	if err != nil {
		log.WithField("error", err).Error("Failed to connect to database")
		os.Exit(1)
	}

	err = db.DB().Ping()
	if err != nil {
		log.WithField("error", err).Error("Failed to ping database")
		os.Exit(1)
	}

	db.AutoMigrate(&domain.Dependency{}, &domain.File{})

	// setup GRPC for LD
	connLD := gRPCConnection(ld)
	defer connLD.Close()
	ldcli := ldprotos.NewUsedLanguagesClient(connLD)

	c := server.NewLibraryUsageAnalyser(log, stor, db, rm, ldcli)
	rp := server.NewResultProvider(log, db)

	protos.RegisterAnalyzerServer(gs, c)
	luaresult.RegisterResultProviderServer(gs, rp)

	// register the reflection service which allows clients to determine the methods
	// for this gRPC service
	reflection.Register(gs)

	// create a TCP socket for inbound server connections
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", 8008))
	if err != nil {
		log.WithField("error", err).Error("Unable to create listener")
		os.Exit(1)
	}

	log.Info("Starting server bind_address ", l.Addr().String())
	// listen for requests
	gs.Serve(l)
}
