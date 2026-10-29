package simulation_test

import (
	"errors"
	"time"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil/simsx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/slashing/simulation"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TestMsgUnjailFactoryGlobalMinSelfDelegation checks that the unjail factory expects the
// delivery to fail when the self-delegation meets the validator's own minimum but not the
// staking module's global minimum, and expects it to succeed when it meets both.
func (suite *SimTestSuite) TestMsgUnjailFactoryGlobalMinSelfDelegation() {
	ctx := suite.ctx.WithBlockTime(time.Now().UTC())

	// jailed validator0 whose jail period has ended
	validator0, err := getTestingValidator0(ctx, suite.stakingKeeper, suite.accounts)
	suite.Require().NoError(err)
	suite.Require().NoError(suite.stakingKeeper.SetValidatorByConsAddr(ctx, validator0))
	consAddr, err := validator0.GetConsAddr()
	suite.Require().NoError(err)
	info := types.NewValidatorSigningInfo(consAddr, int64(4), int64(3), time.Unix(2, 0), false, int64(10))
	suite.Require().NoError(suite.slashingKeeper.SetValidatorSigningInfo(ctx, consAddr, info))
	suite.Require().NoError(suite.stakingKeeper.Jail(ctx, consAddr))

	// self-delegation at or above the validator's own minimum
	_, issuedShares := validator0.AddTokensFromDel(suite.stakingKeeper.TokensFromConsensusPower(ctx, 2))
	selfDelegation := stakingtypes.NewDelegation(suite.accounts[0].Address.String(), validator0.GetOperator(), issuedShares)
	suite.Require().NoError(suite.stakingKeeper.SetDelegation(ctx, selfDelegation))

	valAddr, err := sdk.AccAddressFromHexUnsafe(validator0.GetOperator())
	suite.Require().NoError(err)
	stored, err := suite.stakingKeeper.GetValidator(ctx, valAddr)
	suite.Require().NoError(err)
	selfBond := stored.TokensFromShares(issuedShares).TruncateInt()
	suite.Require().True(selfBond.GTE(stored.GetMinSelfDelegation()))

	setGlobalMin := func(v math.Int) {
		params, err := suite.stakingKeeper.GetParams(ctx)
		suite.Require().NoError(err)
		params.MinSelfDelegation = v
		suite.Require().NoError(suite.stakingKeeper.SetParams(ctx, params))
	}

	// index the accounts by the app address codec's rendering, which GetAccountbyAccAddr looks up
	addressCodec := suite.accountKeeper.AddressCodec()
	accounts := make([]simtypes.Account, len(suite.accounts))
	for i, acc := range suite.accounts {
		acc.AddressBech32, err = addressCodec.BytesToString(acc.Address)
		suite.Require().NoError(err)
		accounts[i] = acc
	}
	testData := simsx.NewChainDataSource(ctx, suite.r, suite.accountKeeper, suite.bankKeeper, addressCodec, accounts...)
	factory := simulation.MsgUnjailFactory(suite.slashingKeeper, suite.stakingKeeper)

	// deliveryResultHandler runs the factory until it picks validator0, the only jailed
	// validator, and returns the delivery result handler it set.
	deliveryResultHandler := func() simsx.SimDeliveryResultHandler {
		skips := make(map[string]int)
		for range 50 {
			reporter := simsx.NewBasicSimulationReporter()
			_, msg := factory.Create()(ctx, testData, reporter)
			if reporter.IsSkipped() {
				skips[reporter.Comment()]++
				continue
			}
			suite.Require().Equal(validator0.GetOperator(), msg.(*types.MsgUnjail).ValidatorAddr)
			return factory.DeliveryResultHandler()
		}
		suite.Require().FailNow("factory never picked the jailed validator", "skips: %v", skips)
		return nil
	}

	setGlobalMin(selfBond.AddRaw(1))
	handler := deliveryResultHandler()
	suite.Require().Error(handler(nil), "unjail succeeding below the global minimum must be reported")
	suite.Require().NoError(handler(errors.New("unjail rejected")), "unjail failing below the global minimum is expected")

	setGlobalMin(selfBond)
	handler = deliveryResultHandler()
	suite.Require().NoError(handler(nil), "unjail succeeding at the global minimum is expected")
}
