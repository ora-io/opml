const child_process = require("child_process")
const fs = require("fs")
const utils = require("./utils")

var config = {
    basedir: "/tmp/cannon",
    programPath: "./mlgo/examples/mnist_mips/mlgo.bin",
    modelPath: "./mlgo/examples/mnist/models/mnist/ggml-model-small-f32-big-endian.bin",
    dataPath: "./mlgo/examples/mnist/models/mnist/input_7"
}

function genGoldenImage(basedir=config.basedir, programPath=config.programPath, modelPath=config.modelPath, dataPath=config.dataPath) {
    var command = "mlvm/mlvm --mipsVMCompatible --outputGolden" + " --basedir="+basedir + " --program="+programPath + " --model="+modelPath + " --data="+dataPath
    console.log(command)
    child_process.execSync(command, {stdio: 'inherit'})
    console.log("golden.json is at " + basedir+"/golden.json")
}


async function deployOPMLContract(basedir=config.basedir) {
    let [c, m, mm] = await utils.deployContract(basedir)
    let json = {
    "Challenge": c.address,
    "MIPS": m.address,
    "MIPSMemory": mm.address,
    }
    console.log("deployed", json)
    fs.writeFileSync(basedir + "/deployed.json", JSON.stringify(json))
    return json
}

// client
async function initiateOPMLRequest(basedir=config.basedir, programPath=config.programPath, modelPath=config.modelPath, dataPath=config.dataPath) {
    genGoldenImage(basedir, programPath, modelPath, dataPath)
    info = await deployOPMLContract(basedir)
    return info
}

// run program
function runProgram(basedir=config.basedir, programPath=config.programPath, modelPath=config.modelPath, dataPath=config.dataPath) {
    var command = "mlvm/mlvm --mipsVMCompatible" + " --basedir="+basedir + " --program="+programPath + " --model="+modelPath + " --data="+dataPath
    console.log(command)
    child_process.execSync(command, {stdio: 'inherit'})
    console.log("checkpoint_final.json is at " + basedir+"/checkpoint_final.json")
}

function getOutput(basedir=config.basedir) {
    let output = fs.readFileSync(basedir+"/output")
    return output
    // return '0x' + output.toString('hex')
}

// submitter
async function uploadResult(basedir=config.basedir) {
    let [c, m, mm] = await utils.deployed(basedir)
    output = getOutput(basedir)
    c.uploadResult(output)
    return output
}

async function getProposedResults(basedir=config.basedir) {
    let [c, m, mm] = await utils.deployed(basedir)
    // console.log(c)
    result = (await c.proposedResults())
    return result
}

// challenger
async function startChallenge(basedir=config.basedir) {
    let [c, m, mm] = await utils.deployed(basedir)
    let startTrie = JSON.parse(fs.readFileSync(basedir+"/golden.json"))
    let finalTrie = JSON.parse(fs.readFileSync(basedir+"/checkpoint_final.json"))
    let preimages = Object.assign({}, startTrie['preimages'], finalTrie['preimages']);
    const finalSystemState = finalTrie['root']
  
    let args = [finalSystemState, finalTrie['step']]
    let cdat = c.interface.encodeFunctionData("initiatePureComputationChallenge", args)
    let nodes = await utils.getTrieNodesForCall(c, c.address, cdat, preimages)
  
    // run "on chain"
    for (n of nodes) {
      await mm.AddTrieNode(n)
    }

    let ret = await c.initiatePureComputationChallenge(...args)
    let receipt = await ret.wait()
    // ChallengeCreated event
    let challengeId = receipt.events[0].args['challengeId'].toNumber()
    console.log("new challenge with id", challengeId)

    return challengeId
}

const RespondState = {
    END: 'END',
    RESPOND: 'RESPOND',
    WAIT: 'WAIT'
}

// challenger and submitter
// return state
async function respond(challengeId, isChallenger, basedir) {
    // console.log("start respond")
    let [c, m, mm] = await utils.deployed(basedir)
    let step = (await c.getStepNumber(challengeId)).toNumber()
    console.log("searching step", step)
  
    if (!(await c.isSearching(challengeId))) {
      console.log("search is done")
      return RespondState.END
    }

    // see if it's proposed or not
    const proposed = await c.getProposedState(challengeId)
    const isProposing = proposed == "0x0000000000000000000000000000000000000000000000000000000000000000"
    if (isProposing != isChallenger) {
        console.log("bad challenger state")
        return RespondState.WAIT
    }
    let thisTrie = utils.getTrieAtStep(basedir, config.programPath, config.modelPath, config.dataPath, step)
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
    return RespondState.RESPOND
}

async function assert(challengeId, isChallenger, basedir=config.basedir) {
    let [c, m, mm] = await utils.deployed(basedir)
    let step = (await c.getStepNumber(challengeId)).toNumber()
    console.log("searching step", step)
  
    if (await c.isSearching(challengeId)) {
      console.log("search is NOT done")
      return
    }

    let cdat
    if (isChallenger) {
      // challenger declare victory
      cdat = c.interface.encodeFunctionData("confirmStateTransition", [challengeId])
    } else {
      // defender declare victory
      // note: not always possible
      cdat = c.interface.encodeFunctionData("denyStateTransition", [challengeId])
    }
  
    let startTrie = utils.getTrieAtStep(basedir, config.programPath, config.modelPath, config.dataPath, step)
    let finalTrie = utils.getTrieAtStep(basedir, config.programPath, config.modelPath, config.dataPath, step+1)
    let preimages = Object.assign({}, startTrie['preimages'], finalTrie['preimages']);
  
    let nodes = await utils.getTrieNodesForCall(c, c.address, cdat, preimages)
    for (n of nodes) {
      await mm.AddTrieNode(n)
    }
  
    let ret
    if (isChallenger) {
      ret = await c.confirmStateTransition(challengeId)
    } else {
      ret = await c.denyStateTransition(challengeId)
    }
  
    let receipt = await ret.wait()
    console.log(receipt.events.map((x) => x.event))
}

module.exports = {
    initiateOPMLRequest, 
    runProgram,
    uploadResult, 
    getProposedResults,
    respond,
    assert,
    config,
    getOutput,
    startChallenge
}