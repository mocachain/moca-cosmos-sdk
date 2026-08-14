package keeper

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	coreaddress "cosmossdk.io/core/address"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/internal/math"
)

func TestDoTallyAndUpdate(t *testing.T) {
	var (
		myAddr      = sdk.AccAddress(bytes.Repeat([]byte{0x01}, 20))
		myOtherAddr = sdk.AccAddress(bytes.Repeat([]byte{0x02}, 20))
	)
	encCfg := moduletestutil.MakeTestEncodingConfig()
	group.RegisterInterfaces(encCfg.InterfaceRegistry)

	storeKey := storetypes.NewKVStoreKey(group.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, storeKey, storetypes.NewTransientStoreKey("transient_test"))
	myAccountKeeper := &mockAccountKeeper{
		AddressCodecFn: func() coreaddress.Codec {
			return address.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
		},
	}
	groupKeeper := NewKeeper(storeKey, encCfg.Codec, nil, myAccountKeeper, group.DefaultConfig())
	noEventsFn := func(proposalID uint64) sdk.Events { return sdk.Events{} }
	type memberVote struct {
		address string
		weight  string
		option  group.VoteOption
	}
	specs := map[string]struct {
		votes           []memberVote
		policy          group.DecisionPolicy
		expStatus       group.ProposalStatus
		expVotesCleared bool
		expEvents       func(proposalID uint64) sdk.Events
	}{
		"proposal accepted": {
			votes: []memberVote{
				{address: myAddr.String(), option: group.VOTE_OPTION_YES, weight: "2"},
				{address: myOtherAddr.String(), option: group.VOTE_OPTION_NO, weight: "1"},
			},
			policy: mockDecisionPolicy{
				AllowFn: func(tallyResult group.TallyResult, totalPower string) (group.DecisionPolicyResult, error) {
					return group.DecisionPolicyResult{Allow: true, Final: true}, nil
				},
			},
			expStatus:       group.PROPOSAL_STATUS_ACCEPTED,
			expVotesCleared: true,
			expEvents:       noEventsFn,
		},
		"proposal rejected": {
			votes: []memberVote{
				{address: myAddr.String(), option: group.VOTE_OPTION_YES, weight: "1"},
				{address: myOtherAddr.String(), option: group.VOTE_OPTION_NO, weight: "2"},
			},
			policy: mockDecisionPolicy{
				AllowFn: func(tallyResult group.TallyResult, totalPower string) (group.DecisionPolicyResult, error) {
					return group.DecisionPolicyResult{Allow: false, Final: true}, nil
				},
			},
			expStatus:       group.PROPOSAL_STATUS_REJECTED,
			expVotesCleared: true,
			expEvents:       noEventsFn,
		},
		"proposal in flight": {
			votes: []memberVote{
				{address: myAddr.String(), option: group.VOTE_OPTION_YES, weight: "1"},
				{address: myOtherAddr.String(), option: group.VOTE_OPTION_NO, weight: "1"},
			},
			policy: mockDecisionPolicy{
				AllowFn: func(tallyResult group.TallyResult, totalPower string) (group.DecisionPolicyResult, error) {
					return group.DecisionPolicyResult{Allow: false, Final: false}, nil
				},
			},
			expStatus:       group.PROPOSAL_STATUS_SUBMITTED,
			expVotesCleared: false,
			expEvents:       noEventsFn,
		},
		"policy errors": {
			votes: []memberVote{
				{address: myAddr.String(), option: group.VOTE_OPTION_YES, weight: "1"},
				{address: myOtherAddr.String(), option: group.VOTE_OPTION_NO, weight: "2"},
			},
			policy: mockDecisionPolicy{
				AllowFn: func(tallyResult group.TallyResult, totalPower string) (group.DecisionPolicyResult, error) {
					return group.DecisionPolicyResult{}, errors.New("my test error")
				},
			},
			expStatus:       group.PROPOSAL_STATUS_REJECTED,
			expVotesCleared: true,
			expEvents: func(proposalID uint64) sdk.Events {
				return sdk.Events{
					sdk.NewEvent("cosmos.group.v1.EventTallyError",
						sdk.Attribute{Key: "error_message", Value: `"my test error"`},
						sdk.Attribute{Key: "proposal_id", Value: fmt.Sprintf(`"%d"`, proposalID)},
					),
				}
			},
		},
	}
	var (
		groupID    uint64
		proposalId uint64
	)
	for name, spec := range specs {
		groupID++
		proposalId++
		t.Run(name, func(t *testing.T) {
			em := sdk.NewEventManager()
			ctx := testCtx.Ctx.WithEventManager(em)
			totalWeight, err := math.NewDecFromString("0")
			require.NoError(t, err)
			// given a group, policy and persisted votes
			for _, v := range spec.votes {
				err := groupKeeper.groupMemberTable.Create(ctx.KVStore(storeKey), &group.GroupMember{
					GroupId: groupID,
					Member:  &group.Member{Address: v.address, Weight: v.weight},
				})
				require.NoError(t, err)
				err = groupKeeper.voteTable.Create(ctx.KVStore(storeKey), &group.Vote{
					ProposalId: proposalId,
					Voter:      v.address,
					Option:     v.option,
				})
				require.NoError(t, err)
			}
			myGroupInfo := group.GroupInfo{
				TotalWeight: totalWeight.String(),
			}
			myPolicy := group.GroupPolicyInfo{GroupId: groupID}
			err = myPolicy.SetDecisionPolicy(spec.policy)
			require.NoError(t, err)

			myProposal := &group.Proposal{
				Id:              proposalId,
				Status:          group.PROPOSAL_STATUS_SUBMITTED,
				VotingPeriodEnd: ctx.BlockTime().Add(time.Hour),
			}

			// when
			gotErr := groupKeeper.doTallyAndUpdate(ctx, myProposal, myGroupInfo, myPolicy)
			// then
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expStatus, myProposal.Status)
			require.Equal(t, spec.expEvents(proposalId), em.Events())
			// and persistent state updated
			persistedVotes, err := groupKeeper.votesByProposal(ctx, groupID)
			require.NoError(t, err)
			if spec.expVotesCleared {
				assert.Empty(t, persistedVotes)
			} else {
				assert.Len(t, persistedVotes, len(spec.votes))
			}
		})
	}
}

// TestIsProposer is a regression test for MOCA-1263 (PR #290 review
// feedback): isProposer used to do exact string membership
// (slices.Contains) against proposal.Proposers, so a proposer submitting
// MsgWithdrawProposal with a validly-different spelling of their own
// address than what was stored on the proposal at submission time was
// wrongly denied. Uses raw sdk.AccAddress byte construction rather than
// string-round-tripping through any shared test fixture, since the
// package's broader shared test harness currently can't run (see
// [[moca-sdk-test-infra-broken-codec-testutil]] Phase 4).
func TestIsProposer(t *testing.T) {
	// 0xAB/0xCD (not 0x01/0x02) so the checksummed rendering actually
	// contains letter characters — an all-digit address would make
	// strings.ToLower a silent no-op and defeat the case-difference cases
	// below.
	addr := sdk.AccAddress(bytes.Repeat([]byte{0xAB}, 20))
	otherAddr := sdk.AccAddress(bytes.Repeat([]byte{0xCD}, 20))

	specs := map[string]struct {
		proposers []string
		address   string
		exp       bool
	}{
		"exact match": {
			proposers: []string{addr.String()},
			address:   addr.String(),
			exp:       true,
		},
		"non-canonical spelling of a stored proposer still matches": {
			proposers: []string{addr.String()},
			address:   strings.ToLower(addr.String()),
			exp:       true,
		},
		"stored proposer in non-canonical spelling still matches a canonical query": {
			proposers: []string{strings.TrimPrefix(addr.String(), "0x")},
			address:   addr.String(),
			exp:       true,
		},
		"different address does not match": {
			proposers: []string{addr.String()},
			address:   otherAddr.String(),
			exp:       false,
		},
		"query address does not decode": {
			proposers: []string{addr.String()},
			address:   "not-an-address",
			exp:       false,
		},
		"stored proposer does not decode, still checks the rest of the list": {
			proposers: []string{"not-an-address", addr.String()},
			address:   addr.String(),
			exp:       true,
		},
		"empty proposers list": {
			proposers: nil,
			address:   addr.String(),
			exp:       false,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got := isProposer(group.Proposal{Proposers: spec.proposers}, spec.address)
			assert.Equal(t, spec.exp, got)
		})
	}
}

// TestWithdrawProposal_NonCanonicalAddressForm is a regression test for
// MOCA-1263 (PR #290 review feedback): WithdrawProposal's group-policy-admin
// check used to compare msg.Address against the stored policy admin as raw
// strings, so a policy admin submitting MsgWithdrawProposal with a validly-
// different spelling of their own address than what was stored at policy
// creation time was wrongly denied. Covers the admin path directly (the
// proposer path is covered in depth by TestIsProposer above).
func TestWithdrawProposal_NonCanonicalAddressForm(t *testing.T) {
	admin := sdk.AccAddress(bytes.Repeat([]byte{0xAB}, 20))
	policyAddr := sdk.AccAddress(bytes.Repeat([]byte{0xCD}, 20))

	encCfg := moduletestutil.MakeTestEncodingConfig()
	group.RegisterInterfaces(encCfg.InterfaceRegistry)

	storeKey := storetypes.NewKVStoreKey(group.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, storeKey, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx
	myAccountKeeper := &mockAccountKeeper{
		AddressCodecFn: func() coreaddress.Codec {
			return address.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
		},
	}
	groupKeeper := NewKeeper(storeKey, encCfg.Codec, nil, myAccountKeeper, group.DefaultConfig())

	policyInfo := &group.GroupPolicyInfo{
		Address: policyAddr.String(),
		Admin:   admin.String(),
		GroupId: 1,
		Version: 1,
	}
	require.NoError(t, policyInfo.SetDecisionPolicy(group.NewThresholdDecisionPolicy("1", time.Hour, 0)))
	require.NoError(t, groupKeeper.groupPolicyTable.Create(ctx.KVStore(storeKey), policyInfo))

	nextID := groupKeeper.proposalTable.Sequence().PeekNextVal(ctx.KVStore(storeKey))
	proposalID, err := groupKeeper.proposalTable.Create(ctx.KVStore(storeKey), &group.Proposal{
		Id:                 nextID,
		GroupPolicyAddress: policyAddr.String(),
		GroupVersion:       1,
		GroupPolicyVersion: 1,
		Status:             group.PROPOSAL_STATUS_SUBMITTED,
		Proposers:          []string{}, // withdrawing via the admin path, not the proposer path
		FinalTallyResult:   group.DefaultTallyResult(),
	})
	require.NoError(t, err)

	// Non-canonical (all-lowercase) spelling of the admin's own address —
	// a valid spelling of the same account, different from the canonical
	// checksummed form stored on the group policy at creation time.
	nonCanonicalAdmin := strings.ToLower(admin.String())
	require.NotEqual(t, admin.String(), nonCanonicalAdmin, "sanity check: fixture must actually differ in spelling from the canonical form")

	_, err = groupKeeper.WithdrawProposal(ctx, &group.MsgWithdrawProposal{
		ProposalId: proposalID,
		Address:    nonCanonicalAdmin,
	})
	require.NoError(t, err)

	got, err := groupKeeper.getProposal(ctx, proposalID)
	require.NoError(t, err)
	assert.Equal(t, group.PROPOSAL_STATUS_WITHDRAWN, got.Status)
}

// TestDoUpdateGroup_NonCanonicalAddressForm is a regression test for
// MOCA-1263 (PR #290 review feedback): doUpdateGroup's admin-authorization
// check — shared by UpdateGroupMembers, UpdateGroupAdmin, and
// UpdateGroupMetadata — used to compare the caller-supplied admin address
// against the stored one with strings.EqualFold, which misses a valid
// spelling that differs in "0x" prefix presence rather than just case.
func TestDoUpdateGroup_NonCanonicalAddressForm(t *testing.T) {
	admin := sdk.AccAddress(bytes.Repeat([]byte{0xAB}, 20))

	encCfg := moduletestutil.MakeTestEncodingConfig()
	group.RegisterInterfaces(encCfg.InterfaceRegistry)

	storeKey := storetypes.NewKVStoreKey(group.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, storeKey, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx
	myAccountKeeper := &mockAccountKeeper{
		AddressCodecFn: func() coreaddress.Codec {
			return address.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
		},
	}
	groupKeeper := NewKeeper(storeKey, encCfg.Codec, nil, myAccountKeeper, group.DefaultConfig())

	nextID := groupKeeper.groupTable.Sequence().PeekNextVal(ctx.KVStore(storeKey))
	groupID, err := groupKeeper.groupTable.Create(ctx.KVStore(storeKey), &group.GroupInfo{
		Id:          nextID,
		Admin:       admin.String(),
		TotalWeight: "0",
		Version:     1,
	})
	require.NoError(t, err)

	// Non-canonical (no "0x" prefix) spelling of the admin's own address —
	// EqualFold alone wouldn't have caught this, since it only normalizes
	// case, not prefix presence.
	nonCanonicalAdmin := strings.TrimPrefix(admin.String(), "0x")
	require.NotEqual(t, admin.String(), nonCanonicalAdmin, "sanity check: fixture must actually differ in spelling from the canonical form")

	ranAction := false
	err = groupKeeper.doUpdateGroup(ctx, groupID, nonCanonicalAdmin, func(*group.GroupInfo) error {
		ranAction = true
		return nil
	}, "test update")
	require.NoError(t, err)
	assert.True(t, ranAction, "action should have run once the admin check passed")
}

// TestDoUpdateGroupPolicy_NonCanonicalAddressForm is a regression test for
// MOCA-1263 (PR #290 review feedback): doUpdateGroupPolicy's admin check —
// shared by UpdateGroupPolicyAdmin, UpdateGroupPolicyMetadata, and
// UpdateGroupPolicyDecisionPolicy — had the same non-canonical-spelling gap
// as doUpdateGroup, fixed the same way.
func TestDoUpdateGroupPolicy_NonCanonicalAddressForm(t *testing.T) {
	admin := sdk.AccAddress(bytes.Repeat([]byte{0xAB}, 20))
	policyAddr := sdk.AccAddress(bytes.Repeat([]byte{0xCD}, 20))

	encCfg := moduletestutil.MakeTestEncodingConfig()
	group.RegisterInterfaces(encCfg.InterfaceRegistry)

	storeKey := storetypes.NewKVStoreKey(group.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, storeKey, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx
	myAccountKeeper := &mockAccountKeeper{
		AddressCodecFn: func() coreaddress.Codec {
			return address.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
		},
	}
	groupKeeper := NewKeeper(storeKey, encCfg.Codec, nil, myAccountKeeper, group.DefaultConfig())

	policyInfo := &group.GroupPolicyInfo{
		Address: policyAddr.String(),
		Admin:   admin.String(),
		GroupId: 1,
		Version: 1,
	}
	require.NoError(t, policyInfo.SetDecisionPolicy(group.NewThresholdDecisionPolicy("1", time.Hour, 0)))
	require.NoError(t, groupKeeper.groupPolicyTable.Create(ctx.KVStore(storeKey), policyInfo))

	// Non-canonical (all-lowercase) spelling of the admin's own address.
	nonCanonicalAdmin := strings.ToLower(admin.String())
	require.NotEqual(t, admin.String(), nonCanonicalAdmin, "sanity check: fixture must actually differ in spelling from the canonical form")

	ranAction := false
	err := groupKeeper.doUpdateGroupPolicy(ctx, policyAddr.String(), nonCanonicalAdmin, func(*group.GroupPolicyInfo) error {
		ranAction = true
		return nil
	}, "test update")
	require.NoError(t, err)
	assert.True(t, ranAction, "action should have run once the admin check passed")
}

var _ group.AccountKeeper = &mockAccountKeeper{}

// mockAccountKeeper is a mock implementation of the AccountKeeper interface for testing purposes.
type mockAccountKeeper struct {
	AddressCodecFn func() coreaddress.Codec
}

func (m mockAccountKeeper) AddressCodec() coreaddress.Codec {
	if m.AddressCodecFn == nil {
		panic("not expected to be called")
	}
	return m.AddressCodecFn()
}

func (m mockAccountKeeper) NewAccount(ctx context.Context, i sdk.AccountI) sdk.AccountI {
	panic("not expected to be called")
}

func (m mockAccountKeeper) GetAccount(ctx context.Context, address sdk.AccAddress) sdk.AccountI {
	panic("not expected to be called")
}

func (m mockAccountKeeper) SetAccount(ctx context.Context, i sdk.AccountI) {
	panic("not expected to be called")
}

func (m mockAccountKeeper) RemoveAccount(ctx context.Context, acc sdk.AccountI) {
	panic("not expected to be called")
}

// mockDecisionPolicy is a mock implementation of a decision policy for testing purposes.
type mockDecisionPolicy struct {
	fakeProtoType
	AllowFn func(tallyResult group.TallyResult, totalPower string) (group.DecisionPolicyResult, error)
}

func (m mockDecisionPolicy) Allow(tallyResult group.TallyResult, totalPower string) (group.DecisionPolicyResult, error) {
	if m.AllowFn == nil {
		panic("not expected to be called")
	}
	return m.AllowFn(tallyResult, totalPower)
}

func (m mockDecisionPolicy) GetVotingPeriod() time.Duration {
	panic("not expected to be called")
}

func (m mockDecisionPolicy) GetMinExecutionPeriod() time.Duration {
	panic("not expected to be called")
}

func (m mockDecisionPolicy) ValidateBasic() error {
	panic("not expected to be called")
}

func (m mockDecisionPolicy) Validate(g group.GroupInfo, config group.Config) error {
	panic("not expected to be called")
}

var (
	_ proto.Marshaler = (*fakeProtoType)(nil)
	_ proto.Message   = (*fakeProtoType)(nil)
)

// fakeProtoType is a struct used for mocking and testing purposes.
// Custom types can be converted into Any and back via internal CachedValue only.
type fakeProtoType struct{}

func (a fakeProtoType) Reset() {}

func (a fakeProtoType) String() string {
	return "testing"
}

func (a fakeProtoType) Marshal() ([]byte, error) {
	return nil, nil
}

func (a fakeProtoType) ProtoMessage() {}
