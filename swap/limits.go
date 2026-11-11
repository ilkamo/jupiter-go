package swap

const (
	BasisPointsPerWhole     = 10000
	MinReferralFeeBps       = 50
	MaxReferralFeeBps       = 255
	MinBuildAccounts        = 1
	MaxBuildAccounts        = 64
	MinBuildBytes           = 1
	MaxBuildBytes           = 1232
	MinBlockhashExpirySlots = 1
	MaxBlockhashExpirySlots = 300
	TransactionVersionV0    = 0
	TransactionVersionV1    = 1
	MaxTransactionVersion   = TransactionVersionV1

	decimalBase          = 10
	lamportAmountBits    = 64
	blockHeightBits      = 64
	executionSuccessCode = 0
)
