// Package createvaltest holds a CreateValidator regression test that must live
// outside x/staking/keeper: that package's own test binary currently fails to
// compile on main for pre-existing reasons unrelated to the code under test
// here (stale golang/mock-generated testutil mocks and NewKeeper /
// NewMsgCreateValidator signature drift in its older upstream-derived test
// files), which would prevent any test added there — internal or external —
// from running at all. This package drives the real keeper's exported
// CreateValidator through a minimal hand-written harness.
package createvaltest

import (
	"bytes"
	"context"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/0xPolygon/polygon-edge/bls"
	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/cometbft/cometbft/votepool"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TestCreateValidator_StoresCanonicalOperatorAddress is a regression test for
// MOCA-1362 (registration / write side): CreateValidator must persist the
// validator's OperatorAddress in canonical EIP-55 checksummed form, so that a
// later lookup that re-derives the canonical string finds it. Gov tally, for
// example, keys its validator-power map by sdk.AccAddress(bz).String(); before
// this fix the raw, as-submitted msg.ValidatorAddress was stored verbatim, so a
// validator registered with a non-canonical (e.g. all-lowercase, prefix-less)
// operator address would be keyed under a spelling the tally never looks up and
// its entire bonded voting power would silently drop out of every tally.
//
// The test registers a validator whose ValidatorAddress is the all-lowercase,
// "0x"-less spelling of an address that renders with hex letters (so the
// canonical and submitted spellings genuinely differ) and asserts the stored
// OperatorAddress is the canonical rendering, not the submitted one.
func TestCreateValidator_StoresCanonicalOperatorAddress(t *testing.T) {
	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx // BlockHeight 0 selects the genesis signer path (signer == self-delegator)
	encCfg := moduletestutil.MakeTestEncodingConfig()

	k := stakingkeeper.NewKeeper(
		encCfg.Codec, storeService,
		fakeAccountKeeper{}, fakeAuthzKeeper{}, fakeBankKeeper{},
		authtypes.NewModuleAddress("gov").String(),
		hexCodec{}, hexCodec{},
	)
	require.NoError(t, k.SetParams(ctx, types.DefaultParams()))
	msgServer := stakingkeeper.NewMsgServerImpl(k)

	// One address, reused as operator + self-delegator + signer + relayer +
	// challenger (as the simulation factory does). 0xAB bytes render with hex
	// letters, so the canonical mixed-case form and the lowercase spelling differ.
	opAcc := sdk.AccAddress(bytes.Repeat([]byte{0xAB}, 20))
	canonical := opAcc.String()
	nonCanonical := strings.ToLower(strings.TrimPrefix(canonical, "0x"))
	require.NotEqual(t, canonical, nonCanonical, "sanity: the two spellings must differ")

	// A valid BLS key + proof-of-possession: the proof signs tmhash.Sum(pubkey)
	// under the votepool DST, exactly as gentx/simulation build it. CreateValidator
	// verifies this via CheckBlsProof before it stores the validator.
	blsSecretKey, err := bls.GenerateBlsKey()
	require.NoError(t, err)
	blsPk := hex.EncodeToString(blsSecretKey.PublicKey().Marshal())
	blsProofBuf, err := blsSecretKey.Sign(tmhash.Sum(blsSecretKey.PublicKey().Marshal()), votepool.DST)
	require.NoError(t, err)
	blsProofBts, err := blsProofBuf.Marshal()
	require.NoError(t, err)
	blsProof := hex.EncodeToString(blsProofBts)

	consPub := ed25519.GenPrivKey().PubKey()

	msg, err := types.NewMsgCreateValidator(
		nonCanonical, consPub,
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 10),
		types.Description{Moniker: "val"},
		types.CommissionRates{
			Rate:          math.LegacyNewDecWithPrec(1, 1),
			MaxRate:       math.LegacyNewDecWithPrec(2, 1),
			MaxChangeRate: math.LegacyNewDecWithPrec(1, 1),
		},
		math.OneInt(),              // minSelfDelegation
		opAcc, opAcc, opAcc, opAcc, // from, selfDelegator, relayer, challenger
		blsPk, blsProof,
	)
	require.NoError(t, err)
	require.Equal(t, nonCanonical, msg.ValidatorAddress, "guard: msg must carry the non-canonical spelling")

	_, err = msgServer.CreateValidator(ctx, msg)
	require.NoError(t, err)

	valAddr, err := sdk.AccAddressFromHexUnsafe(nonCanonical)
	require.NoError(t, err)
	stored, err := k.GetValidator(ctx, valAddr)
	require.NoError(t, err)

	require.Equal(t, canonical, stored.OperatorAddress,
		"CreateValidator must persist the canonical operator address, not the raw submitted spelling")
	require.NotEqual(t, nonCanonical, stored.OperatorAddress)
}

// --- minimal fakes: the keeper's expected keepers, none of which this path
// exercises except the bank keeper's DelegateCoinsFromAccountToModule (invoked
// by the self-delegation), which is a no-op here. ---

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
	return nil
}
func (fakeBankKeeper) BurnCoins(context.Context, string, sdk.Coins) error { panic("unexpected call") }
