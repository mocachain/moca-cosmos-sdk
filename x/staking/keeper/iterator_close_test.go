package keeper_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// The tests in this file drive each keeper method below through every way
// out of its iteration loop (running to the end, stopping early, and
// returning a decode error) and assert that the store iterator the method
// opened is closed on that path. Each test fails if the method's deferred
// Close() is removed.

// corruptRecord is not a valid protobuf encoding (a varint cut short), so the
// keeper fails to decode it. Staging it under a prefix a method iterates
// drives that method down its decode-error return.
var corruptRecord = []byte{0xff}

// testAccAddr returns a distinct 20-byte address for each n.
func testAccAddr(n byte) sdk.AccAddress {
	return sdk.AccAddress(bytes.Repeat([]byte{n}, 20))
}

func TestGetValidatorsClosesIterator(t *testing.T) {
	testCases := []struct {
		name        string
		corrupt     bool
		maxRetrieve uint32
		expLen      int
	}{
		{name: "returns every validator", maxRetrieve: 10, expLen: 2},
		{name: "stops at maxRetrieve", maxRetrieve: 1, expLen: 1},
		{name: "fails to decode a validator", corrupt: true, maxRetrieve: 10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f := newIteratorSpyFixture(t)
			if tc.corrupt {
				require.NoError(t, f.store.Set(stakingtypes.GetValidatorKey(testAccAddr(1)), corruptRecord))
			} else {
				for _, pk := range simtestutil.CreateTestPubKeys(2) {
					validator := stakingtestutil.NewValidator(t, sdk.AccAddress(pk.Address()), pk)
					require.NoError(t, f.keeper.SetValidator(f.ctx, validator))
				}
			}

			requireIteratorsClosed(t, f.tracker, func() {
				validators, err := f.keeper.GetValidators(f.ctx, tc.maxRetrieve)
				if tc.corrupt {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
				require.Len(t, validators, tc.expLen)
			})
		})
	}
}

func TestIterateLastValidatorPowersClosesIterator(t *testing.T) {
	testCases := []struct {
		name    string
		corrupt bool
		stop    bool
		expSeen int
	}{
		{name: "visits every power", expSeen: 2},
		{name: "stops when the handler asks to", stop: true, expSeen: 1},
		{name: "fails to decode a power", corrupt: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f := newIteratorSpyFixture(t)
			if tc.corrupt {
				require.NoError(t, f.store.Set(stakingtypes.GetLastValidatorPowerKey(testAccAddr(1)), corruptRecord))
			} else {
				require.NoError(t, f.keeper.SetLastValidatorPower(f.ctx, testAccAddr(1), 10))
				require.NoError(t, f.keeper.SetLastValidatorPower(f.ctx, testAccAddr(2), 20))
			}

			requireIteratorsClosed(t, f.tracker, func() {
				seen := 0
				err := f.keeper.IterateLastValidatorPowers(f.ctx, func(sdk.AccAddress, int64) bool {
					seen++
					return tc.stop
				})
				if tc.corrupt {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
				require.Equal(t, tc.expSeen, seen)
			})
		})
	}
}

// stageRedelegations stores one redelegation from del to each of dsts, all
// from the same source validator, or a single corrupt record in their place.
func stageRedelegations(t *testing.T, f iteratorSpyFixture, del sdk.AccAddress, corrupt bool, dsts ...sdk.AccAddress) {
	t.Helper()
	src := testAccAddr(0xa)
	if corrupt {
		require.NoError(t, f.store.Set(stakingtypes.GetREDKey(del, src, dsts[0]), corruptRecord))
		return
	}
	for _, dst := range dsts {
		red := stakingtypes.NewRedelegation(del, src, dst, 0, time.Unix(0, 0).UTC(), math.NewInt(5), math.LegacyNewDec(5), 0)
		require.NoError(t, f.keeper.SetRedelegation(f.ctx, red))
	}
}

func TestIterateDelegatorRedelegationsClosesIterator(t *testing.T) {
	testCases := []struct {
		name    string
		corrupt bool
		stop    bool
		expSeen int
	}{
		{name: "visits every redelegation", expSeen: 2},
		{name: "stops when the callback asks to", stop: true, expSeen: 1},
		{name: "fails to decode a redelegation", corrupt: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f := newIteratorSpyFixture(t)
			del := testAccAddr(1)
			stageRedelegations(t, f, del, tc.corrupt, testAccAddr(0xb), testAccAddr(0xc))

			requireIteratorsClosed(t, f.tracker, func() {
				seen := 0
				err := f.keeper.IterateDelegatorRedelegations(f.ctx, del, func(stakingtypes.Redelegation) bool {
					seen++
					return tc.stop
				})
				if tc.corrupt {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
				require.Equal(t, tc.expSeen, seen)
			})
		})
	}
}

func TestGetRedelegationsClosesIterator(t *testing.T) {
	testCases := []struct {
		name        string
		corrupt     bool
		maxRetrieve uint16
		expLen      int
	}{
		{name: "returns every redelegation", maxRetrieve: 10, expLen: 2},
		{name: "stops at maxRetrieve", maxRetrieve: 1, expLen: 1},
		{name: "fails to decode a redelegation", corrupt: true, maxRetrieve: 10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f := newIteratorSpyFixture(t)
			del := testAccAddr(1)
			stageRedelegations(t, f, del, tc.corrupt, testAccAddr(0xb), testAccAddr(0xc))

			requireIteratorsClosed(t, f.tracker, func() {
				reds, err := f.keeper.GetRedelegations(f.ctx, del, tc.maxRetrieve)
				if tc.corrupt {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
				require.Len(t, reds, tc.expLen)
			})
		})
	}
}
