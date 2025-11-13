// Package tmservice provides compatibility for legacy client/grpc/tmservice imports
// This package has been renamed to cmtservice in cosmos-sdk v0.50+
package tmservice

import (
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
)

// Re-export all types and functions from the new cmtservice module
type (
	ServiceServer = cmtservice.ServiceServer
	ServiceClient = cmtservice.ServiceClient
)

// Re-export functions
var (
	NewServiceServer = cmtservice.NewServiceServer
	RegisterServiceServer = cmtservice.RegisterServiceServer
)
