const api = require("./lib")
const fs = require("fs")
const utils = require("./utils")
const child_process = require("child_process")

async function main() {
    const submitterDir = "/tmp/cannon"
    const challengerDir = "/tmp/cannon_fault"

    // init
    child_process.execSync("rm -rf /tmp/cannon/* /tmp/cannon_fault/*", {stdio: 'inherit'})
    child_process.execSync("mkdir -p /tmp/cannon", {stdio: 'inherit'})
    child_process.execSync("mkdir -p /tmp/cannon_fault", {stdio: 'inherit'})

    // === client ===
    // 1. create the golden image
    // 2. deploy the contract
    deployInfo = await api.initiateOPMLRequest(submitterDir, api.config.programPath, api.config.modelPath, api.config.dataPath)
    console.log("contract address: ", deployInfo.Challenge)
    
    // === submitter ===
    // 0. get the data
    // 1. run the program 
    // 2. upload the result
    api.runProgram(submitterDir, api.config.programPath, api.config.modelPath, api.config.dataPath)
    result = await api.uploadResult(submitterDir)
    console.log("submitter upload results: ", result)

    // === challenger ===
    // 0. get the data
    // 1. run the program and find the results incorrect!
    // 2. start challenge
    child_process.execSync("cp /tmp/cannon/golden.json /tmp/cannon_fault/", {stdio: 'inherit'})
    child_process.execSync("cp /tmp/cannon/deployed.json /tmp/cannon_fault/", {stdio: 'inherit'})

    api.runProgram(challengerDir, api.config.programPath, api.config.modelPath, api.config.dataPath)
    result = api.getOutput(challengerDir)
    console.log("challenger's results: ", result)
    
    challengeId = await api.startChallenge(challengerDir)
    console.log("start challenge! challengeId: ", challengeId)


    // === interactive dispute game ===
    for (i = 0; i < 25; i++) {
        console.log("--- STEP ", i, " / 25 ---")
        state = await api.respond(challengeId, true, challengerDir)
        state = await api.respond(challengeId, false, submitterDir)
        if (state == "END") {
            console.log("bisection ends")
            break
        } 
    }

    // === assert ===
    console.log("ASSERTING AS CHALLENGER (should fail)")
    await api.assert(challengeId, true, challengerDir)
    console.log("ASSERTING AS DEFENDER (should pass)")
    await api.assert(challengeId, false, submitterDir)
}

async function test() {
    basedir = "/tmp/cannon_fault"
    let [c, m, mm] = await utils.deployed(basedir)
    // let output = '0xc542de910972c15981876b0495928484a2655f8a548fde03b3ca59bcd60cfcb3'
    step = (await c.getStepNumber(0)).toNumber()
    console.log("step: ", step)
}


main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
