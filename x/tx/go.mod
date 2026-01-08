module cosmossdk.io/x/tx

go 1.23.2

require (
	cosmossdk.io/api v0.7.4
	cosmossdk.io/core v0.0.0-00010101000000-000000000000
	cosmossdk.io/errors v1.0.1
	cosmossdk.io/math v1.3.0
	github.com/cosmos/cosmos-proto v1.0.0-beta.5
	github.com/cosmos/gogoproto v1.7.0
	github.com/google/go-cmp v0.6.0
	github.com/google/gofuzz v1.2.0
	github.com/iancoleman/strcase v0.3.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.9.0
	github.com/tendermint/go-amino v0.16.0
	google.golang.org/protobuf v1.34.2
	gotest.tools/v3 v3.5.1
	pgregory.net/rapid v1.1.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/exp v0.0.0-20240222234643-814bf88cf225 // indirect
	golang.org/x/net v0.27.0 // indirect
	golang.org/x/sys v0.22.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240318140521-94a12d6c2237 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240709173604-40e1e62336c5 // indirect
	google.golang.org/grpc v1.64.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// NOTE: we do not want to replace to the development version of cosmossdk.io/api yet
// Until https://github.com/cosmos/cosmos-sdk/issues/19228 is resolved
// We are tagging x/tx v0.14+ from main and v0.13 from release/v0.50.x and must keep using released versions of x/tx dependencies
replace (
	cosmossdk.io/api => ../../api
	cosmossdk.io/client/v2 => ../../client/v2
	cosmossdk.io/core => ../../core
	cosmossdk.io/errors => ../../errors
	// cosmossdk.io/log => ../../log
	// cosmossdk.io/math => ../../math
	cosmossdk.io/simapp => ../../simapp
	cosmossdk.io/store => ../../store
	cosmossdk.io/tools/confix => ../../tools/confix
	cosmossdk.io/x/circuit => ../../x/circuit
	cosmossdk.io/x/evidence => ../../x/evidence
	cosmossdk.io/x/feegrant => ../../x/feegrant
	cosmossdk.io/x/nft => ../../x/nft
	cosmossdk.io/x/upgrade => ../../x/upgrade
	github.com/0xPolygon/polygon-edge v1.3.3 => github.com/MocaFoundation/polygon-edge v1.3.3-moca.1
	github.com/99designs/keyring => github.com/cosmos/keyring v1.2.0
	github.com/btcsuite/btcd => github.com/btcsuite/btcd v0.22.1
	github.com/cometbft/cometbft => ../../../moca-cometbft
	github.com/cometbft/cometbft-db => ../../../moca-cometbft-db
	github.com/cosmos/cosmos-sdk => ../../.
	github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.9.1
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	github.com/syndtr/goleveldb => github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
	golang.org/x/exp => golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842
)
