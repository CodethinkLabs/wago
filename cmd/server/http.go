package main

import (
	"context"
	"encoding/hex"
	"github.com/CodethinkLabs/wago/pkg/proto"
	"github.com/CodethinkLabs/wago/pkg/wallet"
	"golang.org/x/crypto/ed25519"
	"google.golang.org/grpc"
	"net"
)

type walletServer struct {
	store *wallet.Store
}

func (s walletServer) SubmitTransaction(*proto.Transaction, proto.WalletService_SubmitTransactionServer) error {
	panic("implement me")
}

func (s walletServer) GetBalance(ctx context.Context, request *proto.BalanceRequest) (*proto.Balances, error) {
	balances := proto.Balances{Wallets: make(map[string]*proto.Wallet, 0)}
	for _, publicKey := range request.PublicKeys {
		str := hex.EncodeToString(publicKey)
		protoWallet := proto.Wallet{Currencies: make(map[string]*proto.DecimalAmount, 0)}
		balances.Wallets[str] = &protoWallet
		currencies, ok := s.store.Lookup(ed25519.PublicKey(publicKey))
		if !ok {
			continue
		}
		for name, amount := range currencies {
			protoWallet.Currencies[string(name)] = &proto.DecimalAmount{
				Value:   amount.Value,
				Decimal: int64(amount.Decimal),
			}
		}
	}
	return &balances, nil
}

func (s walletServer) Subscribe(*proto.Empty, proto.WalletService_SubscribeServer) error {
	panic("implement me")
}

func runGRPC(store *wallet.Store, port string) {
	port = ":" + port
	srv, err := net.Listen("tcp", port)
	if err != nil {
		panic(err)
	}

	grpcSrv := grpc.NewServer()
	proto.RegisterWalletServiceServer(grpcSrv, &walletServer{store})
	err = grpcSrv.Serve(srv)
	if err != nil {
		panic(err)
	}
}
