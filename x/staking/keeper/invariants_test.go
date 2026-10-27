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

// iteratorTracker counts the store iterators opened through a spyKVStore and
// how many of them have been closed.
type iteratorTracker struct {
	opened, closed int
}

// spyIterator wraps a real store iterator and reports to its tracker the
// first time it is closed.
type spyIterator struct {
	corestore.Iterator
	tracker *iteratorTracker
	closed  bool
}

func (it *spyIterator) Close() error {
	if !it.closed {
		it.closed = true
		it.tracker.closed++
	}
	return it.Iterator.Close()
}

// spyKVStore wraps a real KVStore and reports every iterator it opens, in
// either direction, to tracker, so a test can tell whether each was closed.
type spyKVStore struct {
	corestore.KVStore
	tracker *iteratorTracker
}

func (s *spyKVStore) Iterator(start, end []byte) (corestore.Iterator, error) {
	return s.track(s.KVStore.Iterator(start, end))
}

func (s *spyKVStore) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	return s.track(s.KVStore.ReverseIterator(start, end))
}

func (s *spyKVStore) track(it corestore.Iterator, err error) (corestore.Iterator, error) {
	if err != nil {
		return nil, err
	}
	s.tracker.opened++
	return &spyIterator{Iterator: it, tracker: s.tracker}, nil
}

// spyKVStoreService wraps a real KVStoreService so every KVStore it opens
// is a spyKVStore.
type spyKVStoreService struct {
	corestore.KVStoreService
	tracker *iteratorTracker
}

func (s *spyKVStoreService) OpenKVStore(ctx context.Context) corestore.KVStore {
	return &spyKVStore{KVStore: s.KVStoreService.OpenKVStore(ctx), tracker: s.tracker}
}

// iteratorSpyFixture is a staking keeper over a fresh store whose iterators
// are all reported to tracker.
type iteratorSpyFixture struct {
	ctx     sdk.Context
	keeper  *stakingkeeper.Keeper
	store   corestore.KVStore // an untracked handle on the keeper's store, for staging raw records
	tracker *iteratorTracker
}

func newIteratorSpyFixture(t *testing.T) iteratorSpyFixture {
	t.Helper()
	key := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
	tracker := &iteratorTracker{}
	storeService := &spyKVStoreService{KVStoreService: runtime.NewKVStoreService(key), tracker: tracker}
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

	return iteratorSpyFixture{ctx: ctx, keeper: keeper, store: runtime.NewKVStoreService(key).OpenKVStore(ctx), tracker: tracker}
}

// requireIteratorsClosed runs call and asserts that it opened at least one
// store iterator and closed every iterator it opened. Iterators opened while
// staging state before call are not counted.
func requireIteratorsClosed(t *testing.T, tracker *iteratorTracker, call func()) {
	t.Helper()
	*tracker = iteratorTracker{}
	call()
	require.Positive(t, tracker.opened, "the call must open a store iterator")
	require.Equal(t, tracker.opened, tracker.closed, "every store iterator the call opened must be closed")
}

// Minimal hand-written fakes for the keeper's three dependency interfaces.
// None of the keeper methods the iterator-closure tests exercise call into
// them, so every method is a stub; only GetModuleAddress must return a
// non-nil address, since NewKeeper checks that at construction.

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
func (fakeBankKeeper) GetSupply(context.Context, string) sdk.Coin               { return sdk.Coin{} }

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
	f := newIteratorSpyFixture(t)

	// Stage a dangling power-index entry: a validator indexed by power but
	// with no corresponding primary record, which is exactly what makes
	// GetValidator error and NonNegativePowerInvariant panic.
	valPubKey := simtestutil.CreateTestPubKeys(1)[0]
	valAddr := sdk.AccAddress(valPubKey.Address().Bytes())
	validator := stakingtestutil.NewValidator(t, valAddr, valPubKey)
	require.NoError(t, f.keeper.SetValidatorByPowerIndex(f.ctx, validator))
	// Deliberately do NOT call SetValidator, so the primary record is
	// missing for the address the power index just wrote.

	invariant := stakingkeeper.NonNegativePowerInvariant(f.keeper)

	// The power store iterator must be closed even though the invariant panics.
	requireIteratorsClosed(t, f.tracker, func() {
		require.Panics(t, func() { invariant(f.ctx) }, "a dangling power-index entry must still panic")
	})
}
