package genutil_test

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

// writeSingleAccountGenesis writes a genesis file containing one account whose
// bank balance entry is stored under storedSpelling (which may be a
// non-canonical rendering of addr), with supply matching the balance. It
// returns the file path.
func writeSingleAccountGenesis(t *testing.T, cdc codec.Codec, addr sdk.AccAddress, storedSpelling string, coins sdk.Coins) string {
	t.Helper()

	authState := authtypes.NewGenesisState(
		authtypes.DefaultParams(),
		[]authtypes.GenesisAccount{authtypes.NewBaseAccountWithAddress(addr)},
	)
	bankState := banktypes.DefaultGenesisState()
	bankState.Balances = []banktypes.Balance{{Address: storedSpelling, Coins: coins}}
	bankState.Supply = coins

	appState := map[string]json.RawMessage{
		authtypes.ModuleName: cdc.MustMarshalJSON(authState),
		banktypes.ModuleName: cdc.MustMarshalJSON(bankState),
	}
	appStateJSON, err := json.Marshal(appState)
	require.NoError(t, err)

	genFile := filepath.Join(t.TempDir(), "genesis.json")
	require.NoError(t, genutil.ExportGenesisFile(genutiltypes.NewAppGenesisWithVersion("test-chain", appStateJSON), genFile))
	return genFile
}

// reloadBankGenesis reads the genesis file back and returns its bank state.
func reloadBankGenesis(t *testing.T, cdc codec.Codec, genFile string) *banktypes.GenesisState {
	t.Helper()
	appState, _, err := genutiltypes.GenesisStateFromGenFile(genFile)
	require.NoError(t, err)
	return banktypes.GetGenesisStateFromAppState(cdc, appState)
}

// assertSingleBalance asserts the bank state holds exactly one balance entry
// for addr (matched by decoded value), with the expected amount, and that
// declared supply equals the sum of all balances.
func assertSingleBalance(t *testing.T, bankState *banktypes.GenesisState, addr sdk.AccAddress, expected sdk.Coins) {
	t.Helper()

	var (
		matches int
		got     sdk.Coins
		total   sdk.Coins
	)
	for _, bal := range bankState.Balances {
		total = total.Add(bal.Coins...)
		balAddr, err := sdk.AccAddressFromHexUnsafe(bal.Address)
		require.NoError(t, err)
		if balAddr.Equals(addr) {
			matches++
			got = bal.Coins
		}
	}
	require.Equal(t, 1, matches, "expected exactly one balance entry for the account, got %d (balances: %v)", matches, bankState.Balances)
	require.True(t, expected.Equal(got), "expected balance %s, got %s", expected, got)
	require.True(t, sdk.Coins(bankState.Supply).Equal(total), "declared supply %s must equal the sum of balances %s", bankState.Supply, total)
}

// TestAddGenesisAccount_NonCanonicalStoredBalance is a regression test for
// MOCA-1263 follow-up work: add-genesis-account's append path looked up the
// existing balance entry to top up by raw string match, so a balance stored
// under a valid non-canonical spelling of the same address was silently left
// untouched while supply was still incremented — leaving the generated file's
// declared supply larger than the sum of its balances.
func TestAddGenesisAccount_NonCanonicalStoredBalance(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{})
	cdc := encCfg.Codec

	// 0xAB pattern so the rendering has letter characters; lowercase +
	// prefix-stripped is then guaranteed to differ from the canonical form.
	addr := sdk.AccAddress(bytes.Repeat([]byte{0xAB}, 20))
	stored := strings.ToLower(strings.TrimPrefix(addr.String(), "0x"))
	require.NotEqual(t, addr.String(), stored, "sanity check: fixture must differ in spelling from the canonical form")

	initial := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100))
	genFile := writeSingleAccountGenesis(t, cdc, addr, stored, initial)

	require.NoError(t, genutil.AddGenesisAccount(cdc, addr, true, genFile, "50"+sdk.DefaultBondDenom, "", 0, 0, ""))

	assertSingleBalance(t, reloadBankGenesis(t, cdc, genFile), addr,
		sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 150)))
}

// TestAddGenesisAccounts_NonCanonicalStoredBalance covers the batch variant,
// whose pre-fix balance-cache was built with the same raw string match: the
// existing entry was missed, so a duplicate balance entry for the same
// account was appended instead of topping up the stored one.
func TestAddGenesisAccounts_NonCanonicalStoredBalance(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{})
	cdc := encCfg.Codec

	addr := sdk.AccAddress(bytes.Repeat([]byte{0xAB}, 20))
	stored := strings.ToLower(strings.TrimPrefix(addr.String(), "0x"))
	require.NotEqual(t, addr.String(), stored, "sanity check: fixture must differ in spelling from the canonical form")

	initial := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100))
	genFile := writeSingleAccountGenesis(t, cdc, addr, stored, initial)

	require.NoError(t, genutil.AddGenesisAccounts(cdc, []genutil.GenesisAccount{{
		Address: addr.String(),
		Coins:   sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 50)),
	}}, true, genFile))

	assertSingleBalance(t, reloadBankGenesis(t, cdc, genFile), addr,
		sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 150)))
}
