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

// GovStakingMetaData contains all meta data concerning the GovStaking contract.
var GovStakingMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"staker\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Delegated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"credentialID\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"requester\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"time\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"unbonding\",\"type\":\"uint256\"}],\"name\":\"NewCredential\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"staker\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Staked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"staker\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"rewardee\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"staking\",\"type\":\"uint256\"}],\"name\":\"StakerRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"staker\",\"type\":\"address\"}],\"name\":\"StakerRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"staker\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Undelegated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"staker\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Unstaked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"credentialID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"requester\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Withdrawn\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"GOV_CONST\",\"outputs\":[{\"internalType\":\"contractGovConst\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"credentialCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"credentials\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"requester\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"requestTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"withdrawableTime\",\"type\":\"uint256\"},{\"internalType\":\"enumGovStaking.WithdrawalStatus\",\"name\":\"status\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_staker\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"delegate\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"delegateTo\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_addr\",\"type\":\"address\"}],\"name\":\"isOperatorOrRewardee\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_staker\",\"type\":\"address\"}],\"name\":\"isStaker\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"_staker\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_rewardee\",\"type\":\"address\"}],\"name\":\"registerStaker\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"stake\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"stakerByOperator\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"stakerByRewardee\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"stakerInfo\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"rewardee\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"staking\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"delegated\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"stakerLength\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"stakers\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalStaking\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_staker\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"undelegate\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"unstake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_cid\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Sigs: map[string]string{
		"e8aaca24": "GOV_CONST()",
		"cd0e35b9": "credentialCount()",
		"e0574e3f": "credentials(uint256)",
		"026e402b": "delegate(address,uint256)",
		"438bb7e5": "delegateTo(address,address)",
		"cdb72f06": "isOperatorOrRewardee(address)",
		"6f1e8533": "isStaker(address)",
		"af716c40": "registerStaker(uint256,address,address)",
		"a694fc3a": "stake(uint256)",
		"9dbccf6a": "stakerByOperator(address)",
		"264ebc38": "stakerByRewardee(address)",
		"4e745f1f": "stakerInfo(address)",
		"5748f6f3": "stakerLength()",
		"fed1252a": "stakers()",
		"165defa4": "totalStaking()",
		"4d99dd16": "undelegate(address,uint256)",
		"2e17de78": "unstake(uint256)",
		"2e1a7d4d": "withdraw(uint256)",
	},
	Bin: "0x608060405234801561001057600080fd5b50611a57806100206000396000f3fe6080604052600436106101095760003560e01c80636f1e853311610095578063cd0e35b911610064578063cd0e35b91461034f578063cdb72f0614610365578063e0574e3f14610385578063e8aaca24146103ed578063fed1252a1461040357600080fd5b80636f1e8533146102c35780639dbccf6a146102f3578063a694fc3a14610329578063af716c401461033c57600080fd5b80632e1a7d4d116100dc5780632e1a7d4d146101ba578063438bb7e5146101da5780634d99dd16146102125780634e745f1f146102325780635748f6f3146102ae57600080fd5b8063026e402b1461010e578063165defa414610123578063264ebc381461014c5780632e17de781461019a575b600080fd5b61012161011c3660046117c4565b610425565b005b34801561012f57600080fd5b5061013960005481565b6040519081526020015b60405180910390f35b34801561015857600080fd5b506101826101673660046117ee565b6005602052600090815260409020546001600160a01b031681565b6040516001600160a01b039091168152602001610143565b3480156101a657600080fd5b506101216101b5366004611809565b610587565b3480156101c657600080fd5b506101216101d5366004611809565b610901565b3480156101e657600080fd5b506101396101f5366004611822565b600660209081526000928352604080842090915290825290205481565b34801561021e57600080fd5b5061012161022d3660046117c4565b610a89565b34801561023e57600080fd5b5061028361024d3660046117ee565b600360208190526000918252604090912080546001820154600283015492909301546001600160a01b0391821693909116919084565b604080516001600160a01b039586168152949093166020850152918301526060820152608001610143565b3480156102ba57600080fd5b50610139610c3a565b3480156102cf57600080fd5b506102e36102de3660046117ee565b610c4b565b6040519015158152602001610143565b3480156102ff57600080fd5b5061018261030e3660046117ee565b6004602052600090815260409020546001600160a01b031681565b610121610337366004611809565b610c5e565b61012161034a366004611855565b610ced565b34801561035b57600080fd5b5061013960075481565b34801561037157600080fd5b506102e36103803660046117ee565b61119f565b34801561039157600080fd5b506103dc6103a0366004611809565b600860205260009081526040902080546001820154600283015460038401546004909401546001600160a01b0390931693919290919060ff1685565b6040516101439594939291906118a7565b3480156103f957600080fd5b5061018261100081565b34801561040f57600080fd5b506104186111e5565b60405161014391906118fa565b8080341461044e5760405162461bcd60e51b815260040161044590611947565b60405180910390fd5b61045733610c4b565b1561049d5760405162461bcd60e51b81526020600482015260166024820152757374616b65722063616e6e6f742064656c656761746560501b6044820152606401610445565b6104a63361119f565b156104fe5760405162461bcd60e51b815260206004820152602260248201527f6f70657261746f72287265776172646565292063616e6e6f742064656c656761604482015261746560f01b6064820152608401610445565b61050a838360016111f1565b3360009081526006602090815260408083206001600160a01b03871684529091528120805484929061053d908490611994565b90915550506040518281526001600160a01b0384169033907fe5541a6b6103d4fa7e021ed54fad39c66f27a76bd13d374cf6240ae6bd0bb72b9060200160405180910390a3505050565b336000908152600460205260409020546001600160a01b0316806105e35760405162461bcd60e51b81526020600482015260136024820152723ab73932b3b4b9ba32b932b21039ba30b5b2b960691b6044820152606401610445565b600082116106245760405162461bcd60e51b815260206004820152600e60248201526d616d6f756e74206973207a65726f60901b6044820152606401610445565b6001600160a01b0381166000908152600360208190526040822090810154600282015491929161065491906119ac565b90508381101561069d5760405162461bcd60e51b8152602060048201526014602482015273696e73756666696369656e742062616c616e636560601b6044820152606401610445565b6110006001600160a01b031663ba631d3f6040518163ffffffff1660e01b8152600401602060405180830381865afa1580156106dd573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061070191906119c3565b61070b85836119ac565b1015610819578381146107735760405162461bcd60e51b815260206004820152602a60248201527f616d6f756e74206d75737420657175616c2062616c616e636520746f2072656d60448201526937bb329039ba30b5b2b960b11b6064820152608401610445565b61077e600184611361565b5033600090815260046020908152604080832080546001600160a01b03199081169091556001868101546001600160a01b0390811686526005855283862080548416905588168086526003948590528386208054841681559182018054909316909255600281018590559092018390555190917fb97b17a738bc54faf4f156f2072573525abf5a6f9fe4e3f78a2f159d7a85180591a2610833565b8382600201600082825461082d91906119ac565b90915550505b8360008082825461084491906119ac565b925050819055506108b8846110006001600160a01b031663fde7f3716040518163ffffffff1660e01b8152600401602060405180830381865afa15801561088f573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906108b391906119c3565b61137d565b826001600160a01b03167f0f5bb82176feb1b5e747e28471aa92156a04d9f3ab9f45f28e2d704232b93f75856040516108f391815260200190565b60405180910390a250505050565b60008181526008602052604090206001600482015460ff16600281111561092a5761092a611891565b1461096c5760405162461bcd60e51b81526020600482015260126024820152711a5b9d985b1a590818dc9959195b9d1a585b60721b6044820152606401610445565b80546001600160a01b031633146109c55760405162461bcd60e51b815260206004820152601b60248201527f6d73672e73656e646572206973206e6f742072657175657374657200000000006044820152606401610445565b8060030154421015610a195760405162461bcd60e51b815260206004820152601860248201527f6e6f74207965742074696d6520746f20776974686472617700000000000000006044820152606401610445565b60018101548154610a35916001600160a01b039091169061149e565b60048101805460ff19166002179055600181015460408051338152602081019290925283917fcf7d23a3cbe4e8b36ff82fd1b05b1b17373dc7804b4ebbd6e2356716ef202372910160405180910390a25050565b3360009081526006602090815260408083206001600160a01b0386168452909152902054811115610af35760405162461bcd60e51b8152602060048201526014602482015273696e73756666696369656e742062616c616e636560601b6044820152606401610445565b610afc82610c4b565b15610b9b576001600160a01b03821660009081526003602081905260408220908101805491928492610b2f9084906119ac565b9250508190555081816002016000828254610b4a91906119ac565b92505081905550610b95826110006001600160a01b031663840c17716040518163ffffffff1660e01b8152600401602060405180830381865afa15801561088f573d6000803e3d6000fd5b50610ba5565b610ba5338261149e565b3360009081526006602090815260408083206001600160a01b038616845290915281208054839290610bd89084906119ac565b9250508190555080600080828254610bf091906119ac565b90915550506040518181526001600160a01b0383169033907f4d10bd049775c77bd7f255195afba5088028ecb3c7c277d393ccff7934f2f92c906020015b60405180910390a35050565b6000610c4660016115bc565b905090565b6000610c586001836115c6565b92915050565b80803414610c7e5760405162461bcd60e51b815260040161044590611947565b336000908152600460205260408120546001600160a01b031690610ca590829085906111f1565b806001600160a01b03167f9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d84604051610ce091815260200190565b60405180910390a2505050565b82803414610d0d5760405162461bcd60e51b815260040161044590611947565b6110006001600160a01b031663ba631d3f6040518163ffffffff1660e01b8152600401602060405180830381865afa158015610d4d573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610d7191906119c3565b8410158015610de357506110006001600160a01b031663129060ab6040518163ffffffff1660e01b8152600401602060405180830381865afa158015610dbb573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610ddf91906119c3565b8411155b610e1f5760405162461bcd60e51b815260206004820152600d60248201526c6f7574206f6620626f756e647360981b6044820152606401610445565b336001600160a01b03841614801590610e415750336001600160a01b03831614155b610e9b5760405162461bcd60e51b815260206004820152602560248201527f6f70657261746f722063616e6e6f74206265207374616b6572206f7220726577604482015264617264656560d81b6064820152608401610445565b6001600160a01b03831615801590610ebb57506001600160a01b03821615155b610ef65760405162461bcd60e51b815260206004820152600c60248201526b7a65726f206164647265737360a01b6044820152606401610445565b816001600160a01b0316836001600160a01b031603610f575760405162461bcd60e51b815260206004820152601960248201527f7374616b65722063616e6e6f74206265207265776172646565000000000000006044820152606401610445565b610f603361119f565b15610fad5760405162461bcd60e51b815260206004820152601e60248201527f6f70657261746f7220697320616c7265616479207265676973746572656400006044820152606401610445565b610fb68361119f565b156110035760405162461bcd60e51b815260206004820152601c60248201527f7374616b657220697320616c72656164792072656769737465726564000000006044820152606401610445565b61100c8261119f565b156110595760405162461bcd60e51b815260206004820152601e60248201527f726577617264656520697320616c7265616479207265676973746572656400006044820152606401610445565b6110646001846115e8565b6110a05760405162461bcd60e51b815260206004820152600d60248201526c7374616b65722065786973747360981b6044820152606401610445565b60408051608081018252338082526001600160a01b0385811660208085018281528587018b81526000606088018181528c871680835260038087528b84209a518b54908a166001600160a01b0319918216178c55955160018c01805491909a1690871617909855925160028a015551979095019690965593835260048152858320805485168617905590825260059052928320805490911690911790558054859190819061114f908490611994565b9091555050604080513381526001600160a01b038481166020830152918101869052908416907f8705598e75e33b571bb9c5bcdbf2506dc4e7b88c0262532269bfee3fd15b3d86906060016108f3565b6001600160a01b03818116600090815260046020526040812054909116151580610c585750506001600160a01b0390811660009081526005602052604090205416151590565b6060610c4660016115fd565b6111fa83610c4b565b61123c5760405162461bcd60e51b81526020600482015260136024820152723ab73932b3b4b9ba32b932b21039ba30b5b2b960691b6044820152606401610445565b6001600160a01b038316600090815260036020908152604091829020825163129060ab60e01b8152925190926110009263129060ab926004808401938290030181865afa158015611291573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906112b591906119c3565b8382600201546112c59190611994565b111561130a5760405162461bcd60e51b8152602060048201526014602482015273657863656564656420746865206d6178696d756d60601b6044820152606401610445565b8260008082825461131b9190611994565b92505081905550828160020160008282546113369190611994565b9091555050811561135b57828160030160008282546113559190611994565b90915550505b50505050565b6000611376836001600160a01b03841661160a565b9392505050565b6040518060a00160405280336001600160a01b0316815260200183815260200142815260200182426113af9190611994565b81526020016001815250600860006007600081546113cc906119dc565b919050819055815260200190815260200160002060008201518160000160006101000a8154816001600160a01b0302191690836001600160a01b0316021790555060208201518160010155604082015181600201556060820151816003015560808201518160040160006101000a81548160ff0219169083600281111561145557611455611891565b021790555050600754604080518581524260208201529081018490523392507f4846f03be8ef87cb6e611b3a3b878a0aadd7c010f3f25707aa472b41de9dc75d90606001610c2e565b804710156114ee5760405162461bcd60e51b815260206004820152601d60248201527f416464726573733a20696e73756666696369656e742062616c616e63650000006044820152606401610445565b6000826001600160a01b03168260405160006040518083038185875af1925050503d806000811461153b576040519150601f19603f3d011682016040523d82523d6000602084013e611540565b606091505b50509050806115b75760405162461bcd60e51b815260206004820152603a60248201527f416464726573733a20756e61626c6520746f2073656e642076616c75652c207260448201527f6563697069656e74206d617920686176652072657665727465640000000000006064820152608401610445565b505050565b6000610c58825490565b6001600160a01b03811660009081526001830160205260408120541515611376565b6000611376836001600160a01b0384166116fd565b606060006113768361174c565b600081815260018301602052604081205480156116f357600061162e6001836119ac565b8554909150600090611642906001906119ac565b90508181146116a7576000866000018281548110611662576116626119f5565b9060005260206000200154905080876000018481548110611685576116856119f5565b6000918252602080832090910192909255918252600188019052604090208390555b85548690806116b8576116b8611a0b565b600190038181906000526020600020016000905590558560010160008681526020019081526020016000206000905560019350505050610c58565b6000915050610c58565b600081815260018301602052604081205461174457508154600181810184556000848152602080822090930184905584548482528286019093526040902091909155610c58565b506000610c58565b60608160000180548060200260200160405190810160405280929190818152602001828054801561179c57602002820191906000526020600020905b815481526020019060010190808311611788575b50505050509050919050565b80356001600160a01b03811681146117bf57600080fd5b919050565b600080604083850312156117d757600080fd5b6117e0836117a8565b946020939093013593505050565b60006020828403121561180057600080fd5b611376826117a8565b60006020828403121561181b57600080fd5b5035919050565b6000806040838503121561183557600080fd5b61183e836117a8565b915061184c602084016117a8565b90509250929050565b60008060006060848603121561186a57600080fd5b8335925061187a602085016117a8565b9150611888604085016117a8565b90509250925092565b634e487b7160e01b600052602160045260246000fd5b6001600160a01b038616815260208101859052604081018490526060810183905260a08101600383106118ea57634e487b7160e01b600052602160045260246000fd5b8260808301529695505050505050565b6020808252825182820181905260009190848201906040850190845b8181101561193b5783516001600160a01b031683529284019291840191600101611916565b50909695505050505050565b6020808252601d908201527f616d6f756e7420616e64206d73672e76616c7565206d69736d61746368000000604082015260600190565b634e487b7160e01b600052601160045260246000fd5b600082198211156119a7576119a761197e565b500190565b6000828210156119be576119be61197e565b500390565b6000602082840312156119d557600080fd5b5051919050565b6000600182016119ee576119ee61197e565b5060010190565b634e487b7160e01b600052603260045260246000fd5b634e487b7160e01b600052603160045260246000fdfea2646970667358221220edb4faf434f7b6ad1d1afa89ff38d064a75b36b637e7078a6ab53bf5e904113464736f6c634300080e0033",
}

// GovStakingABI is the input ABI used to generate the binding from.
// Deprecated: Use GovStakingMetaData.ABI instead.
var GovStakingABI = GovStakingMetaData.ABI

// Deprecated: Use GovStakingMetaData.Sigs instead.
// GovStakingFuncSigs maps the 4-byte function signature to its string representation.
var GovStakingFuncSigs = GovStakingMetaData.Sigs

// GovStakingBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use GovStakingMetaData.Bin instead.
var GovStakingBin = GovStakingMetaData.Bin

// DeployGovStaking deploys a new Ethereum contract, binding an instance of GovStaking to it.
func DeployGovStaking(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *GovStaking, error) {
	parsed, err := GovStakingMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(GovStakingBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &GovStaking{GovStakingCaller: GovStakingCaller{contract: contract}, GovStakingTransactor: GovStakingTransactor{contract: contract}, GovStakingFilterer: GovStakingFilterer{contract: contract}}, nil
}

// GovStaking is an auto generated Go binding around an Ethereum contract.
type GovStaking struct {
	GovStakingCaller     // Read-only binding to the contract
	GovStakingTransactor // Write-only binding to the contract
	GovStakingFilterer   // Log filterer for contract events
}

// GovStakingCaller is an auto generated read-only Go binding around an Ethereum contract.
type GovStakingCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GovStakingTransactor is an auto generated write-only Go binding around an Ethereum contract.
type GovStakingTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GovStakingFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type GovStakingFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GovStakingSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type GovStakingSession struct {
	Contract     *GovStaking       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// GovStakingCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type GovStakingCallerSession struct {
	Contract *GovStakingCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// GovStakingTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type GovStakingTransactorSession struct {
	Contract     *GovStakingTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// GovStakingRaw is an auto generated low-level Go binding around an Ethereum contract.
type GovStakingRaw struct {
	Contract *GovStaking // Generic contract binding to access the raw methods on
}

// GovStakingCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type GovStakingCallerRaw struct {
	Contract *GovStakingCaller // Generic read-only contract binding to access the raw methods on
}

// GovStakingTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type GovStakingTransactorRaw struct {
	Contract *GovStakingTransactor // Generic write-only contract binding to access the raw methods on
}

// NewGovStaking creates a new instance of GovStaking, bound to a specific deployed contract.
func NewGovStaking(address common.Address, backend bind.ContractBackend) (*GovStaking, error) {
	contract, err := bindGovStaking(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &GovStaking{GovStakingCaller: GovStakingCaller{contract: contract}, GovStakingTransactor: GovStakingTransactor{contract: contract}, GovStakingFilterer: GovStakingFilterer{contract: contract}}, nil
}

// NewGovStakingCaller creates a new read-only instance of GovStaking, bound to a specific deployed contract.
func NewGovStakingCaller(address common.Address, caller bind.ContractCaller) (*GovStakingCaller, error) {
	contract, err := bindGovStaking(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &GovStakingCaller{contract: contract}, nil
}

// NewGovStakingTransactor creates a new write-only instance of GovStaking, bound to a specific deployed contract.
func NewGovStakingTransactor(address common.Address, transactor bind.ContractTransactor) (*GovStakingTransactor, error) {
	contract, err := bindGovStaking(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &GovStakingTransactor{contract: contract}, nil
}

// NewGovStakingFilterer creates a new log filterer instance of GovStaking, bound to a specific deployed contract.
func NewGovStakingFilterer(address common.Address, filterer bind.ContractFilterer) (*GovStakingFilterer, error) {
	contract, err := bindGovStaking(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &GovStakingFilterer{contract: contract}, nil
}

// bindGovStaking binds a generic wrapper to an already deployed contract.
func bindGovStaking(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := GovStakingMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GovStaking *GovStakingRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _GovStaking.Contract.GovStakingCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GovStaking *GovStakingRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GovStaking.Contract.GovStakingTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GovStaking *GovStakingRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GovStaking.Contract.GovStakingTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GovStaking *GovStakingCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _GovStaking.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GovStaking *GovStakingTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GovStaking.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GovStaking *GovStakingTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GovStaking.Contract.contract.Transact(opts, method, params...)
}

// GOVCONST is a free data retrieval call binding the contract method 0xe8aaca24.
//
// Solidity: function GOV_CONST() view returns(address)
func (_GovStaking *GovStakingCaller) GOVCONST(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _GovStaking.contract.Call(opts, &out, "GOV_CONST")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GOVCONST is a free data retrieval call binding the contract method 0xe8aaca24.
//
// Solidity: function GOV_CONST() view returns(address)
func (_GovStaking *GovStakingSession) GOVCONST() (common.Address, error) {
	return _GovStaking.Contract.GOVCONST(&_GovStaking.CallOpts)
}

// GOVCONST is a free data retrieval call binding the contract method 0xe8aaca24.
//
// Solidity: function GOV_CONST() view returns(address)
func (_GovStaking *GovStakingCallerSession) GOVCONST() (common.Address, error) {
	return _GovStaking.Contract.GOVCONST(&_GovStaking.CallOpts)
}

// CredentialCount is a free data retrieval call binding the contract method 0xcd0e35b9.
//
// Solidity: function credentialCount() view returns(uint256)
func (_GovStaking *GovStakingCaller) CredentialCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GovStaking.contract.Call(opts, &out, "credentialCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CredentialCount is a free data retrieval call binding the contract method 0xcd0e35b9.
//
// Solidity: function credentialCount() view returns(uint256)
func (_GovStaking *GovStakingSession) CredentialCount() (*big.Int, error) {
	return _GovStaking.Contract.CredentialCount(&_GovStaking.CallOpts)
}

// CredentialCount is a free data retrieval call binding the contract method 0xcd0e35b9.
//
// Solidity: function credentialCount() view returns(uint256)
func (_GovStaking *GovStakingCallerSession) CredentialCount() (*big.Int, error) {
	return _GovStaking.Contract.CredentialCount(&_GovStaking.CallOpts)
}

// Credentials is a free data retrieval call binding the contract method 0xe0574e3f.
//
// Solidity: function credentials(uint256 ) view returns(address requester, uint256 amount, uint256 requestTime, uint256 withdrawableTime, uint8 status)
func (_GovStaking *GovStakingCaller) Credentials(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Requester        common.Address
	Amount           *big.Int
	RequestTime      *big.Int
	WithdrawableTime *big.Int
	Status           uint8
}, error) {
	var out []interface{}
	err := _GovStaking.contract.Call(opts, &out, "credentials", arg0)

	outstruct := new(struct {
		Requester        common.Address
		Amount           *big.Int
		RequestTime      *big.Int
		WithdrawableTime *big.Int
		Status           uint8
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Requester = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.Amount = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.RequestTime = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.WithdrawableTime = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.Status = *abi.ConvertType(out[4], new(uint8)).(*uint8)

	return *outstruct, err

}

// Credentials is a free data retrieval call binding the contract method 0xe0574e3f.
//
// Solidity: function credentials(uint256 ) view returns(address requester, uint256 amount, uint256 requestTime, uint256 withdrawableTime, uint8 status)
func (_GovStaking *GovStakingSession) Credentials(arg0 *big.Int) (struct {
	Requester        common.Address
	Amount           *big.Int
	RequestTime      *big.Int
	WithdrawableTime *big.Int
	Status           uint8
}, error) {
	return _GovStaking.Contract.Credentials(&_GovStaking.CallOpts, arg0)
}

// Credentials is a free data retrieval call binding the contract method 0xe0574e3f.
//
// Solidity: function credentials(uint256 ) view returns(address requester, uint256 amount, uint256 requestTime, uint256 withdrawableTime, uint8 status)
func (_GovStaking *GovStakingCallerSession) Credentials(arg0 *big.Int) (struct {
	Requester        common.Address
	Amount           *big.Int
	RequestTime      *big.Int
	WithdrawableTime *big.Int
	Status           uint8
}, error) {
	return _GovStaking.Contract.Credentials(&_GovStaking.CallOpts, arg0)
}

// DelegateTo is a free data retrieval call binding the contract method 0x438bb7e5.
//
// Solidity: function delegateTo(address , address ) view returns(uint256)
func (_GovStaking *GovStakingCaller) DelegateTo(opts *bind.CallOpts, arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _GovStaking.contract.Call(opts, &out, "delegateTo", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// DelegateTo is a free data retrieval call binding the contract method 0x438bb7e5.
//
// Solidity: function delegateTo(address , address ) view returns(uint256)
func (_GovStaking *GovStakingSession) DelegateTo(arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	return _GovStaking.Contract.DelegateTo(&_GovStaking.CallOpts, arg0, arg1)
}

// DelegateTo is a free data retrieval call binding the contract method 0x438bb7e5.
//
// Solidity: function delegateTo(address , address ) view returns(uint256)
func (_GovStaking *GovStakingCallerSession) DelegateTo(arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	return _GovStaking.Contract.DelegateTo(&_GovStaking.CallOpts, arg0, arg1)
}

// IsOperatorOrRewardee is a free data retrieval call binding the contract method 0xcdb72f06.
//
// Solidity: function isOperatorOrRewardee(address _addr) view returns(bool)
func (_GovStaking *GovStakingCaller) IsOperatorOrRewardee(opts *bind.CallOpts, _addr common.Address) (bool, error) {
	var out []interface{}
	err := _GovStaking.contract.Call(opts, &out, "isOperatorOrRewardee", _addr)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsOperatorOrRewardee is a free data retrieval call binding the contract method 0xcdb72f06.
//
// Solidity: function isOperatorOrRewardee(address _addr) view returns(bool)
func (_GovStaking *GovStakingSession) IsOperatorOrRewardee(_addr common.Address) (bool, error) {
	return _GovStaking.Contract.IsOperatorOrRewardee(&_GovStaking.CallOpts, _addr)
}

// IsOperatorOrRewardee is a free data retrieval call binding the contract method 0xcdb72f06.
//
// Solidity: function isOperatorOrRewardee(address _addr) view returns(bool)
func (_GovStaking *GovStakingCallerSession) IsOperatorOrRewardee(_addr common.Address) (bool, error) {
	return _GovStaking.Contract.IsOperatorOrRewardee(&_GovStaking.CallOpts, _addr)
}

// IsStaker is a free data retrieval call binding the contract method 0x6f1e8533.
//
// Solidity: function isStaker(address _staker) view returns(bool)
func (_GovStaking *GovStakingCaller) IsStaker(opts *bind.CallOpts, _staker common.Address) (bool, error) {
	var out []interface{}
	err := _GovStaking.contract.Call(opts, &out, "isStaker", _staker)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsStaker is a free data retrieval call binding the contract method 0x6f1e8533.
//
// Solidity: function isStaker(address _staker) view returns(bool)
func (_GovStaking *GovStakingSession) IsStaker(_staker common.Address) (bool, error) {
	return _GovStaking.Contract.IsStaker(&_GovStaking.CallOpts, _staker)
}

// IsStaker is a free data retrieval call binding the contract method 0x6f1e8533.
//
// Solidity: function isStaker(address _staker) view returns(bool)
func (_GovStaking *GovStakingCallerSession) IsStaker(_staker common.Address) (bool, error) {
	return _GovStaking.Contract.IsStaker(&_GovStaking.CallOpts, _staker)
}

// StakerByOperator is a free data retrieval call binding the contract method 0x9dbccf6a.
//
// Solidity: function stakerByOperator(address ) view returns(address)
func (_GovStaking *GovStakingCaller) StakerByOperator(opts *bind.CallOpts, arg0 common.Address) (common.Address, error) {
	var out []interface{}
	err := _GovStaking.contract.Call(opts, &out, "stakerByOperator", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// StakerByOperator is a free data retrieval call binding the contract method 0x9dbccf6a.
//
// Solidity: function stakerByOperator(address ) view returns(address)
func (_GovStaking *GovStakingSession) StakerByOperator(arg0 common.Address) (common.Address, error) {
	return _GovStaking.Contract.StakerByOperator(&_GovStaking.CallOpts, arg0)
}

// StakerByOperator is a free data retrieval call binding the contract method 0x9dbccf6a.
//
// Solidity: function stakerByOperator(address ) view returns(address)
func (_GovStaking *GovStakingCallerSession) StakerByOperator(arg0 common.Address) (common.Address, error) {
	return _GovStaking.Contract.StakerByOperator(&_GovStaking.CallOpts, arg0)
}

// StakerByRewardee is a free data retrieval call binding the contract method 0x264ebc38.
//
// Solidity: function stakerByRewardee(address ) view returns(address)
func (_GovStaking *GovStakingCaller) StakerByRewardee(opts *bind.CallOpts, arg0 common.Address) (common.Address, error) {
	var out []interface{}
	err := _GovStaking.contract.Call(opts, &out, "stakerByRewardee", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// StakerByRewardee is a free data retrieval call binding the contract method 0x264ebc38.
//
// Solidity: function stakerByRewardee(address ) view returns(address)
func (_GovStaking *GovStakingSession) StakerByRewardee(arg0 common.Address) (common.Address, error) {
	return _GovStaking.Contract.StakerByRewardee(&_GovStaking.CallOpts, arg0)
}

// StakerByRewardee is a free data retrieval call binding the contract method 0x264ebc38.
//
// Solidity: function stakerByRewardee(address ) view returns(address)
func (_GovStaking *GovStakingCallerSession) StakerByRewardee(arg0 common.Address) (common.Address, error) {
	return _GovStaking.Contract.StakerByRewardee(&_GovStaking.CallOpts, arg0)
}

// StakerInfo is a free data retrieval call binding the contract method 0x4e745f1f.
//
// Solidity: function stakerInfo(address ) view returns(address operator, address rewardee, uint256 staking, uint256 delegated)
func (_GovStaking *GovStakingCaller) StakerInfo(opts *bind.CallOpts, arg0 common.Address) (struct {
	Operator  common.Address
	Rewardee  common.Address
	Staking   *big.Int
	Delegated *big.Int
}, error) {
	var out []interface{}
	err := _GovStaking.contract.Call(opts, &out, "stakerInfo", arg0)

	outstruct := new(struct {
		Operator  common.Address
		Rewardee  common.Address
		Staking   *big.Int
		Delegated *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Operator = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.Rewardee = *abi.ConvertType(out[1], new(common.Address)).(*common.Address)
	outstruct.Staking = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.Delegated = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// StakerInfo is a free data retrieval call binding the contract method 0x4e745f1f.
//
// Solidity: function stakerInfo(address ) view returns(address operator, address rewardee, uint256 staking, uint256 delegated)
func (_GovStaking *GovStakingSession) StakerInfo(arg0 common.Address) (struct {
	Operator  common.Address
	Rewardee  common.Address
	Staking   *big.Int
	Delegated *big.Int
}, error) {
	return _GovStaking.Contract.StakerInfo(&_GovStaking.CallOpts, arg0)
}

// StakerInfo is a free data retrieval call binding the contract method 0x4e745f1f.
//
// Solidity: function stakerInfo(address ) view returns(address operator, address rewardee, uint256 staking, uint256 delegated)
func (_GovStaking *GovStakingCallerSession) StakerInfo(arg0 common.Address) (struct {
	Operator  common.Address
	Rewardee  common.Address
	Staking   *big.Int
	Delegated *big.Int
}, error) {
	return _GovStaking.Contract.StakerInfo(&_GovStaking.CallOpts, arg0)
}

// StakerLength is a free data retrieval call binding the contract method 0x5748f6f3.
//
// Solidity: function stakerLength() view returns(uint256)
func (_GovStaking *GovStakingCaller) StakerLength(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GovStaking.contract.Call(opts, &out, "stakerLength")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StakerLength is a free data retrieval call binding the contract method 0x5748f6f3.
//
// Solidity: function stakerLength() view returns(uint256)
func (_GovStaking *GovStakingSession) StakerLength() (*big.Int, error) {
	return _GovStaking.Contract.StakerLength(&_GovStaking.CallOpts)
}

// StakerLength is a free data retrieval call binding the contract method 0x5748f6f3.
//
// Solidity: function stakerLength() view returns(uint256)
func (_GovStaking *GovStakingCallerSession) StakerLength() (*big.Int, error) {
	return _GovStaking.Contract.StakerLength(&_GovStaking.CallOpts)
}

// Stakers is a free data retrieval call binding the contract method 0xfed1252a.
//
// Solidity: function stakers() view returns(address[])
func (_GovStaking *GovStakingCaller) Stakers(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _GovStaking.contract.Call(opts, &out, "stakers")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// Stakers is a free data retrieval call binding the contract method 0xfed1252a.
//
// Solidity: function stakers() view returns(address[])
func (_GovStaking *GovStakingSession) Stakers() ([]common.Address, error) {
	return _GovStaking.Contract.Stakers(&_GovStaking.CallOpts)
}

// Stakers is a free data retrieval call binding the contract method 0xfed1252a.
//
// Solidity: function stakers() view returns(address[])
func (_GovStaking *GovStakingCallerSession) Stakers() ([]common.Address, error) {
	return _GovStaking.Contract.Stakers(&_GovStaking.CallOpts)
}

// TotalStaking is a free data retrieval call binding the contract method 0x165defa4.
//
// Solidity: function totalStaking() view returns(uint256)
func (_GovStaking *GovStakingCaller) TotalStaking(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GovStaking.contract.Call(opts, &out, "totalStaking")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalStaking is a free data retrieval call binding the contract method 0x165defa4.
//
// Solidity: function totalStaking() view returns(uint256)
func (_GovStaking *GovStakingSession) TotalStaking() (*big.Int, error) {
	return _GovStaking.Contract.TotalStaking(&_GovStaking.CallOpts)
}

// TotalStaking is a free data retrieval call binding the contract method 0x165defa4.
//
// Solidity: function totalStaking() view returns(uint256)
func (_GovStaking *GovStakingCallerSession) TotalStaking() (*big.Int, error) {
	return _GovStaking.Contract.TotalStaking(&_GovStaking.CallOpts)
}

// Delegate is a paid mutator transaction binding the contract method 0x026e402b.
//
// Solidity: function delegate(address _staker, uint256 _amount) payable returns()
func (_GovStaking *GovStakingTransactor) Delegate(opts *bind.TransactOpts, _staker common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _GovStaking.contract.Transact(opts, "delegate", _staker, _amount)
}

// Delegate is a paid mutator transaction binding the contract method 0x026e402b.
//
// Solidity: function delegate(address _staker, uint256 _amount) payable returns()
func (_GovStaking *GovStakingSession) Delegate(_staker common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _GovStaking.Contract.Delegate(&_GovStaking.TransactOpts, _staker, _amount)
}

// Delegate is a paid mutator transaction binding the contract method 0x026e402b.
//
// Solidity: function delegate(address _staker, uint256 _amount) payable returns()
func (_GovStaking *GovStakingTransactorSession) Delegate(_staker common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _GovStaking.Contract.Delegate(&_GovStaking.TransactOpts, _staker, _amount)
}

// RegisterStaker is a paid mutator transaction binding the contract method 0xaf716c40.
//
// Solidity: function registerStaker(uint256 _amount, address _staker, address _rewardee) payable returns()
func (_GovStaking *GovStakingTransactor) RegisterStaker(opts *bind.TransactOpts, _amount *big.Int, _staker common.Address, _rewardee common.Address) (*types.Transaction, error) {
	return _GovStaking.contract.Transact(opts, "registerStaker", _amount, _staker, _rewardee)
}

// RegisterStaker is a paid mutator transaction binding the contract method 0xaf716c40.
//
// Solidity: function registerStaker(uint256 _amount, address _staker, address _rewardee) payable returns()
func (_GovStaking *GovStakingSession) RegisterStaker(_amount *big.Int, _staker common.Address, _rewardee common.Address) (*types.Transaction, error) {
	return _GovStaking.Contract.RegisterStaker(&_GovStaking.TransactOpts, _amount, _staker, _rewardee)
}

// RegisterStaker is a paid mutator transaction binding the contract method 0xaf716c40.
//
// Solidity: function registerStaker(uint256 _amount, address _staker, address _rewardee) payable returns()
func (_GovStaking *GovStakingTransactorSession) RegisterStaker(_amount *big.Int, _staker common.Address, _rewardee common.Address) (*types.Transaction, error) {
	return _GovStaking.Contract.RegisterStaker(&_GovStaking.TransactOpts, _amount, _staker, _rewardee)
}

// Stake is a paid mutator transaction binding the contract method 0xa694fc3a.
//
// Solidity: function stake(uint256 _amount) payable returns()
func (_GovStaking *GovStakingTransactor) Stake(opts *bind.TransactOpts, _amount *big.Int) (*types.Transaction, error) {
	return _GovStaking.contract.Transact(opts, "stake", _amount)
}

// Stake is a paid mutator transaction binding the contract method 0xa694fc3a.
//
// Solidity: function stake(uint256 _amount) payable returns()
func (_GovStaking *GovStakingSession) Stake(_amount *big.Int) (*types.Transaction, error) {
	return _GovStaking.Contract.Stake(&_GovStaking.TransactOpts, _amount)
}

// Stake is a paid mutator transaction binding the contract method 0xa694fc3a.
//
// Solidity: function stake(uint256 _amount) payable returns()
func (_GovStaking *GovStakingTransactorSession) Stake(_amount *big.Int) (*types.Transaction, error) {
	return _GovStaking.Contract.Stake(&_GovStaking.TransactOpts, _amount)
}

// Undelegate is a paid mutator transaction binding the contract method 0x4d99dd16.
//
// Solidity: function undelegate(address _staker, uint256 _amount) returns()
func (_GovStaking *GovStakingTransactor) Undelegate(opts *bind.TransactOpts, _staker common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _GovStaking.contract.Transact(opts, "undelegate", _staker, _amount)
}

// Undelegate is a paid mutator transaction binding the contract method 0x4d99dd16.
//
// Solidity: function undelegate(address _staker, uint256 _amount) returns()
func (_GovStaking *GovStakingSession) Undelegate(_staker common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _GovStaking.Contract.Undelegate(&_GovStaking.TransactOpts, _staker, _amount)
}

// Undelegate is a paid mutator transaction binding the contract method 0x4d99dd16.
//
// Solidity: function undelegate(address _staker, uint256 _amount) returns()
func (_GovStaking *GovStakingTransactorSession) Undelegate(_staker common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _GovStaking.Contract.Undelegate(&_GovStaking.TransactOpts, _staker, _amount)
}

// Unstake is a paid mutator transaction binding the contract method 0x2e17de78.
//
// Solidity: function unstake(uint256 _amount) returns()
func (_GovStaking *GovStakingTransactor) Unstake(opts *bind.TransactOpts, _amount *big.Int) (*types.Transaction, error) {
	return _GovStaking.contract.Transact(opts, "unstake", _amount)
}

// Unstake is a paid mutator transaction binding the contract method 0x2e17de78.
//
// Solidity: function unstake(uint256 _amount) returns()
func (_GovStaking *GovStakingSession) Unstake(_amount *big.Int) (*types.Transaction, error) {
	return _GovStaking.Contract.Unstake(&_GovStaking.TransactOpts, _amount)
}

// Unstake is a paid mutator transaction binding the contract method 0x2e17de78.
//
// Solidity: function unstake(uint256 _amount) returns()
func (_GovStaking *GovStakingTransactorSession) Unstake(_amount *big.Int) (*types.Transaction, error) {
	return _GovStaking.Contract.Unstake(&_GovStaking.TransactOpts, _amount)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 _cid) returns()
func (_GovStaking *GovStakingTransactor) Withdraw(opts *bind.TransactOpts, _cid *big.Int) (*types.Transaction, error) {
	return _GovStaking.contract.Transact(opts, "withdraw", _cid)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 _cid) returns()
func (_GovStaking *GovStakingSession) Withdraw(_cid *big.Int) (*types.Transaction, error) {
	return _GovStaking.Contract.Withdraw(&_GovStaking.TransactOpts, _cid)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 _cid) returns()
func (_GovStaking *GovStakingTransactorSession) Withdraw(_cid *big.Int) (*types.Transaction, error) {
	return _GovStaking.Contract.Withdraw(&_GovStaking.TransactOpts, _cid)
}

// GovStakingDelegatedIterator is returned from FilterDelegated and is used to iterate over the raw logs and unpacked data for Delegated events raised by the GovStaking contract.
type GovStakingDelegatedIterator struct {
	Event *GovStakingDelegated // Event containing the contract specifics and raw log

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
func (it *GovStakingDelegatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GovStakingDelegated)
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
		it.Event = new(GovStakingDelegated)
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
func (it *GovStakingDelegatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GovStakingDelegatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GovStakingDelegated represents a Delegated event raised by the GovStaking contract.
type GovStakingDelegated struct {
	Delegator common.Address
	Staker    common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDelegated is a free log retrieval operation binding the contract event 0xe5541a6b6103d4fa7e021ed54fad39c66f27a76bd13d374cf6240ae6bd0bb72b.
//
// Solidity: event Delegated(address indexed delegator, address indexed staker, uint256 amount)
func (_GovStaking *GovStakingFilterer) FilterDelegated(opts *bind.FilterOpts, delegator []common.Address, staker []common.Address) (*GovStakingDelegatedIterator, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var stakerRule []interface{}
	for _, stakerItem := range staker {
		stakerRule = append(stakerRule, stakerItem)
	}

	logs, sub, err := _GovStaking.contract.FilterLogs(opts, "Delegated", delegatorRule, stakerRule)
	if err != nil {
		return nil, err
	}
	return &GovStakingDelegatedIterator{contract: _GovStaking.contract, event: "Delegated", logs: logs, sub: sub}, nil
}

// WatchDelegated is a free log subscription operation binding the contract event 0xe5541a6b6103d4fa7e021ed54fad39c66f27a76bd13d374cf6240ae6bd0bb72b.
//
// Solidity: event Delegated(address indexed delegator, address indexed staker, uint256 amount)
func (_GovStaking *GovStakingFilterer) WatchDelegated(opts *bind.WatchOpts, sink chan<- *GovStakingDelegated, delegator []common.Address, staker []common.Address) (event.Subscription, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var stakerRule []interface{}
	for _, stakerItem := range staker {
		stakerRule = append(stakerRule, stakerItem)
	}

	logs, sub, err := _GovStaking.contract.WatchLogs(opts, "Delegated", delegatorRule, stakerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GovStakingDelegated)
				if err := _GovStaking.contract.UnpackLog(event, "Delegated", log); err != nil {
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

// ParseDelegated is a log parse operation binding the contract event 0xe5541a6b6103d4fa7e021ed54fad39c66f27a76bd13d374cf6240ae6bd0bb72b.
//
// Solidity: event Delegated(address indexed delegator, address indexed staker, uint256 amount)
func (_GovStaking *GovStakingFilterer) ParseDelegated(log types.Log) (*GovStakingDelegated, error) {
	event := new(GovStakingDelegated)
	if err := _GovStaking.contract.UnpackLog(event, "Delegated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GovStakingNewCredentialIterator is returned from FilterNewCredential and is used to iterate over the raw logs and unpacked data for NewCredential events raised by the GovStaking contract.
type GovStakingNewCredentialIterator struct {
	Event *GovStakingNewCredential // Event containing the contract specifics and raw log

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
func (it *GovStakingNewCredentialIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GovStakingNewCredential)
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
		it.Event = new(GovStakingNewCredential)
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
func (it *GovStakingNewCredentialIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GovStakingNewCredentialIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GovStakingNewCredential represents a NewCredential event raised by the GovStaking contract.
type GovStakingNewCredential struct {
	CredentialID *big.Int
	Requester    common.Address
	Amount       *big.Int
	Time         *big.Int
	Unbonding    *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterNewCredential is a free log retrieval operation binding the contract event 0x4846f03be8ef87cb6e611b3a3b878a0aadd7c010f3f25707aa472b41de9dc75d.
//
// Solidity: event NewCredential(uint256 indexed credentialID, address indexed requester, uint256 amount, uint256 time, uint256 unbonding)
func (_GovStaking *GovStakingFilterer) FilterNewCredential(opts *bind.FilterOpts, credentialID []*big.Int, requester []common.Address) (*GovStakingNewCredentialIterator, error) {

	var credentialIDRule []interface{}
	for _, credentialIDItem := range credentialID {
		credentialIDRule = append(credentialIDRule, credentialIDItem)
	}
	var requesterRule []interface{}
	for _, requesterItem := range requester {
		requesterRule = append(requesterRule, requesterItem)
	}

	logs, sub, err := _GovStaking.contract.FilterLogs(opts, "NewCredential", credentialIDRule, requesterRule)
	if err != nil {
		return nil, err
	}
	return &GovStakingNewCredentialIterator{contract: _GovStaking.contract, event: "NewCredential", logs: logs, sub: sub}, nil
}

// WatchNewCredential is a free log subscription operation binding the contract event 0x4846f03be8ef87cb6e611b3a3b878a0aadd7c010f3f25707aa472b41de9dc75d.
//
// Solidity: event NewCredential(uint256 indexed credentialID, address indexed requester, uint256 amount, uint256 time, uint256 unbonding)
func (_GovStaking *GovStakingFilterer) WatchNewCredential(opts *bind.WatchOpts, sink chan<- *GovStakingNewCredential, credentialID []*big.Int, requester []common.Address) (event.Subscription, error) {

	var credentialIDRule []interface{}
	for _, credentialIDItem := range credentialID {
		credentialIDRule = append(credentialIDRule, credentialIDItem)
	}
	var requesterRule []interface{}
	for _, requesterItem := range requester {
		requesterRule = append(requesterRule, requesterItem)
	}

	logs, sub, err := _GovStaking.contract.WatchLogs(opts, "NewCredential", credentialIDRule, requesterRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GovStakingNewCredential)
				if err := _GovStaking.contract.UnpackLog(event, "NewCredential", log); err != nil {
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

// ParseNewCredential is a log parse operation binding the contract event 0x4846f03be8ef87cb6e611b3a3b878a0aadd7c010f3f25707aa472b41de9dc75d.
//
// Solidity: event NewCredential(uint256 indexed credentialID, address indexed requester, uint256 amount, uint256 time, uint256 unbonding)
func (_GovStaking *GovStakingFilterer) ParseNewCredential(log types.Log) (*GovStakingNewCredential, error) {
	event := new(GovStakingNewCredential)
	if err := _GovStaking.contract.UnpackLog(event, "NewCredential", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GovStakingStakedIterator is returned from FilterStaked and is used to iterate over the raw logs and unpacked data for Staked events raised by the GovStaking contract.
type GovStakingStakedIterator struct {
	Event *GovStakingStaked // Event containing the contract specifics and raw log

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
func (it *GovStakingStakedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GovStakingStaked)
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
		it.Event = new(GovStakingStaked)
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
func (it *GovStakingStakedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GovStakingStakedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GovStakingStaked represents a Staked event raised by the GovStaking contract.
type GovStakingStaked struct {
	Staker common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterStaked is a free log retrieval operation binding the contract event 0x9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d.
//
// Solidity: event Staked(address indexed staker, uint256 amount)
func (_GovStaking *GovStakingFilterer) FilterStaked(opts *bind.FilterOpts, staker []common.Address) (*GovStakingStakedIterator, error) {

	var stakerRule []interface{}
	for _, stakerItem := range staker {
		stakerRule = append(stakerRule, stakerItem)
	}

	logs, sub, err := _GovStaking.contract.FilterLogs(opts, "Staked", stakerRule)
	if err != nil {
		return nil, err
	}
	return &GovStakingStakedIterator{contract: _GovStaking.contract, event: "Staked", logs: logs, sub: sub}, nil
}

// WatchStaked is a free log subscription operation binding the contract event 0x9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d.
//
// Solidity: event Staked(address indexed staker, uint256 amount)
func (_GovStaking *GovStakingFilterer) WatchStaked(opts *bind.WatchOpts, sink chan<- *GovStakingStaked, staker []common.Address) (event.Subscription, error) {

	var stakerRule []interface{}
	for _, stakerItem := range staker {
		stakerRule = append(stakerRule, stakerItem)
	}

	logs, sub, err := _GovStaking.contract.WatchLogs(opts, "Staked", stakerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GovStakingStaked)
				if err := _GovStaking.contract.UnpackLog(event, "Staked", log); err != nil {
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

// ParseStaked is a log parse operation binding the contract event 0x9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d.
//
// Solidity: event Staked(address indexed staker, uint256 amount)
func (_GovStaking *GovStakingFilterer) ParseStaked(log types.Log) (*GovStakingStaked, error) {
	event := new(GovStakingStaked)
	if err := _GovStaking.contract.UnpackLog(event, "Staked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GovStakingStakerRegisteredIterator is returned from FilterStakerRegistered and is used to iterate over the raw logs and unpacked data for StakerRegistered events raised by the GovStaking contract.
type GovStakingStakerRegisteredIterator struct {
	Event *GovStakingStakerRegistered // Event containing the contract specifics and raw log

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
func (it *GovStakingStakerRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GovStakingStakerRegistered)
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
		it.Event = new(GovStakingStakerRegistered)
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
func (it *GovStakingStakerRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GovStakingStakerRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GovStakingStakerRegistered represents a StakerRegistered event raised by the GovStaking contract.
type GovStakingStakerRegistered struct {
	Staker   common.Address
	Operator common.Address
	Rewardee common.Address
	Staking  *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterStakerRegistered is a free log retrieval operation binding the contract event 0x8705598e75e33b571bb9c5bcdbf2506dc4e7b88c0262532269bfee3fd15b3d86.
//
// Solidity: event StakerRegistered(address indexed staker, address operator, address rewardee, uint256 staking)
func (_GovStaking *GovStakingFilterer) FilterStakerRegistered(opts *bind.FilterOpts, staker []common.Address) (*GovStakingStakerRegisteredIterator, error) {

	var stakerRule []interface{}
	for _, stakerItem := range staker {
		stakerRule = append(stakerRule, stakerItem)
	}

	logs, sub, err := _GovStaking.contract.FilterLogs(opts, "StakerRegistered", stakerRule)
	if err != nil {
		return nil, err
	}
	return &GovStakingStakerRegisteredIterator{contract: _GovStaking.contract, event: "StakerRegistered", logs: logs, sub: sub}, nil
}

// WatchStakerRegistered is a free log subscription operation binding the contract event 0x8705598e75e33b571bb9c5bcdbf2506dc4e7b88c0262532269bfee3fd15b3d86.
//
// Solidity: event StakerRegistered(address indexed staker, address operator, address rewardee, uint256 staking)
func (_GovStaking *GovStakingFilterer) WatchStakerRegistered(opts *bind.WatchOpts, sink chan<- *GovStakingStakerRegistered, staker []common.Address) (event.Subscription, error) {

	var stakerRule []interface{}
	for _, stakerItem := range staker {
		stakerRule = append(stakerRule, stakerItem)
	}

	logs, sub, err := _GovStaking.contract.WatchLogs(opts, "StakerRegistered", stakerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GovStakingStakerRegistered)
				if err := _GovStaking.contract.UnpackLog(event, "StakerRegistered", log); err != nil {
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

// ParseStakerRegistered is a log parse operation binding the contract event 0x8705598e75e33b571bb9c5bcdbf2506dc4e7b88c0262532269bfee3fd15b3d86.
//
// Solidity: event StakerRegistered(address indexed staker, address operator, address rewardee, uint256 staking)
func (_GovStaking *GovStakingFilterer) ParseStakerRegistered(log types.Log) (*GovStakingStakerRegistered, error) {
	event := new(GovStakingStakerRegistered)
	if err := _GovStaking.contract.UnpackLog(event, "StakerRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GovStakingStakerRemovedIterator is returned from FilterStakerRemoved and is used to iterate over the raw logs and unpacked data for StakerRemoved events raised by the GovStaking contract.
type GovStakingStakerRemovedIterator struct {
	Event *GovStakingStakerRemoved // Event containing the contract specifics and raw log

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
func (it *GovStakingStakerRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GovStakingStakerRemoved)
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
		it.Event = new(GovStakingStakerRemoved)
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
func (it *GovStakingStakerRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GovStakingStakerRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GovStakingStakerRemoved represents a StakerRemoved event raised by the GovStaking contract.
type GovStakingStakerRemoved struct {
	Staker common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterStakerRemoved is a free log retrieval operation binding the contract event 0xb97b17a738bc54faf4f156f2072573525abf5a6f9fe4e3f78a2f159d7a851805.
//
// Solidity: event StakerRemoved(address indexed staker)
func (_GovStaking *GovStakingFilterer) FilterStakerRemoved(opts *bind.FilterOpts, staker []common.Address) (*GovStakingStakerRemovedIterator, error) {

	var stakerRule []interface{}
	for _, stakerItem := range staker {
		stakerRule = append(stakerRule, stakerItem)
	}

	logs, sub, err := _GovStaking.contract.FilterLogs(opts, "StakerRemoved", stakerRule)
	if err != nil {
		return nil, err
	}
	return &GovStakingStakerRemovedIterator{contract: _GovStaking.contract, event: "StakerRemoved", logs: logs, sub: sub}, nil
}

// WatchStakerRemoved is a free log subscription operation binding the contract event 0xb97b17a738bc54faf4f156f2072573525abf5a6f9fe4e3f78a2f159d7a851805.
//
// Solidity: event StakerRemoved(address indexed staker)
func (_GovStaking *GovStakingFilterer) WatchStakerRemoved(opts *bind.WatchOpts, sink chan<- *GovStakingStakerRemoved, staker []common.Address) (event.Subscription, error) {

	var stakerRule []interface{}
	for _, stakerItem := range staker {
		stakerRule = append(stakerRule, stakerItem)
	}

	logs, sub, err := _GovStaking.contract.WatchLogs(opts, "StakerRemoved", stakerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GovStakingStakerRemoved)
				if err := _GovStaking.contract.UnpackLog(event, "StakerRemoved", log); err != nil {
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

// ParseStakerRemoved is a log parse operation binding the contract event 0xb97b17a738bc54faf4f156f2072573525abf5a6f9fe4e3f78a2f159d7a851805.
//
// Solidity: event StakerRemoved(address indexed staker)
func (_GovStaking *GovStakingFilterer) ParseStakerRemoved(log types.Log) (*GovStakingStakerRemoved, error) {
	event := new(GovStakingStakerRemoved)
	if err := _GovStaking.contract.UnpackLog(event, "StakerRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GovStakingUndelegatedIterator is returned from FilterUndelegated and is used to iterate over the raw logs and unpacked data for Undelegated events raised by the GovStaking contract.
type GovStakingUndelegatedIterator struct {
	Event *GovStakingUndelegated // Event containing the contract specifics and raw log

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
func (it *GovStakingUndelegatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GovStakingUndelegated)
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
		it.Event = new(GovStakingUndelegated)
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
func (it *GovStakingUndelegatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GovStakingUndelegatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GovStakingUndelegated represents a Undelegated event raised by the GovStaking contract.
type GovStakingUndelegated struct {
	Delegator common.Address
	Staker    common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterUndelegated is a free log retrieval operation binding the contract event 0x4d10bd049775c77bd7f255195afba5088028ecb3c7c277d393ccff7934f2f92c.
//
// Solidity: event Undelegated(address indexed delegator, address indexed staker, uint256 amount)
func (_GovStaking *GovStakingFilterer) FilterUndelegated(opts *bind.FilterOpts, delegator []common.Address, staker []common.Address) (*GovStakingUndelegatedIterator, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var stakerRule []interface{}
	for _, stakerItem := range staker {
		stakerRule = append(stakerRule, stakerItem)
	}

	logs, sub, err := _GovStaking.contract.FilterLogs(opts, "Undelegated", delegatorRule, stakerRule)
	if err != nil {
		return nil, err
	}
	return &GovStakingUndelegatedIterator{contract: _GovStaking.contract, event: "Undelegated", logs: logs, sub: sub}, nil
}

// WatchUndelegated is a free log subscription operation binding the contract event 0x4d10bd049775c77bd7f255195afba5088028ecb3c7c277d393ccff7934f2f92c.
//
// Solidity: event Undelegated(address indexed delegator, address indexed staker, uint256 amount)
func (_GovStaking *GovStakingFilterer) WatchUndelegated(opts *bind.WatchOpts, sink chan<- *GovStakingUndelegated, delegator []common.Address, staker []common.Address) (event.Subscription, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var stakerRule []interface{}
	for _, stakerItem := range staker {
		stakerRule = append(stakerRule, stakerItem)
	}

	logs, sub, err := _GovStaking.contract.WatchLogs(opts, "Undelegated", delegatorRule, stakerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GovStakingUndelegated)
				if err := _GovStaking.contract.UnpackLog(event, "Undelegated", log); err != nil {
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

// ParseUndelegated is a log parse operation binding the contract event 0x4d10bd049775c77bd7f255195afba5088028ecb3c7c277d393ccff7934f2f92c.
//
// Solidity: event Undelegated(address indexed delegator, address indexed staker, uint256 amount)
func (_GovStaking *GovStakingFilterer) ParseUndelegated(log types.Log) (*GovStakingUndelegated, error) {
	event := new(GovStakingUndelegated)
	if err := _GovStaking.contract.UnpackLog(event, "Undelegated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GovStakingUnstakedIterator is returned from FilterUnstaked and is used to iterate over the raw logs and unpacked data for Unstaked events raised by the GovStaking contract.
type GovStakingUnstakedIterator struct {
	Event *GovStakingUnstaked // Event containing the contract specifics and raw log

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
func (it *GovStakingUnstakedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GovStakingUnstaked)
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
		it.Event = new(GovStakingUnstaked)
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
func (it *GovStakingUnstakedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GovStakingUnstakedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GovStakingUnstaked represents a Unstaked event raised by the GovStaking contract.
type GovStakingUnstaked struct {
	Staker common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterUnstaked is a free log retrieval operation binding the contract event 0x0f5bb82176feb1b5e747e28471aa92156a04d9f3ab9f45f28e2d704232b93f75.
//
// Solidity: event Unstaked(address indexed staker, uint256 amount)
func (_GovStaking *GovStakingFilterer) FilterUnstaked(opts *bind.FilterOpts, staker []common.Address) (*GovStakingUnstakedIterator, error) {

	var stakerRule []interface{}
	for _, stakerItem := range staker {
		stakerRule = append(stakerRule, stakerItem)
	}

	logs, sub, err := _GovStaking.contract.FilterLogs(opts, "Unstaked", stakerRule)
	if err != nil {
		return nil, err
	}
	return &GovStakingUnstakedIterator{contract: _GovStaking.contract, event: "Unstaked", logs: logs, sub: sub}, nil
}

// WatchUnstaked is a free log subscription operation binding the contract event 0x0f5bb82176feb1b5e747e28471aa92156a04d9f3ab9f45f28e2d704232b93f75.
//
// Solidity: event Unstaked(address indexed staker, uint256 amount)
func (_GovStaking *GovStakingFilterer) WatchUnstaked(opts *bind.WatchOpts, sink chan<- *GovStakingUnstaked, staker []common.Address) (event.Subscription, error) {

	var stakerRule []interface{}
	for _, stakerItem := range staker {
		stakerRule = append(stakerRule, stakerItem)
	}

	logs, sub, err := _GovStaking.contract.WatchLogs(opts, "Unstaked", stakerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GovStakingUnstaked)
				if err := _GovStaking.contract.UnpackLog(event, "Unstaked", log); err != nil {
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

// ParseUnstaked is a log parse operation binding the contract event 0x0f5bb82176feb1b5e747e28471aa92156a04d9f3ab9f45f28e2d704232b93f75.
//
// Solidity: event Unstaked(address indexed staker, uint256 amount)
func (_GovStaking *GovStakingFilterer) ParseUnstaked(log types.Log) (*GovStakingUnstaked, error) {
	event := new(GovStakingUnstaked)
	if err := _GovStaking.contract.UnpackLog(event, "Unstaked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GovStakingWithdrawnIterator is returned from FilterWithdrawn and is used to iterate over the raw logs and unpacked data for Withdrawn events raised by the GovStaking contract.
type GovStakingWithdrawnIterator struct {
	Event *GovStakingWithdrawn // Event containing the contract specifics and raw log

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
func (it *GovStakingWithdrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GovStakingWithdrawn)
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
		it.Event = new(GovStakingWithdrawn)
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
func (it *GovStakingWithdrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GovStakingWithdrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GovStakingWithdrawn represents a Withdrawn event raised by the GovStaking contract.
type GovStakingWithdrawn struct {
	CredentialID *big.Int
	Requester    common.Address
	Amount       *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterWithdrawn is a free log retrieval operation binding the contract event 0xcf7d23a3cbe4e8b36ff82fd1b05b1b17373dc7804b4ebbd6e2356716ef202372.
//
// Solidity: event Withdrawn(uint256 indexed credentialID, address requester, uint256 amount)
func (_GovStaking *GovStakingFilterer) FilterWithdrawn(opts *bind.FilterOpts, credentialID []*big.Int) (*GovStakingWithdrawnIterator, error) {

	var credentialIDRule []interface{}
	for _, credentialIDItem := range credentialID {
		credentialIDRule = append(credentialIDRule, credentialIDItem)
	}

	logs, sub, err := _GovStaking.contract.FilterLogs(opts, "Withdrawn", credentialIDRule)
	if err != nil {
		return nil, err
	}
	return &GovStakingWithdrawnIterator{contract: _GovStaking.contract, event: "Withdrawn", logs: logs, sub: sub}, nil
}

// WatchWithdrawn is a free log subscription operation binding the contract event 0xcf7d23a3cbe4e8b36ff82fd1b05b1b17373dc7804b4ebbd6e2356716ef202372.
//
// Solidity: event Withdrawn(uint256 indexed credentialID, address requester, uint256 amount)
func (_GovStaking *GovStakingFilterer) WatchWithdrawn(opts *bind.WatchOpts, sink chan<- *GovStakingWithdrawn, credentialID []*big.Int) (event.Subscription, error) {

	var credentialIDRule []interface{}
	for _, credentialIDItem := range credentialID {
		credentialIDRule = append(credentialIDRule, credentialIDItem)
	}

	logs, sub, err := _GovStaking.contract.WatchLogs(opts, "Withdrawn", credentialIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GovStakingWithdrawn)
				if err := _GovStaking.contract.UnpackLog(event, "Withdrawn", log); err != nil {
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

// ParseWithdrawn is a log parse operation binding the contract event 0xcf7d23a3cbe4e8b36ff82fd1b05b1b17373dc7804b4ebbd6e2356716ef202372.
//
// Solidity: event Withdrawn(uint256 indexed credentialID, address requester, uint256 amount)
func (_GovStaking *GovStakingFilterer) ParseWithdrawn(log types.Log) (*GovStakingWithdrawn, error) {
	event := new(GovStakingWithdrawn)
	if err := _GovStaking.contract.UnpackLog(event, "Withdrawn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
