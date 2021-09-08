package clerk_test

import (
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/maticnetwork/heimdall/helper"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkAuth "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/maticnetwork/heimdall/app"
	"github.com/maticnetwork/heimdall/clerk"
	"github.com/maticnetwork/heimdall/clerk/types"
	"github.com/maticnetwork/heimdall/common"
	"github.com/maticnetwork/heimdall/helper/mocks"
	hmTypes "github.com/maticnetwork/heimdall/types"
)

//
// Test suite
//

type HandlerTestSuite struct {
	suite.Suite

	app            *app.HeimdallApp
	ctx            sdk.Context
	cliCtx         context.CLIContext
	chainID        string
	handler        sdk.Handler
	contractCaller mocks.IContractCaller
	r              *rand.Rand
	rootChainType  string
}

func (suite *HandlerTestSuite) SetupTest() {
	suite.app, suite.ctx = createTestApp(false)
	suite.contractCaller = mocks.IContractCaller{}
	suite.handler = clerk.NewHandler(suite.app.ClerkKeeper, &suite.contractCaller)

	// fetch chain id
	suite.chainID = suite.app.ChainKeeper.GetParams(suite.ctx).ChainParams.BorChainID

	// random generator
	s1 := rand.NewSource(time.Now().UnixNano())
	suite.r = rand.New(s1)
	suite.rootChainType = hmTypes.RootChainTypeStake

}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

//
// Test cases
//

func (suite *HandlerTestSuite) TestHandleMsgEventRecord() {
	t, app, ctx, chainID, r, rChainType := suite.T(), suite.app, suite.ctx, suite.chainID, suite.r, suite.rootChainType

	// keys and addresses
	_, _, addr1 := sdkAuth.KeyTestPubAddr()

	id := r.Uint64()
	logIndex := r.Uint64()
	blockNumber := r.Uint64()

	// successful message
	msg := types.NewMsgEventRecord(
		hmTypes.BytesToHeimdallAddress(addr1.Bytes()),
		hmTypes.HexToHeimdallHash("123"),
		logIndex,
		blockNumber,
		id,
		hmTypes.BytesToHeimdallAddress(addr1.Bytes()),
		make([]byte, 0),
		chainID,
		rChainType,
	)

	t.Run("Success", func(t *testing.T) {
		result := suite.handler(ctx, msg)
		require.True(t, result.IsOK(), "expected msg record to be ok, got %v", result)

		// there should be no stored event record
		storedEventRecord, err := app.ClerkKeeper.GetEventRecord(ctx, id)
		require.Nil(t, storedEventRecord)
		require.Error(t, err)
	})

	t.Run("ExistingRecord", func(t *testing.T) {
		// store event record in keeper
		app.ClerkKeeper.SetEventRecord(ctx,
			types.NewEventRecord(
				msg.TxHash,
				msg.LogIndex,
				msg.ID,
				msg.ContractAddress,
				msg.Data,
				msg.ChainID,
				time.Now(),
				hmTypes.RootChainTypeStake,
			),
		)

		result := suite.handler(ctx, msg)
		require.False(t, result.IsOK(), "should fail due to existent event record but succeeded")
		require.Equal(t, types.CodeEventRecordAlreadySynced, result.Code)
		// tron chain
		app.ClerkKeeper.SetEventRecord(ctx,
			types.NewEventRecord(
				msg.TxHash,
				msg.LogIndex,
				msg.ID,
				msg.ContractAddress,
				msg.Data,
				msg.ChainID,
				time.Now(),
				hmTypes.RootChainTypeStake,
			),
		)
		result = suite.handler(ctx, msg)
		require.False(t, result.IsOK(), "should fail due to existent event record but succeeded")
		require.Equal(t, types.CodeEventRecordAlreadySynced, result.Code)

	})
}

func (suite *HandlerTestSuite) TestHandleMsgEventRecordSequence() {
	t, app, ctx, chainID, r, rChainType := suite.T(), suite.app, suite.ctx, suite.chainID, suite.r, suite.rootChainType

	_, _, addr1 := sdkAuth.KeyTestPubAddr()
	// eth chain
	msg := types.NewMsgEventRecord(
		hmTypes.BytesToHeimdallAddress(addr1.Bytes()),
		hmTypes.HexToHeimdallHash("123"),
		r.Uint64(),
		r.Uint64(),
		r.Uint64(),
		hmTypes.BytesToHeimdallAddress(addr1.Bytes()),
		make([]byte, 0),
		chainID,
		rChainType,
	)

	// sequence id
	blockNumber := new(big.Int).SetUint64(msg.BlockNumber)
	sequence := helper.CalculateSequence(blockNumber, msg.LogIndex, msg.RootChainType)
	app.ClerkKeeper.SetRecordSequence(ctx, sequence.String())

	result := suite.handler(ctx, msg)
	require.False(t, result.IsOK(), "should fail due to existent sequence but succeeded")
	require.Equal(t, common.CodeOldTx, result.Code)
	// tron chain
	msg = types.NewMsgEventRecord(
		hmTypes.BytesToHeimdallAddress(addr1.Bytes()),
		hmTypes.HexToHeimdallHash("123"),
		r.Uint64(),
		r.Uint64(),
		r.Uint64(),
		hmTypes.BytesToHeimdallAddress(addr1.Bytes()),
		make([]byte, 0),
		chainID,
		hmTypes.RootChainTypeTron,
	)

	// sequence id
	blockNumber = new(big.Int).SetUint64(msg.BlockNumber)
	sequence = helper.CalculateSequence(blockNumber, msg.LogIndex, msg.RootChainType)
	app.ClerkKeeper.SetRecordSequence(ctx, sequence.String())

	result = suite.handler(ctx, msg)
	require.False(t, result.IsOK(), "should fail due to existent sequence but succeeded")
	require.Equal(t, common.CodeOldTx, result.Code)

}

func (suite *HandlerTestSuite) TestHandleMsgEventRecordChainID() {
	t, app, ctx, r, rChainType := suite.T(), suite.app, suite.ctx, suite.r, suite.rootChainType

	_, _, addr1 := sdkAuth.KeyTestPubAddr()

	id := r.Uint64()

	// wrong chain id
	msg := types.NewMsgEventRecord(
		hmTypes.BytesToHeimdallAddress(addr1.Bytes()),
		hmTypes.HexToHeimdallHash("123"),
		r.Uint64(),
		r.Uint64(),
		id,
		hmTypes.BytesToHeimdallAddress(addr1.Bytes()),
		make([]byte, 0),
		"random chain id",
		rChainType,
	)
	result := suite.handler(ctx, msg)
	require.False(t, result.IsOK(), "error invalid bor chain id %v", result.Code)
	require.Equal(t, common.CodeInvalidBorChainID, result.Code)

	// there should be no stored event record
	storedEventRecord, err := app.ClerkKeeper.GetEventRecord(ctx, id)
	require.Nil(t, storedEventRecord)
	require.Error(t, err)
}
