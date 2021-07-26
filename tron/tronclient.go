package tron

import (
	"context"
	"os"

	"github.com/maticnetwork/heimdall/tron/pb"
	"google.golang.org/grpc"
)

// Client defines typed wrappers for the Tron RPC API.
type Client struct {
	client pb.WalletClient
}

// NewClient creates a client that uses the given RPC client.
func NewClient(url string) *Client {
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	//conn, err := grpc.Dial("grpc.shasta.trongrid.io:50051", grpc.WithInsecure())
	if err != nil {
		os.Exit(0)
	}
	return &Client{
		client: pb.NewWalletClient(conn),
	}
}

func (tc *Client) GetNowBlock(ctx context.Context) (int64, error) {
	block, err := tc.client.GetNowBlock2(ctx, &pb.EmptyMessage{})
	if err != nil {
		return 0, err
	}
	return block.BlockHeader.RawData.Number, nil
}
