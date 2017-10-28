package bitcoin

import (
	"reflect"
)

var (
	TRANSACTION_POW = ArrayOfBytes(TRANSACTION_POW_COMPLEXITY, POW_PREFIX)
	BLOCK_POW       = ArrayOfBytes(BLOCK_POW_COMPLEXITY, POW_PREFIX)
)

func CheckProofOfWork(prefix []byte, hash []byte) bool {

	if len(prefix) > 0 {
		return reflect.DeepEqual(prefix, hash[:len(prefix)])
	}
	return true
}
