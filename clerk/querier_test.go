package clerk_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/maticnetwork/heimdall/helper"

	ethTypes "github.com/maticnetwork/bor/core/types"
	"github.com/maticnetwork/heimdall/app"
	"github.com/maticnetwork/heimdall/clerk"
	"github.com/maticnetwork/heimdall/clerk/types"
	"github.com/maticnetwork/heimdall/helper/mocks"
	hmTypes "github.com/maticnetwork/heimdall/types"
	"github.com/maticnetwork/heimdall/types/simulation"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
)

// QuerierTestSuite integrate test suite context object
type QuerierTestSuite struct {
	suite.Suite

	app            *app.HeimdallApp
	ctx            sdk.Context
	querier        sdk.Querier
	contractCaller mocks.IContractCaller
}

// SetupTest setup all necessary things for querier tesing
func (suite *QuerierTestSuite) SetupTest() {
	suite.app, suite.ctx = createTestApp(false)
	suite.contractCaller = mocks.IContractCaller{}
	suite.querier = clerk.NewQuerier(suite.app.ClerkKeeper, &suite.contractCaller)
}

// TestQuerierTestSuite
func TestQuerierTestSuite(t *testing.T) {
	suite.Run(t, new(QuerierTestSuite))
}

// TestInvalidQuery checks request query
func (suite *QuerierTestSuite) TestInvalidQuery() {
	t, _, ctx, querier := suite.T(), suite.app, suite.ctx, suite.querier

	req := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	bz, err := querier(ctx, []string{"other"}, req)
	require.Error(t, err)
	require.Nil(t, bz)

	bz, err = querier(ctx, []string{types.QuerierRoute}, req)
	require.Error(t, err)
	require.Nil(t, bz)

}

func (suite *QuerierTestSuite) TestHandleQueryRecord() {
	t, app, ctx, querier := suite.T(), suite.app, suite.ctx, suite.querier

	path := []string{types.QueryRecord}
	route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryRecord)

	req := abci.RequestQuery{
		Path: route,
		Data: []byte{},
	}
	_, err := querier(ctx, path, req)
	require.Error(t, err, "failed to parse params")

	// test wrong root chain type
	req = abci.RequestQuery{
		Path: route,
		Data: app.Codec().MustMarshalJSON(types.NewQueryRecordRootChainParams(2, "wrongRootChain")),
	}
	_, err = querier(ctx, path, req)
	require.Error(t, err, "should throw wrong rootchainType error")

	// correct root chain type
	req = abci.RequestQuery{
		Path: route,
		Data: app.Codec().MustMarshalJSON(types.NewQueryRecordRootChainParams(2, hmTypes.RootChainTypeEth)),
	}
	_, err = querier(ctx, path, req)
	require.Error(t, err, "could not get state record")

	req = abci.RequestQuery{
		Path: route,
		Data: app.Codec().MustMarshalJSON(types.NewQueryRecordRootChainParams(2, hmTypes.RootChainTypeTron)),
	}
	_, err = querier(ctx, path, req)
	require.Error(t, err, "could not get state record")

	hAddr := hmTypes.BytesToHeimdallAddress([]byte("some-address"))
	hHash := hmTypes.BytesToHeimdallHash([]byte("some-address"))
	testRecord1 := types.NewEventRecord(hHash, 1, 1, hAddr, make([]byte, 0), "1", time.Now(), hmTypes.RootChainTypeEth)

	// eth
	// SetEventRecord
	ck := app.ClerkKeeper
	ck.SetEventRecord(ctx, testRecord1)

	req = abci.RequestQuery{
		Path: route,
		Data: app.Codec().MustMarshalJSON(types.NewQueryRecordRootChainParams(1, hmTypes.RootChainTypeEth)),
	}
	record, err := querier(ctx, path, req)
	require.NoError(t, err)
	require.NotNil(t, record)

	// tron
	testRecord2 := types.NewEventRecord(hHash, 1, 1, hAddr, make([]byte, 0), "1", time.Now(), hmTypes.RootChainTypeTron)
	ck.SetEventRecord(ctx, testRecord2)

	req = abci.RequestQuery{
		Path: route,
		Data: app.Codec().MustMarshalJSON(types.NewQueryRecordParams(2)),
	}
	// SetEventRecord
	record, err = querier(ctx, path, req)
	require.NoError(t, err)
	require.NotNil(t, record)

	var recordTron types.EventRecord
	errs := json.Unmarshal(record, &recordTron)
	require.NoError(t, errs)
	require.Equal(t, recordTron.ID, uint64(2), "msg id should one")
	require.Equal(t, recordTron.RootChainType, hmTypes.RootChainTypeTron, "root chain type should be tron")
}

func (suite *QuerierTestSuite) TestHandleQueryRecordList() {
	t, app, ctx, querier := suite.T(), suite.app, suite.ctx, suite.querier

	path := []string{types.QueryRecordList}
	route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryRecordList)

	req := abci.RequestQuery{
		Path: route,
		Data: []byte{},
	}
	_, err := querier(ctx, path, req)
	require.Error(t, err, "failed to parse params")

	hAddr := hmTypes.BytesToHeimdallAddress([]byte("some-address"))
	hHash := hmTypes.BytesToHeimdallHash([]byte("some-address"))
	testRecord1 := types.NewEventRecord(hHash, 1, 1, hAddr, make([]byte, 0), "1", time.Now(), hmTypes.RootChainTypeEth)
	testRecord2 := types.NewEventRecord(hHash, 1, 1, hAddr, make([]byte, 0), "1", time.Now(), hmTypes.RootChainTypeTron)
	testRecord3 := types.NewEventRecord(hHash, 1, 2, hAddr, make([]byte, 0), "1", time.Now(), hmTypes.RootChainTypeEth)
	testRecord4 := types.NewEventRecord(hHash, 1, 2, hAddr, make([]byte, 0), "1", time.Now(), hmTypes.RootChainTypeTron)

	// SetEventRecord
	ck := app.ClerkKeeper
	ck.SetEventRecord(ctx, testRecord1)
	ck.SetEventRecord(ctx, testRecord2)
	ck.SetEventRecord(ctx, testRecord3)
	ck.SetEventRecord(ctx, testRecord4)

	req = abci.RequestQuery{
		Path: route,
		Data: app.Codec().MustMarshalJSON(hmTypes.NewQueryPaginationParams(1, 4, "")),
	}
	record, err := querier(ctx, path, req)
	require.NoError(t, err)
	require.NotNil(t, record)
	var recordList []types.EventRecord
	errs := json.Unmarshal(record, &recordList)
	require.NoError(t, errs, "should unmarshall EventRecord array")
	require.Equal(t, len(recordList), 4, "should have four record")
	require.Equal(t, recordList[0].ID, uint64(1), "msg id should one")
	require.Equal(t, recordList[1].ID, uint64(2), "msg id should two")
	require.Equal(t, recordList[2].ID, uint64(3), "msg id should three")
	require.Equal(t, recordList[3].ID, uint64(4), "msg id should four")

}

func (suite *QuerierTestSuite) TestHandleQueryRecordSequence() {
	t, app, ctx, querier := suite.T(), suite.app, suite.ctx, suite.querier

	path := []string{types.QueryRecordSequence}
	route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryRecordSequence)

	req := abci.RequestQuery{
		Path: route,
		Data: []byte{},
	}
	_, err := querier(ctx, path, req)
	require.Error(t, err, "failed to parse params")

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	txHash := hmTypes.HexToHeimdallHash("123")
	index := simulation.RandIntBetween(r1, 0, 100)
	logIndex := uint64(index)
	rootChainType := hmTypes.RootChainTypeEth
	chainParams := app.ChainKeeper.GetParams(ctx)
	txreceipt := &ethTypes.Receipt{
		BlockNumber: big.NewInt(1),
	}
	suite.contractCaller.On("GetConfirmedTxReceipt", txHash.EthHash(), chainParams.MainchainTxConfirmations, rootChainType).
		Return(nil, errors.New("err confirmed txn receipt"))

	req = abci.RequestQuery{
		Path: route,
		Data: app.Codec().MustMarshalJSON(types.NewQueryRecordSequenceParams("123", logIndex, rootChainType)),
	}
	_, err = querier(ctx, path, req)
	require.NotNil(t, err, "failed to parse params")

	index = simulation.RandIntBetween(r1, 0, 100)
	logIndex = uint64(index)
	txHash = hmTypes.HexToHeimdallHash("1234")
	suite.contractCaller.On("GetConfirmedTxReceipt", txHash.EthHash(), chainParams.MainchainTxConfirmations, rootChainType).Return(txreceipt, nil)
	req = abci.RequestQuery{
		Path: route,
		Data: app.Codec().MustMarshalJSON(types.NewQueryRecordSequenceParams("1234", logIndex, rootChainType)),
	}
	resp, err := querier(ctx, path, req)
	require.Nil(t, err)
	require.Nil(t, resp)

	testSeq := helper.CalculateSequence(big.NewInt(1), 1, rootChainType).String()
	ck := app.ClerkKeeper
	ck.SetRecordSequence(ctx, testSeq)
	logIndex = uint64(1)
	txHash = hmTypes.HexToHeimdallHash("12345")
	suite.contractCaller.On("GetConfirmedTxReceipt", txHash.EthHash(), chainParams.MainchainTxConfirmations, rootChainType).Return(txreceipt, nil)
	req = abci.RequestQuery{
		Path: route,
		Data: app.Codec().MustMarshalJSON(types.NewQueryRecordSequenceParams("12345", logIndex, rootChainType)),
	}
	resp, err = querier(ctx, path, req)
	require.Nil(t, err)
	require.NotNil(t, resp)

	// tron
	testSeq = helper.CalculateSequence(big.NewInt(1), 1, hmTypes.RootChainTypeTron).String()
	ck.SetRecordSequence(ctx, testSeq)
	suite.contractCaller.On("GetTronTransactionReceipt", txHash.TronHash().String()).Return(txreceipt, nil)
	req = abci.RequestQuery{
		Path: route,
		Data: app.Codec().MustMarshalJSON(types.NewQueryRecordSequenceParams("12345", logIndex, hmTypes.RootChainTypeTron)),
	}
	resp, err = querier(ctx, path, req)
	require.Nil(t, err)
	require.NotNil(t, resp)

}
