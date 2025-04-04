package govwbft

import _ "embed"

var (
	//go:embed govcontracts/GovStaking
	GovStakingContract string
	//go:embed govcontracts/GovNCP
	GovNCPContract string
	//go:embed govcontracts/GovConst
	GovConstContract string
	//go:embed govcontracts/GovRewardeeImp
	GovRewardeeImpContract string
)
