const { expect } = require("chai")
const fs = require("fs")
const { deploy, getTrieNodesForCall } = require("../scripts/lib")

// This test needs preimages to run correctly.
// It is skipped when running `make test_contracts`, but can be run with `make test_challenge`.
describe("Challenge contract", function () {
  if (!fs.existsSync("/tmp/cannon/golden.json")) {
    console.log("golden file doesn't exist, skipping test")
    return
  }

  beforeEach(async function () {
    [c, m, mm] = await deploy()
  })
  it("challenge contract deploys", async function() {
    console.log("Challenge deployed at", c.address)
  })
  it("initiate challenge", async function() {

    let startTrie = JSON.parse(fs.readFileSync("/tmp/cannon/golden.json"))
    let finalTrie = JSON.parse(fs.readFileSync("/tmp/cannon/checkpoint_final.json"))
    let preimages = Object.assign({}, startTrie['preimages'], finalTrie['preimages']);
    const finalSystemState = finalTrie['root']

    let args = [finalSystemState, finalTrie['step']]
    let cdat = c.interface.encodeFunctionData("initiatePureComputationChallenge", args)
    // console.log("contract: ", c)
    // console.log("cdat: ", cdat)
    let nodes = await getTrieNodesForCall(c, c.address, cdat, preimages)
    // console.log("what?")
    // run "on chain"
    for (n of nodes) {
      await mm.AddTrieNode(n)
    }
    let ret = await c.initiatePureComputationChallenge(...args)
    let receipt = await ret.wait()
    // ChallengeCreated event
    let challengeId = receipt.events[0].args['challengeId'].toNumber()
    console.log("new challenge with id", challengeId)
    let challengerResults = await c.challengerResults()
    console.log(challengerResults)

    // the real issue here is from step 0->1 when we write the input hash
    // TODO: prove the challenger wrong?
  }).timeout(200_000)
})
