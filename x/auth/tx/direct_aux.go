package tx

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types/tx"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

var _ signing.SignModeHandler = SignModeDirectAuxHandler{}

// SignModeDirectAuxHandler defines the SIGN_MODE_DIRECT_AUX SignModeHandler.
//
// NOTE: it is not wired into NewTxConfig — config.go registers the
// cosmossdk.io/x/tx directaux handler for SIGN_MODE_DIRECT_AUX. This type is
// exported only so its fee-payer authorization guard can be regression-tested
// from an external test package; the package's own (internal) test binary does
// not currently build due to a pre-existing, unrelated import cycle.
type SignModeDirectAuxHandler struct{}

// DefaultMode implements SignModeHandler.DefaultMode
func (SignModeDirectAuxHandler) DefaultMode() signingtypes.SignMode {
	return signingtypes.SignMode_SIGN_MODE_DIRECT_AUX
}

// Modes implements SignModeHandler.Modes
func (SignModeDirectAuxHandler) Modes() []signingtypes.SignMode {
	return []signingtypes.SignMode{signingtypes.SignMode_SIGN_MODE_DIRECT_AUX}
}

// GetSignBytes implements SignModeHandler.GetSignBytes
func (SignModeDirectAuxHandler) GetSignBytes(
	mode signingtypes.SignMode, data signing.SignerData, tx sdk.Tx,
) ([]byte, error) {
	if mode != signingtypes.SignMode_SIGN_MODE_DIRECT_AUX {
		return nil, fmt.Errorf("expected %s, got %s", signingtypes.SignMode_SIGN_MODE_DIRECT_AUX, mode)
	}

	protoTx, ok := tx.(*wrapper)
	if !ok {
		return nil, fmt.Errorf("can only handle a protobuf Tx, got %T", tx)
	}

	pkAny, err := codectypes.NewAnyWithValue(data.PubKey)
	if err != nil {
		return nil, err
	}

	addr := data.Address
	if addr == "" {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "got empty address in %s handler", signingtypes.SignMode_SIGN_MODE_DIRECT_AUX)
	}

	feePayer := protoTx.FeePayer()

	dataAddr, err := sdk.AccAddressFromHexUnsafe(data.Address)
	if err != nil {
		return nil, err
	}

	// Fee payer cannot use SIGN_MODE_DIRECT_AUX, because SIGN_MODE_DIRECT_AUX
	// does not sign over fees, which would create malleability issues.
	if sdk.AccAddress(feePayer).Equals(dataAddr) {
		return nil, sdkerrors.ErrUnauthorized.Wrapf("fee payer %s cannot sign with %s", sdk.AccAddress(feePayer).String(), signingtypes.SignMode_SIGN_MODE_DIRECT_AUX)
	}

	signDocDirectAux := types.SignDocDirectAux{
		BodyBytes:     protoTx.getBodyBytes(),
		ChainId:       data.ChainID,
		AccountNumber: data.AccountNumber,
		Sequence:      data.Sequence,
		Tip:           protoTx.tx.AuthInfo.Tip,
		PublicKey:     pkAny,
	}

	return signDocDirectAux.Marshal()
}
