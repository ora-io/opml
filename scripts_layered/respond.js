const fs = require("fs")
const { deployed, getTrieNodesForCall, getTrieAtStep } = require("../scripts_layered/lib")

async function main() {
  let [c, m, mm] = await deployed()

  const challengeId = parseInt(process.env.ID)
  const isChallenger = process.env.CHALLENGER == "1"

  console.log("challengeId: ", challengeId)
  console.log("isChallenger: ", isChallenger)

  let step = (await c.getStepNumber(challengeId)).toNumber()
  console.log("searching step: ", step)

  let curLayer = (await c.getCurrentLayer(challengeId)).toNumber()
  console.log("current layer: ", curLayer)

  let nodeID = (await c.getNodeID(challengeId)).toNumber()

  let isSearching = (await c.isSearching(challengeId))


  if (curLayer == 0) {
    nodeID = step
    if (!isSearching) {
      console.log("enter the next layer")
      let startTrie = getTrieAtStep(step, nodeID, false)
      let finalTrie = getTrieAtStep(-1, nodeID, true)
      console.log(challengeId, startTrie['root'], finalTrie['root'], finalTrie['step'])
      ret = await c.toNextLayer(challengeId, startTrie['root'], finalTrie['root'], finalTrie['step'])
      let receipt = await ret.wait()
      console.log("to next layer done", receipt.blockNumber)
      return
    } else {
      const proposed = await c.getProposedState(challengeId)
      const isProposing = proposed == "0x0000000000000000000000000000000000000000000000000000000000000000"
      if (isProposing != isChallenger) {
        console.log("bad challenger state")
        return
      }
      console.log("isProposing", isProposing)
      let thisTrie = getTrieAtStep(step, nodeID, false)
      const root = thisTrie['root']
      console.log("new root", root)

      let ret
      if (isProposing) {
        ret = await c.proposeState(challengeId, root)
      } else {
        ret = await c.respondState(challengeId, root)
      }
      let receipt = await ret.wait()
      console.log("done", receipt.blockNumber)
    }


  } else {
    // curlayer = 1 // mipsvm
    if (!isSearching) {
      console.log("search is done")
      return
    }

    // see if it's proposed or not
    const proposed = await c.getProposedState(challengeId)
    const isProposing = proposed == "0x0000000000000000000000000000000000000000000000000000000000000000"
    if (isProposing != isChallenger) {
        console.log("bad challenger state")
        return
    }
    console.log("isProposing", isProposing)
    let thisTrie = getTrieAtStep(step, nodeID, true)
    const root = thisTrie['root']
    console.log("new root", root)

    let ret
    if (isProposing) {
        ret = await c.proposeState(challengeId, root)
    } else {
        ret = await c.respondState(challengeId, root)
    }
    let receipt = await ret.wait()
    console.log("done", receipt.blockNumber)
  }


}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });