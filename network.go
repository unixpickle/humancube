package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/unixpickle/serializer"
	"github.com/unixpickle/weakai/neuralnet"
	"github.com/unixpickle/weakai/rnn"
)

const serializerTypeNetwork = "github.com/unixpickle/humancube.Network"

const (
	lstmHiddenSize1 = 512
	lstmHiddenSize2 = 512
)

type Network struct {
	Block   rnn.StackedBlock
	MoveMap map[string]int
}

func NewNetwork(inSize int, moveMap map[string]int) *Network {
	netLayer := neuralnet.Network{
		&neuralnet.DenseLayer{InputCount: lstmHiddenSize2, OutputCount: len(moveMap)},
		&neuralnet.SoftmaxLayer{},
	}
	netLayer.Randomize()

	lstmNet1 := rnn.NewLSTM(inSize, lstmHiddenSize1)
	lstmNet2 := rnn.NewLSTM(lstmHiddenSize1, lstmHiddenSize2)
	outputFilter := rnn.NewNetworkBlock(netLayer, 0)

	return &Network{
		Block:   rnn.StackedBlock{lstmNet1, lstmNet2, outputFilter},
		MoveMap: moveMap,
	}
}

func ReadNetwork(path string) (*Network, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	net, err := serializer.DeserializeWithType(data)
	if err != nil {
		return nil, err
	} else if realNet, ok := net.(*Network); ok {
		return realNet, nil
	} else {
		return nil, errors.New("unexpected type of archived data")
	}
}

func DeserializeNetwork(d []byte) (serializer.Serializer, error) {
	slice, err := serializer.DeserializeSlice(d)
	if err != nil {
		return nil, err
	} else if len(slice) != 2 {
		return nil, errors.New("expected two slice elements")
	}
	moveData, ok := slice[0].(serializer.Bytes)
	if !ok {
		return nil, errors.New("expected first slice element to be Bytes")
	}

	var res Network
	if err := json.Unmarshal([]byte(moveData), &res.MoveMap); err != nil {
		return nil, err
	}

	res.Block, ok = slice[1].(rnn.StackedBlock)
	if !ok {
		return nil, errors.New("expected second slice element to be StackedBlock")
	}

	return &res, nil
}

func (n *Network) Serialize() ([]byte, error) {
	moveData, err := json.Marshal(n.MoveMap)
	if err != nil {
		return nil, err
	}
	return serializer.SerializeSlice([]serializer.Serializer{
		serializer.Bytes(moveData),
		n.Block,
	})
}

func (n *Network) SerializerType() string {
	return serializerTypeNetwork
}

func init() {
	serializer.RegisterDeserializer(serializerTypeNetwork, DeserializeNetwork)
}
