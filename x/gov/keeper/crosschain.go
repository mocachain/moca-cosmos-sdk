package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func (k Keeper) RegisterCrossChainSyncParamsApp() error {
	return nil
}

func (k Keeper) SyncParams(_ sdk.Context, _ sdk.ChainID, cpc govv1.CrossChainParamsChange) error {
	if err := cpc.ValidateBasic(); err != nil {
		return err
	}
	return types.ErrCrossChainDisabled
}
