package authz_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"

	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	sdkmath "cosmossdk.io/math"
	txsigning "cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/tx/signing/aminojson"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestMsgGrantGetAuthorization(t *testing.T) {
	require := require.New(t)

	m := authz.MsgGrant{}
	require.Nil(m.GetAuthorization())

	g := authz.GenericAuthorization{Msg: "some_type"}
	var err error
	m.Grant.Authorization, err = cdctypes.NewAnyWithValue(&g)
	require.NoError(err)

	a, err := m.GetAuthorization()
	require.NoError(err)
	require.Equal(a, &g)

	g = authz.GenericAuthorization{Msg: "some_type2"}
	m.SetAuthorization(&g)
	a, err = m.GetAuthorization()
	require.NoError(err)
	require.Equal(a, &g)
}

func TestAminoJSON(t *testing.T) {
	legacyAmino := codec.NewLegacyAmino()
	authz.RegisterLegacyAminoCodec(legacyAmino)
	banktypes.RegisterLegacyAminoCodec(legacyAmino)
	stakingtypes.RegisterLegacyAminoCodec(legacyAmino)
	legacytx.RegressionTestingAminoCodec = legacyAmino

	aminoHandler := aminojson.NewSignModeHandler(aminojson.SignModeHandlerOptions{
		FileResolver: proto.HybridResolver,
	})

	tx := legacytx.StdTx{}
	blockTime := time.Date(1, 1, 1, 1, 1, 1, 1, time.UTC)
	expiresAt := blockTime.Add(time.Hour)
	msgSend := banktypes.MsgSend{FromAddress: "0xD9CFC297C7870c29b6e6e1bEA2619Ae59851566E", ToAddress: "0x35B4c90554c8Ed4C5759B3793B1e8932AA4dF0Af"}
	typeURL := sdk.MsgTypeURL(&msgSend)
	msgSendAny, err := cdctypes.NewAnyWithValue(&msgSend)
	require.NoError(t, err)
	grant, err := authz.NewGrant(blockTime, authz.NewGenericAuthorization(typeURL), &expiresAt)
	require.NoError(t, err)
	sendAuthz := banktypes.NewSendAuthorization(sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1000))), nil)
	sendGrant, err := authz.NewGrant(blockTime, sendAuthz, &expiresAt)
	require.NoError(t, err)
	valAddr, err := sdk.AccAddressFromHexUnsafe("0x46419b6D9590e643dC4ADbCa14e165afdc454AA3")
	require.NoError(t, err)
	stakingAuth, err := stakingtypes.NewStakeAuthorization([]sdk.AccAddress{valAddr}, nil, stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_DELEGATE, &sdk.Coin{Denom: "stake", Amount: sdkmath.NewInt(1000)})
	require.NoError(t, err)
	delegateGrant, err := authz.NewGrant(blockTime, stakingAuth, nil)
	require.NoError(t, err)

	// Amino JSON encoding has changed in authz since v0.46.
	// Before, it was outputting something like:
	// `{"account_number":"1","chain_id":"foo","fee":{"amount":[],"gas":"0"},"memo":"memo","msgs":[{"grant":{"authorization":{"msg":"/cosmos.bank.v1beta1.MsgSend"},"expiration":"0001-01-01T02:01:01.000000001Z"},"grantee":"0x3a3296162941bb4B22F7d7c23091089F82732CcC","granter":"0x3e599c4946ac23b9A6Ff280c76c3d929ebB16AF9"}],"sequence":"1","timeout_height":"1"}`
	//
	// This was a bug. Now, it's as below, See how there's `type` & `value` fields.
	// ref: https://github.com/cosmos/cosmos-sdk/issues/11190
	// ref: https://github.com/cosmos/cosmjs/issues/1026
	tests := []struct {
		msg sdk.Msg
		exp string
	}{
		{
			msg: &authz.MsgGrant{Granter: "0x3e599c4946ac23b9A6Ff280c76c3d929ebB16AF9", Grantee: "0x3a3296162941bb4B22F7d7c23091089F82732CcC", Grant: grant},
			exp: `{"account_number":"1","chain_id":"foo","fee":{"amount":[],"gas":"0"},"memo":"memo","msgs":[{"type":"cosmos-sdk/MsgGrant","value":{"grant":{"authorization":{"type":"cosmos-sdk/GenericAuthorization","value":{"msg":"/cosmos.bank.v1beta1.MsgSend"}},"expiration":"0001-01-01T02:01:01.000000001Z"},"grantee":"0x3a3296162941bb4B22F7d7c23091089F82732CcC","granter":"0x3e599c4946ac23b9A6Ff280c76c3d929ebB16AF9"}}],"sequence":"1","timeout_height":"1"}`,
		},
		{
			msg: &authz.MsgGrant{Granter: "0x3e599c4946ac23b9A6Ff280c76c3d929ebB16AF9", Grantee: "0x3a3296162941bb4B22F7d7c23091089F82732CcC", Grant: sendGrant},
			exp: `{"account_number":"1","chain_id":"foo","fee":{"amount":[],"gas":"0"},"memo":"memo","msgs":[{"type":"cosmos-sdk/MsgGrant","value":{"grant":{"authorization":{"type":"cosmos-sdk/SendAuthorization","value":{"spend_limit":[{"amount":"1000","denom":"stake"}]}},"expiration":"0001-01-01T02:01:01.000000001Z"},"grantee":"0x3a3296162941bb4B22F7d7c23091089F82732CcC","granter":"0x3e599c4946ac23b9A6Ff280c76c3d929ebB16AF9"}}],"sequence":"1","timeout_height":"1"}`,
		},
		{
			msg: &authz.MsgGrant{Granter: "0x3e599c4946ac23b9A6Ff280c76c3d929ebB16AF9", Grantee: "0x3a3296162941bb4B22F7d7c23091089F82732CcC", Grant: delegateGrant},
			exp: `{"account_number":"1","chain_id":"foo","fee":{"amount":[],"gas":"0"},"memo":"memo","msgs":[{"type":"cosmos-sdk/MsgGrant","value":{"grant":{"authorization":{"type":"cosmos-sdk/StakeAuthorization","value":{"Validators":{"type":"cosmos-sdk/StakeAuthorization/AllowList","value":{"allow_list":{"address":["0x46419b6D9590e643dC4ADbCa14e165afdc454AA3"]}}},"authorization_type":1,"max_tokens":{"amount":"1000","denom":"stake"}}}},"grantee":"0x3a3296162941bb4B22F7d7c23091089F82732CcC","granter":"0x3e599c4946ac23b9A6Ff280c76c3d929ebB16AF9"}}],"sequence":"1","timeout_height":"1"}`,
		},
		{
			msg: &authz.MsgRevoke{Granter: "0x3e599c4946ac23b9A6Ff280c76c3d929ebB16AF9", Grantee: "0x3a3296162941bb4B22F7d7c23091089F82732CcC", MsgTypeUrl: typeURL},
			exp: `{"account_number":"1","chain_id":"foo","fee":{"amount":[],"gas":"0"},"memo":"memo","msgs":[{"type":"cosmos-sdk/MsgRevoke","value":{"grantee":"0x3a3296162941bb4B22F7d7c23091089F82732CcC","granter":"0x3e599c4946ac23b9A6Ff280c76c3d929ebB16AF9","msg_type_url":"/cosmos.bank.v1beta1.MsgSend"}}],"sequence":"1","timeout_height":"1"}`,
		},
		{
			msg: &authz.MsgExec{Grantee: "0x3a3296162941bb4B22F7d7c23091089F82732CcC", Msgs: []*cdctypes.Any{msgSendAny}},
			exp: `{"account_number":"1","chain_id":"foo","fee":{"amount":[],"gas":"0"},"memo":"memo","msgs":[{"type":"cosmos-sdk/MsgExec","value":{"grantee":"0x3a3296162941bb4B22F7d7c23091089F82732CcC","msgs":[{"type":"cosmos-sdk/MsgSend","value":{"amount":[],"from_address":"0xD9CFC297C7870c29b6e6e1bEA2619Ae59851566E","to_address":"0x35B4c90554c8Ed4C5759B3793B1e8932AA4dF0Af"}}]}}],"sequence":"1","timeout_height":"1"}`,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tx.Msgs = []sdk.Msg{tt.msg}
			legacyJSON := string(legacytx.StdSignBytes("foo", 1, 1, 1, legacytx.StdFee{}, []sdk.Msg{tt.msg}, "memo"))
			require.Equal(t, tt.exp, legacyJSON)

			legacyAny, err := cdctypes.NewAnyWithValue(tt.msg)
			require.NoError(t, err)
			anyMsg := &anypb.Any{
				TypeUrl: legacyAny.TypeUrl,
				Value:   legacyAny.Value,
			}
			aminoJSON, err := aminoHandler.GetSignBytes(
				context.TODO(),
				txsigning.SignerData{
					Address:       "foo",
					ChainID:       "foo",
					AccountNumber: 1,
					Sequence:      1,
				},
				txsigning.TxData{
					Body: &txv1beta1.TxBody{
						Memo:          "memo",
						Messages:      []*anypb.Any{anyMsg},
						TimeoutHeight: 1,
					},
					AuthInfo: &txv1beta1.AuthInfo{
						Fee: &txv1beta1.Fee{},
					},
				},
			)
			require.NoError(t, err)
			require.Equal(t, tt.exp, string(aminoJSON))
		})
	}
}
