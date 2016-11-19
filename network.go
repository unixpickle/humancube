package humancube

import (
	"encoding/json"
	"errors"

	"github.com/unixpickle/num-analysis/linalg"
	"github.com/unixpickle/serializer"
	"github.com/unixpickle/weakai/neuralnet"
	"github.com/unixpickle/weakai/rnn"
)

const (
	hiddenSize             = 300
	dropoutKeepProbability = 0.5
)

func init() {
	var n Network
	serializer.RegisterTypedDeserializer(n.SerializerType(), DeserializeNetwork)
}

type Network struct {
	Block   rnn.StackedBlock
	MoveMap map[string]int
}

func NewNetwork(inSize int, moveMap map[string]int) *Network {
	netLayer := neuralnet.Network{
		&neuralnet.DropoutLayer{
			KeepProbability: dropoutKeepProbability,
			Training:        false,
		},
		&neuralnet.DenseLayer{InputCount: hiddenSize, OutputCount: len(moveMap)},
		&neuralnet.LogSoftmaxLayer{},
	}
	netLayer.Randomize()

	lstmNet1 := rnn.NewLSTM(inSize, hiddenSize)
	outputFilter := rnn.NewNetworkBlock(netLayer, 0)

	return &Network{
		Block:   rnn.StackedBlock{lstmNet1, outputFilter},
		MoveMap: moveMap,
	}
}

func ReadNetwork(path string) (*Network, error) {
	var net *Network
	err := serializer.LoadAny(path, &net)
	return net, err
}

func DeserializeNetwork(d []byte) (*Network, error) {
	var moveData serializer.Bytes
	var net rnn.StackedBlock

	if err := serializer.DeserializeAny(d, &moveData, &net); err != nil {
		return nil, err
	}
	var moveMap map[string]int
	if err := json.Unmarshal(moveData, &moveMap); err != nil {
		return nil, errors.New("read move map: " + err.Error())
	}
	return &Network{Block: net, MoveMap: moveMap}, nil
}

func (n *Network) OutputMove(out linalg.Vector) string {
	_, idx := out.Max()
	for m, i := range n.MoveMap {
		if i == idx {
			return m
		}
	}
	return "?"
}

func (n *Network) Serialize() ([]byte, error) {
	moveData, err := json.Marshal(n.MoveMap)
	if err != nil {
		return nil, err
	}
	return serializer.SerializeAny(serializer.Bytes(moveData), n.Block)
}

func (n *Network) SerializerType() string {
	return "github.com/unixpickle/humancube.Network"
}

func (n *Network) Dropout(on bool) {
	net := n.Block[len(n.Block)-1].(*rnn.NetworkBlock).Network()
	dropout := net[0].(*neuralnet.DropoutLayer)
	dropout.Training = on
}
