const fs = require("fs")
const { basedir, deployed, getBlockRlp, getTrieNodesForCall } = require("../scripts/lib")

async function main() {
  let [c, m, mm] = await deployed()

  console.log(c.address, m.address, mm.address)
  console.log("basedir: ", basedir)

  // TODO: move this to lib, it's shared with the test
  let startTrie = JSON.parse(fs.readFileSync(basedir+"/golden.json"))
  let finalTrie = JSON.parse(fs.readFileSync(basedir+"/checkpoint_final.json"))
  let preimages = Object.assign({}, startTrie['preimages'], finalTrie['preimages']);
  const finalSystemState = finalTrie['root']

  let args = [finalSystemState, finalTrie['step']]
  let cdat = c.interface.encodeFunctionData("initiatePureComputationChallenge", args)
  let nodes = await getTrieNodesForCall(c, c.address, cdat, preimages)
  console.log("cdat: ", cdat)

  // run "on chain"
  for (n of nodes) {
    await mm.AddTrieNode(n)
  }
// TODO: Setting the gas limit explicitly here shouldn't be necessary, for some
//    weird reason (to be investigated), it is for L2.
//  let ret = await c.initiateChallenge(...args)
  let ret = await c.initiatePureComputationChallenge(...args)
  let receipt = await ret.wait()
  // ChallengeCreated event
  let challengeId = receipt.events[0].args['challengeId'].toNumber()
  console.log("new challenge with id", challengeId)
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
