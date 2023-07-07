#!/usr/bin/env bash

# The following variables can be overridden as environment variables:
# * BLOCK (block whose transition will be challenged)
# * WRONG_BLOCK (block number used by challenger)
# * SKIP_NODE (skip forking a node, useful if you've already forked a node)
#
# Example usage:
# SKIP_NODE=1 BLOCK=13284469 WRONG_BLOCK=13284491 ./demo/challenge_simple.sh

# --- DOC ----------------------------------------------------------------------

# In this example, the challenger will challenge the transition from a block
# (`BLOCK`), but pretends that chain state before another block (`WRONG_BLOCK`)
# is the state before the challenged block. Consequently, the challenger will
# disagree with the defender on every single step of the challenge game, and the
# single step to execute will be the very first MIPS instruction executed. The
# reason is that the initial MIPS state Merkle root is stored on-chain, and
# immediately modified to reflect the fact that the input hash for the block is
# written at address 0x3000000.
#
# (The input hash is automatically validated against the blockhash, so note that
# in this demo the challenger has to provide the correct (`BLOCK`) input hash to
# the `initiateChallenge` function of `Challenge.sol`, but will execute as
# though the input hash was the one derived from `WRONG_BLOCK`.)
#
# Because the challenger uses the wrong inputs, it will assert a post-state
# (Merkle root) for the first MIPS instruction that has the wrong input hash at
# 0x3000000. Hence, the challenge will fail.


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
trap "exit_trap" SIGINT SIGTERM EXIT



# --- CHALLENGE SETUP ----------------------------------------------------------

# AI model (mnist model)
PROGRAM_PATH="./mlgo/mlgo.bin"
MODEL_PATH="./mlgo/examples/mnist/models/mnist/ggml-model-small-f32-big-endian.bin"
DATA_PATH="./mlgo/examples/mnist/models/mnist/input_7"

export PROGRAM_PATH=$PROGRAM_PATH
export MODEL_PATH=$MODEL_PATH
export DATA_PATH=$DATA_PATH

# challenge ID, read by respond.js and assert.js
export ID=0

# clear data from previous runs
rm -rf /tmp/cannon/* /tmp/cannon_fault/*

# stored in /tmp/cannon/golden.json
shout "GENERATING INITIAL MEMORY STATE CHECKPOINT"
mlvm/mlvm --outputGolden --basedir=/tmp/cannon --program="$PROGRAM_PATH" --model="$MODEL_PATH" --data="$DATA_PATH" --mipsVMCompatible

shout "DEPLOYING CONTRACTS"
npx hardhat run scripts/deploy.js --network localhost

# challenger will use same initial memory checkpoint and deployed contracts
cp /tmp/cannon/{golden,deployed}.json /tmp/cannon_fault/

shout "COMPUTING FAKE MIPS FINAL MEMORY CHECKPOINT"
BASEDIR=/tmp/cannon_fault mlvm/mlvm --program="$PROGRAM_PATH" --model="$MODEL_PATH" --data="$DATA_PATH" --mipsVMCompatible


# --- BINARY SEARCH ------------------------------------------------------------

shout "STARTING CHALLENGE"
BASEDIR=/tmp/cannon_fault npx hardhat run scripts/challenge.js --network localhost

shout "BINARY SEARCH"
for i in {1..25}; do
    echo ""
    echo "--- STEP $i / 25 ---"
    echo ""
    BASEDIR=/tmp/cannon_fault CHALLENGER=1 npx hardhat run scripts/respond.js --network localhost
    BASEDIR=/tmp/cannon CHALLENGER=0 npx hardhat run scripts/respond.js --network localhost
done

# --- SINGLE STEP EXECUTION ----------------------------------------------------

shout "ASSERTING AS CHALLENGER (should fail)"
set +e # this should fail!
BASEDIR=/tmp/cannon_fault CHALLENGER=1 npx hardhat run scripts/assert.js --network localhost
set -e

shout "ASSERTING AS DEFENDER (should pass)"
npx hardhat run scripts/assert.js  --network localhost
