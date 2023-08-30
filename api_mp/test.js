const api = require("./lib")
const fs = require("fs")
const utils = require("./utils")
const child_process = require("child_process")

async function main() {
    const submitterDir = "/tmp/cannon"
    const challengerDir = "/tmp/cannon_fault"

    api.config.totalPhase = 3 // test

    var submitterConfig = api.getConfig()
    submitterConfig.basedir = submitterDir

    var challengerConfig = api.getConfig()
    challengerConfig.basedir = challengerDir

    // init
    child_process.execSync("rm -rf /tmp/cannon/* /tmp/cannon_fault/*", {stdio: 'inherit'})
    child_process.execSync("mkdir -p /tmp/cannon/checkpoint /tmp/cannon/data", {stdio: 'inherit'})
    child_process.execSync("mkdir -p /tmp/cannon_fault/checkpoint /tmp/cannon_fault/data", {stdio: 'inherit'})

    // === client ===
    // 1. create the golden image
    // 2. deploy the contract
    submitterConfig.checkpoints = [0]
    submitterConfig.curPhase = 0
    deployInfo = await api.initiateOPMLRequest(submitterConfig)
    console.log("contract address: ", deployInfo.Challenge)
    
    // === submitter ===
    // 0. get the data
    // 1. run the program 
    // 2. upload the result
    api.runProgram(submitterConfig)
    result = await api.uploadResult(submitterDir)
    console.log("submitter upload results: ", result)

    // === challenger ===
    // 0. get the data
    // 1. run the program and find the results incorrect!
    // 2. start challenge
    child_process.execSync("cp /tmp/cannon/checkpoint/[0].json /tmp/cannon_fault/checkpoint/", {stdio: 'inherit'})
    child_process.execSync("cp /tmp/cannon/deployed.json /tmp/cannon_fault/", {stdio: 'inherit'})

    api.runProgram(challengerConfig)
    result = api.getOutput(challengerDir)
    console.log("challenger's results: ", result)
    
    challengeId = await api.startChallenge(challengerConfig)
    console.log("start challenge! challengeId: ", challengeId)


    // === interactive dispute game ===
    for (i = 0; i < 30; i++) {
        console.log("--- STEP ", i, " / 30 ---")
        state = await api.respond(challengeId, true, challengerConfig)
        state = await api.respond(challengeId, false, submitterConfig)
        if (state == "END") {
            console.log("bisection ends")
            break
        } 
    }

    // === assert ===
    console.log("ASSERTING AS CHALLENGER (should fail)")
    await api.assert(challengeId, true, challengerConfig)
    console.log("ASSERTING AS DEFENDER (should pass)")
    await api.assert(challengeId, false, submitterConfig)
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
