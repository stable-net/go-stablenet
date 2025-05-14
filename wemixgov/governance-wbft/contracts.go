package govwbft

import _ "embed"

var (
	//go:embed govcontracts/GovStaking
	GovStakingContract string
	//go:embed govcontracts/GovNCP
	GovNCPContract string
	//go:embed govcontracts/GovConfig
	GovConfigContract string
	//go:embed govcontracts/GovRewardeeImp
	GovRewardeeImpContract string
)
