// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package gov

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

// MultiSigWalletMetaData contains all meta data concerning the MultiSigWallet contract.
var MultiSigWalletMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"_owners\",\"type\":\"address[]\"},{\"internalType\":\"uint256\",\"name\":\"_quorum\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"AddOwner\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"quorum\",\"type\":\"uint256\"}],\"name\":\"ChangeQuorum\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"}],\"name\":\"ConfirmTransaction\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"balance\",\"type\":\"uint256\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"}],\"name\":\"ExecuteTransaction\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"RemoveOwner\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"}],\"name\":\"RevokeConfirmation\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"SubmitTransaction\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"MAX_OWNER_COUNT\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_newOwner\",\"type\":\"address\"}],\"name\":\"addOwner\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_quorum\",\"type\":\"uint256\"}],\"name\":\"changeQuorum\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_transactionId\",\"type\":\"uint256\"}],\"name\":\"confirmTransaction\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_transactionId\",\"type\":\"uint256\"}],\"name\":\"executeTransaction\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOwnerCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOwners\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_transactionId\",\"type\":\"uint256\"}],\"name\":\"getTransaction\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"executed\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"currentNumberOfConfirmations\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getTransactionCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"proposer\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"blockNumber\",\"type\":\"uint256\"}],\"name\":\"getTransactionId\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"isConfirmed\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"isOwner\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"owners\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"proposalHashToTxId\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"quorum\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"removeOwner\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_newOwner\",\"type\":\"address\"}],\"name\":\"replaceOwner\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_transactionId\",\"type\":\"uint256\"}],\"name\":\"revokeConfirmation\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_data\",\"type\":\"bytes\"}],\"name\":\"submitTransaction\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"transactions\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"executed\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"currentNumberOfConfirmations\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
	Sigs: map[string]string{
		"d74f8edd": "MAX_OWNER_COUNT()",
		"7065cb48": "addOwner(address)",
		"d2cd96bd": "changeQuorum(uint256)",
		"c01a8c84": "confirmTransaction(uint256)",
		"ee22610b": "executeTransaction(uint256)",
		"ef18374a": "getOwnerCount()",
		"a0e67e2b": "getOwners()",
		"33ea3dc8": "getTransaction(uint256)",
		"2e7700f0": "getTransactionCount()",
		"e783a1e6": "getTransactionId(address,uint256)",
		"80f59a65": "isConfirmed(uint256,address)",
		"2f54bf6e": "isOwner(address)",
		"025e7c27": "owners(uint256)",
		"c210a5f9": "proposalHashToTxId(bytes32)",
		"1703a018": "quorum()",
		"173825d9": "removeOwner(address)",
		"e20056e6": "replaceOwner(address,address)",
		"20ea8d86": "revokeConfirmation(uint256)",
		"c6427474": "submitTransaction(address,uint256,bytes)",
		"9ace38c2": "transactions(uint256)",
	},
	Bin: "0x60806040523480156200001157600080fd5b50604051620020bb380380620020bb833981016040819052620000349162000301565b60008251116200009e5760405162461bcd60e51b815260206004820152602a60248201527f4d756c74695369673a204f776e657273206c656e677468206d757374206265206044820152696174206c65617374203160b01b60648201526084015b60405180910390fd5b6002825111156200013a57600181118015620000bb575081518111155b6200012f5760405162461bcd60e51b815260206004820152603a60248201527f4d756c74695369673a204e756d626572206f6620636f6e6669726d6174696f6e60448201527f7320646f6573206e6f7420736174697366792071756f72756d2e000000000000606482015260840162000095565b600081905562000140565b60016000555b60005b8251811015620002c5576000838281518110620001645762000164620003db565b6020026020010151905060006001600160a01b0316816001600160a01b031603620001d25760405162461bcd60e51b815260206004820152601c60248201527f4d756c74695369673a204f776e65722063616e6e6f7420626520302e00000000604482015260640162000095565b6001600160a01b03811660009081526004602052604090205460ff16156200024c5760405162461bcd60e51b815260206004820152602660248201527f4d756c74695369673a204f776e65722041646472657373206973206475706c6960448201526531b0ba32b21760d11b606482015260840162000095565b6001600160a01b03166000818152600460205260408120805460ff191660019081179091556003805491820181559091527fc2575a0e9e593c00f959f8c92f12db2869c3395a3b0502d05e2516446f71f85b0180546001600160a01b031916909117905580620002bc81620003f1565b91505062000143565b50505062000419565b634e487b7160e01b600052604160045260246000fd5b80516001600160a01b0381168114620002fc57600080fd5b919050565b600080604083850312156200031557600080fd5b82516001600160401b03808211156200032d57600080fd5b818501915085601f8301126200034257600080fd5b8151602082821115620003595762000359620002ce565b8160051b604051601f19603f83011681018181108682111715620003815762000381620002ce565b604052928352818301935084810182019289841115620003a057600080fd5b948201945b83861015620003c957620003b986620002e4565b85529482019493820193620003a5565b97909101519698969750505050505050565b634e487b7160e01b600052603260045260246000fd5b6000600182016200041257634e487b7160e01b600052601160045260246000fd5b5060010190565b611c9280620004296000396000f3fe6080604052600436106101235760003560e01c8063a0e67e2b116100a0578063d74f8edd11610064578063d74f8edd146103c5578063e20056e6146103da578063e783a1e6146103fa578063ee22610b14610470578063ef18374a1461048357600080fd5b8063a0e67e2b14610316578063c01a8c8414610338578063c210a5f914610358578063c642747414610385578063d2cd96bd146103a557600080fd5b80632f54bf6e116100e75780632f54bf6e1461022157806333ea3dc81461026a5780637065cb481461029b57806380f59a65146102bb5780639ace38c2146102f657600080fd5b8063025e7c27146101695780631703a018146101a6578063173825d9146101ca57806320ea8d86146101ec5780632e7700f01461020c57600080fd5b36610164576040805134815247602082015233917f90890809c654f11d6e72a28fa60149770a0d11ec6c92319d6ceb2bb0a4ea1a15910160405180910390a2005b600080fd5b34801561017557600080fd5b50610189610184366004611763565b610498565b6040516001600160a01b0390911681526020015b60405180910390f35b3480156101b257600080fd5b506101bc60005481565b60405190815260200161019d565b3480156101d657600080fd5b506101ea6101e5366004611798565b6104c2565b005b3480156101f857600080fd5b506101ea610207366004611763565b61070b565b34801561021857600080fd5b506001546101bc565b34801561022d57600080fd5b5061025a61023c366004611798565b6001600160a01b031660009081526004602052604090205460ff1690565b604051901515815260200161019d565b34801561027657600080fd5b5061028a610285366004611763565b6108a9565b60405161019d959493929190611807565b3480156102a757600080fd5b506101ea6102b6366004611798565b6109a4565b3480156102c757600080fd5b5061025a6102d6366004611842565b600560209081526000928352604080842090915290825290205460ff1681565b34801561030257600080fd5b5061028a610311366004611763565b610bef565b34801561032257600080fd5b5061032b610cc8565b60405161019d919061186e565b34801561034457600080fd5b506101ea610353366004611763565b610d2a565b34801561036457600080fd5b506101bc610373366004611763565b60026020526000908152604090205481565b34801561039157600080fd5b506101ea6103a03660046118d1565b610ed4565b3480156103b157600080fd5b506101ea6103c0366004611763565b611111565b3480156103d157600080fd5b506101bc603281565b3480156103e657600080fd5b506101ea6103f536600461199c565b61122f565b34801561040657600080fd5b506101bc6104153660046119c6565b6040516bffffffffffffffffffffffff19606084901b16602082015260348101829052600090600290829060540160405160208183030381529060405280519060200120815260200190815260200160002054905092915050565b6101ea61047e366004611763565b611476565b34801561048f57600080fd5b506003546101bc565b600381815481106104a857600080fd5b6000918252602090912001546001600160a01b0316905081565b6003546001036105215760036000815481106104e0576104e06119f0565b6000918252602090912001546001600160a01b0316331461051c5760405162461bcd60e51b815260040161051390611a06565b60405180910390fd5b610540565b3330146105405760405162461bcd60e51b815260040161051390611a3b565b6001600160a01b038116600090815260046020526040902054819060ff1661057a5760405162461bcd60e51b815260040161051390611a06565b6001600160a01b0382166000908152600460205260408120805460ff191690555b60035481101561069657826001600160a01b0316600382815481106105c2576105c26119f0565b6000918252602090912001546001600160a01b03160361068e57600380546105ec90600190611a92565b815481106105fc576105fc6119f0565b600091825260209091200154600380546001600160a01b039092169183908110610628576106286119f0565b9060005260206000200160006101000a8154816001600160a01b0302191690836001600160a01b03160217905550600380548061066757610667611aa9565b600082815260209020810160001990810180546001600160a01b0319169055019055610696565b60010161059b565b5060035460005411156106d35760035460008181556040517fe5628a724014ba3eb778c374103da76a324247bc76e142fb6c4a597fa8a493db9190a25b6040516001600160a01b038316907fac6e8398676cf37429d530b81144d7079e99f4fe9d28b0d88c4a749ceccbe8cd90600090a25050565b3360009081526004602052604090205460ff1661073a5760405162461bcd60e51b815260040161051390611a06565b6001548190811061075d5760405162461bcd60e51b815260040161051390611abf565b8160018181548110610771576107716119f0565b600091825260209091206003600590920201015460ff16156107a55760405162461bcd60e51b815260040161051390611b04565b6000600184815481106107ba576107ba6119f0565b600091825260208083208784526005808352604080862033875290935291909320549102909101915060ff166108415760405162461bcd60e51b815260206004820152602660248201527f4d756c74695369673a205472616e73616374696f6e206973206e6f7420636f6e604482015265199a5c9b595960d21b6064820152608401610513565b60018160040160008282546108569190611a92565b90915550506000848152600560209081526040808320338085529252808320805460ff191690555186927ff0dca620e2e81f7841d07bcc105e1704fb01475b278a9d4c236e1c62945edd5591a350505050565b60008060606000806000600187815481106108c6576108c66119f0565b6000918252602090912060059091020180546001820154600383015460048401546002850180549596506001600160a01b039094169492939260ff90921691839061091090611b4e565b80601f016020809104026020016040519081016040528092919081815260200182805461093c90611b4e565b80156109895780601f1061095e57610100808354040283529160200191610989565b820191906000526020600020905b81548152906001019060200180831161096c57829003601f168201915b50505050509250955095509550955095505091939590929450565b6003546001036109fa5760036000815481106109c2576109c26119f0565b6000918252602090912001546001600160a01b031633146109f55760405162461bcd60e51b815260040161051390611a06565b610a19565b333014610a195760405162461bcd60e51b815260040161051390611a3b565b806001600160a01b038116610a705760405162461bcd60e51b815260206004820152601c60248201527f4d756c74695369673a204f776e65722063616e6e6f7420626520302e000000006044820152606401610513565b6001600160a01b038216600090815260046020526040902054829060ff1615610adb5760405162461bcd60e51b815260206004820152601f60248201527f4d756c74695369673a204f776e65722063616e206e6f74206163636573732e006044820152606401610513565b600354610ae9906001611b88565b60005460328211158015610afd5750818111155b8015610b0857508015155b8015610b1357508115155b610b5f5760405162461bcd60e51b815260206004820152601e60248201527f4d756c74695369673a20496e76616c696420526571756972656d656e742e00006044820152606401610513565b6001600160a01b038516600081815260046020526040808220805460ff1916600190811790915560038054918201815583527fc2575a0e9e593c00f959f8c92f12db2869c3395a3b0502d05e2516446f71f85b0180546001600160a01b03191684179055517fac1e9ef41b54c676ccf449d83ae6f2624bcdce8f5b93a6b48ce95874c332693d9190a25050505050565b60018181548110610bff57600080fd5b60009182526020909120600590910201805460018201546002830180546001600160a01b039093169450909291610c3590611b4e565b80601f0160208091040260200160405190810160405280929190818152602001828054610c6190611b4e565b8015610cae5780601f10610c8357610100808354040283529160200191610cae565b820191906000526020600020905b815481529060010190602001808311610c9157829003601f168201915b505050506003830154600490930154919260ff1691905085565b60606003805480602002602001604051908101604052809291908181526020018280548015610d2057602002820191906000526020600020905b81546001600160a01b03168152600190910190602001808311610d02575b5050505050905090565b3360009081526004602052604090205460ff16610d595760405162461bcd60e51b815260040161051390611a06565b60015481908110610d7c5760405162461bcd60e51b815260040161051390611abf565b8160018181548110610d9057610d906119f0565b600091825260209091206003600590920201015460ff1615610dc45760405162461bcd60e51b815260040161051390611b04565b6000838152600560209081526040808320338452909152902054839060ff1615610e435760405162461bcd60e51b815260206004820152602a60248201527f4d756c74695369673a205472616e73616374696f6e20697320616c72656164796044820152690818dbdb999a5c9b595960b21b6064820152608401610513565b600060018581548110610e5857610e586119f0565b906000526020600020906005020190506001816004016000828254610e7d9190611b88565b90915550506000858152600560209081526040808320338085529252808320805460ff191660011790555187927f5cbe105e36805f7820e291f799d5794ff948af2a5f664e580382defb6339004191a35050505050565b3360009081526004602052604090205460ff16610f035760405162461bcd60e51b815260040161051390611a06565b60018054600091610f149190611b88565b6040516bffffffffffffffffffffffff193360601b16602082015243603482015290915060009060540160408051601f1981840301815291815281516020928301206000818152600290935291205490915015610fb35760405162461bcd60e51b815260206004820181905260248201527f4475706c69636174652070726f706f73616c20696e2073616d6520626c6f636b6044820152606401610513565b6000818152600260209081526040808320859055805160a0810182526001600160a01b0389811682528184018981529282018881526060830186905260808301869052600180548082018255965282517fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6600590970296870180546001600160a01b0319169190931617825592517fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf7860155915180519194929361109c937fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf801929101906116ca565b50606082015160038201805460ff19169115159190911790556080909101516004909101556040516001600160a01b03861690839033907fd5a05bf70715ad82a09a756320284a1b54c9ff74cd0f8cce6219e79b563fe59d906111029089908990611ba0565b60405180910390a45050505050565b60035460010361116757600360008154811061112f5761112f6119f0565b6000918252602090912001546001600160a01b031633146111625760405162461bcd60e51b815260040161051390611a06565b611186565b3330146111865760405162461bcd60e51b815260040161051390611a3b565b600354816032821180159061119b5750818111155b80156111a657508015155b80156111b157508115155b6111fd5760405162461bcd60e51b815260206004820152601e60248201527f4d756c74695369673a20496e76616c696420526571756972656d656e742e00006044820152606401610513565b600083815560405184917fe5628a724014ba3eb778c374103da76a324247bc76e142fb6c4a597fa8a493db91a2505050565b60035460010361128557600360008154811061124d5761124d6119f0565b6000918252602090912001546001600160a01b031633146112805760405162461bcd60e51b815260040161051390611a06565b6112a4565b3330146112a45760405162461bcd60e51b815260040161051390611a3b565b6001600160a01b038216600090815260046020526040902054829060ff166112de5760405162461bcd60e51b815260040161051390611a06565b6001600160a01b038216600090815260046020526040902054829060ff16156113495760405162461bcd60e51b815260206004820152601f60248201527f4d756c74695369673a204f776e65722063616e206e6f74206163636573732e006044820152606401610513565b60005b6003548110156113dc57846001600160a01b031660038281548110611373576113736119f0565b6000918252602090912001546001600160a01b0316036113d45783600382815481106113a1576113a16119f0565b9060005260206000200160006101000a8154816001600160a01b0302191690836001600160a01b031602179055506113dc565b60010161134c565b506001600160a01b03808516600081815260046020526040808220805460ff1990811690915593871682528082208054909416600117909355915190917fac6e8398676cf37429d530b81144d7079e99f4fe9d28b0d88c4a749ceccbe8cd91a26040516001600160a01b038416907fac1e9ef41b54c676ccf449d83ae6f2624bcdce8f5b93a6b48ce95874c332693d90600090a250505050565b3360009081526004602052604090205460ff166114a55760405162461bcd60e51b815260040161051390611a06565b600154819081106114c85760405162461bcd60e51b815260040161051390611abf565b81600181815481106114dc576114dc6119f0565b600091825260209091206003600590920201015460ff16156115105760405162461bcd60e51b815260040161051390611b04565b600060018481548110611525576115256119f0565b90600052602060002090600502019050600054816004015410156115cc5760405162461bcd60e51b815260206004820152605260248201527f4d756c74695369673a2043757272656e74204e756d626572204f6620436f6e6660448201527f69726d6174696f6e73206d7573742062652067726561746572207468616e206f606482015271391032b8bab0b6103a379038bab7b93ab69760711b608482015260a401610513565b60038101805460ff191660019081179091558154908201546040516000926001600160a01b03169190611603906002860190611bc1565b60006040518083038185875af1925050503d8060008114611640576040519150601f19603f3d011682016040523d82523d6000602084013e611645565b606091505b50509050806116965760405162461bcd60e51b815260206004820152601d60248201527f4d756c74695369673a205472616e73616374696f6e206661696c65642e0000006044820152606401610513565b604051859033907f5445f318f4f5fcfb66592e68e0cc5822aa15664039bd5f0ffde24c5a8142b1ac90600090a35050505050565b8280546116d690611b4e565b90600052602060002090601f0160209004810192826116f8576000855561173e565b82601f1061171157805160ff191683800117855561173e565b8280016001018555821561173e579182015b8281111561173e578251825591602001919060010190611723565b5061174a92915061174e565b5090565b5b8082111561174a576000815560010161174f565b60006020828403121561177557600080fd5b5035919050565b80356001600160a01b038116811461179357600080fd5b919050565b6000602082840312156117aa57600080fd5b6117b38261177c565b9392505050565b6000815180845260005b818110156117e0576020818501810151868301820152016117c4565b818111156117f2576000602083870101525b50601f01601f19169290920160200192915050565b60018060a01b038616815284602082015260a06040820152600061182e60a08301866117ba565b931515606083015250608001529392505050565b6000806040838503121561185557600080fd5b823591506118656020840161177c565b90509250929050565b6020808252825182820181905260009190848201906040850190845b818110156118af5783516001600160a01b03168352928401929184019160010161188a565b50909695505050505050565b634e487b7160e01b600052604160045260246000fd5b6000806000606084860312156118e657600080fd5b6118ef8461177c565b925060208401359150604084013567ffffffffffffffff8082111561191357600080fd5b818601915086601f83011261192757600080fd5b813581811115611939576119396118bb565b604051601f8201601f19908116603f01168101908382118183101715611961576119616118bb565b8160405282815289602084870101111561197a57600080fd5b8260208601602083013760006020848301015280955050505050509250925092565b600080604083850312156119af57600080fd5b6119b88361177c565b91506118656020840161177c565b600080604083850312156119d957600080fd5b6119e28361177c565b946020939093013593505050565b634e487b7160e01b600052603260045260246000fd5b6020808252818101527f4d756c74695369673a204f6e6c79204f776e65722063616e206163636573732e604082015260600190565b60208082526021908201527f4d756c74695369673a204f6e6c792057616c6c65742063616e206163636573736040820152601760f91b606082015260800190565b634e487b7160e01b600052601160045260246000fd5b600082821015611aa457611aa4611a7c565b500390565b634e487b7160e01b600052603160045260246000fd5b60208082526025908201527f4d756c74695369673a205472616e73616374696f6e20646f6573206e6f7420656040820152643c34b9ba1760d91b606082015260800190565b6020808252602a908201527f4d756c74695369673a205472616e73616374696f6e20697320616c72656164796040820152691032bc32b1baba32b21760b11b606082015260800190565b600181811c90821680611b6257607f821691505b602082108103611b8257634e487b7160e01b600052602260045260246000fd5b50919050565b60008219821115611b9b57611b9b611a7c565b500190565b828152604060208201526000611bb960408301846117ba565b949350505050565b600080835481600182811c915080831680611bdd57607f831692505b60208084108203611bfc57634e487b7160e01b86526022600452602486fd5b818015611c105760018114611c2157611c4e565b60ff19861689528489019650611c4e565b60008a81526020902060005b86811015611c465781548b820152908501908301611c2d565b505084890196505b50949897505050505050505056fea26469706673582212205431645aea37a8d2ad62d3fdad0b6b74e94ff93c6a81e704d39395dc0ed80e1564736f6c634300080e0033",
}

// MultiSigWalletABI is the input ABI used to generate the binding from.
// Deprecated: Use MultiSigWalletMetaData.ABI instead.
var MultiSigWalletABI = MultiSigWalletMetaData.ABI

// Deprecated: Use MultiSigWalletMetaData.Sigs instead.
// MultiSigWalletFuncSigs maps the 4-byte function signature to its string representation.
var MultiSigWalletFuncSigs = MultiSigWalletMetaData.Sigs

// MultiSigWalletBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use MultiSigWalletMetaData.Bin instead.
var MultiSigWalletBin = MultiSigWalletMetaData.Bin

// DeployMultiSigWallet deploys a new Ethereum contract, binding an instance of MultiSigWallet to it.
func DeployMultiSigWallet(auth *bind.TransactOpts, backend bind.ContractBackend, _owners []common.Address, _quorum *big.Int) (common.Address, *types.Transaction, *MultiSigWallet, error) {
	parsed, err := MultiSigWalletMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(MultiSigWalletBin), backend, _owners, _quorum)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &MultiSigWallet{MultiSigWalletCaller: MultiSigWalletCaller{contract: contract}, MultiSigWalletTransactor: MultiSigWalletTransactor{contract: contract}, MultiSigWalletFilterer: MultiSigWalletFilterer{contract: contract}}, nil
}

// MultiSigWallet is an auto generated Go binding around an Ethereum contract.
type MultiSigWallet struct {
	MultiSigWalletCaller     // Read-only binding to the contract
	MultiSigWalletTransactor // Write-only binding to the contract
	MultiSigWalletFilterer   // Log filterer for contract events
}

// MultiSigWalletCaller is an auto generated read-only Go binding around an Ethereum contract.
type MultiSigWalletCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MultiSigWalletTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MultiSigWalletTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MultiSigWalletFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MultiSigWalletFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MultiSigWalletSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MultiSigWalletSession struct {
	Contract     *MultiSigWallet   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MultiSigWalletCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MultiSigWalletCallerSession struct {
	Contract *MultiSigWalletCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// MultiSigWalletTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MultiSigWalletTransactorSession struct {
	Contract     *MultiSigWalletTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// MultiSigWalletRaw is an auto generated low-level Go binding around an Ethereum contract.
type MultiSigWalletRaw struct {
	Contract *MultiSigWallet // Generic contract binding to access the raw methods on
}

// MultiSigWalletCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MultiSigWalletCallerRaw struct {
	Contract *MultiSigWalletCaller // Generic read-only contract binding to access the raw methods on
}

// MultiSigWalletTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MultiSigWalletTransactorRaw struct {
	Contract *MultiSigWalletTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMultiSigWallet creates a new instance of MultiSigWallet, bound to a specific deployed contract.
func NewMultiSigWallet(address common.Address, backend bind.ContractBackend) (*MultiSigWallet, error) {
	contract, err := bindMultiSigWallet(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &MultiSigWallet{MultiSigWalletCaller: MultiSigWalletCaller{contract: contract}, MultiSigWalletTransactor: MultiSigWalletTransactor{contract: contract}, MultiSigWalletFilterer: MultiSigWalletFilterer{contract: contract}}, nil
}

// NewMultiSigWalletCaller creates a new read-only instance of MultiSigWallet, bound to a specific deployed contract.
func NewMultiSigWalletCaller(address common.Address, caller bind.ContractCaller) (*MultiSigWalletCaller, error) {
	contract, err := bindMultiSigWallet(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletCaller{contract: contract}, nil
}

// NewMultiSigWalletTransactor creates a new write-only instance of MultiSigWallet, bound to a specific deployed contract.
func NewMultiSigWalletTransactor(address common.Address, transactor bind.ContractTransactor) (*MultiSigWalletTransactor, error) {
	contract, err := bindMultiSigWallet(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletTransactor{contract: contract}, nil
}

// NewMultiSigWalletFilterer creates a new log filterer instance of MultiSigWallet, bound to a specific deployed contract.
func NewMultiSigWalletFilterer(address common.Address, filterer bind.ContractFilterer) (*MultiSigWalletFilterer, error) {
	contract, err := bindMultiSigWallet(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletFilterer{contract: contract}, nil
}

// bindMultiSigWallet binds a generic wrapper to an already deployed contract.
func bindMultiSigWallet(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := MultiSigWalletMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MultiSigWallet *MultiSigWalletRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MultiSigWallet.Contract.MultiSigWalletCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MultiSigWallet *MultiSigWalletRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.MultiSigWalletTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MultiSigWallet *MultiSigWalletRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.MultiSigWalletTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MultiSigWallet *MultiSigWalletCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MultiSigWallet.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MultiSigWallet *MultiSigWalletTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MultiSigWallet *MultiSigWalletTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.contract.Transact(opts, method, params...)
}

// MAXOWNERCOUNT is a free data retrieval call binding the contract method 0xd74f8edd.
//
// Solidity: function MAX_OWNER_COUNT() view returns(uint256)
func (_MultiSigWallet *MultiSigWalletCaller) MAXOWNERCOUNT(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _MultiSigWallet.contract.Call(opts, &out, "MAX_OWNER_COUNT")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXOWNERCOUNT is a free data retrieval call binding the contract method 0xd74f8edd.
//
// Solidity: function MAX_OWNER_COUNT() view returns(uint256)
func (_MultiSigWallet *MultiSigWalletSession) MAXOWNERCOUNT() (*big.Int, error) {
	return _MultiSigWallet.Contract.MAXOWNERCOUNT(&_MultiSigWallet.CallOpts)
}

// MAXOWNERCOUNT is a free data retrieval call binding the contract method 0xd74f8edd.
//
// Solidity: function MAX_OWNER_COUNT() view returns(uint256)
func (_MultiSigWallet *MultiSigWalletCallerSession) MAXOWNERCOUNT() (*big.Int, error) {
	return _MultiSigWallet.Contract.MAXOWNERCOUNT(&_MultiSigWallet.CallOpts)
}

// GetOwnerCount is a free data retrieval call binding the contract method 0xef18374a.
//
// Solidity: function getOwnerCount() view returns(uint256)
func (_MultiSigWallet *MultiSigWalletCaller) GetOwnerCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _MultiSigWallet.contract.Call(opts, &out, "getOwnerCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetOwnerCount is a free data retrieval call binding the contract method 0xef18374a.
//
// Solidity: function getOwnerCount() view returns(uint256)
func (_MultiSigWallet *MultiSigWalletSession) GetOwnerCount() (*big.Int, error) {
	return _MultiSigWallet.Contract.GetOwnerCount(&_MultiSigWallet.CallOpts)
}

// GetOwnerCount is a free data retrieval call binding the contract method 0xef18374a.
//
// Solidity: function getOwnerCount() view returns(uint256)
func (_MultiSigWallet *MultiSigWalletCallerSession) GetOwnerCount() (*big.Int, error) {
	return _MultiSigWallet.Contract.GetOwnerCount(&_MultiSigWallet.CallOpts)
}

// GetOwners is a free data retrieval call binding the contract method 0xa0e67e2b.
//
// Solidity: function getOwners() view returns(address[])
func (_MultiSigWallet *MultiSigWalletCaller) GetOwners(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _MultiSigWallet.contract.Call(opts, &out, "getOwners")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetOwners is a free data retrieval call binding the contract method 0xa0e67e2b.
//
// Solidity: function getOwners() view returns(address[])
func (_MultiSigWallet *MultiSigWalletSession) GetOwners() ([]common.Address, error) {
	return _MultiSigWallet.Contract.GetOwners(&_MultiSigWallet.CallOpts)
}

// GetOwners is a free data retrieval call binding the contract method 0xa0e67e2b.
//
// Solidity: function getOwners() view returns(address[])
func (_MultiSigWallet *MultiSigWalletCallerSession) GetOwners() ([]common.Address, error) {
	return _MultiSigWallet.Contract.GetOwners(&_MultiSigWallet.CallOpts)
}

// GetTransaction is a free data retrieval call binding the contract method 0x33ea3dc8.
//
// Solidity: function getTransaction(uint256 _transactionId) view returns(address to, uint256 value, bytes data, bool executed, uint256 currentNumberOfConfirmations)
func (_MultiSigWallet *MultiSigWalletCaller) GetTransaction(opts *bind.CallOpts, _transactionId *big.Int) (struct {
	To                           common.Address
	Value                        *big.Int
	Data                         []byte
	Executed                     bool
	CurrentNumberOfConfirmations *big.Int
}, error) {
	var out []interface{}
	err := _MultiSigWallet.contract.Call(opts, &out, "getTransaction", _transactionId)

	outstruct := new(struct {
		To                           common.Address
		Value                        *big.Int
		Data                         []byte
		Executed                     bool
		CurrentNumberOfConfirmations *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.To = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.Value = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.Data = *abi.ConvertType(out[2], new([]byte)).(*[]byte)
	outstruct.Executed = *abi.ConvertType(out[3], new(bool)).(*bool)
	outstruct.CurrentNumberOfConfirmations = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetTransaction is a free data retrieval call binding the contract method 0x33ea3dc8.
//
// Solidity: function getTransaction(uint256 _transactionId) view returns(address to, uint256 value, bytes data, bool executed, uint256 currentNumberOfConfirmations)
func (_MultiSigWallet *MultiSigWalletSession) GetTransaction(_transactionId *big.Int) (struct {
	To                           common.Address
	Value                        *big.Int
	Data                         []byte
	Executed                     bool
	CurrentNumberOfConfirmations *big.Int
}, error) {
	return _MultiSigWallet.Contract.GetTransaction(&_MultiSigWallet.CallOpts, _transactionId)
}

// GetTransaction is a free data retrieval call binding the contract method 0x33ea3dc8.
//
// Solidity: function getTransaction(uint256 _transactionId) view returns(address to, uint256 value, bytes data, bool executed, uint256 currentNumberOfConfirmations)
func (_MultiSigWallet *MultiSigWalletCallerSession) GetTransaction(_transactionId *big.Int) (struct {
	To                           common.Address
	Value                        *big.Int
	Data                         []byte
	Executed                     bool
	CurrentNumberOfConfirmations *big.Int
}, error) {
	return _MultiSigWallet.Contract.GetTransaction(&_MultiSigWallet.CallOpts, _transactionId)
}

// GetTransactionCount is a free data retrieval call binding the contract method 0x2e7700f0.
//
// Solidity: function getTransactionCount() view returns(uint256)
func (_MultiSigWallet *MultiSigWalletCaller) GetTransactionCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _MultiSigWallet.contract.Call(opts, &out, "getTransactionCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTransactionCount is a free data retrieval call binding the contract method 0x2e7700f0.
//
// Solidity: function getTransactionCount() view returns(uint256)
func (_MultiSigWallet *MultiSigWalletSession) GetTransactionCount() (*big.Int, error) {
	return _MultiSigWallet.Contract.GetTransactionCount(&_MultiSigWallet.CallOpts)
}

// GetTransactionCount is a free data retrieval call binding the contract method 0x2e7700f0.
//
// Solidity: function getTransactionCount() view returns(uint256)
func (_MultiSigWallet *MultiSigWalletCallerSession) GetTransactionCount() (*big.Int, error) {
	return _MultiSigWallet.Contract.GetTransactionCount(&_MultiSigWallet.CallOpts)
}

// GetTransactionId is a free data retrieval call binding the contract method 0xe783a1e6.
//
// Solidity: function getTransactionId(address proposer, uint256 blockNumber) view returns(uint256)
func (_MultiSigWallet *MultiSigWalletCaller) GetTransactionId(opts *bind.CallOpts, proposer common.Address, blockNumber *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _MultiSigWallet.contract.Call(opts, &out, "getTransactionId", proposer, blockNumber)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTransactionId is a free data retrieval call binding the contract method 0xe783a1e6.
//
// Solidity: function getTransactionId(address proposer, uint256 blockNumber) view returns(uint256)
func (_MultiSigWallet *MultiSigWalletSession) GetTransactionId(proposer common.Address, blockNumber *big.Int) (*big.Int, error) {
	return _MultiSigWallet.Contract.GetTransactionId(&_MultiSigWallet.CallOpts, proposer, blockNumber)
}

// GetTransactionId is a free data retrieval call binding the contract method 0xe783a1e6.
//
// Solidity: function getTransactionId(address proposer, uint256 blockNumber) view returns(uint256)
func (_MultiSigWallet *MultiSigWalletCallerSession) GetTransactionId(proposer common.Address, blockNumber *big.Int) (*big.Int, error) {
	return _MultiSigWallet.Contract.GetTransactionId(&_MultiSigWallet.CallOpts, proposer, blockNumber)
}

// IsConfirmed is a free data retrieval call binding the contract method 0x80f59a65.
//
// Solidity: function isConfirmed(uint256 , address ) view returns(bool)
func (_MultiSigWallet *MultiSigWalletCaller) IsConfirmed(opts *bind.CallOpts, arg0 *big.Int, arg1 common.Address) (bool, error) {
	var out []interface{}
	err := _MultiSigWallet.contract.Call(opts, &out, "isConfirmed", arg0, arg1)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsConfirmed is a free data retrieval call binding the contract method 0x80f59a65.
//
// Solidity: function isConfirmed(uint256 , address ) view returns(bool)
func (_MultiSigWallet *MultiSigWalletSession) IsConfirmed(arg0 *big.Int, arg1 common.Address) (bool, error) {
	return _MultiSigWallet.Contract.IsConfirmed(&_MultiSigWallet.CallOpts, arg0, arg1)
}

// IsConfirmed is a free data retrieval call binding the contract method 0x80f59a65.
//
// Solidity: function isConfirmed(uint256 , address ) view returns(bool)
func (_MultiSigWallet *MultiSigWalletCallerSession) IsConfirmed(arg0 *big.Int, arg1 common.Address) (bool, error) {
	return _MultiSigWallet.Contract.IsConfirmed(&_MultiSigWallet.CallOpts, arg0, arg1)
}

// IsOwner is a free data retrieval call binding the contract method 0x2f54bf6e.
//
// Solidity: function isOwner(address _owner) view returns(bool)
func (_MultiSigWallet *MultiSigWalletCaller) IsOwner(opts *bind.CallOpts, _owner common.Address) (bool, error) {
	var out []interface{}
	err := _MultiSigWallet.contract.Call(opts, &out, "isOwner", _owner)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsOwner is a free data retrieval call binding the contract method 0x2f54bf6e.
//
// Solidity: function isOwner(address _owner) view returns(bool)
func (_MultiSigWallet *MultiSigWalletSession) IsOwner(_owner common.Address) (bool, error) {
	return _MultiSigWallet.Contract.IsOwner(&_MultiSigWallet.CallOpts, _owner)
}

// IsOwner is a free data retrieval call binding the contract method 0x2f54bf6e.
//
// Solidity: function isOwner(address _owner) view returns(bool)
func (_MultiSigWallet *MultiSigWalletCallerSession) IsOwner(_owner common.Address) (bool, error) {
	return _MultiSigWallet.Contract.IsOwner(&_MultiSigWallet.CallOpts, _owner)
}

// Owners is a free data retrieval call binding the contract method 0x025e7c27.
//
// Solidity: function owners(uint256 ) view returns(address)
func (_MultiSigWallet *MultiSigWalletCaller) Owners(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _MultiSigWallet.contract.Call(opts, &out, "owners", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owners is a free data retrieval call binding the contract method 0x025e7c27.
//
// Solidity: function owners(uint256 ) view returns(address)
func (_MultiSigWallet *MultiSigWalletSession) Owners(arg0 *big.Int) (common.Address, error) {
	return _MultiSigWallet.Contract.Owners(&_MultiSigWallet.CallOpts, arg0)
}

// Owners is a free data retrieval call binding the contract method 0x025e7c27.
//
// Solidity: function owners(uint256 ) view returns(address)
func (_MultiSigWallet *MultiSigWalletCallerSession) Owners(arg0 *big.Int) (common.Address, error) {
	return _MultiSigWallet.Contract.Owners(&_MultiSigWallet.CallOpts, arg0)
}

// ProposalHashToTxId is a free data retrieval call binding the contract method 0xc210a5f9.
//
// Solidity: function proposalHashToTxId(bytes32 ) view returns(uint256)
func (_MultiSigWallet *MultiSigWalletCaller) ProposalHashToTxId(opts *bind.CallOpts, arg0 [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _MultiSigWallet.contract.Call(opts, &out, "proposalHashToTxId", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ProposalHashToTxId is a free data retrieval call binding the contract method 0xc210a5f9.
//
// Solidity: function proposalHashToTxId(bytes32 ) view returns(uint256)
func (_MultiSigWallet *MultiSigWalletSession) ProposalHashToTxId(arg0 [32]byte) (*big.Int, error) {
	return _MultiSigWallet.Contract.ProposalHashToTxId(&_MultiSigWallet.CallOpts, arg0)
}

// ProposalHashToTxId is a free data retrieval call binding the contract method 0xc210a5f9.
//
// Solidity: function proposalHashToTxId(bytes32 ) view returns(uint256)
func (_MultiSigWallet *MultiSigWalletCallerSession) ProposalHashToTxId(arg0 [32]byte) (*big.Int, error) {
	return _MultiSigWallet.Contract.ProposalHashToTxId(&_MultiSigWallet.CallOpts, arg0)
}

// Quorum is a free data retrieval call binding the contract method 0x1703a018.
//
// Solidity: function quorum() view returns(uint256)
func (_MultiSigWallet *MultiSigWalletCaller) Quorum(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _MultiSigWallet.contract.Call(opts, &out, "quorum")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Quorum is a free data retrieval call binding the contract method 0x1703a018.
//
// Solidity: function quorum() view returns(uint256)
func (_MultiSigWallet *MultiSigWalletSession) Quorum() (*big.Int, error) {
	return _MultiSigWallet.Contract.Quorum(&_MultiSigWallet.CallOpts)
}

// Quorum is a free data retrieval call binding the contract method 0x1703a018.
//
// Solidity: function quorum() view returns(uint256)
func (_MultiSigWallet *MultiSigWalletCallerSession) Quorum() (*big.Int, error) {
	return _MultiSigWallet.Contract.Quorum(&_MultiSigWallet.CallOpts)
}

// Transactions is a free data retrieval call binding the contract method 0x9ace38c2.
//
// Solidity: function transactions(uint256 ) view returns(address to, uint256 value, bytes data, bool executed, uint256 currentNumberOfConfirmations)
func (_MultiSigWallet *MultiSigWalletCaller) Transactions(opts *bind.CallOpts, arg0 *big.Int) (struct {
	To                           common.Address
	Value                        *big.Int
	Data                         []byte
	Executed                     bool
	CurrentNumberOfConfirmations *big.Int
}, error) {
	var out []interface{}
	err := _MultiSigWallet.contract.Call(opts, &out, "transactions", arg0)

	outstruct := new(struct {
		To                           common.Address
		Value                        *big.Int
		Data                         []byte
		Executed                     bool
		CurrentNumberOfConfirmations *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.To = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.Value = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.Data = *abi.ConvertType(out[2], new([]byte)).(*[]byte)
	outstruct.Executed = *abi.ConvertType(out[3], new(bool)).(*bool)
	outstruct.CurrentNumberOfConfirmations = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Transactions is a free data retrieval call binding the contract method 0x9ace38c2.
//
// Solidity: function transactions(uint256 ) view returns(address to, uint256 value, bytes data, bool executed, uint256 currentNumberOfConfirmations)
func (_MultiSigWallet *MultiSigWalletSession) Transactions(arg0 *big.Int) (struct {
	To                           common.Address
	Value                        *big.Int
	Data                         []byte
	Executed                     bool
	CurrentNumberOfConfirmations *big.Int
}, error) {
	return _MultiSigWallet.Contract.Transactions(&_MultiSigWallet.CallOpts, arg0)
}

// Transactions is a free data retrieval call binding the contract method 0x9ace38c2.
//
// Solidity: function transactions(uint256 ) view returns(address to, uint256 value, bytes data, bool executed, uint256 currentNumberOfConfirmations)
func (_MultiSigWallet *MultiSigWalletCallerSession) Transactions(arg0 *big.Int) (struct {
	To                           common.Address
	Value                        *big.Int
	Data                         []byte
	Executed                     bool
	CurrentNumberOfConfirmations *big.Int
}, error) {
	return _MultiSigWallet.Contract.Transactions(&_MultiSigWallet.CallOpts, arg0)
}

// AddOwner is a paid mutator transaction binding the contract method 0x7065cb48.
//
// Solidity: function addOwner(address _newOwner) returns()
func (_MultiSigWallet *MultiSigWalletTransactor) AddOwner(opts *bind.TransactOpts, _newOwner common.Address) (*types.Transaction, error) {
	return _MultiSigWallet.contract.Transact(opts, "addOwner", _newOwner)
}

// AddOwner is a paid mutator transaction binding the contract method 0x7065cb48.
//
// Solidity: function addOwner(address _newOwner) returns()
func (_MultiSigWallet *MultiSigWalletSession) AddOwner(_newOwner common.Address) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.AddOwner(&_MultiSigWallet.TransactOpts, _newOwner)
}

// AddOwner is a paid mutator transaction binding the contract method 0x7065cb48.
//
// Solidity: function addOwner(address _newOwner) returns()
func (_MultiSigWallet *MultiSigWalletTransactorSession) AddOwner(_newOwner common.Address) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.AddOwner(&_MultiSigWallet.TransactOpts, _newOwner)
}

// ChangeQuorum is a paid mutator transaction binding the contract method 0xd2cd96bd.
//
// Solidity: function changeQuorum(uint256 _quorum) returns()
func (_MultiSigWallet *MultiSigWalletTransactor) ChangeQuorum(opts *bind.TransactOpts, _quorum *big.Int) (*types.Transaction, error) {
	return _MultiSigWallet.contract.Transact(opts, "changeQuorum", _quorum)
}

// ChangeQuorum is a paid mutator transaction binding the contract method 0xd2cd96bd.
//
// Solidity: function changeQuorum(uint256 _quorum) returns()
func (_MultiSigWallet *MultiSigWalletSession) ChangeQuorum(_quorum *big.Int) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.ChangeQuorum(&_MultiSigWallet.TransactOpts, _quorum)
}

// ChangeQuorum is a paid mutator transaction binding the contract method 0xd2cd96bd.
//
// Solidity: function changeQuorum(uint256 _quorum) returns()
func (_MultiSigWallet *MultiSigWalletTransactorSession) ChangeQuorum(_quorum *big.Int) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.ChangeQuorum(&_MultiSigWallet.TransactOpts, _quorum)
}

// ConfirmTransaction is a paid mutator transaction binding the contract method 0xc01a8c84.
//
// Solidity: function confirmTransaction(uint256 _transactionId) returns()
func (_MultiSigWallet *MultiSigWalletTransactor) ConfirmTransaction(opts *bind.TransactOpts, _transactionId *big.Int) (*types.Transaction, error) {
	return _MultiSigWallet.contract.Transact(opts, "confirmTransaction", _transactionId)
}

// ConfirmTransaction is a paid mutator transaction binding the contract method 0xc01a8c84.
//
// Solidity: function confirmTransaction(uint256 _transactionId) returns()
func (_MultiSigWallet *MultiSigWalletSession) ConfirmTransaction(_transactionId *big.Int) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.ConfirmTransaction(&_MultiSigWallet.TransactOpts, _transactionId)
}

// ConfirmTransaction is a paid mutator transaction binding the contract method 0xc01a8c84.
//
// Solidity: function confirmTransaction(uint256 _transactionId) returns()
func (_MultiSigWallet *MultiSigWalletTransactorSession) ConfirmTransaction(_transactionId *big.Int) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.ConfirmTransaction(&_MultiSigWallet.TransactOpts, _transactionId)
}

// ExecuteTransaction is a paid mutator transaction binding the contract method 0xee22610b.
//
// Solidity: function executeTransaction(uint256 _transactionId) payable returns()
func (_MultiSigWallet *MultiSigWalletTransactor) ExecuteTransaction(opts *bind.TransactOpts, _transactionId *big.Int) (*types.Transaction, error) {
	return _MultiSigWallet.contract.Transact(opts, "executeTransaction", _transactionId)
}

// ExecuteTransaction is a paid mutator transaction binding the contract method 0xee22610b.
//
// Solidity: function executeTransaction(uint256 _transactionId) payable returns()
func (_MultiSigWallet *MultiSigWalletSession) ExecuteTransaction(_transactionId *big.Int) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.ExecuteTransaction(&_MultiSigWallet.TransactOpts, _transactionId)
}

// ExecuteTransaction is a paid mutator transaction binding the contract method 0xee22610b.
//
// Solidity: function executeTransaction(uint256 _transactionId) payable returns()
func (_MultiSigWallet *MultiSigWalletTransactorSession) ExecuteTransaction(_transactionId *big.Int) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.ExecuteTransaction(&_MultiSigWallet.TransactOpts, _transactionId)
}

// RemoveOwner is a paid mutator transaction binding the contract method 0x173825d9.
//
// Solidity: function removeOwner(address _owner) returns()
func (_MultiSigWallet *MultiSigWalletTransactor) RemoveOwner(opts *bind.TransactOpts, _owner common.Address) (*types.Transaction, error) {
	return _MultiSigWallet.contract.Transact(opts, "removeOwner", _owner)
}

// RemoveOwner is a paid mutator transaction binding the contract method 0x173825d9.
//
// Solidity: function removeOwner(address _owner) returns()
func (_MultiSigWallet *MultiSigWalletSession) RemoveOwner(_owner common.Address) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.RemoveOwner(&_MultiSigWallet.TransactOpts, _owner)
}

// RemoveOwner is a paid mutator transaction binding the contract method 0x173825d9.
//
// Solidity: function removeOwner(address _owner) returns()
func (_MultiSigWallet *MultiSigWalletTransactorSession) RemoveOwner(_owner common.Address) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.RemoveOwner(&_MultiSigWallet.TransactOpts, _owner)
}

// ReplaceOwner is a paid mutator transaction binding the contract method 0xe20056e6.
//
// Solidity: function replaceOwner(address _owner, address _newOwner) returns()
func (_MultiSigWallet *MultiSigWalletTransactor) ReplaceOwner(opts *bind.TransactOpts, _owner common.Address, _newOwner common.Address) (*types.Transaction, error) {
	return _MultiSigWallet.contract.Transact(opts, "replaceOwner", _owner, _newOwner)
}

// ReplaceOwner is a paid mutator transaction binding the contract method 0xe20056e6.
//
// Solidity: function replaceOwner(address _owner, address _newOwner) returns()
func (_MultiSigWallet *MultiSigWalletSession) ReplaceOwner(_owner common.Address, _newOwner common.Address) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.ReplaceOwner(&_MultiSigWallet.TransactOpts, _owner, _newOwner)
}

// ReplaceOwner is a paid mutator transaction binding the contract method 0xe20056e6.
//
// Solidity: function replaceOwner(address _owner, address _newOwner) returns()
func (_MultiSigWallet *MultiSigWalletTransactorSession) ReplaceOwner(_owner common.Address, _newOwner common.Address) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.ReplaceOwner(&_MultiSigWallet.TransactOpts, _owner, _newOwner)
}

// RevokeConfirmation is a paid mutator transaction binding the contract method 0x20ea8d86.
//
// Solidity: function revokeConfirmation(uint256 _transactionId) returns()
func (_MultiSigWallet *MultiSigWalletTransactor) RevokeConfirmation(opts *bind.TransactOpts, _transactionId *big.Int) (*types.Transaction, error) {
	return _MultiSigWallet.contract.Transact(opts, "revokeConfirmation", _transactionId)
}

// RevokeConfirmation is a paid mutator transaction binding the contract method 0x20ea8d86.
//
// Solidity: function revokeConfirmation(uint256 _transactionId) returns()
func (_MultiSigWallet *MultiSigWalletSession) RevokeConfirmation(_transactionId *big.Int) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.RevokeConfirmation(&_MultiSigWallet.TransactOpts, _transactionId)
}

// RevokeConfirmation is a paid mutator transaction binding the contract method 0x20ea8d86.
//
// Solidity: function revokeConfirmation(uint256 _transactionId) returns()
func (_MultiSigWallet *MultiSigWalletTransactorSession) RevokeConfirmation(_transactionId *big.Int) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.RevokeConfirmation(&_MultiSigWallet.TransactOpts, _transactionId)
}

// SubmitTransaction is a paid mutator transaction binding the contract method 0xc6427474.
//
// Solidity: function submitTransaction(address _to, uint256 _value, bytes _data) returns()
func (_MultiSigWallet *MultiSigWalletTransactor) SubmitTransaction(opts *bind.TransactOpts, _to common.Address, _value *big.Int, _data []byte) (*types.Transaction, error) {
	return _MultiSigWallet.contract.Transact(opts, "submitTransaction", _to, _value, _data)
}

// SubmitTransaction is a paid mutator transaction binding the contract method 0xc6427474.
//
// Solidity: function submitTransaction(address _to, uint256 _value, bytes _data) returns()
func (_MultiSigWallet *MultiSigWalletSession) SubmitTransaction(_to common.Address, _value *big.Int, _data []byte) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.SubmitTransaction(&_MultiSigWallet.TransactOpts, _to, _value, _data)
}

// SubmitTransaction is a paid mutator transaction binding the contract method 0xc6427474.
//
// Solidity: function submitTransaction(address _to, uint256 _value, bytes _data) returns()
func (_MultiSigWallet *MultiSigWalletTransactorSession) SubmitTransaction(_to common.Address, _value *big.Int, _data []byte) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.SubmitTransaction(&_MultiSigWallet.TransactOpts, _to, _value, _data)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_MultiSigWallet *MultiSigWalletTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MultiSigWallet.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_MultiSigWallet *MultiSigWalletSession) Receive() (*types.Transaction, error) {
	return _MultiSigWallet.Contract.Receive(&_MultiSigWallet.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_MultiSigWallet *MultiSigWalletTransactorSession) Receive() (*types.Transaction, error) {
	return _MultiSigWallet.Contract.Receive(&_MultiSigWallet.TransactOpts)
}

// MultiSigWalletAddOwnerIterator is returned from FilterAddOwner and is used to iterate over the raw logs and unpacked data for AddOwner events raised by the MultiSigWallet contract.
type MultiSigWalletAddOwnerIterator struct {
	Event *MultiSigWalletAddOwner // Event containing the contract specifics and raw log

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
func (it *MultiSigWalletAddOwnerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiSigWalletAddOwner)
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
		it.Event = new(MultiSigWalletAddOwner)
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
func (it *MultiSigWalletAddOwnerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiSigWalletAddOwnerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiSigWalletAddOwner represents a AddOwner event raised by the MultiSigWallet contract.
type MultiSigWalletAddOwner struct {
	NewOwner common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterAddOwner is a free log retrieval operation binding the contract event 0xac1e9ef41b54c676ccf449d83ae6f2624bcdce8f5b93a6b48ce95874c332693d.
//
// Solidity: event AddOwner(address indexed newOwner)
func (_MultiSigWallet *MultiSigWalletFilterer) FilterAddOwner(opts *bind.FilterOpts, newOwner []common.Address) (*MultiSigWalletAddOwnerIterator, error) {

	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _MultiSigWallet.contract.FilterLogs(opts, "AddOwner", newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletAddOwnerIterator{contract: _MultiSigWallet.contract, event: "AddOwner", logs: logs, sub: sub}, nil
}

// WatchAddOwner is a free log subscription operation binding the contract event 0xac1e9ef41b54c676ccf449d83ae6f2624bcdce8f5b93a6b48ce95874c332693d.
//
// Solidity: event AddOwner(address indexed newOwner)
func (_MultiSigWallet *MultiSigWalletFilterer) WatchAddOwner(opts *bind.WatchOpts, sink chan<- *MultiSigWalletAddOwner, newOwner []common.Address) (event.Subscription, error) {

	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _MultiSigWallet.contract.WatchLogs(opts, "AddOwner", newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiSigWalletAddOwner)
				if err := _MultiSigWallet.contract.UnpackLog(event, "AddOwner", log); err != nil {
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

// ParseAddOwner is a log parse operation binding the contract event 0xac1e9ef41b54c676ccf449d83ae6f2624bcdce8f5b93a6b48ce95874c332693d.
//
// Solidity: event AddOwner(address indexed newOwner)
func (_MultiSigWallet *MultiSigWalletFilterer) ParseAddOwner(log types.Log) (*MultiSigWalletAddOwner, error) {
	event := new(MultiSigWalletAddOwner)
	if err := _MultiSigWallet.contract.UnpackLog(event, "AddOwner", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MultiSigWalletChangeQuorumIterator is returned from FilterChangeQuorum and is used to iterate over the raw logs and unpacked data for ChangeQuorum events raised by the MultiSigWallet contract.
type MultiSigWalletChangeQuorumIterator struct {
	Event *MultiSigWalletChangeQuorum // Event containing the contract specifics and raw log

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
func (it *MultiSigWalletChangeQuorumIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiSigWalletChangeQuorum)
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
		it.Event = new(MultiSigWalletChangeQuorum)
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
func (it *MultiSigWalletChangeQuorumIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiSigWalletChangeQuorumIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiSigWalletChangeQuorum represents a ChangeQuorum event raised by the MultiSigWallet contract.
type MultiSigWalletChangeQuorum struct {
	Quorum *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterChangeQuorum is a free log retrieval operation binding the contract event 0xe5628a724014ba3eb778c374103da76a324247bc76e142fb6c4a597fa8a493db.
//
// Solidity: event ChangeQuorum(uint256 indexed quorum)
func (_MultiSigWallet *MultiSigWalletFilterer) FilterChangeQuorum(opts *bind.FilterOpts, quorum []*big.Int) (*MultiSigWalletChangeQuorumIterator, error) {

	var quorumRule []interface{}
	for _, quorumItem := range quorum {
		quorumRule = append(quorumRule, quorumItem)
	}

	logs, sub, err := _MultiSigWallet.contract.FilterLogs(opts, "ChangeQuorum", quorumRule)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletChangeQuorumIterator{contract: _MultiSigWallet.contract, event: "ChangeQuorum", logs: logs, sub: sub}, nil
}

// WatchChangeQuorum is a free log subscription operation binding the contract event 0xe5628a724014ba3eb778c374103da76a324247bc76e142fb6c4a597fa8a493db.
//
// Solidity: event ChangeQuorum(uint256 indexed quorum)
func (_MultiSigWallet *MultiSigWalletFilterer) WatchChangeQuorum(opts *bind.WatchOpts, sink chan<- *MultiSigWalletChangeQuorum, quorum []*big.Int) (event.Subscription, error) {

	var quorumRule []interface{}
	for _, quorumItem := range quorum {
		quorumRule = append(quorumRule, quorumItem)
	}

	logs, sub, err := _MultiSigWallet.contract.WatchLogs(opts, "ChangeQuorum", quorumRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiSigWalletChangeQuorum)
				if err := _MultiSigWallet.contract.UnpackLog(event, "ChangeQuorum", log); err != nil {
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

// ParseChangeQuorum is a log parse operation binding the contract event 0xe5628a724014ba3eb778c374103da76a324247bc76e142fb6c4a597fa8a493db.
//
// Solidity: event ChangeQuorum(uint256 indexed quorum)
func (_MultiSigWallet *MultiSigWalletFilterer) ParseChangeQuorum(log types.Log) (*MultiSigWalletChangeQuorum, error) {
	event := new(MultiSigWalletChangeQuorum)
	if err := _MultiSigWallet.contract.UnpackLog(event, "ChangeQuorum", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MultiSigWalletConfirmTransactionIterator is returned from FilterConfirmTransaction and is used to iterate over the raw logs and unpacked data for ConfirmTransaction events raised by the MultiSigWallet contract.
type MultiSigWalletConfirmTransactionIterator struct {
	Event *MultiSigWalletConfirmTransaction // Event containing the contract specifics and raw log

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
func (it *MultiSigWalletConfirmTransactionIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiSigWalletConfirmTransaction)
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
		it.Event = new(MultiSigWalletConfirmTransaction)
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
func (it *MultiSigWalletConfirmTransactionIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiSigWalletConfirmTransactionIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiSigWalletConfirmTransaction represents a ConfirmTransaction event raised by the MultiSigWallet contract.
type MultiSigWalletConfirmTransaction struct {
	Owner   common.Address
	TxIndex *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterConfirmTransaction is a free log retrieval operation binding the contract event 0x5cbe105e36805f7820e291f799d5794ff948af2a5f664e580382defb63390041.
//
// Solidity: event ConfirmTransaction(address indexed owner, uint256 indexed txIndex)
func (_MultiSigWallet *MultiSigWalletFilterer) FilterConfirmTransaction(opts *bind.FilterOpts, owner []common.Address, txIndex []*big.Int) (*MultiSigWalletConfirmTransactionIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var txIndexRule []interface{}
	for _, txIndexItem := range txIndex {
		txIndexRule = append(txIndexRule, txIndexItem)
	}

	logs, sub, err := _MultiSigWallet.contract.FilterLogs(opts, "ConfirmTransaction", ownerRule, txIndexRule)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletConfirmTransactionIterator{contract: _MultiSigWallet.contract, event: "ConfirmTransaction", logs: logs, sub: sub}, nil
}

// WatchConfirmTransaction is a free log subscription operation binding the contract event 0x5cbe105e36805f7820e291f799d5794ff948af2a5f664e580382defb63390041.
//
// Solidity: event ConfirmTransaction(address indexed owner, uint256 indexed txIndex)
func (_MultiSigWallet *MultiSigWalletFilterer) WatchConfirmTransaction(opts *bind.WatchOpts, sink chan<- *MultiSigWalletConfirmTransaction, owner []common.Address, txIndex []*big.Int) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var txIndexRule []interface{}
	for _, txIndexItem := range txIndex {
		txIndexRule = append(txIndexRule, txIndexItem)
	}

	logs, sub, err := _MultiSigWallet.contract.WatchLogs(opts, "ConfirmTransaction", ownerRule, txIndexRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiSigWalletConfirmTransaction)
				if err := _MultiSigWallet.contract.UnpackLog(event, "ConfirmTransaction", log); err != nil {
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

// ParseConfirmTransaction is a log parse operation binding the contract event 0x5cbe105e36805f7820e291f799d5794ff948af2a5f664e580382defb63390041.
//
// Solidity: event ConfirmTransaction(address indexed owner, uint256 indexed txIndex)
func (_MultiSigWallet *MultiSigWalletFilterer) ParseConfirmTransaction(log types.Log) (*MultiSigWalletConfirmTransaction, error) {
	event := new(MultiSigWalletConfirmTransaction)
	if err := _MultiSigWallet.contract.UnpackLog(event, "ConfirmTransaction", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MultiSigWalletDepositIterator is returned from FilterDeposit and is used to iterate over the raw logs and unpacked data for Deposit events raised by the MultiSigWallet contract.
type MultiSigWalletDepositIterator struct {
	Event *MultiSigWalletDeposit // Event containing the contract specifics and raw log

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
func (it *MultiSigWalletDepositIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiSigWalletDeposit)
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
		it.Event = new(MultiSigWalletDeposit)
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
func (it *MultiSigWalletDepositIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiSigWalletDepositIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiSigWalletDeposit represents a Deposit event raised by the MultiSigWallet contract.
type MultiSigWalletDeposit struct {
	Sender  common.Address
	Amount  *big.Int
	Balance *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterDeposit is a free log retrieval operation binding the contract event 0x90890809c654f11d6e72a28fa60149770a0d11ec6c92319d6ceb2bb0a4ea1a15.
//
// Solidity: event Deposit(address indexed sender, uint256 amount, uint256 balance)
func (_MultiSigWallet *MultiSigWalletFilterer) FilterDeposit(opts *bind.FilterOpts, sender []common.Address) (*MultiSigWalletDepositIterator, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _MultiSigWallet.contract.FilterLogs(opts, "Deposit", senderRule)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletDepositIterator{contract: _MultiSigWallet.contract, event: "Deposit", logs: logs, sub: sub}, nil
}

// WatchDeposit is a free log subscription operation binding the contract event 0x90890809c654f11d6e72a28fa60149770a0d11ec6c92319d6ceb2bb0a4ea1a15.
//
// Solidity: event Deposit(address indexed sender, uint256 amount, uint256 balance)
func (_MultiSigWallet *MultiSigWalletFilterer) WatchDeposit(opts *bind.WatchOpts, sink chan<- *MultiSigWalletDeposit, sender []common.Address) (event.Subscription, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _MultiSigWallet.contract.WatchLogs(opts, "Deposit", senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiSigWalletDeposit)
				if err := _MultiSigWallet.contract.UnpackLog(event, "Deposit", log); err != nil {
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

// ParseDeposit is a log parse operation binding the contract event 0x90890809c654f11d6e72a28fa60149770a0d11ec6c92319d6ceb2bb0a4ea1a15.
//
// Solidity: event Deposit(address indexed sender, uint256 amount, uint256 balance)
func (_MultiSigWallet *MultiSigWalletFilterer) ParseDeposit(log types.Log) (*MultiSigWalletDeposit, error) {
	event := new(MultiSigWalletDeposit)
	if err := _MultiSigWallet.contract.UnpackLog(event, "Deposit", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MultiSigWalletExecuteTransactionIterator is returned from FilterExecuteTransaction and is used to iterate over the raw logs and unpacked data for ExecuteTransaction events raised by the MultiSigWallet contract.
type MultiSigWalletExecuteTransactionIterator struct {
	Event *MultiSigWalletExecuteTransaction // Event containing the contract specifics and raw log

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
func (it *MultiSigWalletExecuteTransactionIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiSigWalletExecuteTransaction)
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
		it.Event = new(MultiSigWalletExecuteTransaction)
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
func (it *MultiSigWalletExecuteTransactionIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiSigWalletExecuteTransactionIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiSigWalletExecuteTransaction represents a ExecuteTransaction event raised by the MultiSigWallet contract.
type MultiSigWalletExecuteTransaction struct {
	Owner   common.Address
	TxIndex *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterExecuteTransaction is a free log retrieval operation binding the contract event 0x5445f318f4f5fcfb66592e68e0cc5822aa15664039bd5f0ffde24c5a8142b1ac.
//
// Solidity: event ExecuteTransaction(address indexed owner, uint256 indexed txIndex)
func (_MultiSigWallet *MultiSigWalletFilterer) FilterExecuteTransaction(opts *bind.FilterOpts, owner []common.Address, txIndex []*big.Int) (*MultiSigWalletExecuteTransactionIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var txIndexRule []interface{}
	for _, txIndexItem := range txIndex {
		txIndexRule = append(txIndexRule, txIndexItem)
	}

	logs, sub, err := _MultiSigWallet.contract.FilterLogs(opts, "ExecuteTransaction", ownerRule, txIndexRule)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletExecuteTransactionIterator{contract: _MultiSigWallet.contract, event: "ExecuteTransaction", logs: logs, sub: sub}, nil
}

// WatchExecuteTransaction is a free log subscription operation binding the contract event 0x5445f318f4f5fcfb66592e68e0cc5822aa15664039bd5f0ffde24c5a8142b1ac.
//
// Solidity: event ExecuteTransaction(address indexed owner, uint256 indexed txIndex)
func (_MultiSigWallet *MultiSigWalletFilterer) WatchExecuteTransaction(opts *bind.WatchOpts, sink chan<- *MultiSigWalletExecuteTransaction, owner []common.Address, txIndex []*big.Int) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var txIndexRule []interface{}
	for _, txIndexItem := range txIndex {
		txIndexRule = append(txIndexRule, txIndexItem)
	}

	logs, sub, err := _MultiSigWallet.contract.WatchLogs(opts, "ExecuteTransaction", ownerRule, txIndexRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiSigWalletExecuteTransaction)
				if err := _MultiSigWallet.contract.UnpackLog(event, "ExecuteTransaction", log); err != nil {
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

// ParseExecuteTransaction is a log parse operation binding the contract event 0x5445f318f4f5fcfb66592e68e0cc5822aa15664039bd5f0ffde24c5a8142b1ac.
//
// Solidity: event ExecuteTransaction(address indexed owner, uint256 indexed txIndex)
func (_MultiSigWallet *MultiSigWalletFilterer) ParseExecuteTransaction(log types.Log) (*MultiSigWalletExecuteTransaction, error) {
	event := new(MultiSigWalletExecuteTransaction)
	if err := _MultiSigWallet.contract.UnpackLog(event, "ExecuteTransaction", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MultiSigWalletRemoveOwnerIterator is returned from FilterRemoveOwner and is used to iterate over the raw logs and unpacked data for RemoveOwner events raised by the MultiSigWallet contract.
type MultiSigWalletRemoveOwnerIterator struct {
	Event *MultiSigWalletRemoveOwner // Event containing the contract specifics and raw log

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
func (it *MultiSigWalletRemoveOwnerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiSigWalletRemoveOwner)
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
		it.Event = new(MultiSigWalletRemoveOwner)
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
func (it *MultiSigWalletRemoveOwnerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiSigWalletRemoveOwnerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiSigWalletRemoveOwner represents a RemoveOwner event raised by the MultiSigWallet contract.
type MultiSigWalletRemoveOwner struct {
	Owner common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterRemoveOwner is a free log retrieval operation binding the contract event 0xac6e8398676cf37429d530b81144d7079e99f4fe9d28b0d88c4a749ceccbe8cd.
//
// Solidity: event RemoveOwner(address indexed owner)
func (_MultiSigWallet *MultiSigWalletFilterer) FilterRemoveOwner(opts *bind.FilterOpts, owner []common.Address) (*MultiSigWalletRemoveOwnerIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _MultiSigWallet.contract.FilterLogs(opts, "RemoveOwner", ownerRule)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletRemoveOwnerIterator{contract: _MultiSigWallet.contract, event: "RemoveOwner", logs: logs, sub: sub}, nil
}

// WatchRemoveOwner is a free log subscription operation binding the contract event 0xac6e8398676cf37429d530b81144d7079e99f4fe9d28b0d88c4a749ceccbe8cd.
//
// Solidity: event RemoveOwner(address indexed owner)
func (_MultiSigWallet *MultiSigWalletFilterer) WatchRemoveOwner(opts *bind.WatchOpts, sink chan<- *MultiSigWalletRemoveOwner, owner []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _MultiSigWallet.contract.WatchLogs(opts, "RemoveOwner", ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiSigWalletRemoveOwner)
				if err := _MultiSigWallet.contract.UnpackLog(event, "RemoveOwner", log); err != nil {
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

// ParseRemoveOwner is a log parse operation binding the contract event 0xac6e8398676cf37429d530b81144d7079e99f4fe9d28b0d88c4a749ceccbe8cd.
//
// Solidity: event RemoveOwner(address indexed owner)
func (_MultiSigWallet *MultiSigWalletFilterer) ParseRemoveOwner(log types.Log) (*MultiSigWalletRemoveOwner, error) {
	event := new(MultiSigWalletRemoveOwner)
	if err := _MultiSigWallet.contract.UnpackLog(event, "RemoveOwner", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MultiSigWalletRevokeConfirmationIterator is returned from FilterRevokeConfirmation and is used to iterate over the raw logs and unpacked data for RevokeConfirmation events raised by the MultiSigWallet contract.
type MultiSigWalletRevokeConfirmationIterator struct {
	Event *MultiSigWalletRevokeConfirmation // Event containing the contract specifics and raw log

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
func (it *MultiSigWalletRevokeConfirmationIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiSigWalletRevokeConfirmation)
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
		it.Event = new(MultiSigWalletRevokeConfirmation)
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
func (it *MultiSigWalletRevokeConfirmationIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiSigWalletRevokeConfirmationIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiSigWalletRevokeConfirmation represents a RevokeConfirmation event raised by the MultiSigWallet contract.
type MultiSigWalletRevokeConfirmation struct {
	Owner   common.Address
	TxIndex *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRevokeConfirmation is a free log retrieval operation binding the contract event 0xf0dca620e2e81f7841d07bcc105e1704fb01475b278a9d4c236e1c62945edd55.
//
// Solidity: event RevokeConfirmation(address indexed owner, uint256 indexed txIndex)
func (_MultiSigWallet *MultiSigWalletFilterer) FilterRevokeConfirmation(opts *bind.FilterOpts, owner []common.Address, txIndex []*big.Int) (*MultiSigWalletRevokeConfirmationIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var txIndexRule []interface{}
	for _, txIndexItem := range txIndex {
		txIndexRule = append(txIndexRule, txIndexItem)
	}

	logs, sub, err := _MultiSigWallet.contract.FilterLogs(opts, "RevokeConfirmation", ownerRule, txIndexRule)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletRevokeConfirmationIterator{contract: _MultiSigWallet.contract, event: "RevokeConfirmation", logs: logs, sub: sub}, nil
}

// WatchRevokeConfirmation is a free log subscription operation binding the contract event 0xf0dca620e2e81f7841d07bcc105e1704fb01475b278a9d4c236e1c62945edd55.
//
// Solidity: event RevokeConfirmation(address indexed owner, uint256 indexed txIndex)
func (_MultiSigWallet *MultiSigWalletFilterer) WatchRevokeConfirmation(opts *bind.WatchOpts, sink chan<- *MultiSigWalletRevokeConfirmation, owner []common.Address, txIndex []*big.Int) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var txIndexRule []interface{}
	for _, txIndexItem := range txIndex {
		txIndexRule = append(txIndexRule, txIndexItem)
	}

	logs, sub, err := _MultiSigWallet.contract.WatchLogs(opts, "RevokeConfirmation", ownerRule, txIndexRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiSigWalletRevokeConfirmation)
				if err := _MultiSigWallet.contract.UnpackLog(event, "RevokeConfirmation", log); err != nil {
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

// ParseRevokeConfirmation is a log parse operation binding the contract event 0xf0dca620e2e81f7841d07bcc105e1704fb01475b278a9d4c236e1c62945edd55.
//
// Solidity: event RevokeConfirmation(address indexed owner, uint256 indexed txIndex)
func (_MultiSigWallet *MultiSigWalletFilterer) ParseRevokeConfirmation(log types.Log) (*MultiSigWalletRevokeConfirmation, error) {
	event := new(MultiSigWalletRevokeConfirmation)
	if err := _MultiSigWallet.contract.UnpackLog(event, "RevokeConfirmation", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MultiSigWalletSubmitTransactionIterator is returned from FilterSubmitTransaction and is used to iterate over the raw logs and unpacked data for SubmitTransaction events raised by the MultiSigWallet contract.
type MultiSigWalletSubmitTransactionIterator struct {
	Event *MultiSigWalletSubmitTransaction // Event containing the contract specifics and raw log

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
func (it *MultiSigWalletSubmitTransactionIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiSigWalletSubmitTransaction)
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
		it.Event = new(MultiSigWalletSubmitTransaction)
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
func (it *MultiSigWalletSubmitTransactionIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiSigWalletSubmitTransactionIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiSigWalletSubmitTransaction represents a SubmitTransaction event raised by the MultiSigWallet contract.
type MultiSigWalletSubmitTransaction struct {
	Owner   common.Address
	TxIndex *big.Int
	To      common.Address
	Value   *big.Int
	Data    []byte
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterSubmitTransaction is a free log retrieval operation binding the contract event 0xd5a05bf70715ad82a09a756320284a1b54c9ff74cd0f8cce6219e79b563fe59d.
//
// Solidity: event SubmitTransaction(address indexed owner, uint256 indexed txIndex, address indexed to, uint256 value, bytes data)
func (_MultiSigWallet *MultiSigWalletFilterer) FilterSubmitTransaction(opts *bind.FilterOpts, owner []common.Address, txIndex []*big.Int, to []common.Address) (*MultiSigWalletSubmitTransactionIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var txIndexRule []interface{}
	for _, txIndexItem := range txIndex {
		txIndexRule = append(txIndexRule, txIndexItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _MultiSigWallet.contract.FilterLogs(opts, "SubmitTransaction", ownerRule, txIndexRule, toRule)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletSubmitTransactionIterator{contract: _MultiSigWallet.contract, event: "SubmitTransaction", logs: logs, sub: sub}, nil
}

// WatchSubmitTransaction is a free log subscription operation binding the contract event 0xd5a05bf70715ad82a09a756320284a1b54c9ff74cd0f8cce6219e79b563fe59d.
//
// Solidity: event SubmitTransaction(address indexed owner, uint256 indexed txIndex, address indexed to, uint256 value, bytes data)
func (_MultiSigWallet *MultiSigWalletFilterer) WatchSubmitTransaction(opts *bind.WatchOpts, sink chan<- *MultiSigWalletSubmitTransaction, owner []common.Address, txIndex []*big.Int, to []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var txIndexRule []interface{}
	for _, txIndexItem := range txIndex {
		txIndexRule = append(txIndexRule, txIndexItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _MultiSigWallet.contract.WatchLogs(opts, "SubmitTransaction", ownerRule, txIndexRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiSigWalletSubmitTransaction)
				if err := _MultiSigWallet.contract.UnpackLog(event, "SubmitTransaction", log); err != nil {
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

// ParseSubmitTransaction is a log parse operation binding the contract event 0xd5a05bf70715ad82a09a756320284a1b54c9ff74cd0f8cce6219e79b563fe59d.
//
// Solidity: event SubmitTransaction(address indexed owner, uint256 indexed txIndex, address indexed to, uint256 value, bytes data)
func (_MultiSigWallet *MultiSigWalletFilterer) ParseSubmitTransaction(log types.Log) (*MultiSigWalletSubmitTransaction, error) {
	event := new(MultiSigWalletSubmitTransaction)
	if err := _MultiSigWallet.contract.UnpackLog(event, "SubmitTransaction", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
