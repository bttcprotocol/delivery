package processor

import (
	"context"
	cliContext "github.com/cosmos/cosmos-sdk/client/context"
	"github.com/maticnetwork/heimdall/bridge/setu/util"
	"github.com/stretchr/testify/mock"
	"github.com/tendermint/tendermint/libs/log"
	"testing"
	"time"
)

// MockLogger is a mock implementation of the Logger interface
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Debug(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Error(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func TestStartPolling(t *testing.T) {
	logger := util.Logger().With("service", "processor", "module", "checkpoint")
	if logger == nil {
		logger = log.NewNopLogger()
	}
	// Setup mock CheckpointProcessor
	cp := &CheckpointProcessor{
		BaseProcessor: BaseProcessor{
			Logger: logger,
			cliCtx: cliContext.CLIContext{},
		},
	}
	ackCtx, cancelNoACKPolling := context.WithCancel(context.Background())
	cp.cancelNoACKPolling = cancelNoACKPolling
	go cp.startPolling(ackCtx)
	time.Sleep(10 * time.Second)
	cancelNoACKPolling()
	logger.Info("Polling cancelled and test Stopped")
	//need to see the log "Polling cancelled while waiting for checkpoint params"

}

// function GetTronDynamicCheckpointFeature should change to the following format temporarily : var GetTronDynamicCheckpointFeature = func(cliCtx cliContext.CLIContext) (*featureManagerTypes.PlainFeatureData, error)
//func TestGetTronDynamicCheckpointProposal(t *testing.T) {
//	// Setup mock logger
//	mockLogger := new(MockLogger)
//	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()
//
//	logger := util.Logger().With("service", "processor", "module", "checkpoint")
//	if logger == nil {
//		logger = log.NewNopLogger()
//	}
//	// Setup mock CheckpointProcessor
//	cp := &CheckpointProcessor{
//		BaseProcessor: BaseProcessor{
//			Logger: logger,
//			cliCtx: cliContext.CLIContext{},
//		},
//	}
//
//	// Test case 1: Successful feature fetch
//	t.Run("Successful feature fetch", func(t *testing.T) {
//		// Mock util.GetTronDynamicCheckpointFeature
//		originalGetTronDynamicCheckpointFeature := util.GetTronDynamicCheckpointFeature
//		defer func() { util.GetTronDynamicCheckpointFeature = originalGetTronDynamicCheckpointFeature }()
//		util.GetTronDynamicCheckpointFeature = func(cliCtx cliContext.CLIContext) (*featureManagerTypes.PlainFeatureData, error) {
//			return &featureManagerTypes.PlainFeatureData{
//				IsOpen: true,
//				IntConf: map[string]int{
//					"maxLength": 1024,
//				},
//			}, nil
//		}
//
//		isOpen, maxLength := cp.getTronDynamicCheckpointProposal()
//		assert.True(t, isOpen)
//		assert.Equal(t, 1024, maxLength)
//	})
//
//	// Test case 2: Error fetching feature
//	t.Run("Error fetching feature", func(t *testing.T) {
//		// Mock util.GetTronDynamicCheckpointFeature
//		originalGetTronDynamicCheckpointFeature := util.GetTronDynamicCheckpointFeature
//		defer func() { util.GetTronDynamicCheckpointFeature = originalGetTronDynamicCheckpointFeature }()
//		util.GetTronDynamicCheckpointFeature = func(cliCtx cliContext.CLIContext) (*featureManagerTypes.PlainFeatureData, error) {
//			return nil, errors.New("some error")
//		}
//
//		isOpen, maxLength := cp.getTronDynamicCheckpointProposal()
//		assert.False(t, isOpen)
//		assert.Equal(t, 0, maxLength)
//		//mockLogger.AssertCalled(t, "Error", "Error while fetching dynamic checkpoint feature", "error", errors.New("some error"))
//	})
//}

//// function GetCheckpointParams should change to the following format temporarily : var GetCheckpointParams = func(cliCtx cliContext.CLIContext) (*checkpointTypes.Params, error)
//// so do the function GetTronDynamicCheckpointFeature
//func TestGetCheckpointPollTimeWithTronDynamic(t *testing.T) {
//	// Setup mock logger
//	mockLogger := new(MockLogger)
//	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()
//
//	logger := util.Logger().With("service", "processor", "module", "checkpoint")
//	if logger == nil {
//		logger = log.NewNopLogger()
//	}
//	// Setup mock CheckpointProcessor
//	cp := &CheckpointProcessor{
//		BaseProcessor: BaseProcessor{
//			Logger: logger,
//			cliCtx: cliContext.CLIContext{},
//		},
//	}
//
//	// Test case 1: Successful feature fetch
//	t.Run("Successful feature fetch", func(t *testing.T) {
//		// Mock util.GetTronDynamicCheckpointFeature
//		originalGetTronDynamicCheckpointFeature := util.GetTronDynamicCheckpointFeature
//		defer func() { util.GetTronDynamicCheckpointFeature = originalGetTronDynamicCheckpointFeature }()
//		util.GetTronDynamicCheckpointFeature = func(cliCtx cliContext.CLIContext) (*featureManagerTypes.PlainFeatureData, error) {
//			return &featureManagerTypes.PlainFeatureData{
//				IsOpen: true,
//				IntConf: map[string]int{
//					"maxLength": 1024,
//				},
//			}, nil
//		}
//
//		originalGetCheckpointParams := util.GetCheckpointParams
//		defer func() { util.GetCheckpointParams = originalGetCheckpointParams }()
//		util.GetCheckpointParams = func(cliCtx cliContext.CLIContext) (*checkpointTypes.Params, error) {
//			return &checkpointTypes.Params{
//				CheckpointBufferTime:   1000 * time.Second,
//				CheckpointPollInterval: 10 * time.Minute,
//			}, nil
//		}
//
//		pollTime, err := cp.GetCheckpointPollTime()
//		assert.Equal(t, 10*time.Minute, pollTime)
//		assert.Nil(t, err)
//	})
//}
//
//// function GetCheckpointParams should change to the following format temporarily : var GetCheckpointParams = func(cliCtx cliContext.CLIContext) (*checkpointTypes.Params, error)
//// so do the function GetTronDynamicCheckpointFeature
//func TestGetCheckpointPollTimeWithoutTronDynamic(t *testing.T) {
//	helper.SetTestConfig(helper.GetDefaultHeimdallConfig())
//	// Setup mock logger
//	mockLogger := new(MockLogger)
//	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()
//
//	logger := util.Logger().With("service", "processor", "module", "checkpoint")
//	if logger == nil {
//		logger = log.NewNopLogger()
//	}
//	// Setup mock CheckpointProcessor
//	cp := &CheckpointProcessor{
//		BaseProcessor: BaseProcessor{
//			Logger: logger,
//			cliCtx: cliContext.CLIContext{},
//		},
//	}
//
//	// Test case 1: Successful feature fetch
//	t.Run("Successful feature fetch", func(t *testing.T) {
//		// Mock util.GetTronDynamicCheckpointFeature
//		originalGetTronDynamicCheckpointFeature := util.GetTronDynamicCheckpointFeature
//		defer func() { util.GetTronDynamicCheckpointFeature = originalGetTronDynamicCheckpointFeature }()
//		util.GetTronDynamicCheckpointFeature = func(cliCtx cliContext.CLIContext) (*featureManagerTypes.PlainFeatureData, error) {
//			return &featureManagerTypes.PlainFeatureData{
//				IsOpen: false,
//			}, nil
//		}
//
//		originalGetCheckpointParams := util.GetCheckpointParams
//		defer func() { util.GetCheckpointParams = originalGetCheckpointParams }()
//		util.GetCheckpointParams = func(cliCtx cliContext.CLIContext) (*checkpointTypes.Params, error) {
//			return &checkpointTypes.Params{
//				CheckpointBufferTime:   1000 * time.Second,
//				CheckpointPollInterval: 10 * time.Minute,
//			}, nil
//		}
//
//		pollTime, err := cp.GetCheckpointPollTime()
//		assert.Equal(t, 30*time.Minute, pollTime)
//		assert.Nil(t, err)
//	})
//}
