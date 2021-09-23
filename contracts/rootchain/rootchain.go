// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package rootchain

import (
	"math/big"
	"strings"

	ethereum "github.com/maticnetwork/bor"
	"github.com/maticnetwork/bor/accounts/abi"
	"github.com/maticnetwork/bor/accounts/abi/bind"
	"github.com/maticnetwork/bor/common"
	"github.com/maticnetwork/bor/core/types"
	"github.com/maticnetwork/bor/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = abi.U256
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// RootchainABI is the input ABI used to generate the binding from.
const RootchainABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"headerBlocks\",\"outputs\":[{\"name\":\"root\",\"type\":\"bytes32\"},{\"name\":\"start\",\"type\":\"uint256\"},{\"name\":\"end\",\"type\":\"uint256\"},{\"name\":\"createdAt\",\"type\":\"uint256\"},{\"name\":\"proposer\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"data\",\"type\":\"bytes\"},{\"name\":\"sigs\",\"type\":\"uint256[3][]\"}],\"name\":\"submitCheckpoint\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"data\",\"type\":\"bytes\"},{\"name\":\"sigs\",\"type\":\"bytes\"}],\"name\":\"submitHeaderBlock\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getLastChildBlock\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"currentHeaderBlock\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"proposer\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"headerBlockId\",\"type\":\"uint256\"},{\"indexed\":true,\"name\":\"reward\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"start\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"end\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"root\",\"type\":\"bytes32\"}],\"name\":\"NewHeaderBlock\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"rootChainId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"activationHeight\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"txConfirmations\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"rootChainAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"stateSenderAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"stakingManagerAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"stakingInfoAddress\",\"type\":\"address\"}],\"name\":\"NewChain\",\"type\":\"event\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"chainMap\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"rootChainId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"activationHeight\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"txConfirmations\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"rootChainAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"stateSenderAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"stakingManagerAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"stakingInfoAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"timeStamp\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"rootChainId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"activationHeight\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"txConfirmations\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"rootChainAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"stateSenderAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"stakingManagerAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"stakingInfoAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"timeStamp\",\"type\":\"uint256\"}],\"name\":\"setChainInfo\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// Rootchain is an auto generated Go binding around an Ethereum contract.
type Rootchain struct {
	RootchainCaller     // Read-only binding to the contract
	RootchainTransactor // Write-only binding to the contract
	RootchainFilterer   // Log filterer for contract events
}

// RootchainCaller is an auto generated read-only Go binding around an Ethereum contract.
type RootchainCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RootchainTransactor is an auto generated write-only Go binding around an Ethereum contract.
type RootchainTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RootchainFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type RootchainFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RootchainSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type RootchainSession struct {
	Contract     *Rootchain        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// RootchainCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type RootchainCallerSession struct {
	Contract *RootchainCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// RootchainTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type RootchainTransactorSession struct {
	Contract     *RootchainTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// RootchainRaw is an auto generated low-level Go binding around an Ethereum contract.
type RootchainRaw struct {
	Contract *Rootchain // Generic contract binding to access the raw methods on
}

// RootchainCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type RootchainCallerRaw struct {
	Contract *RootchainCaller // Generic read-only contract binding to access the raw methods on
}

// RootchainTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type RootchainTransactorRaw struct {
	Contract *RootchainTransactor // Generic write-only contract binding to access the raw methods on
}

// NewRootchain creates a new instance of Rootchain, bound to a specific deployed contract.
func NewRootchain(address common.Address, backend bind.ContractBackend) (*Rootchain, error) {
	contract, err := bindRootchain(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Rootchain{RootchainCaller: RootchainCaller{contract: contract}, RootchainTransactor: RootchainTransactor{contract: contract}, RootchainFilterer: RootchainFilterer{contract: contract}}, nil
}

// NewRootchainCaller creates a new read-only instance of Rootchain, bound to a specific deployed contract.
func NewRootchainCaller(address common.Address, caller bind.ContractCaller) (*RootchainCaller, error) {
	contract, err := bindRootchain(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &RootchainCaller{contract: contract}, nil
}

// NewRootchainTransactor creates a new write-only instance of Rootchain, bound to a specific deployed contract.
func NewRootchainTransactor(address common.Address, transactor bind.ContractTransactor) (*RootchainTransactor, error) {
	contract, err := bindRootchain(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &RootchainTransactor{contract: contract}, nil
}

// NewRootchainFilterer creates a new log filterer instance of Rootchain, bound to a specific deployed contract.
func NewRootchainFilterer(address common.Address, filterer bind.ContractFilterer) (*RootchainFilterer, error) {
	contract, err := bindRootchain(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &RootchainFilterer{contract: contract}, nil
}

// bindRootchain binds a generic wrapper to an already deployed contract.
func bindRootchain(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(RootchainABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Rootchain *RootchainRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Rootchain.Contract.RootchainCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Rootchain *RootchainRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Rootchain.Contract.RootchainTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Rootchain *RootchainRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Rootchain.Contract.RootchainTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Rootchain *RootchainCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Rootchain.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Rootchain *RootchainTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Rootchain.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Rootchain *RootchainTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Rootchain.Contract.contract.Transact(opts, method, params...)
}

// ChainMap is a free data retrieval call binding the contract method 0xdf35ce79.
//
// Solidity: function chainMap(uint256 ) constant returns(uint256 rootChainId, uint256 activationHeight, uint256 txConfirmations, address rootChainAddress, address stateSenderAddress, address stakingManagerAddress, address stakingInfoAddress, uint256 timeStamp)
func (_Rootchain *RootchainCaller) ChainMap(opts *bind.CallOpts, arg0 *big.Int) (struct {
	RootChainId           *big.Int
	ActivationHeight      *big.Int
	TxConfirmations       *big.Int
	RootChainAddress      common.Address
	StateSenderAddress    common.Address
	StakingManagerAddress common.Address
	StakingInfoAddress    common.Address
	TimeStamp             *big.Int
}, error) {
	ret := new(struct {
		RootChainId           *big.Int
		ActivationHeight      *big.Int
		TxConfirmations       *big.Int
		RootChainAddress      common.Address
		StateSenderAddress    common.Address
		StakingManagerAddress common.Address
		StakingInfoAddress    common.Address
		TimeStamp             *big.Int
	})
	out := ret
	err := _Rootchain.contract.Call(opts, out, "chainMap", arg0)
	return *ret, err
}

// ChainMap is a free data retrieval call binding the contract method 0xdf35ce79.
//
// Solidity: function chainMap(uint256 ) constant returns(uint256 rootChainId, uint256 activationHeight, uint256 txConfirmations, address rootChainAddress, address stateSenderAddress, address stakingManagerAddress, address stakingInfoAddress, uint256 timeStamp)
func (_Rootchain *RootchainSession) ChainMap(arg0 *big.Int) (struct {
	RootChainId           *big.Int
	ActivationHeight      *big.Int
	TxConfirmations       *big.Int
	RootChainAddress      common.Address
	StateSenderAddress    common.Address
	StakingManagerAddress common.Address
	StakingInfoAddress    common.Address
	TimeStamp             *big.Int
}, error) {
	return _Rootchain.Contract.ChainMap(&_Rootchain.CallOpts, arg0)
}

// ChainMap is a free data retrieval call binding the contract method 0xdf35ce79.
//
// Solidity: function chainMap(uint256 ) constant returns(uint256 rootChainId, uint256 activationHeight, uint256 txConfirmations, address rootChainAddress, address stateSenderAddress, address stakingManagerAddress, address stakingInfoAddress, uint256 timeStamp)
func (_Rootchain *RootchainCallerSession) ChainMap(arg0 *big.Int) (struct {
	RootChainId           *big.Int
	ActivationHeight      *big.Int
	TxConfirmations       *big.Int
	RootChainAddress      common.Address
	StateSenderAddress    common.Address
	StakingManagerAddress common.Address
	StakingInfoAddress    common.Address
	TimeStamp             *big.Int
}, error) {
	return _Rootchain.Contract.ChainMap(&_Rootchain.CallOpts, arg0)
}

// CurrentHeaderBlock is a free data retrieval call binding the contract method 0xec7e4855.
//
// Solidity: function currentHeaderBlock() constant returns(uint256)
func (_Rootchain *RootchainCaller) CurrentHeaderBlock(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Rootchain.contract.Call(opts, out, "currentHeaderBlock")
	return *ret0, err
}

// CurrentHeaderBlock is a free data retrieval call binding the contract method 0xec7e4855.
//
// Solidity: function currentHeaderBlock() constant returns(uint256)
func (_Rootchain *RootchainSession) CurrentHeaderBlock() (*big.Int, error) {
	return _Rootchain.Contract.CurrentHeaderBlock(&_Rootchain.CallOpts)
}

// CurrentHeaderBlock is a free data retrieval call binding the contract method 0xec7e4855.
//
// Solidity: function currentHeaderBlock() constant returns(uint256)
func (_Rootchain *RootchainCallerSession) CurrentHeaderBlock() (*big.Int, error) {
	return _Rootchain.Contract.CurrentHeaderBlock(&_Rootchain.CallOpts)
}

// GetLastChildBlock is a free data retrieval call binding the contract method 0xb87e1b66.
//
// Solidity: function getLastChildBlock() constant returns(uint256)
func (_Rootchain *RootchainCaller) GetLastChildBlock(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Rootchain.contract.Call(opts, out, "getLastChildBlock")
	return *ret0, err
}

// GetLastChildBlock is a free data retrieval call binding the contract method 0xb87e1b66.
//
// Solidity: function getLastChildBlock() constant returns(uint256)
func (_Rootchain *RootchainSession) GetLastChildBlock() (*big.Int, error) {
	return _Rootchain.Contract.GetLastChildBlock(&_Rootchain.CallOpts)
}

// GetLastChildBlock is a free data retrieval call binding the contract method 0xb87e1b66.
//
// Solidity: function getLastChildBlock() constant returns(uint256)
func (_Rootchain *RootchainCallerSession) GetLastChildBlock() (*big.Int, error) {
	return _Rootchain.Contract.GetLastChildBlock(&_Rootchain.CallOpts)
}

// HeaderBlocks is a free data retrieval call binding the contract method 0x41539d4a.
//
// Solidity: function headerBlocks(uint256 ) constant returns(bytes32 root, uint256 start, uint256 end, uint256 createdAt, address proposer)
func (_Rootchain *RootchainCaller) HeaderBlocks(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Root      [32]byte
	Start     *big.Int
	End       *big.Int
	CreatedAt *big.Int
	Proposer  common.Address
}, error) {
	ret := new(struct {
		Root      [32]byte
		Start     *big.Int
		End       *big.Int
		CreatedAt *big.Int
		Proposer  common.Address
	})
	out := ret
	err := _Rootchain.contract.Call(opts, out, "headerBlocks", arg0)
	return *ret, err
}

// HeaderBlocks is a free data retrieval call binding the contract method 0x41539d4a.
//
// Solidity: function headerBlocks(uint256 ) constant returns(bytes32 root, uint256 start, uint256 end, uint256 createdAt, address proposer)
func (_Rootchain *RootchainSession) HeaderBlocks(arg0 *big.Int) (struct {
	Root      [32]byte
	Start     *big.Int
	End       *big.Int
	CreatedAt *big.Int
	Proposer  common.Address
}, error) {
	return _Rootchain.Contract.HeaderBlocks(&_Rootchain.CallOpts, arg0)
}

// HeaderBlocks is a free data retrieval call binding the contract method 0x41539d4a.
//
// Solidity: function headerBlocks(uint256 ) constant returns(bytes32 root, uint256 start, uint256 end, uint256 createdAt, address proposer)
func (_Rootchain *RootchainCallerSession) HeaderBlocks(arg0 *big.Int) (struct {
	Root      [32]byte
	Start     *big.Int
	End       *big.Int
	CreatedAt *big.Int
	Proposer  common.Address
}, error) {
	return _Rootchain.Contract.HeaderBlocks(&_Rootchain.CallOpts, arg0)
}

// SetChainInfo is a paid mutator transaction binding the contract method 0x139073e8.
//
// Solidity: function setChainInfo(uint256 rootChainId, uint256 activationHeight, uint256 txConfirmations, address rootChainAddress, address stateSenderAddress, address stakingManagerAddress, address stakingInfoAddress, uint256 timeStamp) returns()
func (_Rootchain *RootchainTransactor) SetChainInfo(opts *bind.TransactOpts, rootChainId *big.Int, activationHeight *big.Int, txConfirmations *big.Int, rootChainAddress common.Address, stateSenderAddress common.Address, stakingManagerAddress common.Address, stakingInfoAddress common.Address, timeStamp *big.Int) (*types.Transaction, error) {
	return _Rootchain.contract.Transact(opts, "setChainInfo", rootChainId, activationHeight, txConfirmations, rootChainAddress, stateSenderAddress, stakingManagerAddress, stakingInfoAddress, timeStamp)
}

// SetChainInfo is a paid mutator transaction binding the contract method 0x139073e8.
//
// Solidity: function setChainInfo(uint256 rootChainId, uint256 activationHeight, uint256 txConfirmations, address rootChainAddress, address stateSenderAddress, address stakingManagerAddress, address stakingInfoAddress, uint256 timeStamp) returns()
func (_Rootchain *RootchainSession) SetChainInfo(rootChainId *big.Int, activationHeight *big.Int, txConfirmations *big.Int, rootChainAddress common.Address, stateSenderAddress common.Address, stakingManagerAddress common.Address, stakingInfoAddress common.Address, timeStamp *big.Int) (*types.Transaction, error) {
	return _Rootchain.Contract.SetChainInfo(&_Rootchain.TransactOpts, rootChainId, activationHeight, txConfirmations, rootChainAddress, stateSenderAddress, stakingManagerAddress, stakingInfoAddress, timeStamp)
}

// SetChainInfo is a paid mutator transaction binding the contract method 0x139073e8.
//
// Solidity: function setChainInfo(uint256 rootChainId, uint256 activationHeight, uint256 txConfirmations, address rootChainAddress, address stateSenderAddress, address stakingManagerAddress, address stakingInfoAddress, uint256 timeStamp) returns()
func (_Rootchain *RootchainTransactorSession) SetChainInfo(rootChainId *big.Int, activationHeight *big.Int, txConfirmations *big.Int, rootChainAddress common.Address, stateSenderAddress common.Address, stakingManagerAddress common.Address, stakingInfoAddress common.Address, timeStamp *big.Int) (*types.Transaction, error) {
	return _Rootchain.Contract.SetChainInfo(&_Rootchain.TransactOpts, rootChainId, activationHeight, txConfirmations, rootChainAddress, stateSenderAddress, stakingManagerAddress, stakingInfoAddress, timeStamp)
}

// SubmitCheckpoint is a paid mutator transaction binding the contract method 0x4e43e495.
//
// Solidity: function submitCheckpoint(bytes data, uint256[3][] sigs) returns()
func (_Rootchain *RootchainTransactor) SubmitCheckpoint(opts *bind.TransactOpts, data []byte, sigs [][3]*big.Int) (*types.Transaction, error) {
	return _Rootchain.contract.Transact(opts, "submitCheckpoint", data, sigs)
}

// SubmitCheckpoint is a paid mutator transaction binding the contract method 0x4e43e495.
//
// Solidity: function submitCheckpoint(bytes data, uint256[3][] sigs) returns()
func (_Rootchain *RootchainSession) SubmitCheckpoint(data []byte, sigs [][3]*big.Int) (*types.Transaction, error) {
	return _Rootchain.Contract.SubmitCheckpoint(&_Rootchain.TransactOpts, data, sigs)
}

// SubmitCheckpoint is a paid mutator transaction binding the contract method 0x4e43e495.
//
// Solidity: function submitCheckpoint(bytes data, uint256[3][] sigs) returns()
func (_Rootchain *RootchainTransactorSession) SubmitCheckpoint(data []byte, sigs [][3]*big.Int) (*types.Transaction, error) {
	return _Rootchain.Contract.SubmitCheckpoint(&_Rootchain.TransactOpts, data, sigs)
}

// SubmitHeaderBlock is a paid mutator transaction binding the contract method 0x6a791f11.
//
// Solidity: function submitHeaderBlock(bytes data, bytes sigs) returns()
func (_Rootchain *RootchainTransactor) SubmitHeaderBlock(opts *bind.TransactOpts, data []byte, sigs []byte) (*types.Transaction, error) {
	return _Rootchain.contract.Transact(opts, "submitHeaderBlock", data, sigs)
}

// SubmitHeaderBlock is a paid mutator transaction binding the contract method 0x6a791f11.
//
// Solidity: function submitHeaderBlock(bytes data, bytes sigs) returns()
func (_Rootchain *RootchainSession) SubmitHeaderBlock(data []byte, sigs []byte) (*types.Transaction, error) {
	return _Rootchain.Contract.SubmitHeaderBlock(&_Rootchain.TransactOpts, data, sigs)
}

// SubmitHeaderBlock is a paid mutator transaction binding the contract method 0x6a791f11.
//
// Solidity: function submitHeaderBlock(bytes data, bytes sigs) returns()
func (_Rootchain *RootchainTransactorSession) SubmitHeaderBlock(data []byte, sigs []byte) (*types.Transaction, error) {
	return _Rootchain.Contract.SubmitHeaderBlock(&_Rootchain.TransactOpts, data, sigs)
}

// RootchainNewChainIterator is returned from FilterNewChain and is used to iterate over the raw logs and unpacked data for NewChain events raised by the Rootchain contract.
type RootchainNewChainIterator struct {
	Event *RootchainNewChain // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *RootchainNewChainIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RootchainNewChain)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(RootchainNewChain)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *RootchainNewChainIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RootchainNewChainIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RootchainNewChain represents a NewChain event raised by the Rootchain contract.
type RootchainNewChain struct {
	RootChainId           *big.Int
	ActivationHeight      *big.Int
	TxConfirmations       *big.Int
	RootChainAddress      common.Address
	StateSenderAddress    common.Address
	StakingManagerAddress common.Address
	StakingInfoAddress    common.Address
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterNewChain is a free log retrieval operation binding the contract event 0x46d8e11a9f6eb349dbb840ef59be3b9f104a0ec5ef156698d105d99a2a33c00c.
//
// Solidity: event NewChain(uint256 indexed rootChainId, uint256 activationHeight, uint256 txConfirmations, address rootChainAddress, address stateSenderAddress, address stakingManagerAddress, address stakingInfoAddress)
func (_Rootchain *RootchainFilterer) FilterNewChain(opts *bind.FilterOpts, rootChainId []*big.Int) (*RootchainNewChainIterator, error) {

	var rootChainIdRule []interface{}
	for _, rootChainIdItem := range rootChainId {
		rootChainIdRule = append(rootChainIdRule, rootChainIdItem)
	}

	logs, sub, err := _Rootchain.contract.FilterLogs(opts, "NewChain", rootChainIdRule)
	if err != nil {
		return nil, err
	}
	return &RootchainNewChainIterator{contract: _Rootchain.contract, event: "NewChain", logs: logs, sub: sub}, nil
}

// WatchNewChain is a free log subscription operation binding the contract event 0x46d8e11a9f6eb349dbb840ef59be3b9f104a0ec5ef156698d105d99a2a33c00c.
//
// Solidity: event NewChain(uint256 indexed rootChainId, uint256 activationHeight, uint256 txConfirmations, address rootChainAddress, address stateSenderAddress, address stakingManagerAddress, address stakingInfoAddress)
func (_Rootchain *RootchainFilterer) WatchNewChain(opts *bind.WatchOpts, sink chan<- *RootchainNewChain, rootChainId []*big.Int) (event.Subscription, error) {

	var rootChainIdRule []interface{}
	for _, rootChainIdItem := range rootChainId {
		rootChainIdRule = append(rootChainIdRule, rootChainIdItem)
	}

	logs, sub, err := _Rootchain.contract.WatchLogs(opts, "NewChain", rootChainIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RootchainNewChain)
				if err := _Rootchain.contract.UnpackLog(event, "NewChain", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseNewChain is a log parse operation binding the contract event 0x46d8e11a9f6eb349dbb840ef59be3b9f104a0ec5ef156698d105d99a2a33c00c.
//
// Solidity: event NewChain(uint256 indexed rootChainId, uint256 activationHeight, uint256 txConfirmations, address rootChainAddress, address stateSenderAddress, address stakingManagerAddress, address stakingInfoAddress)
func (_Rootchain *RootchainFilterer) ParseNewChain(log types.Log) (*RootchainNewChain, error) {
	event := new(RootchainNewChain)
	if err := _Rootchain.contract.UnpackLog(event, "NewChain", log); err != nil {
		return nil, err
	}
	return event, nil
}

// RootchainNewHeaderBlockIterator is returned from FilterNewHeaderBlock and is used to iterate over the raw logs and unpacked data for NewHeaderBlock events raised by the Rootchain contract.
type RootchainNewHeaderBlockIterator struct {
	Event *RootchainNewHeaderBlock // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *RootchainNewHeaderBlockIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RootchainNewHeaderBlock)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(RootchainNewHeaderBlock)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *RootchainNewHeaderBlockIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RootchainNewHeaderBlockIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RootchainNewHeaderBlock represents a NewHeaderBlock event raised by the Rootchain contract.
type RootchainNewHeaderBlock struct {
	Proposer      common.Address
	HeaderBlockId *big.Int
	Reward        *big.Int
	Start         *big.Int
	End           *big.Int
	Root          [32]byte
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterNewHeaderBlock is a free log retrieval operation binding the contract event 0xba5de06d22af2685c6c7765f60067f7d2b08c2d29f53cdf14d67f6d1c9bfb527.
//
// Solidity: event NewHeaderBlock(address indexed proposer, uint256 indexed headerBlockId, uint256 indexed reward, uint256 start, uint256 end, bytes32 root)
func (_Rootchain *RootchainFilterer) FilterNewHeaderBlock(opts *bind.FilterOpts, proposer []common.Address, headerBlockId []*big.Int, reward []*big.Int) (*RootchainNewHeaderBlockIterator, error) {

	var proposerRule []interface{}
	for _, proposerItem := range proposer {
		proposerRule = append(proposerRule, proposerItem)
	}
	var headerBlockIdRule []interface{}
	for _, headerBlockIdItem := range headerBlockId {
		headerBlockIdRule = append(headerBlockIdRule, headerBlockIdItem)
	}
	var rewardRule []interface{}
	for _, rewardItem := range reward {
		rewardRule = append(rewardRule, rewardItem)
	}

	logs, sub, err := _Rootchain.contract.FilterLogs(opts, "NewHeaderBlock", proposerRule, headerBlockIdRule, rewardRule)
	if err != nil {
		return nil, err
	}
	return &RootchainNewHeaderBlockIterator{contract: _Rootchain.contract, event: "NewHeaderBlock", logs: logs, sub: sub}, nil
}

// WatchNewHeaderBlock is a free log subscription operation binding the contract event 0xba5de06d22af2685c6c7765f60067f7d2b08c2d29f53cdf14d67f6d1c9bfb527.
//
// Solidity: event NewHeaderBlock(address indexed proposer, uint256 indexed headerBlockId, uint256 indexed reward, uint256 start, uint256 end, bytes32 root)
func (_Rootchain *RootchainFilterer) WatchNewHeaderBlock(opts *bind.WatchOpts, sink chan<- *RootchainNewHeaderBlock, proposer []common.Address, headerBlockId []*big.Int, reward []*big.Int) (event.Subscription, error) {

	var proposerRule []interface{}
	for _, proposerItem := range proposer {
		proposerRule = append(proposerRule, proposerItem)
	}
	var headerBlockIdRule []interface{}
	for _, headerBlockIdItem := range headerBlockId {
		headerBlockIdRule = append(headerBlockIdRule, headerBlockIdItem)
	}
	var rewardRule []interface{}
	for _, rewardItem := range reward {
		rewardRule = append(rewardRule, rewardItem)
	}

	logs, sub, err := _Rootchain.contract.WatchLogs(opts, "NewHeaderBlock", proposerRule, headerBlockIdRule, rewardRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RootchainNewHeaderBlock)
				if err := _Rootchain.contract.UnpackLog(event, "NewHeaderBlock", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseNewHeaderBlock is a log parse operation binding the contract event 0xba5de06d22af2685c6c7765f60067f7d2b08c2d29f53cdf14d67f6d1c9bfb527.
//
// Solidity: event NewHeaderBlock(address indexed proposer, uint256 indexed headerBlockId, uint256 indexed reward, uint256 start, uint256 end, bytes32 root)
func (_Rootchain *RootchainFilterer) ParseNewHeaderBlock(log types.Log) (*RootchainNewHeaderBlock, error) {
	event := new(RootchainNewHeaderBlock)
	if err := _Rootchain.contract.UnpackLog(event, "NewHeaderBlock", log); err != nil {
		return nil, err
	}
	return event, nil
}
