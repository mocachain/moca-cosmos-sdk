package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type BankKeeper interface {
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
}

type StakingKeeper interface {
	BondDenom(ctx context.Context) (string, error)
}
