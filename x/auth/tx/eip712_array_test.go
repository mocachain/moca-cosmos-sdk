package tx

import (
	"bytes"
	"context"
	"encoding/hex"
	"testing"

	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/stretchr/testify/require"

	txsigning "cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/group"
)

const arrayFxChainID = "moca_5151-1"

func arrayFxAddr(b byte) sdk.AccAddress { return sdk.AccAddress(bytes.Repeat([]byte{b}, 20)) }

func arrayFxSignerData() signing.SignerData {
	return signing.SignerData{Address: arrayFxAddr(0x11).String(), ChainID: arrayFxChainID, AccountNumber: 7, Sequence: 3}
}

func arrayFxTx(t *testing.T, msg sdk.Msg) sdk.Tx {
	t.Helper()

	b := NewTxConfig(codec.NewProtoCodec(codectypes.NewInterfaceRegistry()),
		[]signingtypes.SignMode{signingtypes.SignMode_SIGN_MODE_EIP_712}).NewTxBuilder()
	require.NoError(t, b.SetMsgs(msg))
	b.SetMemo("eip712 array fixture")
	b.SetGasLimit(200000)
	b.SetTimeoutHeight(1000)
	b.SetFeeAmount(sdk.NewCoins(sdk.NewInt64Coin("amoca", 150)))
	b.SetFeePayer(arrayFxAddr(0x11))
	return b.GetTx()
}

// The digest was produced on base revision 17daa50fcc and is pinned identically in x/auth/signing.
func TestEIP712HandlerArrayDigest(t *testing.T) {
	coins := func(a, b int64) sdk.Coins {
		return sdk.NewCoins(sdk.NewInt64Coin("amoca", a), sdk.NewInt64Coin("bmoca", b))
	}
	tx := arrayFxTx(t, banktypes.NewMsgMultiSend(
		banktypes.Input{Address: arrayFxAddr(1).String(), Coins: coins(10, 11)},
		[]banktypes.Output{
			{Address: arrayFxAddr(2).String(), Coins: coins(4, 5)},
			{Address: arrayFxAddr(3).String(), Coins: coins(6, 6)},
		},
	))

	got, err := signModeEip712Handler{}.GetSignBytes(signingtypes.SignMode_SIGN_MODE_EIP_712, arrayFxSignerData(), tx)
	require.NoError(t, err)
	require.Equal(t, "60342e9c086d67aeaf943d03ee53f6546e0cfba272f87a944e7d93dd03484bf2", hex.EncodeToString(got))

	sd := arrayFxSignerData()
	adapter, err := signing.EIP712GetSignBytes(context.Background(),
		txsigning.SignerData{Address: sd.Address, ChainID: sd.ChainID, AccountNumber: sd.AccountNumber, Sequence: sd.Sequence}, tx)
	require.NoError(t, err)
	require.Equal(t, adapter, got)
}

func TestEIP712HandlerRejectsHeterogeneousElements(t *testing.T) {
	cases := []struct {
		name  string
		msg   sdk.Msg
		field string
	}{
		{
			name: "later denom unit carries aliases the first lacks",
			msg: &banktypes.Metadata{DenomUnits: []*banktypes.DenomUnit{
				{Denom: "amoca"}, {Denom: "moca", Exponent: 18, Aliases: []string{"MOCA"}},
			}},
			field: `"aliases"`,
		},
		{
			name: "later group member leaves a template field unset",
			msg: &group.QueryGroupMembersResponse{
				Members:    []*group.GroupMember{{GroupId: 1, Member: &group.Member{Address: arrayFxAddr(1).String()}}, {GroupId: 2}},
				Pagination: &query.PageResponse{Total: 2},
			},
			field: `"member"`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tx := arrayFxTx(t, tc.msg)
			_, err := signModeEip712Handler{}.GetSignBytes(signingtypes.SignMode_SIGN_MODE_EIP_712, arrayFxSignerData(), tx)
			require.ErrorContains(t, err, tc.field)

			// the CLI preview path
			chainID, err := sdk.ParseChainID(arrayFxChainID)
			require.NoError(t, err)
			msgTypes, signDoc, err := GetMsgTypes(arrayFxSignerData(), tx, chainID)
			require.NoError(t, err)
			_, err = WrapTxToTypedData(signDoc, msgTypes, apitypes.TypedDataDomain{})
			require.ErrorContains(t, err, tc.field)
		})
	}
}
