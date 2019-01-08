package matrixstate

import (
	"encoding/json"

	"encoding/binary"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/pkg/errors"
)

func encodeAccount(account common.Address) ([]byte, error) {
	data, err := json.Marshal(account)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func decodeAccount(data []byte) (common.Address, error) {
	msg := common.Address{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return common.Address{}, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	return msg, nil
}

func encodeAccounts(accounts []common.Address) ([]byte, error) {
	data, err := json.Marshal(accounts)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func decodeAccounts(data []byte) ([]common.Address, error) {
	msg := make([]common.Address, 0)
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	//todo 测试 data为空切片时， msg返回什么
	return msg, nil
}

func encodeUint64(num uint64) []byte {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, num)
	return data
}

func decodeUint64(data []byte) (uint64, error) {
	if len(data) < 8 { // todo data < 8 可以解码吗？
		log.Error(logInfo, "decode uint64 failed", "data size is not enough", "size", len(data))
		return 0, ErrDataSize
	}
	return binary.BigEndian.Uint64(data[:8]), nil
}
