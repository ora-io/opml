#!/usr/bin/env bash

# --- SCRIPT SETUP -------------------------------------------------------------

shout() {
    echo ""
    echo "----------------------------------------"
    echo "$1"
    echo "----------------------------------------"
    echo ""
}

# Exit if any command fails.
set -e

exit_trap() {
    # Print an error if the last command failed
    # (in which case the script is exiting because of set -e).
    [[ $? == 0 ]] && return
    echo "----------------------------------------"
    echo "EARLY EXIT: SCRIPT FAILED"
    echo "----------------------------------------"

    # Kill (send SIGTERM) to the whole process group, also killing
    # any background processes.
    # I think the trap command resets SIGTERM before resending it to the whole
    # group. (cf. https://stackoverflow.com/a/2173421)
    trap - SIGTERM && kill -- -$$
}
# trap "exit_trap" SIGINT SIGTERM EXIT



# --- CHALLENGE SETUP ----------------------------------------------------------

workdir=$(cd $(dirname $0);cd ..; pwd)

# AI model (7B-LLaMA model)
PROGRAM_PATH="./mlgo/ml_mips/ml_mips.bin"
MODEL_NAME="MNIST"
MODEL_PATH="$workdir/mlgo/examples/mnist/models/mnist/ggml-model-small-f32.bin" 
# PROMPT="How to combine AI and blockchain?"
DATA_PATH="./mlgo/examples/mnist/models/mnist/input_7"

export PROGRAM_PATH=$PROGRAM_PATH
export MODEL_NAME=$MODEL_NAME
export MODEL_PATH=$MODEL_PATH
export DATA_PATH=$DATA_PATH
export PROMPT=$PROMPT

# challenge ID, read by respond.js and assert.js
export ID=0

# clear data from previous runs
rm -rf /tmp/cannon/* /tmp/cannon_fault/*
mkdir -p /tmp/cannon/data
mkdir -p /tmp/cannon/checkpoint
mkdir -p /tmp/cannon_fault/data
mkdir -p /tmp/cannon_fault/checkpoint

# stored in /tmp/cannon/golden.json
shout "GENERATING INITIAL MEMORY STATE CHECKPOINT"
mlvm/mlvm --basedir=/tmp/cannon --program="$PROGRAM_PATH" --modelName="$MODEL_NAME" --model="$MODEL_PATH" --prompt="$PROMPT" --data="$DATA_PATH" --target=0 --nodeID 0

shout "DEPLOYING CONTRACTS"
npx hardhat run scripts_layered/deploy.js --network localhost

# challenger will use same initial memory checkpoint and deployed contracts
cp /tmp/cannon/deployed.json /tmp/cannon_fault/
cp -r /tmp/cannon/checkpoint /tmp/cannon_fault/
cp -r /tmp/cannon/data /tmp/cannon_fault/

# shout "COMPUTING FAKE MIPS FINAL MEMORY CHECKPOINT"
# BASEDIR=/tmp/cannon_fault mlvm/mlvm --program="$PROGRAM_PATH" --model="$MODEL_PATH" --data="$DATA_PATH"


# --- BINARY SEARCH ------------------------------------------------------------

shout "STARTING CHALLENGE"
BASEDIR=/tmp/cannon_fault npx hardhat run scripts_layered/challenge.js --network localhost

shout "BINARY SEARCH"
for i in $(seq 1 1 25); do
# for i in {1..40}; do
    echo ""
    echo "--- STEP $i / 40 ---"
    echo ""
    # bug: https://github.com/ethereum-optimism/cannon/issues/99
    BASEDIR=/tmp/cannon_fault CHALLENGER=1 npx hardhat run scripts_layered/respond.js --network localhost || BASEDIR=/tmp/cannon_fault CHALLENGER=1 npx hardhat run scripts_layered/respond.js --network localhost
    BASEDIR=/tmp/cannon CHALLENGER=0 npx hardhat run scripts_layered/respond.js --network localhost || BASEDIR=/tmp/cannon CHALLENGER=0 npx hardhat run scripts_layered/respond.js --network localhost
done

# --- SINGLE STEP EXECUTION ----------------------------------------------------

shout "ASSERTING AS CHALLENGER (should fail)"
set +e # this should fail!
BASEDIR=/tmp/cannon_fault CHALLENGER=1 npx hardhat run scripts_layered/assert.js --network localhost
set -e

shout "ASSERTING AS DEFENDER (should pass)"
npx hardhat run scripts_layered/assert.js  --network localhost
