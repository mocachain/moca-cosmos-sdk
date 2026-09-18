package keeper_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	corestore "cosmossdk.io/core/store"
	storetypes "cosmossdk.io/store/types"

	authcodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// spyIterator wraps a real store iterator and records whether Close was
// called on it.
type spyIterator struct {
	corestore.Iterator
	closed *bool
}

func (it *spyIterator) Close() error {
	*it.closed = true
	return it.Iterator.Close()
}

// spyKVStore wraps a real KVStore and reports every ReverseIterator it
// opens through closed, so a test can tell whether the iterator was closed.
type spyKVStore struct {
	corestore.KVStore
	closed *bool
}

func (s *spyKVStore) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	it, err := s.KVStore.ReverseIterator(start, end)
	if err != nil {
		return nil, err
	}
	*s.closed = false
	return &spyIterator{Iterator: it, closed: s.closed}, nil
}

// spyKVStoreService wraps a real KVStoreService so every KVStore it opens
// is a spyKVStore.
type spyKVStoreService struct {
	corestore.KVStoreService
	closed *bool
}

func (s *spyKVStoreService) OpenKVStore(ctx context.Context) corestore.KVStore {
	return &spyKVStore{KVStore: s.KVStoreService.OpenKVStore(ctx), closed: s.closed}
}

// Minimal hand-written fakes for the keeper's three dependency interfaces.
// NonNegativePowerInvariant never calls into any of them, so every method
// is a stub; only GetModuleAddress must return a non-nil address, since
// NewKeeper checks that at construction.

type fakeAccountKeeper struct{}

func (fakeAccountKeeper) IterateAccounts(context.Context, func(sdk.AccountI) bool) {}
func (fakeAccountKeeper) GetAccount(context.Context, sdk.AccAddress) sdk.AccountI  { return nil }
func (fakeAccountKeeper) GetModuleAddress(name string) sdk.AccAddress {
	return authtypes.NewModuleAddress(name)
}
func (fakeAccountKeeper) GetModuleAccount(context.Context, string) sdk.ModuleAccountI { return nil }
func (fakeAccountKeeper) SetModuleAccount(context.Context, sdk.ModuleAccountI)        {}

type fakeAuthzKeeper struct{}

func (fakeAuthzKeeper) GetGrant(context.Context, sdk.AccAddress, sdk.AccAddress, string) (authz.Grant, bool) {
	return authz.Grant{}, false
}

func (fakeAuthzKeeper) Update(context.Context, sdk.AccAddress, sdk.AccAddress, authz.Authorization) error {
	return nil
}

func (fakeAuthzKeeper) DeleteGrant(context.Context, sdk.AccAddress, sdk.AccAddress, string) error {
	return nil
}

type fakeBankKeeper struct{}

func (fakeBankKeeper) GetAllBalances(context.Context, sdk.AccAddress) sdk.Coins { return nil }
func (fakeBankKeeper) GetBalance(context.Context, sdk.AccAddress, string) sdk.Coin {
	return sdk.Coin{}
}
func (fakeBankKeeper) LockedCoins(context.Context, sdk.AccAddress) sdk.Coins    { return nil }
func (fakeBankKeeper) SpendableCoins(context.Context, sdk.AccAddress) sdk.Coins { return nil }
func (fakeBankKeeper) GetSupply(context.Context, string) sdk.Coin              { return sdk.Coin{} }
func (fakeBankKeeper) SendCoinsFromModuleToModule(context.Context, string, string, sdk.Coins) error {
	return nil
}

func (fakeBankKeeper) UndelegateCoinsFromModuleToAccount(context.Context, string, sdk.AccAddress, sdk.Coins) error {
	return nil
}

func (fakeBankKeeper) DelegateCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error {
	return nil
}

func (fakeBankKeeper) BurnCoins(context.Context, string, sdk.Coins) error { return nil }

// TestNonNegativePowerInvariantClosesIteratorOnPanic proves that the power
// store iterator opened by NonNegativePowerInvariant is closed even when the
// invariant panics partway through (a dangling power-index entry whose
// validator record is missing). On the pre-fix code the iterator is left
// open when this happens because iterator.Close() sits after the loop,
// unreached by the panic.
func TestNonNegativePowerInvariantClosesIteratorOnPanic(t *testing.T) {
	key := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
	iteratorClosed := new(bool)
	storeService := &spyKVStoreService{KVStoreService: runtime.NewKVStoreService(key), closed: iteratorClosed}
	testCtx := sdktestutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx
	encCfg := moduletestutil.MakeTestEncodingConfig()

	keeper := stakingkeeper.NewKeeper(
		encCfg.Codec,
		storeService,
		fakeAccountKeeper{},
		fakeAuthzKeeper{},
		fakeBankKeeper{},
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		authcodec.NewBech32Codec(sdk.Bech32PrefixValAddr),
		authcodec.NewBech32Codec(sdk.Bech32PrefixConsAddr),
	)
	require.NoError(t, keeper.SetParams(ctx, stakingtypes.DefaultParams()))

	// Stage a dangling power-index entry: a validator indexed by power but
	// with no corresponding primary record, which is exactly what makes
	// GetValidator error and NonNegativePowerInvariant panic.
	valPubKey := simtestutil.CreateTestPubKeys(1)[0]
	valAddr := sdk.AccAddress(valPubKey.Address().Bytes())
	validator := stakingtestutil.NewValidator(t, valAddr, valPubKey)
	require.NoError(t, keeper.SetValidatorByPowerIndex(ctx, validator))
	// Deliberately do NOT call SetValidator, so the primary record is
	// missing for the address the power index just wrote.

	invariant := stakingkeeper.NonNegativePowerInvariant(keeper)

	require.Panics(t, func() { invariant(ctx) }, "a dangling power-index entry must still panic")
	require.True(t, *iteratorClosed, "the power store iterator must be closed even though the invariant panicked")
}
