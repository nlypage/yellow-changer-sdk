package errorz

import "errors"

var (
	CurrencyNotFound = errors.New("currency not found")
	InvalidNetwork   = errors.New("invalid network")
	InvalidAddress   = errors.New("invalid address")
)
