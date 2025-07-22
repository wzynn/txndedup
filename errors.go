package txndedup

import "errors"

var (
	ErrInvalidTimeWindow         = errors.New("invalid time window")
	ErrInvalidCleanupInterval    = errors.New("invalid cleanup interval")
	ErrMissingRedisConfig        = errors.New("missing redis config")
	ErrUnsupportedStorageType    = errors.New("unsupported storage type")
	ErrInvalidTransactionRequest = errors.New("invalid transaction request")
	ErrTransactionNotFound       = errors.New("transaction not found")
)
