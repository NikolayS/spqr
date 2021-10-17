package router

import (
	context "context"

	"github.com/pg-sharding/spqr/router/pkg"
	"github.com/pg-sharding/spqr/router/pkg/console"
	"github.com/pg-sharding/spqr/router/pkg/qrouter"
	"github.com/pg-sharding/spqr/router/pkg/rrouter"
	proto "github.com/pg-sharding/spqr/router/protos"
	"google.golang.org/grpc"
)

type RouterConn struct {
	proto.UnimplementedQueryServiceServer

	Console console.Console
}

func (s RouterConn) Process(ctx context.Context, request *proto.QueryExecuteRequest) (*proto.QueryExecuteResponse, error) {
	_ = s.Console.ProcessQuery(request.Query, rrouter.NewFakeClient())

	return &proto.QueryExecuteResponse{}, nil
}

func NewSpqrConn(c console.Console) *RouterConn {
	return &RouterConn{
		Console: c,
	}
}

var _ proto.QueryServiceServer = RouterConn{}

type KeyRangeService struct {
	proto.UnimplementedKeyRangeServiceServer

	impl  pkg.Router
	qimpl qrouter.Qrouter
}

func (k KeyRangeService) LockKeyRange(ctx context.Context, in *proto.LockKeyRangeRequest, opts ...grpc.CallOption) (*proto.LockKeyRangeReply, error) {
	_ = k.qimpl.Lock(in.Krid)

	return nil, nil
}

func (k KeyRangeService) UnlockKeyRange(ctx context.Context, in *proto.UnlockKeyRangeRequest, opts ...grpc.CallOption) (*proto.UnlockKeyRangeReply, error) {
	panic("implement me")
}

func (k KeyRangeService) SplitKeyRange(ctx context.Context, in *proto.SplitKeyRangeRequest, opts ...grpc.CallOption) (*proto.SplitKeyRangeReply, error) {
	panic("implement me")
}

func (k KeyRangeService) AddShardingColumn(ctx context.Context, in *proto.AddShardingColumnRequest, opts ...grpc.CallOption) (*proto.AddShardingColumnReply, error) {
	panic("implement me")
}

func (k KeyRangeService) AddLocalTable(ctx context.Context, in *proto.AddLocalTableRequest, opts ...grpc.CallOption) (*proto.AddLocalTableReply, error) {
	panic("implement me")
}

func (k KeyRangeService) ListKeyRange(ctx context.Context, in *proto.ListKeyRangeRequest, opts ...grpc.CallOption) (*proto.KeyRangeReply, error) {
	var krs []*proto.KeyRange
	for _, el := range k.qimpl.KeyRanges() {
		krs = append(krs, el.ToProto())
	}
	return &proto.KeyRangeReply{
		KeyRanges: krs,
	}, nil
}

var _ proto.KeyRangeServiceClient = KeyRangeService{}
