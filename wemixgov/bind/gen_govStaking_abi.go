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
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"staker\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Delegated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"credentialID\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"requester\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"time\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"unbonding\",\"type\":\"uint256\"}],\"name\":\"NewCredential\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"staker\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Staked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"staker\",\"type\":\"address\"}],\"name\":\"StakerDeactivated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"staker\",\"type\":\"address\"}],\"name\":\"StakerReactivated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"staker\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"rewardee\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"staking\",\"type\":\"uint256\"}],\"name\":\"StakerRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"staker\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Undelegated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"staker\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Unstaked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"credentialID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"requester\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Withdrawn\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"GOV_CONST\",\"outputs\":[{\"internalType\":\"contractGovConst\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"afterStabilization\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"credentialCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"credentials\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"requester\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"requestTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"withdrawableTime\",\"type\":\"uint256\"},{\"internalType\":\"enumGovStaking.WithdrawalStatus\",\"name\":\"status\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_staker\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"delegate\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"delegateTo\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_addr\",\"type\":\"address\"}],\"name\":\"isOperatorOrRewardee\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_staker\",\"type\":\"address\"}],\"name\":\"isStaker\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"_staker\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_rewardee\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"_blsPK\",\"type\":\"bytes\"}],\"name\":\"registerStaker\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"stake\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"stakerByOperator\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"stakerByRewardee\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"stakerInfo\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"rewardee\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"staking\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"delegated\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"blsPubKey\",\"type\":\"bytes\"},{\"internalType\":\"enumGovStaking.StakerState\",\"name\":\"state\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"stakerLength\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"stakers\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalStaking\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_staker\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"undelegate\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"unstake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_cid\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Sigs: map[string]string{
		"e8aaca24": "GOV_CONST()",
		"d617246e": "afterStabilization()",
		"cd0e35b9": "credentialCount()",
		"e0574e3f": "credentials(uint256)",
		"026e402b": "delegate(address,uint256)",
		"438bb7e5": "delegateTo(address,address)",
		"cdb72f06": "isOperatorOrRewardee(address)",
		"6f1e8533": "isStaker(address)",
		"650c1ea2": "registerStaker(uint256,address,address,bytes)",
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
	Bin: "0x608060405234801561001057600080fd5b5061208c806100206000396000f3fe6080604052600436106101145760003560e01c8063650c1ea2116100a0578063cdb72f0611610064578063cdb72f0614610326578063d617246e14610346578063e0574e3f14610360578063e8aaca24146103c8578063fed1252a146103de57600080fd5b8063650c1ea2146102845780636f1e8533146102975780639dbccf6a146102c7578063a694fc3a146102fd578063cd0e35b91461031057600080fd5b80632e1a7d4d116100e75780632e1a7d4d146101c5578063438bb7e5146101e55780634d99dd161461021d5780634e745f1f1461023d5780635748f6f31461026f57600080fd5b8063026e402b14610119578063165defa41461012e578063264ebc38146101575780632e17de78146101a5575b600080fd5b61012c610127366004611c8d565b610400565b005b34801561013a57600080fd5b5061014460005481565b6040519081526020015b60405180910390f35b34801561016357600080fd5b5061018d610172366004611cb7565b6005602052600090815260409020546001600160a01b031681565b6040516001600160a01b03909116815260200161014e565b3480156101b157600080fd5b5061012c6101c0366004611cd2565b610563565b3480156101d157600080fd5b5061012c6101e0366004611cd2565b610820565b3480156101f157600080fd5b50610144610200366004611ceb565b600660209081526000928352604080842090915290825290205481565b34801561022957600080fd5b5061012c610238366004611c8d565b6109a8565b34801561024957600080fd5b5061025d610258366004611cb7565b610b61565b60405161014e96959493929190611d62565b34801561027b57600080fd5b50610144610c30565b61012c610292366004611def565b610c41565b3480156102a357600080fd5b506102b76102b2366004611cb7565b6112e3565b604051901515815260200161014e565b3480156102d357600080fd5b5061018d6102e2366004611cb7565b6004602052600090815260409020546001600160a01b031681565b61012c61030b366004611cd2565b6112f6565b34801561031c57600080fd5b5061014460075481565b34801561033257600080fd5b506102b7610341366004611cb7565b6113a0565b34801561035257600080fd5b506009546102b79060ff1681565b34801561036c57600080fd5b506103b761037b366004611cd2565b600860205260009081526040902080546001820154600283015460038401546004909401546001600160a01b0390931693919290919060ff1685565b60405161014e959493929190611e8a565b3480156103d457600080fd5b5061018d61100081565b3480156103ea57600080fd5b506103f36113e6565b60405161014e9190611ec8565b808034146104295760405162461bcd60e51b815260040161042090611f15565b60405180910390fd5b610432336112e3565b156104785760405162461bcd60e51b81526020600482015260166024820152757374616b65722063616e6e6f742064656c656761746560501b6044820152606401610420565b610481336113a0565b156104d95760405162461bcd60e51b815260206004820152602260248201527f6f70657261746f72287265776172646565292063616e6e6f742064656c656761604482015261746560f01b6064820152608401610420565b6104e5838360016113f2565b3360009081526006602090815260408083206001600160a01b038716845290915281208054849290610518908490611f62565b90915550506040518281526001600160a01b0384169033907fe5541a6b6103d4fa7e021ed54fad39c66f27a76bd13d374cf6240ae6bd0bb72b906020015b60405180910390a3505050565b336000908152600460205260409020546001600160a01b0316806105995760405162461bcd60e51b815260040161042090611f7a565b600082116105da5760405162461bcd60e51b815260206004820152600e60248201526d616d6f756e74206973207a65726f60901b6044820152606401610420565b6001600160a01b0381166000908152600360208190526040822090810154600282015491929161060a9190611fa7565b9050838110156106535760405162461bcd60e51b8152602060048201526014602482015273696e73756666696369656e742062616c616e636560601b6044820152606401610420565b6110006001600160a01b031663ba631d3f6040518163ffffffff1660e01b8152600401602060405180830381865afa158015610693573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906106b79190611fbe565b6106c18583611fa7565b10156107375783811461072d5760405162461bcd60e51b815260206004820152602e60248201527f616d6f756e74206d75737420657175616c2062616c616e636520746f2064656160448201526d31ba34bb30ba329039ba30b5b2b960911b6064820152608401610420565b61073783836116ae565b836000808282546107489190611fa7565b92505081905550838260020160008282546107639190611fa7565b925050819055506107d7846110006001600160a01b031663fde7f3716040518163ffffffff1660e01b8152600401602060405180830381865afa1580156107ae573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906107d29190611fbe565b61171d565b826001600160a01b03167f0f5bb82176feb1b5e747e28471aa92156a04d9f3ab9f45f28e2d704232b93f758560405161081291815260200190565b60405180910390a250505050565b60008181526008602052604090206001600482015460ff16600281111561084957610849611d1e565b1461088b5760405162461bcd60e51b81526020600482015260126024820152711a5b9d985b1a590818dc9959195b9d1a585b60721b6044820152606401610420565b80546001600160a01b031633146108e45760405162461bcd60e51b815260206004820152601b60248201527f6d73672e73656e646572206973206e6f742072657175657374657200000000006044820152606401610420565b80600301544210156109385760405162461bcd60e51b815260206004820152601860248201527f6e6f74207965742074696d6520746f20776974686472617700000000000000006044820152606401610420565b60018101548154610954916001600160a01b0390911690611845565b60048101805460ff19166002179055600181015460408051338152602081019290925283917fcf7d23a3cbe4e8b36ff82fd1b05b1b17373dc7804b4ebbd6e2356716ef202372910160405180910390a25050565b3360009081526006602090815260408083206001600160a01b0386168452909152902054811115610a125760405162461bcd60e51b8152602060048201526014602482015273696e73756666696369656e742062616c616e636560601b6044820152606401610420565b6001600160a01b03821660009081526003602081905260408220908101805491928492610a40908490611fa7565b9250508190555081816002016000828254610a5b9190611fa7565b90915550503360009081526006602090815260408083206001600160a01b038716845290915281208054849290610a93908490611fa7565b9091555060029050600582015460ff166002811115610ab457610ab4611d1e565b14610ac857610ac33383611845565b610b24565b81600080828254610ad99190611fa7565b92505081905550610b24826110006001600160a01b031663840c17716040518163ffffffff1660e01b8152600401602060405180830381865afa1580156107ae573d6000803e3d6000fd5b6040518281526001600160a01b0384169033907f4d10bd049775c77bd7f255195afba5088028ecb3c7c277d393ccff7934f2f92c90602001610556565b6003602081905260009182526040909120805460018201546002830154938301546004840180546001600160a01b039485169694909316949192610ba490611fd7565b80601f0160208091040260200160405190810160405280929190818152602001828054610bd090611fd7565b8015610c1d5780601f10610bf257610100808354040283529160200191610c1d565b820191906000526020600020905b815481529060010190602001808311610c0057829003601f168201915b5050506005909301549192505060ff1686565b6000610c3c6001611963565b905090565b84803414610c615760405162461bcd60e51b815260040161042090611f15565b6110006001600160a01b031663ba631d3f6040518163ffffffff1660e01b8152600401602060405180830381865afa158015610ca1573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610cc59190611fbe565b8610158015610d3757506110006001600160a01b031663129060ab6040518163ffffffff1660e01b8152600401602060405180830381865afa158015610d0f573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610d339190611fbe565b8611155b610d735760405162461bcd60e51b815260206004820152600d60248201526c6f7574206f6620626f756e647360981b6044820152606401610420565b336001600160a01b03861614801590610d955750336001600160a01b03851614155b610def5760405162461bcd60e51b815260206004820152602560248201527f6f70657261746f722063616e6e6f74206265207374616b6572206f7220726577604482015264617264656560d81b6064820152608401610420565b6001600160a01b03851615801590610e0f57506001600160a01b03841615155b610e4a5760405162461bcd60e51b815260206004820152600c60248201526b7a65726f206164647265737360a01b6044820152606401610420565b836001600160a01b0316856001600160a01b031603610eab5760405162461bcd60e51b815260206004820152601960248201527f7374616b65722063616e6e6f74206265207265776172646565000000000000006044820152606401610420565b610eb4336113a0565b15610f015760405162461bcd60e51b815260206004820152601e60248201527f6f70657261746f7220697320616c7265616479207265676973746572656400006044820152606401610420565b610f0a856113a0565b15610f575760405162461bcd60e51b815260206004820152601c60248201527f7374616b657220697320616c72656164792072656769737465726564000000006044820152606401610420565b610f60846113a0565b15610fad5760405162461bcd60e51b815260206004820152601e60248201527f726577617264656520697320616c7265616479207265676973746572656400006044820152606401610420565b6110006001600160a01b0316638280a25a6040518163ffffffff1660e01b8152600401602060405180830381865afa158015610fed573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906110119190611fbe565b82146110585760405162461bcd60e51b8152602060048201526016602482015275696e76616c696420626c73207075626c6963206b657960501b6044820152606401610420565b61106360018661196d565b61109f5760405162461bcd60e51b815260206004820152600d60248201526c7374616b65722065786973747360981b6044820152606401610420565b6040518060c00160405280336001600160a01b03168152602001856001600160a01b031681526020018781526020016000815260200184848080601f016020809104026020016040519081016040528093929190818152602001838380828437600092019190915250505090825250602001600290526001600160a01b03808716600090815260036020818152604092839020855181549086166001600160a01b03199182161782558683015160018301805491909716911617909455918401516002840155606084015190830155608083015180516111859260048501920190611bd8565b5060a082015160058201805460ff191660018360028111156111a9576111a9611d1e565b02179055505033600090815260046020908152604080832080546001600160a01b03808c166001600160a01b031992831681179093558a1685526005909352908320805490921617905580548892508190611205908490611f62565b925050819055506110006001600160a01b031663decf02066040518163ffffffff1660e01b8152600401602060405180830381865afa15801561124c573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906112709190611fbe565b61127a6001611963565b1061128d576009805460ff191660011790555b604080513381526001600160a01b038681166020830152918101889052908616907f8705598e75e33b571bb9c5bcdbf2506dc4e7b88c0262532269bfee3fd15b3d869060600160405180910390a2505050505050565b60006112f0600183611989565b92915050565b808034146113165760405162461bcd60e51b815260040161042090611f15565b336000908152600460205260409020546001600160a01b03168061134c5760405162461bcd60e51b815260040161042090611f7a565b611358818460006113f2565b806001600160a01b03167f9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d8460405161139391815260200190565b60405180910390a2505050565b6001600160a01b038181166000908152600460205260408120549091161515806112f05750506001600160a01b0390811660009081526005602052604090205416151590565b6060610c3c60016119ab565b6001600160a01b03831660009081526003602052604090206002600582015460ff16600281111561142557611425611d1e565b1461159e576000600582015460ff16600281111561144557611445611d1e565b036114625760405162461bcd60e51b815260040161042090611f7a565b81156114bb5760405162461bcd60e51b815260206004820152602260248201527f63616e6e6f742064656c656761746520746f20696e616374697665207374616b60448201526132b960f11b6064820152608401610420565b6110006001600160a01b031663ba631d3f6040518163ffffffff1660e01b8152600401602060405180830381865afa1580156114fb573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061151f9190611fbe565b8310156115945760405162461bcd60e51b815260206004820152603a60248201527f746f2072656163746976617465207374616b65722c205f616d6f756e74206d7560448201527f7374206265206174206c6561737420746865206d696e696d756d0000000000006064820152608401610420565b61159e84826119b8565b6110006001600160a01b031663129060ab6040518163ffffffff1660e01b8152600401602060405180830381865afa1580156115de573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906116029190611fbe565b8382600201546116129190611f62565b11156116575760405162461bcd60e51b8152602060048201526014602482015273657863656564656420746865206d6178696d756d60601b6044820152606401610420565b826000808282546116689190611f62565b92505081905550828160020160008282546116839190611f62565b909155505081156116a857828160030160008282546116a29190611f62565b90915550505b50505050565b60058101805460ff191660019081179091556116ca9083611a25565b5080600301546000808282546116e09190611fa7565b90915550506040516001600160a01b038316907f35ba4655eac3567283864ab2ed68f93c94fc31c7179e6ff46e534b2b2d7d1ccc90600090a25050565b6040518060a00160405280336001600160a01b03168152602001838152602001428152602001824261174f9190611f62565b815260200160018152506008600060076000815461176c90612011565b919050819055815260200190815260200160002060008201518160000160006101000a8154816001600160a01b0302191690836001600160a01b0316021790555060208201518160010155604082015181600201556060820151816003015560808201518160040160006101000a81548160ff021916908360028111156117f5576117f5611d1e565b021790555050600754604080518581524260208201529081018490523392507f4846f03be8ef87cb6e611b3a3b878a0aadd7c010f3f25707aa472b41de9dc75d9060600160405180910390a35050565b804710156118955760405162461bcd60e51b815260206004820152601d60248201527f416464726573733a20696e73756666696369656e742062616c616e63650000006044820152606401610420565b6000826001600160a01b03168260405160006040518083038185875af1925050503d80600081146118e2576040519150601f19603f3d011682016040523d82523d6000602084013e6118e7565b606091505b505090508061195e5760405162461bcd60e51b815260206004820152603a60248201527f416464726573733a20756e61626c6520746f2073656e642076616c75652c207260448201527f6563697069656e74206d617920686176652072657665727465640000000000006064820152608401610420565b505050565b60006112f0825490565b6000611982836001600160a01b038416611a3a565b9392505050565b6001600160a01b03811660009081526001830160205260408120541515611982565b6060600061198283611a89565b60058101805460ff191660021790556119d260018361196d565b5080600301546000808282546119e89190611f62565b90915550506040516001600160a01b038316907fd2ba2793e35e0b66263a892ba0ae35ec7001e3726cc69422b07afafff7fad01a90600090a25050565b6000611982836001600160a01b038416611ae5565b6000818152600183016020526040812054611a81575081546001818101845560008481526020808220909301849055845484825282860190935260409020919091556112f0565b5060006112f0565b606081600001805480602002602001604051908101604052809291908181526020018280548015611ad957602002820191906000526020600020905b815481526020019060010190808311611ac5575b50505050509050919050565b60008181526001830160205260408120548015611bce576000611b09600183611fa7565b8554909150600090611b1d90600190611fa7565b9050818114611b82576000866000018281548110611b3d57611b3d61202a565b9060005260206000200154905080876000018481548110611b6057611b6061202a565b6000918252602080832090910192909255918252600188019052604090208390555b8554869080611b9357611b93612040565b6001900381819060005260206000200160009055905585600101600086815260200190815260200160002060009055600193505050506112f0565b60009150506112f0565b828054611be490611fd7565b90600052602060002090601f016020900481019282611c065760008555611c4c565b82601f10611c1f57805160ff1916838001178555611c4c565b82800160010185558215611c4c579182015b82811115611c4c578251825591602001919060010190611c31565b50611c58929150611c5c565b5090565b5b80821115611c585760008155600101611c5d565b80356001600160a01b0381168114611c8857600080fd5b919050565b60008060408385031215611ca057600080fd5b611ca983611c71565b946020939093013593505050565b600060208284031215611cc957600080fd5b61198282611c71565b600060208284031215611ce457600080fd5b5035919050565b60008060408385031215611cfe57600080fd5b611d0783611c71565b9150611d1560208401611c71565b90509250929050565b634e487b7160e01b600052602160045260246000fd5b60038110611d5257634e487b7160e01b600052602160045260246000fd5b50565b611d5e81611d34565b9052565b600060018060a01b03808916835260208189168185015287604085015286606085015260c06080850152855191508160c085015260005b82811015611db55786810182015185820160e001528101611d99565b82811115611dc757600060e084870101525b5050601f01601f1916820160e0019050611de460a0830184611d55565b979650505050505050565b600080600080600060808688031215611e0757600080fd5b85359450611e1760208701611c71565b9350611e2560408701611c71565b9250606086013567ffffffffffffffff80821115611e4257600080fd5b818801915088601f830112611e5657600080fd5b813581811115611e6557600080fd5b896020828501011115611e7757600080fd5b9699959850939650602001949392505050565b6001600160a01b038616815260208101859052604081018490526060810183905260a08101611eb883611d34565b8260808301529695505050505050565b6020808252825182820181905260009190848201906040850190845b81811015611f095783516001600160a01b031683529284019291840191600101611ee4565b50909695505050505050565b6020808252601d908201527f616d6f756e7420616e64206d73672e76616c7565206d69736d61746368000000604082015260600190565b634e487b7160e01b600052601160045260246000fd5b60008219821115611f7557611f75611f4c565b500190565b6020808252601390820152723ab73932b3b4b9ba32b932b21039ba30b5b2b960691b604082015260600190565b600082821015611fb957611fb9611f4c565b500390565b600060208284031215611fd057600080fd5b5051919050565b600181811c90821680611feb57607f821691505b60208210810361200b57634e487b7160e01b600052602260045260246000fd5b50919050565b60006001820161202357612023611f4c565b5060010190565b634e487b7160e01b600052603260045260246000fd5b634e487b7160e01b600052603160045260246000fdfea2646970667358221220bf55112499f734e0860ca88bf7fd2bfc2d93e51b7204c4ba35818196cbb24d3164736f6c634300080e0033",
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

// AfterStabilization is a free data retrieval call binding the contract method 0xd617246e.
//
// Solidity: function afterStabilization() view returns(bool)
func (_GovStaking *GovStakingCaller) AfterStabilization(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _GovStaking.contract.Call(opts, &out, "afterStabilization")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// AfterStabilization is a free data retrieval call binding the contract method 0xd617246e.
//
// Solidity: function afterStabilization() view returns(bool)
func (_GovStaking *GovStakingSession) AfterStabilization() (bool, error) {
	return _GovStaking.Contract.AfterStabilization(&_GovStaking.CallOpts)
}

// AfterStabilization is a free data retrieval call binding the contract method 0xd617246e.
//
// Solidity: function afterStabilization() view returns(bool)
func (_GovStaking *GovStakingCallerSession) AfterStabilization() (bool, error) {
	return _GovStaking.Contract.AfterStabilization(&_GovStaking.CallOpts)
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
// Solidity: function stakerInfo(address ) view returns(address operator, address rewardee, uint256 staking, uint256 delegated, bytes blsPubKey, uint8 state)
func (_GovStaking *GovStakingCaller) StakerInfo(opts *bind.CallOpts, arg0 common.Address) (struct {
	Operator  common.Address
	Rewardee  common.Address
	Staking   *big.Int
	Delegated *big.Int
	BlsPubKey []byte
	State     uint8
}, error) {
	var out []interface{}
	err := _GovStaking.contract.Call(opts, &out, "stakerInfo", arg0)

	outstruct := new(struct {
		Operator  common.Address
		Rewardee  common.Address
		Staking   *big.Int
		Delegated *big.Int
		BlsPubKey []byte
		State     uint8
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Operator = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.Rewardee = *abi.ConvertType(out[1], new(common.Address)).(*common.Address)
	outstruct.Staking = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.Delegated = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.BlsPubKey = *abi.ConvertType(out[4], new([]byte)).(*[]byte)
	outstruct.State = *abi.ConvertType(out[5], new(uint8)).(*uint8)

	return *outstruct, err

}

// StakerInfo is a free data retrieval call binding the contract method 0x4e745f1f.
//
// Solidity: function stakerInfo(address ) view returns(address operator, address rewardee, uint256 staking, uint256 delegated, bytes blsPubKey, uint8 state)
func (_GovStaking *GovStakingSession) StakerInfo(arg0 common.Address) (struct {
	Operator  common.Address
	Rewardee  common.Address
	Staking   *big.Int
	Delegated *big.Int
	BlsPubKey []byte
	State     uint8
}, error) {
	return _GovStaking.Contract.StakerInfo(&_GovStaking.CallOpts, arg0)
}

// StakerInfo is a free data retrieval call binding the contract method 0x4e745f1f.
//
// Solidity: function stakerInfo(address ) view returns(address operator, address rewardee, uint256 staking, uint256 delegated, bytes blsPubKey, uint8 state)
func (_GovStaking *GovStakingCallerSession) StakerInfo(arg0 common.Address) (struct {
	Operator  common.Address
	Rewardee  common.Address
	Staking   *big.Int
	Delegated *big.Int
	BlsPubKey []byte
	State     uint8
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

// RegisterStaker is a paid mutator transaction binding the contract method 0x650c1ea2.
//
// Solidity: function registerStaker(uint256 _amount, address _staker, address _rewardee, bytes _blsPK) payable returns()
func (_GovStaking *GovStakingTransactor) RegisterStaker(opts *bind.TransactOpts, _amount *big.Int, _staker common.Address, _rewardee common.Address, _blsPK []byte) (*types.Transaction, error) {
	return _GovStaking.contract.Transact(opts, "registerStaker", _amount, _staker, _rewardee, _blsPK)
}

// RegisterStaker is a paid mutator transaction binding the contract method 0x650c1ea2.
//
// Solidity: function registerStaker(uint256 _amount, address _staker, address _rewardee, bytes _blsPK) payable returns()
func (_GovStaking *GovStakingSession) RegisterStaker(_amount *big.Int, _staker common.Address, _rewardee common.Address, _blsPK []byte) (*types.Transaction, error) {
	return _GovStaking.Contract.RegisterStaker(&_GovStaking.TransactOpts, _amount, _staker, _rewardee, _blsPK)
}

// RegisterStaker is a paid mutator transaction binding the contract method 0x650c1ea2.
//
// Solidity: function registerStaker(uint256 _amount, address _staker, address _rewardee, bytes _blsPK) payable returns()
func (_GovStaking *GovStakingTransactorSession) RegisterStaker(_amount *big.Int, _staker common.Address, _rewardee common.Address, _blsPK []byte) (*types.Transaction, error) {
	return _GovStaking.Contract.RegisterStaker(&_GovStaking.TransactOpts, _amount, _staker, _rewardee, _blsPK)
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

// GovStakingStakerDeactivatedIterator is returned from FilterStakerDeactivated and is used to iterate over the raw logs and unpacked data for StakerDeactivated events raised by the GovStaking contract.
type GovStakingStakerDeactivatedIterator struct {
	Event *GovStakingStakerDeactivated // Event containing the contract specifics and raw log

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
func (it *GovStakingStakerDeactivatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GovStakingStakerDeactivated)
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
		it.Event = new(GovStakingStakerDeactivated)
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
func (it *GovStakingStakerDeactivatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GovStakingStakerDeactivatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GovStakingStakerDeactivated represents a StakerDeactivated event raised by the GovStaking contract.
type GovStakingStakerDeactivated struct {
	Staker common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterStakerDeactivated is a free log retrieval operation binding the contract event 0x35ba4655eac3567283864ab2ed68f93c94fc31c7179e6ff46e534b2b2d7d1ccc.
//
// Solidity: event StakerDeactivated(address indexed staker)
func (_GovStaking *GovStakingFilterer) FilterStakerDeactivated(opts *bind.FilterOpts, staker []common.Address) (*GovStakingStakerDeactivatedIterator, error) {

	var stakerRule []interface{}
	for _, stakerItem := range staker {
		stakerRule = append(stakerRule, stakerItem)
	}

	logs, sub, err := _GovStaking.contract.FilterLogs(opts, "StakerDeactivated", stakerRule)
	if err != nil {
		return nil, err
	}
	return &GovStakingStakerDeactivatedIterator{contract: _GovStaking.contract, event: "StakerDeactivated", logs: logs, sub: sub}, nil
}

// WatchStakerDeactivated is a free log subscription operation binding the contract event 0x35ba4655eac3567283864ab2ed68f93c94fc31c7179e6ff46e534b2b2d7d1ccc.
//
// Solidity: event StakerDeactivated(address indexed staker)
func (_GovStaking *GovStakingFilterer) WatchStakerDeactivated(opts *bind.WatchOpts, sink chan<- *GovStakingStakerDeactivated, staker []common.Address) (event.Subscription, error) {

	var stakerRule []interface{}
	for _, stakerItem := range staker {
		stakerRule = append(stakerRule, stakerItem)
	}

	logs, sub, err := _GovStaking.contract.WatchLogs(opts, "StakerDeactivated", stakerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GovStakingStakerDeactivated)
				if err := _GovStaking.contract.UnpackLog(event, "StakerDeactivated", log); err != nil {
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

// ParseStakerDeactivated is a log parse operation binding the contract event 0x35ba4655eac3567283864ab2ed68f93c94fc31c7179e6ff46e534b2b2d7d1ccc.
//
// Solidity: event StakerDeactivated(address indexed staker)
func (_GovStaking *GovStakingFilterer) ParseStakerDeactivated(log types.Log) (*GovStakingStakerDeactivated, error) {
	event := new(GovStakingStakerDeactivated)
	if err := _GovStaking.contract.UnpackLog(event, "StakerDeactivated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GovStakingStakerReactivatedIterator is returned from FilterStakerReactivated and is used to iterate over the raw logs and unpacked data for StakerReactivated events raised by the GovStaking contract.
type GovStakingStakerReactivatedIterator struct {
	Event *GovStakingStakerReactivated // Event containing the contract specifics and raw log

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
func (it *GovStakingStakerReactivatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GovStakingStakerReactivated)
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
		it.Event = new(GovStakingStakerReactivated)
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
func (it *GovStakingStakerReactivatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GovStakingStakerReactivatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GovStakingStakerReactivated represents a StakerReactivated event raised by the GovStaking contract.
type GovStakingStakerReactivated struct {
	Staker common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterStakerReactivated is a free log retrieval operation binding the contract event 0xd2ba2793e35e0b66263a892ba0ae35ec7001e3726cc69422b07afafff7fad01a.
//
// Solidity: event StakerReactivated(address indexed staker)
func (_GovStaking *GovStakingFilterer) FilterStakerReactivated(opts *bind.FilterOpts, staker []common.Address) (*GovStakingStakerReactivatedIterator, error) {

	var stakerRule []interface{}
	for _, stakerItem := range staker {
		stakerRule = append(stakerRule, stakerItem)
	}

	logs, sub, err := _GovStaking.contract.FilterLogs(opts, "StakerReactivated", stakerRule)
	if err != nil {
		return nil, err
	}
	return &GovStakingStakerReactivatedIterator{contract: _GovStaking.contract, event: "StakerReactivated", logs: logs, sub: sub}, nil
}

// WatchStakerReactivated is a free log subscription operation binding the contract event 0xd2ba2793e35e0b66263a892ba0ae35ec7001e3726cc69422b07afafff7fad01a.
//
// Solidity: event StakerReactivated(address indexed staker)
func (_GovStaking *GovStakingFilterer) WatchStakerReactivated(opts *bind.WatchOpts, sink chan<- *GovStakingStakerReactivated, staker []common.Address) (event.Subscription, error) {

	var stakerRule []interface{}
	for _, stakerItem := range staker {
		stakerRule = append(stakerRule, stakerItem)
	}

	logs, sub, err := _GovStaking.contract.WatchLogs(opts, "StakerReactivated", stakerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GovStakingStakerReactivated)
				if err := _GovStaking.contract.UnpackLog(event, "StakerReactivated", log); err != nil {
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

// ParseStakerReactivated is a log parse operation binding the contract event 0xd2ba2793e35e0b66263a892ba0ae35ec7001e3726cc69422b07afafff7fad01a.
//
// Solidity: event StakerReactivated(address indexed staker)
func (_GovStaking *GovStakingFilterer) ParseStakerReactivated(log types.Log) (*GovStakingStakerReactivated, error) {
	event := new(GovStakingStakerReactivated)
	if err := _GovStaking.contract.UnpackLog(event, "StakerReactivated", log); err != nil {
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
