// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contracts

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// ContractsMetaData contains all meta data concerning the Contracts contract.
var ContractsMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_initialValue\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"oldValue\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newValue\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"changer\",\"type\":\"address\"}],\"name\":\"CountChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"resetter\",\"type\":\"address\"}],\"name\":\"CountReset\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"add\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decrement\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getInfo\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"_count\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"increment\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"reset\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"setCount\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"subtract\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x6080604052348015600e575f5ffd5b5060405161055a38038061055a833981016040819052602b916084565b5f818155600180546001600160a01b03191633908117909155604080519283526020830184905290917f0c71f6d459e7ac4df0a4abe9377b1b730c147cf19846f02685321aee0fd4c663910160405180910390a250609a565b5f602082840312156093575f5ffd5b5051919050565b6104b3806100a75f395ff3fe608060405234801561000f575f5ffd5b5060043610610090575f3560e01c80638da5cb5b116100635780638da5cb5b146100eb578063a87d942c14610116578063d09de08a14610126578063d14e62b81461012e578063d826f88f14610141575f5ffd5b80631003e2d2146100945780631dc05f17146100a95780632baeceb7146100bc5780635a9b0b89146100c4575b5f5ffd5b6100a76100a2366004610406565b610149565b005b6100a76100b7366004610406565b610193565b6100a761020a565b5f54600154604080519283526001600160a01b039091166020830152015b60405180910390f35b6001546100fe906001600160a01b031681565b6040516001600160a01b0390911681526020016100e2565b5f546040519081526020016100e2565b6100a76102b0565b6100a761013c366004610406565b6102c2565b6100a7610352565b5f80549082908061015a8385610431565b90915550505f5460405133915f51602061045e5f395f51905f529161018791858252602082015260400190565b60405180910390a25050565b805f5410156101f95760405162461bcd60e51b815260206004820152602760248201527f436f756e7465723a20696e73756666696369656e7420636f756e7420746f20736044820152661d589d1c9858dd60ca1b60648201526084015b60405180910390fd5b5f80549082908061015a838561044a565b5f5f54116102665760405162461bcd60e51b8152602060048201526024808201527f436f756e7465723a2063616e6e6f742064656372656d656e742062656c6f77206044820152637a65726f60e01b60648201526084016101f0565b5f80549060019080610278838561044a565b90915550505f5460405133915f51602061045e5f395f51905f52916102a591858252602082015260400190565b60405180910390a250565b5f805490600190806102788385610431565b6001546001600160a01b031633146103265760405162461bcd60e51b815260206004820152602160248201527f436f756e7465723a206f6e6c79206f776e65722063616e2073657420636f756e6044820152601d60fa1b60648201526084016101f0565b5f805490829055604080518281526020810184905233915f51602061045e5f395f51905f529101610187565b6001546001600160a01b031633146103ac5760405162461bcd60e51b815260206004820152601d60248201527f436f756e7465723a206f6e6c79206f776e65722063616e20726573657400000060448201526064016101f0565b5f80805560405133917fa5ee6258204973c56c5a39c4ac31e61723f410d84f9e8117ba52b76b7cea990c91a25f805460408051918252602082019290925233915f51602061045e5f395f51905f52910160405180910390a2565b5f60208284031215610416575f5ffd5b5035919050565b634e487b7160e01b5f52601160045260245ffd5b808201808211156104445761044461041d565b92915050565b818103818111156104445761044461041d56fe0c71f6d459e7ac4df0a4abe9377b1b730c147cf19846f02685321aee0fd4c663a2646970667358221220ca0f37070f290b2bad0cff92ef4b4b7aba84a5e587a00bf679c5670c0327e7ef64736f6c634300081e0033",
}

// ContractsABI is the input ABI used to generate the binding from.
// Deprecated: Use ContractsMetaData.ABI instead.
var ContractsABI = ContractsMetaData.ABI

// ContractsBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ContractsMetaData.Bin instead.
var ContractsBin = ContractsMetaData.Bin

// DeployContracts deploys a new Ethereum contract, binding an instance of Contracts to it.
func DeployContracts(auth *bind.TransactOpts, backend bind.ContractBackend, _initialValue *big.Int) (common.Address, *types.Transaction, *Contracts, error) {
	parsed, err := ContractsMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ContractsBin), backend, _initialValue)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Contracts{ContractsCaller: ContractsCaller{contract: contract}, ContractsTransactor: ContractsTransactor{contract: contract}, ContractsFilterer: ContractsFilterer{contract: contract}}, nil
}

// Contracts is an auto generated Go binding around an Ethereum contract.
type Contracts struct {
	ContractsCaller     // Read-only binding to the contract
	ContractsTransactor // Write-only binding to the contract
	ContractsFilterer   // Log filterer for contract events
}

// ContractsCaller is an auto generated read-only Go binding around an Ethereum contract.
type ContractsCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractsTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ContractsTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractsFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ContractsFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractsSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ContractsSession struct {
	Contract     *Contracts        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ContractsCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ContractsCallerSession struct {
	Contract *ContractsCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// ContractsTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ContractsTransactorSession struct {
	Contract     *ContractsTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// ContractsRaw is an auto generated low-level Go binding around an Ethereum contract.
type ContractsRaw struct {
	Contract *Contracts // Generic contract binding to access the raw methods on
}

// ContractsCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ContractsCallerRaw struct {
	Contract *ContractsCaller // Generic read-only contract binding to access the raw methods on
}

// ContractsTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ContractsTransactorRaw struct {
	Contract *ContractsTransactor // Generic write-only contract binding to access the raw methods on
}

// NewContracts creates a new instance of Contracts, bound to a specific deployed contract.
func NewContracts(address common.Address, backend bind.ContractBackend) (*Contracts, error) {
	contract, err := bindContracts(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Contracts{ContractsCaller: ContractsCaller{contract: contract}, ContractsTransactor: ContractsTransactor{contract: contract}, ContractsFilterer: ContractsFilterer{contract: contract}}, nil
}

// NewContractsCaller creates a new read-only instance of Contracts, bound to a specific deployed contract.
func NewContractsCaller(address common.Address, caller bind.ContractCaller) (*ContractsCaller, error) {
	contract, err := bindContracts(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ContractsCaller{contract: contract}, nil
}

// NewContractsTransactor creates a new write-only instance of Contracts, bound to a specific deployed contract.
func NewContractsTransactor(address common.Address, transactor bind.ContractTransactor) (*ContractsTransactor, error) {
	contract, err := bindContracts(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ContractsTransactor{contract: contract}, nil
}

// NewContractsFilterer creates a new log filterer instance of Contracts, bound to a specific deployed contract.
func NewContractsFilterer(address common.Address, filterer bind.ContractFilterer) (*ContractsFilterer, error) {
	contract, err := bindContracts(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ContractsFilterer{contract: contract}, nil
}

// bindContracts binds a generic wrapper to an already deployed contract.
func bindContracts(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ContractsMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contracts *ContractsRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Contracts.Contract.ContractsCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contracts *ContractsRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contracts.Contract.ContractsTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contracts *ContractsRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contracts.Contract.ContractsTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contracts *ContractsCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Contracts.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contracts *ContractsTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contracts.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contracts *ContractsTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contracts.Contract.contract.Transact(opts, method, params...)
}

// GetCount is a free data retrieval call binding the contract method 0xa87d942c.
//
// Solidity: function getCount() view returns(uint256)
func (_Contracts *ContractsCaller) GetCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Contracts.contract.Call(opts, &out, "getCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetCount is a free data retrieval call binding the contract method 0xa87d942c.
//
// Solidity: function getCount() view returns(uint256)
func (_Contracts *ContractsSession) GetCount() (*big.Int, error) {
	return _Contracts.Contract.GetCount(&_Contracts.CallOpts)
}

// GetCount is a free data retrieval call binding the contract method 0xa87d942c.
//
// Solidity: function getCount() view returns(uint256)
func (_Contracts *ContractsCallerSession) GetCount() (*big.Int, error) {
	return _Contracts.Contract.GetCount(&_Contracts.CallOpts)
}

// GetInfo is a free data retrieval call binding the contract method 0x5a9b0b89.
//
// Solidity: function getInfo() view returns(uint256 _count, address _owner)
func (_Contracts *ContractsCaller) GetInfo(opts *bind.CallOpts) (struct {
	Count *big.Int
	Owner common.Address
}, error) {
	var out []interface{}
	err := _Contracts.contract.Call(opts, &out, "getInfo")

	outstruct := new(struct {
		Count *big.Int
		Owner common.Address
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Count = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Owner = *abi.ConvertType(out[1], new(common.Address)).(*common.Address)

	return *outstruct, err

}

// GetInfo is a free data retrieval call binding the contract method 0x5a9b0b89.
//
// Solidity: function getInfo() view returns(uint256 _count, address _owner)
func (_Contracts *ContractsSession) GetInfo() (struct {
	Count *big.Int
	Owner common.Address
}, error) {
	return _Contracts.Contract.GetInfo(&_Contracts.CallOpts)
}

// GetInfo is a free data retrieval call binding the contract method 0x5a9b0b89.
//
// Solidity: function getInfo() view returns(uint256 _count, address _owner)
func (_Contracts *ContractsCallerSession) GetInfo() (struct {
	Count *big.Int
	Owner common.Address
}, error) {
	return _Contracts.Contract.GetInfo(&_Contracts.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Contracts *ContractsCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Contracts.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Contracts *ContractsSession) Owner() (common.Address, error) {
	return _Contracts.Contract.Owner(&_Contracts.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Contracts *ContractsCallerSession) Owner() (common.Address, error) {
	return _Contracts.Contract.Owner(&_Contracts.CallOpts)
}

// Add is a paid mutator transaction binding the contract method 0x1003e2d2.
//
// Solidity: function add(uint256 _value) returns()
func (_Contracts *ContractsTransactor) Add(opts *bind.TransactOpts, _value *big.Int) (*types.Transaction, error) {
	return _Contracts.contract.Transact(opts, "add", _value)
}

// Add is a paid mutator transaction binding the contract method 0x1003e2d2.
//
// Solidity: function add(uint256 _value) returns()
func (_Contracts *ContractsSession) Add(_value *big.Int) (*types.Transaction, error) {
	return _Contracts.Contract.Add(&_Contracts.TransactOpts, _value)
}

// Add is a paid mutator transaction binding the contract method 0x1003e2d2.
//
// Solidity: function add(uint256 _value) returns()
func (_Contracts *ContractsTransactorSession) Add(_value *big.Int) (*types.Transaction, error) {
	return _Contracts.Contract.Add(&_Contracts.TransactOpts, _value)
}

// Decrement is a paid mutator transaction binding the contract method 0x2baeceb7.
//
// Solidity: function decrement() returns()
func (_Contracts *ContractsTransactor) Decrement(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contracts.contract.Transact(opts, "decrement")
}

// Decrement is a paid mutator transaction binding the contract method 0x2baeceb7.
//
// Solidity: function decrement() returns()
func (_Contracts *ContractsSession) Decrement() (*types.Transaction, error) {
	return _Contracts.Contract.Decrement(&_Contracts.TransactOpts)
}

// Decrement is a paid mutator transaction binding the contract method 0x2baeceb7.
//
// Solidity: function decrement() returns()
func (_Contracts *ContractsTransactorSession) Decrement() (*types.Transaction, error) {
	return _Contracts.Contract.Decrement(&_Contracts.TransactOpts)
}

// Increment is a paid mutator transaction binding the contract method 0xd09de08a.
//
// Solidity: function increment() returns()
func (_Contracts *ContractsTransactor) Increment(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contracts.contract.Transact(opts, "increment")
}

// Increment is a paid mutator transaction binding the contract method 0xd09de08a.
//
// Solidity: function increment() returns()
func (_Contracts *ContractsSession) Increment() (*types.Transaction, error) {
	return _Contracts.Contract.Increment(&_Contracts.TransactOpts)
}

// Increment is a paid mutator transaction binding the contract method 0xd09de08a.
//
// Solidity: function increment() returns()
func (_Contracts *ContractsTransactorSession) Increment() (*types.Transaction, error) {
	return _Contracts.Contract.Increment(&_Contracts.TransactOpts)
}

// Reset is a paid mutator transaction binding the contract method 0xd826f88f.
//
// Solidity: function reset() returns()
func (_Contracts *ContractsTransactor) Reset(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contracts.contract.Transact(opts, "reset")
}

// Reset is a paid mutator transaction binding the contract method 0xd826f88f.
//
// Solidity: function reset() returns()
func (_Contracts *ContractsSession) Reset() (*types.Transaction, error) {
	return _Contracts.Contract.Reset(&_Contracts.TransactOpts)
}

// Reset is a paid mutator transaction binding the contract method 0xd826f88f.
//
// Solidity: function reset() returns()
func (_Contracts *ContractsTransactorSession) Reset() (*types.Transaction, error) {
	return _Contracts.Contract.Reset(&_Contracts.TransactOpts)
}

// SetCount is a paid mutator transaction binding the contract method 0xd14e62b8.
//
// Solidity: function setCount(uint256 _value) returns()
func (_Contracts *ContractsTransactor) SetCount(opts *bind.TransactOpts, _value *big.Int) (*types.Transaction, error) {
	return _Contracts.contract.Transact(opts, "setCount", _value)
}

// SetCount is a paid mutator transaction binding the contract method 0xd14e62b8.
//
// Solidity: function setCount(uint256 _value) returns()
func (_Contracts *ContractsSession) SetCount(_value *big.Int) (*types.Transaction, error) {
	return _Contracts.Contract.SetCount(&_Contracts.TransactOpts, _value)
}

// SetCount is a paid mutator transaction binding the contract method 0xd14e62b8.
//
// Solidity: function setCount(uint256 _value) returns()
func (_Contracts *ContractsTransactorSession) SetCount(_value *big.Int) (*types.Transaction, error) {
	return _Contracts.Contract.SetCount(&_Contracts.TransactOpts, _value)
}

// Subtract is a paid mutator transaction binding the contract method 0x1dc05f17.
//
// Solidity: function subtract(uint256 _value) returns()
func (_Contracts *ContractsTransactor) Subtract(opts *bind.TransactOpts, _value *big.Int) (*types.Transaction, error) {
	return _Contracts.contract.Transact(opts, "subtract", _value)
}

// Subtract is a paid mutator transaction binding the contract method 0x1dc05f17.
//
// Solidity: function subtract(uint256 _value) returns()
func (_Contracts *ContractsSession) Subtract(_value *big.Int) (*types.Transaction, error) {
	return _Contracts.Contract.Subtract(&_Contracts.TransactOpts, _value)
}

// Subtract is a paid mutator transaction binding the contract method 0x1dc05f17.
//
// Solidity: function subtract(uint256 _value) returns()
func (_Contracts *ContractsTransactorSession) Subtract(_value *big.Int) (*types.Transaction, error) {
	return _Contracts.Contract.Subtract(&_Contracts.TransactOpts, _value)
}

// ContractsCountChangedIterator is returned from FilterCountChanged and is used to iterate over the raw logs and unpacked data for CountChanged events raised by the Contracts contract.
type ContractsCountChangedIterator struct {
	Event *ContractsCountChanged // Event containing the contract specifics and raw log

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
func (it *ContractsCountChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractsCountChanged)
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
		it.Event = new(ContractsCountChanged)
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
func (it *ContractsCountChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractsCountChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractsCountChanged represents a CountChanged event raised by the Contracts contract.
type ContractsCountChanged struct {
	OldValue *big.Int
	NewValue *big.Int
	Changer  common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterCountChanged is a free log retrieval operation binding the contract event 0x0c71f6d459e7ac4df0a4abe9377b1b730c147cf19846f02685321aee0fd4c663.
//
// Solidity: event CountChanged(uint256 oldValue, uint256 newValue, address indexed changer)
func (_Contracts *ContractsFilterer) FilterCountChanged(opts *bind.FilterOpts, changer []common.Address) (*ContractsCountChangedIterator, error) {

	var changerRule []interface{}
	for _, changerItem := range changer {
		changerRule = append(changerRule, changerItem)
	}

	logs, sub, err := _Contracts.contract.FilterLogs(opts, "CountChanged", changerRule)
	if err != nil {
		return nil, err
	}
	return &ContractsCountChangedIterator{contract: _Contracts.contract, event: "CountChanged", logs: logs, sub: sub}, nil
}

// WatchCountChanged is a free log subscription operation binding the contract event 0x0c71f6d459e7ac4df0a4abe9377b1b730c147cf19846f02685321aee0fd4c663.
//
// Solidity: event CountChanged(uint256 oldValue, uint256 newValue, address indexed changer)
func (_Contracts *ContractsFilterer) WatchCountChanged(opts *bind.WatchOpts, sink chan<- *ContractsCountChanged, changer []common.Address) (event.Subscription, error) {

	var changerRule []interface{}
	for _, changerItem := range changer {
		changerRule = append(changerRule, changerItem)
	}

	logs, sub, err := _Contracts.contract.WatchLogs(opts, "CountChanged", changerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractsCountChanged)
				if err := _Contracts.contract.UnpackLog(event, "CountChanged", log); err != nil {
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

// ParseCountChanged is a log parse operation binding the contract event 0x0c71f6d459e7ac4df0a4abe9377b1b730c147cf19846f02685321aee0fd4c663.
//
// Solidity: event CountChanged(uint256 oldValue, uint256 newValue, address indexed changer)
func (_Contracts *ContractsFilterer) ParseCountChanged(log types.Log) (*ContractsCountChanged, error) {
	event := new(ContractsCountChanged)
	if err := _Contracts.contract.UnpackLog(event, "CountChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractsCountResetIterator is returned from FilterCountReset and is used to iterate over the raw logs and unpacked data for CountReset events raised by the Contracts contract.
type ContractsCountResetIterator struct {
	Event *ContractsCountReset // Event containing the contract specifics and raw log

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
func (it *ContractsCountResetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractsCountReset)
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
		it.Event = new(ContractsCountReset)
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
func (it *ContractsCountResetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractsCountResetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractsCountReset represents a CountReset event raised by the Contracts contract.
type ContractsCountReset struct {
	Resetter common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterCountReset is a free log retrieval operation binding the contract event 0xa5ee6258204973c56c5a39c4ac31e61723f410d84f9e8117ba52b76b7cea990c.
//
// Solidity: event CountReset(address indexed resetter)
func (_Contracts *ContractsFilterer) FilterCountReset(opts *bind.FilterOpts, resetter []common.Address) (*ContractsCountResetIterator, error) {

	var resetterRule []interface{}
	for _, resetterItem := range resetter {
		resetterRule = append(resetterRule, resetterItem)
	}

	logs, sub, err := _Contracts.contract.FilterLogs(opts, "CountReset", resetterRule)
	if err != nil {
		return nil, err
	}
	return &ContractsCountResetIterator{contract: _Contracts.contract, event: "CountReset", logs: logs, sub: sub}, nil
}

// WatchCountReset is a free log subscription operation binding the contract event 0xa5ee6258204973c56c5a39c4ac31e61723f410d84f9e8117ba52b76b7cea990c.
//
// Solidity: event CountReset(address indexed resetter)
func (_Contracts *ContractsFilterer) WatchCountReset(opts *bind.WatchOpts, sink chan<- *ContractsCountReset, resetter []common.Address) (event.Subscription, error) {

	var resetterRule []interface{}
	for _, resetterItem := range resetter {
		resetterRule = append(resetterRule, resetterItem)
	}

	logs, sub, err := _Contracts.contract.WatchLogs(opts, "CountReset", resetterRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractsCountReset)
				if err := _Contracts.contract.UnpackLog(event, "CountReset", log); err != nil {
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

// ParseCountReset is a log parse operation binding the contract event 0xa5ee6258204973c56c5a39c4ac31e61723f410d84f9e8117ba52b76b7cea990c.
//
// Solidity: event CountReset(address indexed resetter)
func (_Contracts *ContractsFilterer) ParseCountReset(log types.Log) (*ContractsCountReset, error) {
	event := new(ContractsCountReset)
	if err := _Contracts.contract.UnpackLog(event, "CountReset", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
