package bitcoin

const (
	BLOCKCHAIN_DEFAULT_PORT = 9200

	NETWORK_KEY_SIZE = 80

	TRANSACTION_HEADER_SIZE = NETWORK_KEY_SIZE /* from key */ +
		NETWORK_KEY_SIZE /* to key */ +
		4 /* int32 timestamp */ +
		32 /* sha256 payload hash */ +
		4 /* int32 payload length */ +
		4 /* int32 nonce */

	BLOCK_HEADER_SIZE = NETWORK_KEY_SIZE /* origin key */ +
		4 /* int32 timestamp */ +
		32 /* prev block hash */ +
		32 /* merkel tree hash */ +
		4 /* int32 nonce */

	KEY_POW_COMPLEXITY         = 0
	BLOCK_POW_COMPLEXITY       = 2
	TRANSACTION_POW_COMPLEXITY = 1

	POW_PREFIX = 0

	KEY_SIZE = 28
	IP_SIZE  = 21

	MESSAGE_TYPE_SIZE    = 1
	MESSAGE_OPTIONS_SIZE = 4
)

const (
	MESSAGE_GET_NODES = iota + 20
	MESSAGE_SEND_NODES

	MESSAGE_GET_TRANSACTION
	MESSAGE_SEND_TRANSACTION

	MESSAGE_GET_BLOCK
	MESSAGE_SEND_BLOCK
)
