package types

import (
	"context"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

type StakingKeeper interface {
	GetLastValidators(ctx context.Context) (validators []types.Validator, err error)
	GetHistoricalInfo(ctx context.Context, height int64) (types.HistoricalInfo, error)
	BondDenom(ctx context.Context) (string, error)
}

type CrossChainKeeper interface {
	CreateRawIBCPackageWithFee(ctx context.Context, destChainId sdk.ChainID, channelID sdk.ChannelID,
		packageType sdk.CrossChainPackageType, packageLoad []byte, relayerFee, ackRelayerFee *big.Int,
	) (uint64, error)
	GetCrossChainApp(channelID sdk.ChannelID) sdk.CrossChainApplication
	GetSrcChainID() sdk.ChainID
	IsDestChainSupported(chainID sdk.ChainID) bool
	GetReceiveSequence(ctx context.Context, chainId sdk.ChainID, channelID sdk.ChannelID) uint64
	IncrReceiveSequence(ctx context.Context, chainId sdk.ChainID, channelID sdk.ChannelID)
	GetDestBscChainID() sdk.ChainID
	GetDestOpChainID() sdk.ChainID
	GetDestPolygonChainID() sdk.ChainID
	GetDestScrollChainID() sdk.ChainID
	GetDestLineaChainID() sdk.ChainID
	GetDestMantleChainID() sdk.ChainID
	GetDestArbitrumChainID() sdk.ChainID
	GetDestOptimismChainID() sdk.ChainID
	GetDestBaseChainID() sdk.ChainID
}

type BankKeeper interface {
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}
