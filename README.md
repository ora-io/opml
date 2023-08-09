# OPML: OPtimistic Machine Learning on Blockchain


---

OPML enables off-chain AI model inference using optimistic approach with an on chain interactive dispute engine implementing fault proofs.

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

The script files [`demo/challenge_simple.sh`](demo/challenge_simple.sh) presents an example scenario (a DNN model for MNIST) demonstrating the whole process of a fault proof, including the challenge game and single step verification.

To test the example, we should first start a local node
```shell
npx hardhat node
```
Then we can run 
```shell
sh ./demo/challenge_simple.sh
```

## License

This code is MIT licensed.

Part of this code is borrowed from `ethereum-optimism/cannon`

Note: This code is unaudited. It in NO WAY should be used to secure any money until a lot more
testing and auditing are done. 