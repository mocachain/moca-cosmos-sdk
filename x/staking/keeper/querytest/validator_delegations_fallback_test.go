// Package querytest holds staking query tests that must live outside
// x/staking/keeper: that package's own test binary currently fails to compile
// on main for pre-existing reasons unrelated to the code under test here
// (stale golang/mock-generated testutil mocks and ValAddress/AccAddress drift
// in its older test files), which would prevent any test added there — internal
// or external — from running at all. This package exercises the keeper strictly
// through its exported API.
package querytest

import (
	"bytes"
	"context"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TestValidatorDelegationsLegacyFallback_NonCanonicalValidatorAddr is a
// regression test for the legacy fallback inside the ValidatorDelegations
// query: its per-record filter compared the stored delegation's validator
// address against the request's ValidatorAddr with strings.EqualFold, which
// normalizes case but not "0x"-prefix presence, silently dropping matching
// delegations for a validly-spelled query.
//
// The fallback only runs when the primary, index-based scan errors, so the
// test stages exactly that: a dangling by-validator index entry pointing at a
// record holding non-protobuf bytes (the index/record drift this fallback
// exists to paper over). Key geometry makes the two scans diverge — the
// by-validator index orders raw delegator keys, so the corrupt pointer's
// 32-byte zero key sorts first and the primary fails immediately, while the
// legacy scan orders length-prefixed keys, so the same corrupt record's 0x20
// length byte sorts after every real 20-byte (0x14) delegation. Key-based
// pagination with Limit=1 then stops the legacy scan after the real record,
// before it would reach the corrupt one (unlike offset pagination, which
// decodes every record up to the limit check).
func TestValidatorDelegationsLegacyFallback_NonCanonicalValidatorAddr(t *testing.T) {
	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx
	encCfg := moduletestutil.MakeTestEncodingConfig()

	k := stakingkeeper.NewKeeper(
		encCfg.Codec, storeService,
		fakeAccountKeeper{}, fakeAuthzKeeper{}, fakeBankKeeper{},
		authtypes.NewModuleAddress("gov").String(),
		hexCodec{}, hexCodec{},
	)
	require.NoError(t, k.SetParams(ctx, types.DefaultParams()))

	valAcc := sdk.AccAddress(bytes.Repeat([]byte{0xAA}, 20))
	delAcc := sdk.AccAddress(bytes.Repeat([]byte{0xBB}, 20))

	require.NoError(t, k.SetValidator(ctx, types.Validator{
		OperatorAddress: valAcc.String(),
		Status:          types.Bonded,
		Tokens:          math.NewInt(1000),
		DelegatorShares: math.LegacyNewDec(1000),
	}))

	store := testCtx.Ctx.KVStore(key)

	// The real delegation, stored under the validator's canonical rendering.
	del := types.Delegation{
		DelegatorAddress: delAcc.String(),
		ValidatorAddress: valAcc.String(),
		Shares:           math.LegacyNewDec(1000),
	}
	store.Set(types.GetDelegationKey(delAcc, valAcc), encCfg.Codec.MustMarshal(&del))

	// The corrupt pointer that forces the primary scan to error: an index
	// entry whose 32-byte delegator key leads to non-protobuf bytes.
	badDel := sdk.AccAddress(bytes.Repeat([]byte{0x00}, 32))
	store.Set(append(types.GetDelegationsByValPrefixKey(valAcc), badDel...), []byte{0x1})
	store.Set(types.GetDelegationKey(badDel, valAcc), []byte{0xFF})

	nonCanonical := strings.ToLower(strings.TrimPrefix(valAcc.String(), "0x"))
	require.NotEqual(t, valAcc.String(), nonCanonical, "sanity check: fixture must actually differ in spelling from the canonical form")

	resp, err := stakingkeeper.NewQuerier(k).ValidatorDelegations(ctx, &types.QueryValidatorDelegationsRequest{
		ValidatorAddr: nonCanonical,
		Pagination:    &query.PageRequest{Key: []byte{0x00}, Limit: 1},
	})
	require.NoError(t, err)
	require.Len(t, resp.DelegationResponses, 1, "the legacy fallback must return the matching delegation for a non-canonical validator address spelling")
	require.Equal(t, delAcc.String(), resp.DelegationResponses[0].Delegation.DelegatorAddress)
	require.Equal(t, valAcc.String(), resp.DelegationResponses[0].Delegation.ValidatorAddress)
}

type hexCodec struct{}

func (hexCodec) StringToBytes(text string) ([]byte, error) {
	return hex.DecodeString(strings.TrimPrefix(text, "0x"))
}

func (hexCodec) BytesToString(bz []byte) (string, error) {
	return "0x" + hex.EncodeToString(bz), nil
}

type fakeAccountKeeper struct{}

func (fakeAccountKeeper) IterateAccounts(context.Context, func(sdk.AccountI) (stop bool)) {
	panic("unexpected call")
}

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

type fakeAuthzKeeper struct{}

func (fakeAuthzKeeper) GetGrant(context.Context, sdk.AccAddress, sdk.AccAddress, string) (authz.Grant, bool) {
	panic("unexpected call")
}

func (fakeAuthzKeeper) Update(context.Context, sdk.AccAddress, sdk.AccAddress, authz.Authorization) error {
	panic("unexpected call")
}

func (fakeAuthzKeeper) DeleteGrant(context.Context, sdk.AccAddress, sdk.AccAddress, string) error {
	panic("unexpected call")
}

type fakeBankKeeper struct{}

func (fakeBankKeeper) GetAllBalances(context.Context, sdk.AccAddress) sdk.Coins {
	panic("unexpected call")
}
func (fakeBankKeeper) GetBalance(context.Context, sdk.AccAddress, string) sdk.Coin {
	panic("unexpected call")
}
func (fakeBankKeeper) LockedCoins(context.Context, sdk.AccAddress) sdk.Coins {
	panic("unexpected call")
}
func (fakeBankKeeper) SpendableCoins(context.Context, sdk.AccAddress) sdk.Coins {
	panic("unexpected call")
}
func (fakeBankKeeper) GetSupply(context.Context, string) sdk.Coin { panic("unexpected call") }
func (fakeBankKeeper) SendCoinsFromModuleToModule(context.Context, string, string, sdk.Coins) error {
	panic("unexpected call")
}
func (fakeBankKeeper) UndelegateCoinsFromModuleToAccount(context.Context, string, sdk.AccAddress, sdk.Coins) error {
	panic("unexpected call")
}
func (fakeBankKeeper) DelegateCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error {
	panic("unexpected call")
}
func (fakeBankKeeper) BurnCoins(context.Context, string, sdk.Coins) error { panic("unexpected call") }
