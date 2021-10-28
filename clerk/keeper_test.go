package clerk_test

import (
	"testing"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/maticnetwork/heimdall/app"
	"github.com/maticnetwork/heimdall/clerk"
	"github.com/maticnetwork/heimdall/clerk/types"
	hmTypes "github.com/maticnetwork/heimdall/types"
)

//
// Test suite
//

// KeeperTestSuite integrate test suite context object
type KeeperTestSuite struct {
	suite.Suite

	app *app.HeimdallApp
	ctx sdk.Context
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app, suite.ctx = createTestApp(false)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

//
// Tests
//

func (suite *KeeperTestSuite) TestHasGetSetEventRecords() {
	t, app, ctx := suite.T(), suite.app, suite.ctx

	hAddr := hmTypes.BytesToHeimdallAddress([]byte("some-address"))
	hHash := hmTypes.BytesToHeimdallHash([]byte("some-address"))
	EthRecord := types.NewEventRecord(hHash, 1, 1, hAddr, make([]byte, 0), "1", time.Now(), hmTypes.RootChainTypeEth)

	// SetEventRecord
	ck := app.ClerkKeeper
	err := ck.SetEventRecord(ctx, EthRecord)
	require.Nil(t, err)

	TronRecord := types.NewEventRecord(hHash, 1, 1, hAddr, make([]byte, 0), "1", time.Now(), hmTypes.RootChainTypeTron)

	err = ck.SetEventRecord(ctx, TronRecord)
	require.Nil(t, err)

	// GetEventRecord
	latestID := ck.GetLatestID(ctx)
	respRecord, err := ck.GetEventRecord(ctx, latestID)
	require.Nil(t, err)
	require.Equal(t, (*respRecord).ID, latestID)

	respRecord, err = ck.GetRootChainEventRecord(ctx, EthRecord.ID, EthRecord.RootChainType)
	require.Nil(t, err)
	require.Equal(t, (*respRecord).ID, EthRecord.ID)

	respRecord, err = ck.GetRootChainEventRecord(ctx, TronRecord.ID, TronRecord.RootChainType)
	require.Nil(t, err)
	require.Equal(t, (*respRecord).ID, TronRecord.ID)

	respRecord, err = ck.GetEventRecord(ctx, latestID+1)
	require.NotNil(t, err)

	respRecord, err = ck.GetRootChainEventRecord(ctx, EthRecord.ID+1, EthRecord.RootChainType)
	require.NotNil(t, err)

	respRecord, err = ck.GetRootChainEventRecord(ctx, TronRecord.ID+1, TronRecord.RootChainType)
	require.NotNil(t, err)

	// HasEventRecord
	recordPresent := ck.HasEventRecord(ctx, latestID)
	require.True(t, recordPresent)

	recordPresent = ck.HasRootChainEventRecord(ctx, EthRecord.RootChainType, EthRecord.ID)
	require.True(t, recordPresent)

	recordPresent = ck.HasRootChainEventRecord(ctx, TronRecord.RootChainType, TronRecord.ID)
	require.True(t, recordPresent)

	recordPresent = ck.HasEventRecord(ctx, latestID+1)
	require.False(t, recordPresent)

	recordPresent = ck.HasRootChainEventRecord(ctx, EthRecord.RootChainType, EthRecord.ID+1)
	require.False(t, recordPresent)

	recordPresent = ck.HasRootChainEventRecord(ctx, TronRecord.RootChainType, TronRecord.ID+1)
	require.False(t, recordPresent)

	recordList := ck.GetAllEventRecords(ctx)
	require.Len(t, recordList, 2)
}

func (suite *KeeperTestSuite) TestGetEventRecordList() {
	t, app, ctx := suite.T(), suite.app, suite.ctx
	var i uint64

	hAddr := hmTypes.BytesToHeimdallAddress([]byte("some-address"))
	hHash := hmTypes.BytesToHeimdallHash([]byte("some-address"))
	ck := app.ClerkKeeper
	for i = 0; i < 60; i++ {
		testRecord := types.NewEventRecord(hHash, i, i, hAddr, make([]byte, 0), "1", time.Now(), hmTypes.RootChainTypeStake)
		ck.SetEventRecord(ctx, testRecord)
	}

	recordList, _ := ck.GetEventRecordList(ctx, 1, 20)
	require.Len(t, recordList, 20)

	recordList, _ = ck.GetEventRecordList(ctx, 2, 20)
	require.Len(t, recordList, 20)

	recordList, _ = ck.GetEventRecordList(ctx, 3, 30)
	require.Len(t, recordList, 0)

	recordList, _ = ck.GetEventRecordList(ctx, 1, 70)
	require.Len(t, recordList, 50)

	recordList, _ = ck.GetEventRecordList(ctx, 2, 60)
	require.Len(t, recordList, 10)
}

func (suite *KeeperTestSuite) TestGetEventRecordListTime() {
	t, app, ctx := suite.T(), suite.app, suite.ctx
	var i uint64

	hAddr := hmTypes.BytesToHeimdallAddress([]byte("some-address"))
	hHash := hmTypes.BytesToHeimdallHash([]byte("some-address"))
	ck := app.ClerkKeeper
	for i = 0; i < 30; i++ {
		testRecord := types.NewEventRecord(hHash, i, i, hAddr, make([]byte, 0), "1", time.Unix(int64(i), 0), hmTypes.RootChainTypeStake)
		ck.SetEventRecord(ctx, testRecord)
	}
	ctx = ctx.WithBlockHeader(abci.Header{
		Time: time.Unix(30, 0),
	})
	recordList, err := ck.GetEventRecordListWithTime(ctx, time.Unix(1, 0), time.Unix(6, 0), 0, 0)
	require.NoError(t, err)
	require.Len(t, recordList, 5)
	require.Equal(t, int64(5), recordList[len(recordList)-1].RecordTime.Unix())

	recordList, err = ck.GetEventRecordListWithTime(ctx, time.Unix(1, 0), time.Unix(6, 0), 1, 1)
	require.NoError(t, err)
	require.Len(t, recordList, 1)

	recordList, err = ck.GetEventRecordListWithTime(ctx, time.Unix(10, 0), time.Unix(20, 0), 0, 0)
	require.NoError(t, err)
	require.Len(t, recordList, 10)
	require.Equal(t, int64(10), recordList[0].RecordTime.Unix())
	require.Equal(t, int64(19), recordList[len(recordList)-1].RecordTime.Unix())
}

func (suite *KeeperTestSuite) TestGetEventRecordKey() {
	t, _, _ := suite.T(), suite.app, suite.ctx

	hAddr := hmTypes.BytesToHeimdallAddress([]byte("some-address"))
	hHash := hmTypes.BytesToHeimdallHash([]byte("some-address"))
	testRecord1 := types.NewEventRecord(hHash, 1, 1, hAddr, make([]byte, 0), "1", time.Now(), hmTypes.RootChainTypeStake)

	respKey := clerk.GetEventRecordKey(testRecord1.ID)
	require.Equal(t, respKey, []byte{17, 49})
}

func (suite *KeeperTestSuite) TestGetRootToHeimdallIdKey() {
	t, _, _ := suite.T(), suite.app, suite.ctx

	hAddr := hmTypes.BytesToHeimdallAddress([]byte("some-address"))
	hHash := hmTypes.BytesToHeimdallHash([]byte("some-address"))
	testRecord1 := types.NewEventRecord(hHash, 1, 1, hAddr, make([]byte, 0), "1", time.Now(), hmTypes.RootChainTypeEth)

	respKey := clerk.GetRootToHeimdallIdKey(testRecord1.RootChainType, testRecord1.ID)
	require.Equal(t, respKey, []byte{25, 101, 116, 104, 49})

}

func (suite *KeeperTestSuite) TestSetHasGetRecordSequence() {
	t, app, ctx := suite.T(), suite.app, suite.ctx

	testSeq := "testseq"
	ck := app.ClerkKeeper
	ck.SetRecordSequence(ctx, testSeq)
	found := ck.HasRecordSequence(ctx, testSeq)
	require.True(t, found)

	found = ck.HasRecordSequence(ctx, "testSeq")
	require.False(t, found)

	recordSequences := ck.GetRecordSequences(ctx)
	require.Len(t, recordSequences, 1)
}
