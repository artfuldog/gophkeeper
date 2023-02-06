package grpcapi

import (
	"context"
	"log"
	"net"
	"testing"

	"github.com/artfuldog/gophkeeper/internal/mocks/mockauth"
	"github.com/artfuldog/gophkeeper/internal/mocks/mockdb"
	"github.com/artfuldog/gophkeeper/internal/mocks/mocklogger"
	"github.com/artfuldog/gophkeeper/internal/pb"
	"github.com/golang/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

var (
	mockAny = gomock.Any()
	testCtx = context.Background()
)

type TestSuiteGRPCServer struct {
	MockCtrl    *gomock.Controller
	Conn        *grpc.ClientConn
	DB          *mockdb.MockDB
	Authorizer  *mockauth.MockA
	UsersClient pb.UsersClient
	ItemsClient pb.ItemsClient
}

type TestGRCPServices struct {
	usersService *UsersService
	itemsService *ItemsService
}

func NewTestSuiteGRPCServer(t *testing.T, opt ...grpc.ServerOption) (ts *TestSuiteGRPCServer, err error) {
	ts = new(TestSuiteGRPCServer)

	ts.MockCtrl = gomock.NewController(t)

	ts.DB = mockdb.NewMockDB(ts.MockCtrl)
	ts.Authorizer = mockauth.NewMockA(ts.MockCtrl)
	logger := mocklogger.NewMockLogger()

	services := &TestGRCPServices{
		usersService: NewUsersService(ts.DB, logger, ts.Authorizer),
		itemsService: NewItemsService(ts.DB, logger),
	}

	ts.Conn, err = createTestGRPCBufConn(context.Background(), services, opt...)
	if err != nil {
		return
	}

	ts.UsersClient = pb.NewUsersClient(ts.Conn)
	ts.ItemsClient = pb.NewItemsClient(ts.Conn)

	return
}

func (ts *TestSuiteGRPCServer) Stop() {
	ts.MockCtrl.Finish()
	if err := ts.Conn.Close(); err != nil {
		log.Printf("failed close connection: %v", err)
	}
}

func testGRPCdialer(s *TestGRCPServices, opt ...grpc.ServerOption) func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)

	server := grpc.NewServer(opt...)

	pb.RegisterUsersServer(server, s.usersService)
	pb.RegisterItemsServer(server, s.itemsService)

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Printf("gRPC-server error: %v", err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func createTestGRPCBufConn(ctx context.Context, s *TestGRCPServices, opt ...grpc.ServerOption) (*grpc.ClientConn, error) {
	return grpc.DialContext(ctx, "", grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(testGRPCdialer(s, opt...)))
}
