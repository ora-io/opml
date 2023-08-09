# Tutorial

In this tutorial, you will learn how to run a simple handwritten digit recognition DNN model (MNIST) in OPML.

## Train

First train a DNN model using Pytorch, the training part is shown in `mlgo/examples/mnist/trainning/mnist.ipynb`
and then save the model at `mlgo/examples/mnist/models/mnist/mnist-small.state_dict`

## Model Format Conversion

Convert the Pytorch model to GGML format, only save the hyperparameters used plus the model weights and biases. Run `mlgo/examples/mnist/convert-h5-to-ggml.py` to convert your pytorch model. The output format is:

- magic constant (int32)
- repeated list of tensors
- number of dimensions of tensor (int32)
- tensor dimension (int32 repeated)
- values of tensor (int32)

Note that the model is saved in big-endian, making it easy to process in the big-endian MIPS-32 VM. 

## Construct ML Program in MIPS VM

Write a program for DNN model inference, the source code is at `mlgo/examples/mnist_mips`

Note that the model, input, and output are all stored in the VM memory.

Go supports compilation to MIPS. However, the generated executable is in ELF format. We'd like to get a pure sequence of MIPS instructions instead.
To build a ML program in MIPS VM, just run `mlgo/examples/mnist_mips/build.sh`

## Construct VM Image

The user who proposes a ML inference request should first construct a initial VM image

```shell
mlvm/mlvm --outputGolden --basedir=/tmp/cannon --program="$PROGRAM_PATH" --model="$MODEL_PATH" --data="$DATA_PATH" --mipsVMCompatible
```

The initial VM image is saved at `/tmp/cannon/golden.json`, organized in a Merkle trie tree format.

## On-chain Dispute

(For more details, please refer to `demo/challenge_simple.sh`)

First start a local node
```shell
npx hardhat node
```
Then deploy the smart contract with the initial VM image
```shell
npx hardhat run scripts/deploy.js --network localhost
```
Then the challenger can start the dispute game
```shell
BASEDIR=/tmp/cannon_fault npx hardhat run scripts/challenge.js --network localhost
```
Then the submitter and the challenger will interactively find the dispute point using bisection protocol
```shell
for i in {1..25}; do
    echo ""
    echo "--- STEP $i / 25 ---"
    echo ""
    BASEDIR=/tmp/cannon_fault CHALLENGER=1 npx hardhat run scripts/respond.js --network localhost
    BASEDIR=/tmp/cannon CHALLENGER=0 npx hardhat run scripts/respond.js --network localhost
done
```
Finally, the bisection protocol will help to locate the dispute step, the step will be sent to the arbitration contract on the blockchain.
```shell
npx hardhat run scripts/assert.js  --network localhost
```