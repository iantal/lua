package service

import (
	"context"
	"log"
	"net"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	ldprotos "github.com/iantal/ld/protos/ld"

	"github.com/iantal/lua/internal/util"
	"github.com/iantal/lua/protos/lua"
	"github.com/jinzhu/gorm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type mockLDServer struct {
	ldprotos.UnimplementedUsedLanguagesServer
}

func (m *mockLDServer) Breakdown(ctx context.Context, req *ldprotos.BreakdownRequest) (*ldprotos.BreakdownResponse, error) {
	if req.CommitHash == "b" {
		return nil, status.Errorf(codes.InvalidArgument, "Error")
	}
	return &ldprotos.BreakdownResponse{
		Breakdown: []*ldprotos.Language{
			{
				Name: "Bash",
				Files: []string{
					"a.sh",
					"ab.sh",
					"abc.sh",
					"b.sh",
				},
			},
			{
				Name: "Java",
				Files: []string{
					"clients/src/main/java/org/apache/kafka/clients/ApiVersion.java",
					"clients/src/main/java/org/apache/kafka/clients/ApiVersions.java",
					"clients/src/main/java/org/apache/kafka/clients/ClientDnsLookup.java",
					"clients/src/main/java/org/apache/kafka/clients/ClientRequest.java",
					"clients/src/main/java/org/apache/kafka/clients/ClientResponse.java",
				},
			},
			{
				Name: "HTML",
				Files: []string{
					"core/src/main/scala/kafka/log/package.html",
					"core/src/main/scala/kafka/network/package.html",
					"core/src/main/scala/kafka/server/package.html",
				},
			},
		},
	}, nil
}

func dialer() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)

	server := grpc.NewServer()

	ldprotos.RegisterUsedLanguagesServer(server, &mockLDServer{})

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestFilterJavaFiles(t *testing.T) {

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	ldcli := ldprotos.NewUsedLanguagesClient(conn)

	db, _, _ := sqlmock.New()           // mock sql.DB
	gdb, _ := gorm.Open("postgres", db) // open gorm db

	a := NewAnalyzer(util.NewLogger(), nil, gdb, "", ldcli)

	r := a.getFilesByLanguage("282119ba-7f0a-478f-9d94-bb59dfbaefa7", "a", []*lua.LuaLibrary{})

	expected := []string{
		"clients/src/main/java/org/apache/kafka/clients/ApiVersion.java",
		"clients/src/main/java/org/apache/kafka/clients/ApiVersions.java",
		"clients/src/main/java/org/apache/kafka/clients/ClientDnsLookup.java",
		"clients/src/main/java/org/apache/kafka/clients/ClientRequest.java",
		"clients/src/main/java/org/apache/kafka/clients/ClientResponse.java",
	}

	actual := []string{}
	for _, res := range r {
		actual = append(actual, res.File.Name)
	}

	if !equal(actual, expected) {
		t.Errorf("expected %s, actual %s", expected, actual)
	}

}

func TestLDError(t *testing.T) {

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	ldcli := ldprotos.NewUsedLanguagesClient(conn)

	db, _, _ := sqlmock.New()           // mock sql.DB
	gdb, _ := gorm.Open("postgres", db) // open gorm db

	a := NewAnalyzer(util.NewLogger(), nil, gdb, "", ldcli)

	r := a.getFilesByLanguage("282119ba-7f0a-478f-9d94-bb59dfbaefa7", "b", []*lua.LuaLibrary{})

	expected := []string{}

	actual := []string{}
	for _, res := range r {
		actual = append(actual, res.File.Name)
	}

	if !equal(actual, expected) {
		t.Errorf("expected %s, actual %s", expected, actual)
	}

}
