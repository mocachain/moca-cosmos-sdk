// This file only used to generate mocks

package testutil

import (
	context "context"
	"math/big"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// AccountKeeper extends gov's actual expected AccountKeeper with additional
// methods used in tests.
type AccountKeeper interface {
	types.AccountKeeper

	IterateAccounts(ctx context.Context, cb func(account sdk.AccountI) (stop bool))
}

// BankKeeper extends gov's actual expected BankKeeper with additional
// methods used in tests.
type BankKeeper interface {
	bankkeeper.Keeper
}

// StakingKeeper extends gov's actual expected StakingKeeper with additional
// methods used in tests.
type StakingKeeper interface {
	types.StakingKeeper

	BondDenom(ctx context.Context) (string, error)
	TokensFromConsensusPower(ctx context.Context, power int64) math.Int
}

// CrossChainKeeper defines the expected crossChain keeper
type CrossChainKeeper interface {
	GetDestBscChainID() sdk.ChainID
	CreateRawIBCPackageWithFee(ctx context.Context, destChainId sdk.ChainID, channelID sdk.ChannelID, packageType sdk.CrossChainPackageType,
		packageLoad []byte, relayerFee, ackRelayerFee *big.Int,
	) (uint64, error)

	RegisterChannel(name string, id sdk.ChannelID, app sdk.CrossChainApplication) error

	GetSendSequence(ctx context.Context, destChainId sdk.ChainID, channelID sdk.ChannelID) uint64

	GetReceiveSequence(ctx context.Context, destChainId sdk.ChainID, channelID sdk.ChannelID) uint64

	IsDestChainSupported(chainID sdk.ChainID) bool

	GetDestOpChainID() sdk.ChainID
	GetDestPolygonChainID() sdk.ChainID
	GetDestScrollChainID() sdk.ChainID
	GetDestLineaChainID() sdk.ChainID
	GetDestMantleChainID() sdk.ChainID
	GetDestArbitrumChainID() sdk.ChainID
	GetDestOptimismChainID() sdk.ChainID
	GetDestBaseChainID() sdk.ChainID
}

// DistributionKeeper defines the expected distribution keeper
type DistributionKeeper interface {
	FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error
}
