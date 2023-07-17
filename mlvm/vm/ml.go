package vm

import (
	"fmt"
	"io/ioutil"

	llama "mlgo/examples/llama/llama_go"
	"mlgo/examples/mnist"
	"mlgo/ml"
)


func LLAMA(nodeID int) ([]byte, int, error){
	modelFile := "./mlgo/examples/llama/models/llama-7b-fp32.bin"
	prompt := "How to combine AI and blockchain?"
	threadCount := 32
	ctx, err := llama.LoadModel(modelFile, true)
	fmt.Println("Load Model Finish")
	if err != nil {
		fmt.Println("load model error: ", err)
		return nil, 0, err
	}
	embd := ml.Tokenize(ctx.Vocab, prompt, true)
	graph, mlctx, err := llama.ExpandGraph(ctx, embd, uint32(len(embd)), 0, threadCount)
	ml.GraphComputeByNodes(mlctx, graph, nodeID)
	envBytes := ml.SaveComputeNodeEnvToBytes(uint32(nodeID), graph.Nodes[nodeID], graph, true)
	return envBytes, int(graph.NodesCount), nil
}

func MNIST(nodeID int) ([]byte, int, error) {
	threadCount := 1
	modelFile := "../../mlgo/examples/mnist/models/mnist/ggml-model-small-f32.bin"
	model, err := mnist.LoadModel(modelFile)
	if err != nil {
		fmt.Println("Load model error: ", err)
		return nil, 0, err
	}
	// load input
	input, err := MNIST_Input(false)
	if err != nil {
		fmt.Println("Load input data error: ", err)
		return nil, 0, err
	}
	graph, ctx := mnist.ExpandGraph(model, threadCount, input)
	ml.GraphComputeByNodes(ctx, graph, nodeID)
	envBytes := ml.SaveComputeNodeEnvToBytes(uint32(nodeID), graph.Nodes[nodeID], graph, true)
	return envBytes, int(graph.NodesCount), nil
}

func MNIST_Input(show bool) ([]float32, error) {
	dataFile := "../../mlgo/examples/mnist/models/mnist/input_7"
	buf, err := ioutil.ReadFile(dataFile)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	digits := make([]float32, 784)

	// render the digit in ASCII
	var c string
	for row := 0; row < 28; row++{
		for col := 0; col < 28; col++ {
			digits[row*28 + col] = float32(buf[row*28 + col])
			if buf[row*28 + col] > 230 {
				c += "*"
			} else {
				c += "_"
			}
		}
		c += "\n"
	}
	if show {
		fmt.Println(c)
	}


	return digits, nil
}