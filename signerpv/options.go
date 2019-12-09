package signerpv

type Option func(pv *SignerPV)

func RetryLimit(limit int) Option {
	return func(pv *SignerPV) {
		pv.retryLimit = limit
	}
}

func MalleableSigCheck(check bool) Option {
	return func(pv *SignerPV) {
		pv.malleableSigCheck = check
	}
}
