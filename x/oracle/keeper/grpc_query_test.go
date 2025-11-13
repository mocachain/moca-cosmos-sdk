package keeper_test

import (
	gocontext "context"

	"github.com/zkMeLabs/moca-cosmos-sdk/x/oracle/types"
)

func (s *TestSuite) TestQueryParams() {
	res, err := s.queryClient.Params(gocontext.Background(), &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Equal(s.oracleKeeper.GetParams(s.ctx), res.GetParams())
}
