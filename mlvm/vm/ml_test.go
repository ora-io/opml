package vm

import (
	"fmt"
	llama "mlgo/examples/llama/llama_go"
	"mlgo/examples/mnist"
	"mlgo/ml"
	"testing"
)

func TestMNIST(t *testing.T) {
	threadCount := 1
	modelFile := "../../mlgo/examples/mnist/models/mnist/ggml-model-small-f32.bin"
	model, err := mnist.LoadModel(modelFile)
	if err != nil {
		fmt.Println("Load model error: ", err)
		return 
	}
	// load input
	input, err := MNIST_Input(true)
	if err != nil {
		fmt.Println("Load input data error: ", err)
		return 
	}
	graph, _ := mnist.ExpandGraph(model, threadCount, input)
	fmt.Println("graph.nodeNum: ", graph.NodesCount)
}

func TestLLAMA(t *testing.T) {
	modelFile := "../../mlgo/examples/llama/models/llama-7b-fp32.bin"
	prompt := "Why Golang is so popular?"
	threadCount := 32
	ctx, err := llama.LoadModel(modelFile, true)
	fmt.Println("Load Model Finish")
	if err != nil {
		fmt.Println("load model error: ", err)
		return 
	}
	embd := ml.Tokenize(ctx.Vocab, prompt, true)
	graph, _, _ := llama.ExpandGraph(ctx, embd, uint32(len(embd)), 0, threadCount)
	fmt.Println("graph.nodeCount: ", graph.NodesCount)
}