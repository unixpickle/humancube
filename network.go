package main

import (
	"encoding/json"
	"errors"

	"github.com/unixpickle/serializer"
	"github.com/unixpickle/weakai/rnn"
	"github.com/unixpickle/weakai/rnn/lstm"
	"github.com/unixpickle/weakai/rnn/softmax"
)

const serializerTypeNetwork = "github.com/unixpickle/humancube.Network"

const (
	lstmHiddenSize1 = 40
	lstmHiddenSize2 = 40
)

type Network struct {
	RNN     rnn.RNN
	MoveMap map[string]int
}

func NewNetwork(inSize int, moveMap map[string]int) *Network {
	lstmNet1 := lstm.NewNet(rnn.ReLU{}, inSize, lstmHiddenSize1, lstmHiddenSize1)
	lstmNet2 := lstm.NewNet(rnn.ReLU{}, lstmHiddenSize1, lstmHiddenSize2, len(moveMap))
	softmaxLayer := softmax.NewSoftmax(len(moveMap))
	net := rnn.DeepRNN{lstmNet1, lstmNet2, softmaxLayer}
	return &Network{
		RNN:     net,
		MoveMap: moveMap,
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

	res.RNN, ok = slice[1].(rnn.RNN)
	if !ok {
		return nil, err
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
		n.RNN,
	})
}

func (n *Network) SerializerType() string {
	return serializerTypeNetwork
}

func init() {
	serializer.RegisterDeserializer(serializerTypeNetwork, DeserializeNetwork)
}
