import argparse

import hashlib

def parseArg(): 
    parser = argparse.ArgumentParser()

    parser.add_argument(
        "--mp_checkpoints",
        type=str,
        nargs="?",
        default="[]",
        help="checkpoints, e.g. [1,2]"
    )
    parser.add_argument(
        "--mp_stepCount",
        type=str,
        nargs="?",
        default="[]",
        help="stepCount: [1,2]"
    )
    parser.add_argument(
        "--mp_execOutputPath",
        type=str,
        nargs="?",
        default="/tmp/cannon/output.tmp",
        help="path for saving the output"
    )

    opt = parser.parse_args()
    return opt


import json
class MPJTree():
    def __init__(self, root, checkpoints=[], stepCount=[], preimages=dict()) -> None:
        self.root = root
        self.checkpoints = checkpoints
        self.stepCount = stepCount
        self.preimages = preimages

    def toJSONFile(self, fn):
        data = {
            "root": self.root,
            "checkpoints": self.checkpoints,
            "stepCount": self.stepCount,
            "preimages": self.preimages
        }
        with open(fn, "w") as f:
            json.dump(data, f) 
            print("json file is saved at ", fn)       

def getRoot(checkpoints):
    data = {
        "checkpoints": checkpoints,
    }
    root = "0x" + hashlib.sha256(json.dumps(data).encode()).hexdigest()
    return root

def main():
    args = parseArg()
    # root = hash(0x1cc).to_bytes(32, "big").hex()
    # root = "0x" + hashlib.sha256(b"0x1cc").hexdigest()
    root = getRoot(args.mp_checkpoints)
    # mock for stepCount
    for i in range(len(args.mp_stepCount)):
        if (args.mp_stepCount[i] == 0): 
            args.mp_stepCount[i] = (i + 1) * 10
    
    data = MPJTree(root=root, checkpoints=args.mp_checkpoints, stepCount=args.mp_stepCount)
    data.toJSONFile(args.mp_execOutputPath)

if __name__ == "__main__":
    main()
