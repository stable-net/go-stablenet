package gov

type GovContract string

const (
	CONTRACT_REGISTRY          = "Registry"
	CONTRACT_GOV               = "Gov"
	CONTRACT_GOV_IMP           = "GovImp"
	CONTRACT_NCPEXIT           = "NCPExit"
	CONTRACT_NCPEXIT_IMP       = "NCPExitImp"
	CONTRACT_STAKING           = "Staking"
	CONTRACT_STAKING_IMP       = "StakingImp"
	CONTRACT_BALLOTSTORAGE     = "BallotStorage"
	CONTRACT_BALLOTSTORAGE_IMP = "BallotStorageImp"
	CONTRACT_ENVSTORAGE        = "EnvStorage"
	CONTRACT_ENVSTORAGE_IMP    = "EnvStorageImp"
)

type GovDomain string

const (
	DOMAIN_Gov           = "GovernanceContract"
	DOMAIN_NCPExit       = "NCPExit"
	DOMAIN_Staking       = "Staking"
	DOMAIN_BallotStorage = "BallotStorage"
	DOMAIN_EnvStorage    = "EnvStorage"

	DOMAIN_StakingReward = "StakingReward"
	DOMAIN_Ecosystem     = "Ecosystem"
	DOMAIN_Maintenance   = "Maintenance"
	DOMAIN_FeeCollector  = "FeeCollector"
)

const (
	CONTRACT_GOV_STAKING      = "GovStaking"
	CONTRACT_GOV_NCP          = "GovNCP"
	CONTRACT_GOV_CONST        = "GovConst"
	CONTRACT_GOV_REWARDEE_IMP = "GovRewardeeImp"
	CONTRACT_GOV_REWARDEE     = "GovRewardee"
	CONTRACT_MULTISIG_WALLET  = "MultiSigWallet"
	CONTRACT_OPERATOR_SAMPLE  = "OperatorSample"
)
