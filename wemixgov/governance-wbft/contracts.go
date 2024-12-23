package govwbft

import _ "embed"

// contract codes for Chapel upgrade
var (
	//go:embed govcontracts/GovStaking
	GovStakingContract string
	//go:embed govcontracts/GovNCP
	GovNCPContract string
	//go:embed govcontracts/GovConst
	GovConstContract string
)
