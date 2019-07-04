package main

import (
	"context"
	"encoding/hex"
	"net"

	"github.com/CodethinkLabs/wago/pkg/proto"
	"github.com/CodethinkLabs/wago/pkg/wallet"
	"golang.org/x/crypto/ed25519"
	"google.golang.org/grpc"
)

type walletServer struct {
	store *wallet.Store
}

func (s walletServer) CreateCurrency(c *proto.Create, r proto.WalletService_CreateCurrencyServer) error {
	trans, err := wallet.NewTransaction(
		nil, c.Update.Dest,
		wallet.Currency(c.Update.Currency),
		wallet.DecimalAmount{Value: c.Update.Amount.Value, Decimal: int8(c.Update.Amount.Decimal)},
		true,
	)

	if err != nil {
		err = r.Send(&proto.TransactionUpdate{Status: proto.TransactionUpdate_INVALIDATED, Message: err.Error()})
		return err
	}

	err = s.store.Propose(trans)
	if err != nil {
		err = r.Send(&proto.TransactionUpdate{Status: proto.TransactionUpdate_INVALIDATED, Message: err.Error()})
		return err
	}

	// todo(arlyon) provide accurate updates
	return r.Send(&proto.TransactionUpdate{Status: proto.TransactionUpdate_COMMITTED})
}

func (s walletServer) SubmitTransaction(t *proto.Transaction, r proto.WalletService_SubmitTransactionServer) error {
	trans, err := wallet.NewTransaction(
		t.Update.Src, t.Update.Dest,
		wallet.Currency(t.Update.Currency),
		wallet.DecimalAmount{Value: t.Update.Amount.Value, Decimal: int8(t.Update.Amount.Decimal)},
		false,
	)
	if err != nil {
		return err
	}

	copy(trans.Sig[:], t.Sig)

	err = s.store.Propose(trans)
	if err != nil {
		return err
	}

	// todo(arlyon) provide accurate updates
	err = r.Send(&proto.TransactionUpdate{Status: proto.TransactionUpdate_COMMITTED})
	if err != nil {
		return err
	}

	return nil
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
