# OPML: OPtimistic Machine Learning on Blockchain


---

OPML enables off-chain AI model inference using optimistic approach with an on chain interactive dispute engine implementing EVM-equivalent fault proofs.

## Directory Layout

```
mlgo -- A tensor library for machine learning in pure Golang that can run on MIPS.
mlvm -- A MIPS runtime with ML execution
contracts -- A Merkleized MIPS processor on chain + the challenge logic
```

## Building

Pre-requisites: Go, Node.js, Make, and CMake.

```
make build
```

## Examples

### MNIST

The script files [`demo/challenge_simple.sh`](demo/challenge_simple.sh) presents an example scenario (a DNN model for MNIST) demonstrating the whole process of a fault proof, including the challenge game and single step verification.

To test the example, we should first start a local node
```shell
npx hardhat node
```
Then we can run 
```shell
sh ./demo/challenge_simple.sh
```

### 7B-LLaMA

**Note**: This part is still under development!

Before testing 7B-LLaMA, please refer to `mlgo/examples/llama/README.md` and download the model of 7B-LLaMA.

After that, we can first start a local node
```shell
npx hardhat node
```
Then we can run 
```shell
sh ./demo/challenge_llama.sh
```

Note: when running `sh ./demo/challenge_llama.sh`, you may encounter such an error in console "SocketError: other side closed". Just ignore it. :) This is a special "feature" of hardhat when running the JS script that takes too long time. I have fixed it in the script. Although you can see an error in console, the script should run correctly.

## License

Most of this code is MIT licensed, minigeth is LGPL3.

Part of this code is borrowed from `ethereum-optimism/cannon`

Note: This code is unaudited. It in NO WAY should be used to secure any money until a lot more
testing and auditing are done. I have deployed this nowhere, have advised against deploying it, and
make no guarantees of security of ANY KIND.
