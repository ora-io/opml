const child_process = require("child_process")
const fs = require("fs")

async function getTrieNodesForCall(c, caddress, cdat, preimages) {
    let nodes = []
    while (1) {
        try {
        // TODO: make this eth call?
        // needs something like initiateChallengeWithTrieNodesj
        let calldata = c.interface.encodeFunctionData("callWithTrieNodes", [caddress, cdat, nodes])
        ret = await ethers.provider.call({
            to:c.address,
            data:calldata
        });
        break
        } catch(e) {
        let missing = e.toString().split("'")[1]
        if (missing == undefined) {
            // other kind of error from HTTPProvider
            missing = e.error.message.toString().split("execution reverted: ")[1]
        }
        if (missing !== undefined && missing.length == 64) {
            console.log("requested node", missing)
            let node = preimages["0x"+missing]
            if (node === undefined) {
            throw("node not found")
            }
            const bin = Uint8Array.from(Buffer.from(node, 'base64').toString('binary'), c => c.charCodeAt(0))
            nodes.push(bin)
            continue
        } else if (missing !== undefined && missing.length == 128) {
            let hash = missing.slice(0, 64)
            let offset = parseInt(missing.slice(64, 128), 16)
            console.log("requested hash oracle", hash, offset)
            throw new MissingHashError(hash, offset)
        } else {
            console.log(e)
            break
        }
        }
    }
    return nodes
}

function getTrieAtStep(config) {

    var command = "mlvm/mlvm --mp" + " --basedir="+config.basedir + " --program="+config.programPath + " --model="+config.modelPath + " --data="+config.dataPath + " --modelName="+config.modelName + " --curPhase="+config.curPhase + " --totalPhase="+config.totalPhase + " --checkpoints="+JSON.stringify(config.checkpoints) + " --stepCount="+JSON.stringify(config.stepCount) + " --execCommand="+JSON.stringify(config.execCommand)

    let fn  = config.basedir + "/checkpoint/" + JSON.stringify(config.checkpoints) + ".json"
  
    console.log("getTrieAtStep fn: ", fn)

    if (!fs.existsSync(fn)) {
      console.log(command)
      child_process.execSync(command)
      // child_process.execSync(command, {stdio: 'inherit'})
    }
  
    return JSON.parse(fs.readFileSync(fn))
  }
// function getTrieAtStep(basedir, programPath, modelPath, dataPath, step) {
//     // console.log("getTrieAtStep step: ", step)
//     const fn = basedir+"/checkpoint_"+step.toString()+".json"
  
//     if (!fs.existsSync(fn)) {
//       // console.log("running mipsevm")
//       const command = "mlvm/mlvm --mipsVMCompatible" + " --basedir="+basedir + " --target="+step.toString() + " --program="+programPath + " --model="+modelPath + " --data="+dataPath
//       console.log(command)
//       child_process.execSync(command)
//     //   child_process.execSync(command, {stdio: 'inherit'})
//     }
  
//     return JSON.parse(fs.readFileSync(fn))
// }

async function deployContract(basedir) {
    const MIPS = await ethers.getContractFactory("MIPS")
    const m = await MIPS.deploy()
    const mm = await ethers.getContractAt("MIPSMemory", await m.m())

    let startTrie = JSON.parse(fs.readFileSync(basedir+"/checkpoint/[0].json"))
    let goldenRoot = startTrie["root"]
    console.log("goldenRoot is", goldenRoot)

    const Challenge = await ethers.getContractFactory("MPChallenge")
    const c = await Challenge.deploy(m.address, goldenRoot)

    return [c,m,mm]
}

async function deployed(basedir) {
    let addresses = JSON.parse(fs.readFileSync(basedir+"/deployed.json"))
    const c = await ethers.getContractAt("MPChallenge", addresses["MPChallenge"])
    const m = await ethers.getContractAt("MIPS", addresses["MIPS"])
    const mm = await ethers.getContractAt("MIPSMemory", addresses["MIPSMemory"])
    return [c,m,mm]
}

module.exports = {
    getTrieNodesForCall,
    getTrieAtStep,
    deployed,
    deployContract
}