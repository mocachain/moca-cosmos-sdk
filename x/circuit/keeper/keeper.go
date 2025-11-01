package keeper

import (
	context "context"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/x/circuit/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper defines the circuit module's keeper.
type Keeper struct {
	cdc          codec.BinaryCodec
	storeService store.KVStoreService

	authority []byte

	Schema collections.Schema
	// Permissions contains the permissions for each account
	Permissions collections.Map[[]byte, types.Permissions]
	// DisableList contains the message URLs that are disabled
	DisableList collections.KeySet[string]
}

// NewKeeper constructs a new Circuit Keeper instance
func NewKeeper(cdc codec.BinaryCodec, storeService store.KVStoreService, authority string) Keeper {
	auth, err := sdk.AccAddressFromHexUnsafe(authority)
	if err != nil {
		panic(err)
	}

	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		cdc:          cdc,
		storeService: storeService,
		authority:    auth,
		Permissions: collections.NewMap(
			sb,
			types.AccountPermissionPrefix,
			"permissions",
			collections.BytesKey,
			codec.CollValue[types.Permissions](cdc),
		),
		DisableList: collections.NewKeySet(
			sb,
			types.DisableListPrefix,
			"disable_list",
			collections.StringKey,
		),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	return k
}

func (k *Keeper) GetAuthority() []byte {
	return k.authority
}

// IsAllowed returns true when msg URL is not found in the DisableList for given context, else false.
func (k *Keeper) IsAllowed(ctx context.Context, msgURL string) (bool, error) {
	has, err := k.DisableList.Has(ctx, msgURL)
	return !has, err
}
