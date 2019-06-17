package cli

import (
	"wago/pkg/wallet"
)

func SendTransactionCommand(store wallet.WalletStore, ctx commandContext, create bool) {
	trans := wallet.NewTransaction(ctx.SrcPublic, ctx.DstPublic, ctx.Currency, wallet.DecimalAmount{Value: ctx.Value, Decimal: ctx.Decimal}, create)
	trans.Sign(ctx.SrcPrivate)

	err := store.Propose(trans)
	if err != nil {
		panic(err)
	}
}
