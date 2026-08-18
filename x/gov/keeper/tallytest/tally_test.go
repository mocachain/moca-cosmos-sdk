// Package tallytest holds a gov Tally regression test that must live outside
// x/gov/keeper: that package's own test binary does not compile on this branch
// for pre-existing reasons unrelated to the code under test here (bech32-oriented
// testutil fixtures and mock drift against moca's hex address decoder), which
// would prevent any test added there from running at all. This package drives the
// real gov keeper's exported Tally through a minimal hand-written harness whose
// only meaningful collaborator is a fake staking keeper.
package tallytest

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TestTally_NonCanonicalValidatorOperatorAddress is a regression test for
// MOCA-1362 (tally read side): Tally keys its validator-power map by the
// canonical EIP-55 rendering of each validator's address and looks delegations
// up by the same canonical rendering. Before the fix the map was keyed by the
// raw, as-stored OperatorAddress string, so a validator whose operator address
// was stored in a non-canonical (e.g. all-lowercase) spelling was keyed under a
// string the canonical delegation lookup never matched — dropping its entire
// bonded voting power out of the tally with no error.
//
// The test stages exactly that: a bonded validator whose stored OperatorAddress
// is the lowercase, prefix-less spelling of an address that renders with hex
// letters, with a full self-delegation whose ValidatorAddress is the canonical
// rendering (as any lookup that re-derives the canonical form produces). It
// then has that validator self-vote Yes and asserts the vote is counted and the
// proposal passes. Against the pre-fix code the map key and the lookup key
// disagree, the power is dropped, and the proposal fails quorum.
func TestTally_NonCanonicalValidatorOperatorAddress(t *testing.T) {
	key := storetypes.NewKVStoreKey(govtypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx
	encCfg := moduletestutil.MakeTestEncodingConfig()

	// 0xAB bytes render with hex letters, so the canonical (mixed-case) form and
	// the lowercase spelling genuinely differ as strings.
	opAcc := sdk.AccAddress(bytes.Repeat([]byte{0xAB}, 20))
	canonical := opAcc.String()
	nonCanonical := strings.ToLower(strings.TrimPrefix(canonical, "0x"))
	require.NotEqual(t, canonical, nonCanonical, "sanity: the two spellings must differ")

	sk := fakeStakingKeeper{
		validator: stakingtypes.Validator{
			OperatorAddress: nonCanonical, // stored non-canonically — the bug's premise
			Status:          stakingtypes.Bonded,
			Tokens:          math.NewInt(1000),
			DelegatorShares: math.LegacyNewDec(1000),
		},
		delegation: stakingtypes.Delegation{
			DelegatorAddress: canonical,
			ValidatorAddress: canonical, // canonical, as a re-derived lookup would render it
			Shares:           math.LegacyNewDec(1000),
		},
		voter:       opAcc,
		totalBonded: math.NewInt(1000),
	}

	k := govkeeper.NewKeeper(
		encCfg.Codec, storeService,
		fakeAccountKeeper{}, nil, sk, nil, nil, // bank / crosschain / distr keepers are never touched by Tally
		nil, govtypes.DefaultConfig(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	require.NoError(t, k.Params.Set(ctx, v1.DefaultParams()))

	// The validator self-votes Yes.
	proposalID := uint64(1)
	require.NoError(t, k.Votes.Set(ctx, collections.Join(proposalID, opAcc), v1.Vote{
		ProposalId: proposalID,
		Voter:      canonical,
		Options:    []*v1.WeightedVoteOption{{Option: v1.OptionYes, Weight: "1"}},
	}))

	passes, _, results, err := k.Tally(ctx, v1.Proposal{Id: proposalID})
	require.NoError(t, err)

	require.Equal(t, "1000", results.YesCount,
		"the validator's bonded power must be tallied despite its non-canonical stored operator address")
	require.True(t, passes, "proposal should pass once the non-canonical validator's Yes vote is counted")
}

// fakeStakingKeeper yields a single bonded validator and, for the matching
// voter, a single self-delegation. It implements exactly the 3-method gov
// StakingKeeper interface that Tally depends on.
type fakeStakingKeeper struct {
	validator   stakingtypes.Validator
	delegation  stakingtypes.Delegation
	voter       sdk.AccAddress
	totalBonded math.Int
}

func (f fakeStakingKeeper) IterateBondedValidatorsByPower(_ context.Context, fn func(index int64, validator stakingtypes.ValidatorI) (stop bool)) error {
	fn(0, f.validator)
	return nil
}

func (f fakeStakingKeeper) IterateDelegations(_ context.Context, delegator sdk.AccAddress, fn func(index int64, delegation stakingtypes.DelegationI) (stop bool)) error {
	if delegator.Equals(f.voter) {
		fn(0, f.delegation)
	}
	return nil
}

func (f fakeStakingKeeper) TotalBondedTokens(context.Context) (math.Int, error) {
	return f.totalBonded, nil
}

// fakeAccountKeeper satisfies the gov AccountKeeper interface; only
// GetModuleAddress is exercised (by NewKeeper, to confirm the gov module
// account is set).
type fakeAccountKeeper struct{}

func (fakeAccountKeeper) GetAccount(context.Context, sdk.AccAddress) sdk.AccountI {
	panic("unexpected call")
}
func (fakeAccountKeeper) GetModuleAddress(name string) sdk.AccAddress {
	return authtypes.NewModuleAddress(name)
}
func (fakeAccountKeeper) GetModuleAccount(context.Context, string) sdk.ModuleAccountI {
	panic("unexpected call")
}
func (fakeAccountKeeper) SetModuleAccount(context.Context, sdk.ModuleAccountI) {
	panic("unexpected call")
}
