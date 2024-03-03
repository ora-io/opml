# OPML: OPtimistic Machine Learning on Blockchain

OPML enables off-chain AI model inference using optimistic approach with an on chain interactive dispute engine implementing fault proofs.

For more in-depth information, refer to the [project wiki](https://github.com/hyperoracle/opml/wiki).

You can also find a tutorial on building a straightforward handwritten digit recognition DNN model (MNIST) within OPML in the [`docs/tutorial.md`](docs/tutorial.md).

## Building

Pre-requisites: Go (Go 1.19), Node.js, Make, and CMake.

```
git clone https://github.com/ora-io/opml.git --recursive

make build
```

## Examples

The script files [`demo/challenge_simple.sh`](demo/challenge_simple.sh) presents an example scenario (a DNN model for MNIST) demonstrating the whole process of a fault proof, including the challenge game and single step verification.

To test the example, we should first start a local node:

```shell
npx hardhat node
```

Then we can run:

```shell
bash ./demo/challenge_simple.sh
```

A large language model, the llama example is provided in the branch ["llama"](https://github.com/hyperoracle/opml/tree/llama) (It also works for llama 2).

## Roadmap

🔨 = Pending

🛠 = Work In Progress

✅ = Feature complete


| Feature |  Status |
| ------- |  :------: |
| **Supported Model** |    |
| DNN for MNIST | ✅ |
| LLaMA | ✅ |
| General DNN Model (Onnx Support) | 🛠 |
| Traditional ML Algorithm (Decision Tree, KNN etc) | 🔨 |
| **Mode** |    |
| Inference| ✅ |
| Training | 🔨 |
| Fine-tuning | 🔨 |
| **Optimization** |    |
| ZK Fault Proof with zkOracle and zkWASM | 🛠 |
| GPU Acceleration | 🛠 |
| High Performance VM | 🛠 |
| **Functionality** |    |
| User-Friendly SDK| 🛠 |

## Project Structure

```
mlgo -- A tensor library for machine learning in pure Golang that can run on MIPS.
mlvm -- A MIPS runtime with ML execution
contracts -- A Merkleized MIPS processor on chain + the challenge logic
```

## License

This code is MIT licensed.

Part of this code is borrowed from `ethereum-optimism/cannon`

Note: This code is unaudited. It in NO WAY should be used to secure any money until a lot more
testing and auditing are done. 
