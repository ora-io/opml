// SPDX-License-Identifier: MIT
pragma solidity ^0.7.3;
pragma experimental ABIEncoderV2;

import "./lib/Lib_RLPReader.sol";

/// @notice MIPS virtual machine interface
interface IMIPS {
  /// @notice Given a MIPS state hash (includes code & registers), execute the next instruction and returns
  ///         the update state hash.
  function Step(bytes32 stateHash) external returns (bytes32);

  /// @notice Returns the associated MIPS memory contract.
  function m() external pure returns (IMIPSMemory);
}

/// @notice MIPS memory (really "state", including registers and memory-mapped I/O)
interface IMIPSMemory {
  /// @notice Adds a `(hash(anything) => anything)` entry to the mapping that underpins all the
  ///         Merkle tries that this contract deals with (where "state hash" = Merkle root of such
  ///         a trie).
  /// @param anything node data to add to the trie
  function AddTrieNode(bytes calldata anything) external;

  function ReadMemory(bytes32 stateHash, uint32 addr) external view returns (uint32);
  function ReadBytes32(bytes32 stateHash, uint32 addr) external view returns (bytes32);
  function ReadMemoryToBytes(bytes32 stateHash, uint32 addr) external view returns (bytes memory);

  /// @notice Write 32 bits at the given address and returns the updated state hash.
  function WriteMemory(bytes32 stateHash, uint32 addr, uint32 val) external returns (bytes32);

  /// @notice Write 32 bytes at the given address and returns the updated state hash.
  function WriteBytes32(bytes32 stateHash, uint32 addr, bytes32 val) external returns (bytes32);
}

/// @notice Implementation of the challenge game, which allows a challenger to challenge an L1 block
///         by asserting a different state root for the transition implied by the block's
///         transactions. The challenger plays against a defender (the owner of this contract),
///         which we assume acts honestly. The challenger and the defender perform a binary search
///         over the execution trace of the fault proof program (in this case minigeth), in order
///         to determine a single execution step that they disagree on, at which point that step
///         can be executed on-chain in order to determine if the challenge is valid.
contract MPChallenge {
  address payable immutable owner;

  IMIPS immutable mips;
  IMIPSMemory immutable mem;

  /// @notice State hash of the fault proof program's initial MIPS state.
  bytes32 public immutable globalStartState;

  constructor(IMIPS _mips, bytes32 _globalStartState) {
    owner = msg.sender;
    mips = _mips;
    mem = _mips.m();
    globalStartState = _globalStartState;
  }

  struct ChallengeData {
    // Left bound of the binary search of i-th layer: challenger & defender agree on all steps <= L[i].
    mapping(uint256 => uint256) L;
    // Right bound of the binary search of i-th layer: challenger & defender disagree on all steps >= R[i].
    mapping(uint256 => uint256) R;
    // Maps step numbers to asserted state hashes for the challenger.
    mapping(uint256 => mapping(uint256 => bytes32)) assertedState; 
    // Maps step numbers to asserted state hashes for the defender.
    mapping(uint256 => mapping(uint256 => bytes32)) defendedState;
    // Address of the challenger.
    address payable challenger;
    // Current challenge's layer
    uint256 currentLayer;
    // Number of total layers
    uint256 totalLayer;
    // nodeID
    uint256 nodeID;
  }

  /// @notice ID if the last created challenged, incremented for new challenge IDs.
  uint256 public lastChallengeId = 0;

  /// @notice Maps challenge IDs to challenge data.
  mapping(uint256 => ChallengeData) public challenges;

  /// @notice Emitted when a new challenge is created.
  event ChallengeCreated(uint256 challengeId);


  /// @notice proposer's results
  bytes public proposedResults;

  /// @notice challenger's results
  bytes public challengerResults;

  /// @notice Proposer should first upload the results and stake some money, waiting for the challenge
  ///         process. Note that the results can only be set once (TODO)
  function uploadResult(bytes calldata data) public {
    require(data.length % 32 == 0, "the result should 32-align");
    proposedResults = data;
  }


  /// @notice Challenges the pure computation without accessing to the blockchain data
  ///         Before calling this, it is necessary to have loaded all the trie node necessary to
  ///         write the input hash in the Merkleized initial MIPS state, and to read the output hash
  ///         and machine state from the Merkleized final MIPS state (i.e. `finalSystemState`). Use
  ///         `MIPSMemory.AddTrieNode` for this purpose. Use `callWithTrieNodes` to figure out
  ///         which nodes you need.
  /// @param finalSystemState The state hash of the fault proof program's final MIPS state.
  /// @param stepCount The number of steps (MIPS instructions) taken to execute the fault proof
  ///        program.
  /// @return The challenge identifier
  function initiatePureComputationChallenge(
      bytes32 finalSystemState, uint256 stepCount, uint256 totalLayer)
    external
    returns (uint256)
  {
    // Write input hash at predefined memory address.
    bytes32 startState = globalStartState;

    // Confirm that `finalSystemState` asserts the state you claim and that the machine is stopped.
    // require(mem.ReadMemory(finalSystemState, 0xC0000080) == 0x5EAD0000,
    //     "the final MIPS machine state is not stopped (PC != 0x5EAD0000)");
    
    // maybe we do not need that, since it is binded with evm smart contract?
    // require(mem.ReadMemory(finalSystemState, 0x30000800) == 0x1337f00d,
    //     "the final state root has not been written a the predefined MIPS memory location");

    // the challenger should upload his results on chain
    // for a valid challenge, the challenge results != proposer results!
    // bytes memory result = mem.ReadMemoryToBytes(finalSystemState, 0x32000000);
    // require(keccak256(result) != keccak256(proposedResults), "the challenger's results should be different from the proposed results");
    // challengerResults = result;

    uint256 challengeId = lastChallengeId++;
    ChallengeData storage c = challenges[challengeId];

    // A NEW CHALLENGER APPEARS
    c.challenger = msg.sender;
    // c.blockNumberN = blockNumberN; // no need to set the blockNumber
    c.assertedState[0][0] = startState;
    c.defendedState[0][0] = startState;
    c.assertedState[0][stepCount] = finalSystemState;
    c.totalLayer = totalLayer;
    c.currentLayer = 0;
    c.L[0] = 0;
    c.R[0] = stepCount;
    c.nodeID = 0;

    emit ChallengeCreated(challengeId);
    return challengeId;
  }

  function toNextLayer(uint256 challengeId, bytes32 startState, bytes32 finalState, uint256 stepCount) external returns (uint256) {
    ChallengeData storage c = challenges[challengeId];
    require(c.challenger != address(0), "invalid challenge");
    require(c.L[c.currentLayer] + 1 == c.R[c.currentLayer], "the current layer should end");
    require(c.currentLayer < c.totalLayer - 1, "the current layer is the last layer");
    c.nodeID = c.L[c.currentLayer];
    c.currentLayer += 1;
    c.assertedState[c.currentLayer][0] = startState;
    c.defendedState[c.currentLayer][0] = startState;
    c.assertedState[c.currentLayer][stepCount] = finalState;
    c.L[c.currentLayer] = 0;
    c.R[c.currentLayer] = stepCount;
    return c.currentLayer;
  }


  /// @notice Calling `initiateChallenge`, `confirmStateTransition` or `denyStateTransition requires
  ///         some trie nodes to have been supplied beforehand (see these functions for details).
  ///         This function can be used to figure out which nodes are needed, as memory-accessing
  ///         functions in MIPSMemory.sol will revert with the missing node ID when a node is
  ///         missing. Therefore, you can call this function repeatedly via `eth_call`, and
  ///         iteratively build the list of required node until the call succeeds.
  /// @param target The contract to call to (usually this contract)
  /// @param dat The data to include in the call (usually the calldata for a call to
  ///        one of the aforementionned functions)
  /// @param nodes The nodes to add the MIPS state trie before making the call
  function callWithTrieNodes(address target, bytes calldata dat, bytes[] calldata nodes) public {
    for (uint i = 0; i < nodes.length; i++) {
      mem.AddTrieNode(nodes[i]);
    }
    (bool success, bytes memory revertData) = target.call(dat);
    if (!success) {
      uint256 revertDataLength = revertData.length;
      assembly {
        let revertDataStart := add(revertData, 32)
        revert(revertDataStart, revertDataLength)
      }
    }
  }

  /// @notice Indicates whether the given challenge is still searching (true), or if the single step
  ///         of disagreement has been found (false).
  function isSearching(uint256 challengeId) view public returns (bool) {
    ChallengeData storage c = challenges[challengeId];
    require(c.challenger != address(0), "invalid challenge");
    return c.L[c.currentLayer] + 1 != c.R[c.currentLayer];
  }

  /// @notice Returns the next step number where the challenger and the defender must compared
  ///         state hash, namely the midpoint between the current left and right bounds of the
  ///         binary search.
  function getStepNumber(uint256 challengeId) view public returns (uint256) {
    ChallengeData storage c = challenges[challengeId];
    require(c.challenger != address(0), "invalid challenge");
    return (c.L[c.currentLayer]+c.R[c.currentLayer])/2;
  }

  function getNodeID(uint256 challengeId) view public returns (uint256) {
    ChallengeData storage c = challenges[challengeId];
    require(c.challenger != address(0), "invalid challenge");
    return c.nodeID;
  }

  function getCurrentLayer(uint256 challengeId) view public returns (uint256) {
    ChallengeData storage c = challenges[challengeId];
    require(c.challenger != address(0), "invalid challenge");
    return c.currentLayer;    
  }

  /// @notice Returns the last state hash proposed by the challenger during the binary search.
  function getProposedState(uint256 challengeId) view public returns (bytes32) {
    ChallengeData storage c = challenges[challengeId];
    require(c.challenger != address(0), "invalid challenge");
    uint256 stepNumber = getStepNumber(challengeId);
    return c.assertedState[c.currentLayer][stepNumber];
  }

  /// @notice The challenger can call this function to submit the state hash for the next step
  ///         in the binary search (cf. `getStepNumber`).
  function proposeState(uint256 challengeId, bytes32 stateHash) external {
    ChallengeData storage c = challenges[challengeId];
    require(c.challenger != address(0), "invalid challenge");
    require(c.challenger == msg.sender, "must be challenger");
    require(isSearching(challengeId), "must be searching");

    uint256 stepNumber = getStepNumber(challengeId);
    require(c.assertedState[c.currentLayer][stepNumber] == bytes32(0), "state already proposed");
    c.assertedState[c.currentLayer][stepNumber] = stateHash;
  }

  /// @notice The defender can call this function to submit the state hash for the next step
  ///         in the binary search (cf. `getStepNumber`). He can only do this after the challenger
  ///         has submitted his own state hash for this step.
  ///         If the defender believes there are less steps in the execution of the fault proof
  ///         program than the current step number, he should submit the final state hash.
  function respondState(uint256 challengeId, bytes32 stateHash) external {
    ChallengeData storage c = challenges[challengeId];
    require(c.challenger != address(0), "invalid challenge");
    require(owner == msg.sender, "must be owner");
    require(isSearching(challengeId), "must be searching");

    uint256 stepNumber = getStepNumber(challengeId);
    require(c.assertedState[c.currentLayer][stepNumber] != bytes32(0), "challenger state not proposed");
    require(c.defendedState[c.currentLayer][stepNumber] == bytes32(0), "state already proposed");

    // Technically, we don't have to save these states, but we have to if we want to let the
    // defender terminate the proof early (and not via a timeout) after the binary search completes.
    c.defendedState[c.currentLayer][stepNumber] = stateHash;

    // update binary search bounds
    if (c.assertedState[c.currentLayer][stepNumber] == c.defendedState[c.currentLayer][stepNumber]) {
      c.L[c.currentLayer] = stepNumber; // agree
    } else {
      c.R[c.currentLayer] = stepNumber; // disagree
    }
  }

  /// @notice Emitted when the challenger can provably be shown to be correct about his assertion.
  event ChallengerWins(uint256 challengeId);

  /// @notice Emitted when the challenger can provably be shown to be wrong about his assertion.
  event ChallengerLoses(uint256 challengeId);

  /// @notice Emitted when the challenger should lose if he does not generate a `ChallengerWins`
  ///         event in a timely manner (TBD). This occurs in a specific scenario when we can't
  ///         explicitly verify that the defender is right (cf. `denyStateTransition).
  event ChallengerLosesByDefault(uint256 challengeId);

  /// @notice Anybody can call this function to confirm that the single execution step that the
  ///         challenger and defender disagree on does indeed yield the result asserted by the
  ///         challenger, leading to him winning the challenge.
  ///         Before calling this function, you need to add trie nodes so that the MIPS state can be
  ///         read/written by the single step execution. Use `MIPSMemory.AddTrieNode` for this
  ///         purpose. Use `callWithTrieNodes` to figure out which nodes you need.
  ///         You will also need to supply any preimage that the step tries to access with
  ///         `MIPSMemory.AddPreimage`. See `scripts/assert.js` for details on how this can be
  ///         done.
  function confirmStateTransition(uint256 challengeId) external {
    ChallengeData storage c = challenges[challengeId];
    require(c.challenger != address(0), "invalid challenge");
    require(c.currentLayer == c.totalLayer - 1, "do not reach the last layer");
    require(!isSearching(challengeId), "binary search not finished");

    bytes32 stepState = mips.Step(c.assertedState[c.currentLayer][c.L[c.currentLayer]]);
    require(stepState == c.assertedState[c.currentLayer][c.R[c.currentLayer]], "wrong asserted state for challenger");

    // pay out bounty!!
    (bool sent, ) = c.challenger.call{value: address(this).balance}("");
    require(sent, "Failed to send Ether");

    emit ChallengerWins(challengeId);
  }

  /// @notice Anybody can call this function to confirm that the single execution step that the
  ///         challenger and defender disagree on does indeed yield the result asserted by the
  ///         defender, leading to the challenger losing the challenge.
  ///         Before calling this function, you need to add trie nodes so that the MIPS state can be
  ///         read/written by the single step execution. Use `MIPSMemory.AddTrieNode` for this
  ///         purpose. Use `callWithTrieNodes` to figure out which nodes you need.
  ///         You will also need to supply any preimage that the step tries to access with
  ///         `MIPSMemory.AddPreimage`. See `scripts/assert.js` for details on how this can be
  ///         done.
  function denyStateTransition(uint256 challengeId) external {
    ChallengeData storage c = challenges[challengeId];
    require(c.challenger != address(0), "invalid challenge");
    require(c.currentLayer == c.totalLayer - 1, "do not reach the last layer");
    require(!isSearching(challengeId), "binary search not finished");

    // We run this before the next check so that if executing the final step somehow
    // causes a revert, then at least we do not emit `ChallengerLosesByDefault` when we know that
    // the challenger can't win (even if right) because of the revert.
    bytes32 stepState = mips.Step(c.defendedState[c.currentLayer][c.L[c.currentLayer]]);

    // If the challenger always agrees with the defender during the search, we end up with:
    // c.L + 1 == c.R == stepCount (from `initiateChallenge`)
    // In this case, the defender didn't assert his state hash for c.R, which makes
    // `c.defendedState[c.R]` zero. This means we can't verify that the defender right about the
    // final execution step.
    // The solution is to emit `ChallengerLosesByDefault` to signify the challenger should lose
    // if he can't emit `ChallengerWins` in a timely manner.
    if (c.defendedState[c.currentLayer][c.R[c.currentLayer]] == bytes32(0)) {
      emit ChallengerLosesByDefault(challengeId);
      return;
    }

    require(stepState == c.defendedState[c.currentLayer][c.R[c.currentLayer]], "wrong asserted state for defender");

    // consider the challenger mocked
    emit ChallengerLoses(challengeId);
  }

  /// @notice Allow sending money to the contract (without calldata).
  receive() external payable {}

  /// @notice Allows the owner to withdraw funds from the contract.
  function withdraw() external {
    require(msg.sender == owner);
    owner.transfer(address(this).balance);
  }
}
