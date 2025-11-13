package upgrade

// Compatibility module for legacy x/upgrade imports
// In Cosmos SDK v0.50, x/upgrade has been moved to cosmossdk.io/x/upgrade
// This module provides minimal compatibility for existing code

import (
	"context"
	"encoding/json"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

var (
	_ module.AppModule = AppModule{}
)

// AppModule implements an application module for the upgrade module.
type AppModule struct{}

// NewAppModule creates a new AppModule object
func NewAppModule() AppModule {
	return AppModule{}
}

// Name returns the upgrade module's name.
func (AppModule) Name() string { return "upgrade" }

// RegisterLegacyAminoCodec registers the upgrade module's types on the given LegacyAmino codec.
func (AppModule) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {}

// RegisterInterfaces registers the module's interface types
func (AppModule) RegisterInterfaces(_ codec.InterfaceRegistry) {}

// DefaultGenesis returns default genesis state as raw bytes for the upgrade module.
func (AppModule) DefaultGenesis(_ codec.JSONCodec) json.RawMessage { return nil }

// ValidateGenesis performs genesis state validation for the upgrade module.
func (AppModule) ValidateGenesis(_ codec.JSONCodec, _ client.TxEncodingConfig, _ json.RawMessage) error {
	return nil
}

// RegisterServices registers module services.
func (AppModule) RegisterServices(_ module.Configurator) {}

// InitGenesis performs genesis initialization for the upgrade module.
func (am AppModule) InitGenesis(_ sdk.Context, _ codec.JSONCodec, _ json.RawMessage) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the upgrade module.
func (am AppModule) ExportGenesis(_ sdk.Context, _ codec.JSONCodec) json.RawMessage {
	return nil
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// BeginBlock returns the begin blocker for the upgrade module.
func (am AppModule) BeginBlock(_ context.Context) error {
	return nil
}

// EndBlock returns the end blocker for the upgrade module.
func (am AppModule) EndBlock(_ context.Context) error {
	return nil
}