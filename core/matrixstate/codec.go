package matrixstate

import (
	"encoding/json"
	"github.com/matrix/go-matrix/mc"
	"github.com/pkg/errors"
)

func (self *keyManager) initCodec() {
	self.codecMap[MSPTopologyGraph] = new(TopologyGraphCodec)
	self.codecMap[MSPElectGraph] = new(ElectGraphCodec)
	self.codecMap[MSPElectOnlineState] = new(ElectOnlineStateCodec)
}

type codec interface {
	encodeFn(msg interface{}) ([]byte, error)
	decodeFn(data []byte) (interface{}, error)
}

////////////////////////////////////////////////////////////////////////
// key = MSPTopologyGraph
type TopologyGraphCodec struct {
}

func (TopologyGraphCodec) encodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (TopologyGraphCodec) decodeFn(data []byte) (interface{}, error) {
	msg := new(mc.ElectGraph)
	err := json.Unmarshal(data, msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	if msg == nil {
		return nil, errors.New("msg is nil")
	}
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// key = MSPElectGraph
type ElectGraphCodec struct {
}

func (ElectGraphCodec) encodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (ElectGraphCodec) decodeFn(data []byte) (interface{}, error) {
	msg := new(mc.ElectGraph)
	err := json.Unmarshal(data, msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	if msg == nil {
		return nil, errors.New("msg is nil")
	}
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// key = MSPElectOnlineState
type ElectOnlineStateCodec struct {
}

func (ElectOnlineStateCodec) encodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (ElectOnlineStateCodec) decodeFn(data []byte) (interface{}, error) {
	msg := new(mc.ElectOnlineStatus)
	err := json.Unmarshal(data, msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	if msg == nil {
		return nil, errors.New("msg is nil")
	}
	return msg, nil
}
