package keeper_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtestutil "github.com/cosmos/cosmos-sdk/x/gov/testutil"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// TestChargeDeposit_NonCanonicalCommunityPoolDest is a regression test for
// MOCA-1263: ChargeDeposit must route cancellation charges to
// distrKeeper.FundCommunityPool whenever the configured ProposalCancelDest
// decodes to the distribution module account, regardless of which valid
// spelling (case, 0x-prefix) of that address was configured. It must never
// fall through to a raw bankKeeper.SendCoinsFromModuleToAccount for that
// case, since that path moves the module's bank balance without updating
// the community pool's tracked ledger (x/distribution's FeePool).
//
// This is a hand-rolled minimal harness rather than setupGovKeeper/
// trackMockBalances from common_test.go: ChargeDeposit only touches
// Deposits/bankKeeper/authKeeper/distrKeeper (not Params or Proposals), and
// a purpose-built gomock expectation (FundCommunityPool exactly once,
// SendCoinsFromModuleToAccount never) is the only way to distinguish the
// two routing paths — trackMockBalances's shared balance map credits the
// same canonical-string key either way, so it can't tell them apart.
func TestChargeDeposit_NonCanonicalCommunityPoolDest(t *testing.T) {
	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx
	encCfg := moduletestutil.MakeTestEncodingConfig()

	ctrl := gomock.NewController(t)
	acctKeeper := govtestutil.NewMockAccountKeeper(ctrl)
	bankKeeper := govtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := govtestutil.NewMockStakingKeeper(ctrl)
	distributionKeeper := govtestutil.NewMockDistributionKeeper(ctrl)

	acctKeeper.EXPECT().GetModuleAddress(types.ModuleName).Return(govAcct).AnyTimes()
	acctKeeper.EXPECT().GetModuleAddress(disttypes.ModuleName).Return(distAcct).AnyTimes()

	msr := baseapp.NewMsgServiceRouter()
	govKeeper := keeper.NewKeeper(encCfg.Codec, storeService, acctKeeper, bankKeeper, stakingKeeper, distributionKeeper, msr, types.DefaultConfig(), govAcct.String())

	_, _, depositor := testdata.KeyTestPubAddr()
	const proposalID = uint64(1)
	depositAmount := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1000)))

	require.NoError(t, govKeeper.Deposits.Set(ctx, collections.Join(proposalID, depositor), v1.Deposit{
		ProposalId: proposalID,
		Depositor:  depositor.String(),
		Amount:     depositAmount,
	}))

	// Non-canonical (all-lowercase) spelling of the distribution module
	// address. distAcct.String() always renders the one EIP-55 checksummed
	// form, so this differs in spelling while decoding to the same bytes —
	// exactly the kind of input a real ProposalCancelDest param can hold,
	// since param validation (x/gov/types/v1/params.go) only checks
	// decodability, not canonical form.
	nonCanonicalDest := strings.ToLower(distAcct.String())
	require.NotEqual(t, distAcct.String(), nonCanonicalDest, "sanity check: fixture must actually differ in spelling from the canonical form")

	// cancel ratio of 1 makes the entire deposit the cancellation charge, so
	// there's no "remaining amount" refund and thus no unrelated call to
	// SendCoinsFromModuleToAccount to account for.
	distributionKeeper.EXPECT().FundCommunityPool(gomock.Any(), depositAmount, gomock.Any()).Times(1).Return(nil)
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	bankKeeper.EXPECT().BurnCoins(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	err := govKeeper.ChargeDeposit(ctx, proposalID, nonCanonicalDest, "1")
	require.NoError(t, err)
}
