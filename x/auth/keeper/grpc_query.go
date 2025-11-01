package keeper

import (
	"context"
	"sort"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

var _ types.QueryServer = queryServer{}

func NewQueryServer(k AccountKeeper) types.QueryServer {
	return queryServer{k: k}
}

type queryServer struct{ k AccountKeeper }

func (s queryServer) AccountAddressByID(ctx context.Context, req *types.QueryAccountAddressByIDRequest) (*types.QueryAccountAddressByIDResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if req.Id != 0 { // ignoring `0` case since it is default value.
		return nil, status.Error(codes.InvalidArgument, "requesting with id isn't supported, try to request using account-id")
	}

	accID := req.AccountId

	address, err := s.k.Accounts.Indexes.Number.MatchExact(ctx, accID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "account address not found with account number %d", accID)
	}

	return &types.QueryAccountAddressByIDResponse{AccountAddress: address.String()}, nil
}

func (s queryServer) Accounts(ctx context.Context, req *types.QueryAccountsRequest) (*types.QueryAccountsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// offset query is not allowed
	if err := query.CheckOffsetQueryNotAllowed(sdkCtx, req.Pagination); err != nil {
		return nil, err
	}
	accounts, pageRes, err := query.CollectionPaginate(
		ctx,
		s.k.Accounts,
		req.Pagination,
		func(_ sdk.AccAddress, value sdk.AccountI) (*codectypes.Any, error) {
			return codectypes.NewAnyWithValue(value)
		},
	)

	return &types.QueryAccountsResponse{Accounts: accounts, Pagination: pageRes}, err
}

// Account returns account details based on address
func (s queryServer) Account(ctx context.Context, req *types.QueryAccountRequest) (*types.QueryAccountResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "Address cannot be empty")
	}

	addr, err := sdk.AccAddressFromHexUnsafe(req.Address)
	if err != nil {
		return nil, err
	}
	account := s.k.GetAccount(ctx, addr)
	if account == nil {
		return nil, status.Errorf(codes.NotFound, "account %s not found", req.Address)
	}

	any, err := codectypes.NewAnyWithValue(account)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &types.QueryAccountResponse{Account: any}, nil
}

// Params returns parameters of auth module
func (s queryServer) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	params := s.k.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// ModuleAccounts returns all the existing Module Accounts
func (s queryServer) ModuleAccounts(c context.Context, req *types.QueryModuleAccountsRequest) (*types.QueryModuleAccountsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	// For deterministic output, sort the permAddrs by module name.
	sortedPermAddrs := make([]string, 0, len(s.k.permAddrs))
	for moduleName := range s.k.permAddrs {
		sortedPermAddrs = append(sortedPermAddrs, moduleName)
	}
	sort.Strings(sortedPermAddrs)

	modAccounts := make([]*codectypes.Any, 0, len(s.k.permAddrs))

	for _, moduleName := range sortedPermAddrs {
		account := s.k.GetModuleAccount(ctx, moduleName)
		if account == nil {
			return nil, status.Errorf(codes.NotFound, "account %s not found", moduleName)
		}
		any, err := codectypes.NewAnyWithValue(account)
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		modAccounts = append(modAccounts, any)
	}

	return &types.QueryModuleAccountsResponse{Accounts: modAccounts}, nil
}

// ModuleAccountByName returns module account by module name
func (s queryServer) ModuleAccountByName(c context.Context, req *types.QueryModuleAccountByNameRequest) (*types.QueryModuleAccountByNameResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if len(req.Name) == 0 {
		return nil, status.Error(codes.InvalidArgument, "module name is empty")
	}

	ctx := sdk.UnwrapSDKContext(c)
	moduleName := req.Name

	account := s.k.GetModuleAccount(ctx, moduleName)
	if account == nil {
		return nil, status.Errorf(codes.NotFound, "account %s not found", moduleName)
	}
	any, err := codectypes.NewAnyWithValue(account)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &types.QueryModuleAccountByNameResponse{Account: any}, nil
}

// AccountInfo implements the AccountInfo query.
func (s queryServer) AccountInfo(ctx context.Context, req *types.QueryAccountInfoRequest) (*types.QueryAccountInfoResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	addr, err := sdk.AccAddressFromHexUnsafe(req.Address)
	if err != nil {
		return nil, err
	}

	account := s.k.GetAccount(ctx, addr)
	if account == nil {
		return nil, status.Errorf(codes.NotFound, "account %s not found", req.Address)
	}

	// if there is no public key, avoid serializing the nil value
	pubKey := account.GetPubKey()
	var pkAny *codectypes.Any
	if pubKey != nil {
		pkAny, err = codectypes.NewAnyWithValue(account.GetPubKey())
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
	}

	return &types.QueryAccountInfoResponse{
		Info: &types.BaseAccount{
			Address:       req.Address,
			PubKey:        pkAny,
			AccountNumber: account.GetAccountNumber(),
			Sequence:      account.GetSequence(),
		},
	}, nil
}
