var express = require("express")
var router = express.Router()

const child_process = require("child_process")
const api = require("../api_mp/lib")


router.get("/", function(req, res){
	res.send("hello world!");
});

/*
req={modelName, prompt}
res= {
    "MPChallenge": c.address,
    "MIPS": m.address,
    "MIPSMemory": mm.address,
    }
*/
router.post("/opMLRequest", async function(req, res){
    api.config.totalPhase = 3 // test

    console.log(req.query)
    updateConfig(req.query)
    initDir()
    info = await deployContract()
    res.send(info)
    console.log("deployed info: ", info)
})


/*
req={}
res=result
*/
router.post("/submitterUploadResult", async function(req, res) {
    submitterConfig = getSubmitterConfig()
    api.runProgram(submitterConfig)
    result = await api.uploadResult(submitterConfig.basedir)
    res.send(result)
    console.log("submitter upload results: ", result)
})


/*
res={}
res={result, challengeId}
*/
router.post("/startChallenge", async function(req, res) {
    // === challenger ===
    // 0. get the data
    // 1. run the program and find the results incorrect!
    // 2. start challenge
    challengerConfig = getChallengerConfig()

    child_process.execSync("cp /tmp/cannon/checkpoint/[0].json /tmp/cannon_fault/checkpoint/", {stdio: 'inherit'})
    child_process.execSync("cp /tmp/cannon/deployed.json /tmp/cannon_fault/", {stdio: 'inherit'})

    api.runProgram(challengerConfig)
    result = api.getOutput(challengerConfig.basedir)
    console.log("challenger's results: ", result)
    
    challengeId = await api.startChallenge(challengerConfig)
    console.log("start challenge! challengeId: ", challengeId)
    api.config.challengeId = challengeId

    data = {result: result, challengeId: challengeId}
    console.log("data: ", data)
    res.send(data)
})


/*
req={challengeId}
res={
        config: config,
        state: RespondState.RESPOND,
        root: null,
    }
*/
router.post("/challengerRespond", async function(req, res) {
    challengerConfig = getChallengerConfig()
    challengeId = req.challengeId ? req.challengeId : challengerConfig.challengeId
    result = await api.respond(challengeId, true, challengerConfig)
    console.log(result)
    res.send(result)
})

/*
req={challengeId}
res={
        config: config,
        state: RespondState.RESPOND,
        root: null,
    }
*/
router.post("/submitterRespond", async function(req, res) {
    submitterConfig = getSubmitterConfig()
    challengeId = req.challengeId ? req.challengeId : submitterConfig.challengeId
    result = await api.respond(challengeId, false, submitterConfig)
    console.log(result)
    res.send(result)
})

/*
req={challengeId}
res=events
*/
router.post("/challengerAssert", async function(req, res) {
    challengerConfig = getChallengerConfig()
    challengeId = req.challengeId ? req.challengeId : challengerConfig.challengeId
    result = await api.assert(challengeId, true, challengerConfig)
    res.send(result)
})

/*
req={challengeId}
res=events
*/
router.post("/submitterAssert", async function(req, res) {
    submitterConfig = getSubmitterConfig()
    challengeId = req.challengeId ? req.challengeId : submitterConfig.challengeId
    result = await api.assert(challengeId, false, submitterConfig)
    res.send(result)
})


function initDir() {
    // init
    child_process.execSync("rm -rf /tmp/cannon/* /tmp/cannon_fault/*", {stdio: 'inherit'})
    child_process.execSync("mkdir -p /tmp/cannon/checkpoint /tmp/cannon/data", {stdio: 'inherit'})
    child_process.execSync("mkdir -p /tmp/cannon_fault/checkpoint /tmp/cannon_fault/data", {stdio: 'inherit'})
}

async function deployContract() {
    submitterConfig = getSubmitterConfig()
    // === client ===
    // 1. create the golden image
    // 2. deploy the contract
    submitterConfig.checkpoints = [0]
    submitterConfig.curPhase = 0
    deployInfo = await api.initiateOPMLRequest(submitterConfig)
    return deployInfo
}

function updateConfig(query) {
    api.config.modelName = query.modelName ? query.modelName : api.config.modelName
    api.config.prompt = query.prompt ? query.prompt : api.config.prompt
}

function getClientConfig() {
    config = api.getConfig()
    return config
}

function getSubmitterConfig() {
    const submitterDir = "/tmp/cannon"
    var submitterConfig = api.getConfig()
    submitterConfig.basedir = submitterDir
    return submitterConfig
}

function getChallengerConfig() {
    const challengerDir = "/tmp/cannon_fault"
    var challengerConfig = api.getConfig()
    challengerConfig.basedir = challengerDir
    return challengerConfig
}


module.exports = router;