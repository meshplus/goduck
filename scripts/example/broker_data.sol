pragma solidity >=0.6.9 <=0.7.6;
pragma experimental ABIEncoderV2;

contract BrokerData {
    struct Proposal {
        uint64 approve;
        uint64 reject;
        address[] votedAdmins;
        bool exist;
    }

    struct CallFunc {
        string func;
        bytes[] args;
    }

    struct InterchainInvoke {
        bool encrypt;
        string[] group;
        CallFunc callFunc;
        CallFunc callback;
        CallFunc rollback;
    }

    struct Receipt {
        bool encrypt;
        uint64 typ;
        bytes[][] results;
        bool[] multiStatus;
    }

    struct multiInvokeArgs {
        string contractAddr;
        CallFunc invokeFunc;
        bytes[] arg;
    }

    address[] bxhSigners;

    string[] outServicePairs;
    string[] inServicePairs;
    string[] callbackServicePairs;

    mapping(string => uint64) outCounter;
    mapping(string => uint64) callbackCounter;
    mapping(string => uint64) inCounter;
    mapping(string => uint64) dstRollbackCounter;

    mapping(string => mapping(uint64 => InterchainInvoke)) outMessages;
    mapping(string => mapping(uint64 => Receipt)) receiptMessages;

    mapping(address => Proposal) localProposal;

    address public BrokerAddr;
    address[] public admins;
    uint64 public adminThreshold;

    // Authority control. Only the administrator can audit the contract
    modifier onlyAdmin {
        bool flag = false;
        for (uint i = 0; i < admins.length; i++) {
            if (msg.sender == admins[i]) {flag = true;}
        }

        require(flag == true, "Invoker are not in admin list");
        _;
    }

    modifier onlyBroker {
        require(msg.sender == BrokerAddr, "Invoker are not the Broker");
        _;
    }

    constructor(address[] memory _admins,
        uint64 _adminThreshold) {
        admins = _admins;
        adminThreshold = _adminThreshold;
    }

    function initialize() public onlyBroker {
        for (uint i = 0; i < inServicePairs.length; i++) {
            inCounter[inServicePairs[i]] = 0;
        }
        for (uint j = 0; j < outServicePairs.length; j++) {
            outCounter[outServicePairs[j]] = 0;
        }
        for (uint k = 0; k < callbackServicePairs.length; k++) {
            callbackCounter[callbackServicePairs[k]] = 0;
        }
        for (uint m = 0; m < inServicePairs.length; m++) {
            dstRollbackCounter[inServicePairs[m]] = 0;
        }
        delete outServicePairs;
        delete inServicePairs;
        delete callbackServicePairs;
    }

    function setAdmins(address[] memory _admins, uint64 _adminThreshold) public onlyAdmin {
        admins = _admins;
        adminThreshold = _adminThreshold;
    }

    function register() public {
        require(tx.origin != msg.sender, "register not by contract");
        if (BrokerAddr == msg.sender || localProposal[msg.sender].exist) {
            return;
        }

        localProposal[msg.sender] = Proposal(0, 0, new address[](admins.length), true);
    }

    function audit(address addr, int64 status) public onlyAdmin returns (bool) {
        uint result = vote(addr, status);

        if (result == 0) {
            return false;
        }

        if (result == 1) {
            delete localProposal[addr];
            BrokerAddr = addr;
        } else {
            delete localProposal[addr];
        }

        return true;
    }

    // return value explain:
    // 0: vote is not finished
    // 1: approve the proposal
    // 2: reject the proposal
    function vote(address addr, int64 status) private returns (uint) {
        require(localProposal[addr].exist, "the proposal does not exist");
        require(status == 0 || status == 1, "vote status should be 0 or 1");

        for (uint i = 0; i < localProposal[addr].votedAdmins.length; i++) {
            require(localProposal[addr].votedAdmins[i] != msg.sender, "current use has voted the proposal");
        }

        localProposal[addr].votedAdmins[localProposal[addr].reject + localProposal[addr].approve] = msg.sender;
        if (status == 0) {
            localProposal[addr].reject++;
            if (localProposal[addr].reject == admins.length - adminThreshold + 1) {
                return 2;
            }
        } else {
            localProposal[addr].approve++;
            if (localProposal[addr].approve == adminThreshold) {
                return 1;
            }
        }

        return 0;
    }


    function invokeIndexUpdate(string memory srcFullID, string memory dstFullID, uint64 index, uint64 reqType) public onlyBroker returns(bool) {
        string memory servicePair = genServicePair(srcFullID, dstFullID);
        if (reqType == 0) {
            if (inCounter[servicePair] + 1 != index) {
                return false;
            }
            markInCounter(servicePair);
        } else if (reqType == 1) {
            // invoke src callback or rollback
            if (callbackCounter[servicePair] + 1 != index) {
                return false;
            }
            markCallbackCounter(servicePair, index);
        } else if (reqType == 2) {
            // invoke dst rollback
            if (dstRollbackCounter[servicePair] + 1 > index) {
                Receipt memory receipt = receiptMessages[servicePair][index];
                if (receipt.typ != 1) {
                    return false;
                }
            }
            markDstRollbackCounter(servicePair, index);
            if (inCounter[servicePair] + 1 == index) {
                markInCounter(servicePair);
            }
        }
        return true;
    }

    function genServicePair(string memory from, string memory to) private pure returns (string memory) {
        return string(abi.encodePacked(from, "-", to));
    }

    // The helper functions that help document Meta information.
    function markCallbackCounter(string memory servicePair, uint64 index) private {
        if (callbackCounter[servicePair] == 0) {
            callbackServicePairs.push(servicePair);
        }
        callbackCounter[servicePair] = index;
    }

    function markDstRollbackCounter(string memory servicePair, uint64 index) private {
        dstRollbackCounter[servicePair] = index;
    }

    function markInCounter(string memory servicePair) private {
        inCounter[servicePair]++;
        if (inCounter[servicePair] == 1) {
            inServicePairs.push(servicePair);
        }
    }

    function markOutCounter(string memory servicePair) public onlyBroker returns(uint64) {
        outCounter[servicePair]++;
        if (outCounter[servicePair] == 1) {
            outServicePairs.push(servicePair);
        }
        return outCounter[servicePair];
    }

    // The helper functions that help plugin query.
    function getOuterMeta() public view onlyBroker returns (string[] memory, uint64[] memory) {
        uint64[] memory indices = new uint64[](outServicePairs.length);
        for (uint64 i = 0; i < outServicePairs.length; i++) {
            indices[i] = outCounter[outServicePairs[i]];
        }

        return (outServicePairs, indices);
    }

    function getOutMessage(string memory outServicePair, uint64 idx) public view onlyBroker returns (string memory, bytes[] memory, bool, string[] memory) {
        InterchainInvoke memory invoke = outMessages[outServicePair][idx];
        return (invoke.callFunc.func, invoke.callFunc.args, invoke.encrypt, invoke.group);
    }

    function getReceiptMessage(string memory inServicePair, uint64 idx) public view onlyBroker returns (bytes[][] memory, uint64, bool, bool[] memory)  {
        Receipt memory receipt = receiptMessages[inServicePair][idx];
        return (receipt.results, receipt.typ, receipt.encrypt, receipt.multiStatus);
    }

    function getInnerMeta() public view onlyBroker returns (string[] memory, uint64[] memory) {
        uint64[] memory indices = new uint64[](inServicePairs.length);
        for (uint i = 0; i < inServicePairs.length; i++) {
            indices[i] = inCounter[inServicePairs[i]];
        }

        return (inServicePairs, indices);
    }

    function getCallbackMeta() public view onlyBroker returns (string[] memory, uint64[] memory) {
        uint64[] memory indices = new uint64[](callbackServicePairs.length);
        for (uint64 i = 0; i < callbackServicePairs.length; i++) {
            indices[i] = callbackCounter[callbackServicePairs[i]];
        }

        return (callbackServicePairs, indices);
    }

    function getDstRollbackMeta() public view onlyBroker returns (string[] memory, uint64[] memory) {
        uint64[] memory indices = new uint64[](inServicePairs.length);
        for (uint i = 0; i < inServicePairs.length; i++) {
            indices[i] = dstRollbackCounter[inServicePairs[i]];
        }

        return (inServicePairs, indices);
    }

    function getOutCounter(string memory servicePair) public view onlyBroker returns(uint64) {
        return outCounter[servicePair];
    }

    function getInCounter(string memory servicePair) public view onlyBroker returns(uint64) {
        return inCounter[servicePair];
    }

    function getCallbackCounter(string memory servicePair) public view onlyBroker returns(uint64) {
        return callbackCounter[servicePair];
    }

    function getCallbackMessage(string memory servicePair, uint64 index) public view onlyBroker returns(string memory, bytes[] memory) {
        InterchainInvoke memory invoke = outMessages[servicePair][index];
        return (invoke.callback.func, invoke.callback.args);
    }

    function getRollbackMessage(string memory servicePair, uint64 index) public view onlyBroker returns(string memory, bytes[] memory) {
        InterchainInvoke memory invoke = outMessages[servicePair][index];
        return (invoke.rollback.func, invoke.rollback.args);
    }

    function setReceiptMessage(string memory servicePair, uint64 index, bool isEncrypt, uint64 typ, bytes[][] memory results, bool[] memory multiStatus) public onlyBroker {
        receiptMessages[servicePair][index] = Receipt(isEncrypt, typ, results, multiStatus);
    }

    function setOutMessage(string memory servicePair,
        bool isEncrypt,
        string[] memory group,
        string memory funcCall,
        bytes[] memory args,
        string memory funcCb,
        bytes[] memory argsCb,
        string memory funcRb,
        bytes[] memory argsRb) public onlyBroker {
        outMessages[servicePair][outCounter[servicePair]] = InterchainInvoke(isEncrypt, group,
            CallFunc(funcCall, args),
            CallFunc(funcCb, argsCb),
            CallFunc(funcRb, argsRb));
    }

    // The helper functions that help check multisign.
    function checkInterchainMultiSigns(string memory srcFullID,
        string memory dstFullID,
        uint64 index,
        uint64 typ,
        string memory callFunc,
        bytes[] memory args,
        uint64 txStatus,
        bytes[] memory multiSignatures,
        address[] memory validators,
        uint64 valThreshold) public returns(bool) {
        bytes memory packed = abi.encodePacked(srcFullID, dstFullID, index, typ);
        bytes memory funcPacked = abi.encodePacked(callFunc);

        funcPacked = abi.encodePacked(funcPacked, uint64(0));
        for (uint i = 0; i < args.length; i++) {
            funcPacked = abi.encodePacked(funcPacked, args[i]);
        }
        packed = abi.encodePacked(packed, keccak256(funcPacked), txStatus);
        bytes32 hash = keccak256(packed);

        //require(checkMultiSigns(hash, multiSignatures), "invalid multi-signature");
        return checkMultiSigns(hash, multiSignatures, validators, valThreshold);
    }

    function checkMultiInterchainMultiSigns(string memory srcFullID,
        string memory dstFullID,
        uint64 index,
        uint64 typ,
        string memory callFunc,
        bytes[][] memory args,
        uint64 txStatus,
        bytes[] memory multiSignatures,
        address[] memory validators,
        uint64 valThreshold) public returns (bool) {
        bytes memory packed = abi.encodePacked(srcFullID, dstFullID, index, typ);
        bytes memory funcPacked = abi.encodePacked(callFunc);
        funcPacked = abi.encodePacked(funcPacked, uint64(1));
        if (args.length == 0) {
            funcPacked = abi.encodePacked(funcPacked, uint64(0));
        } else {
            funcPacked = abi.encodePacked(funcPacked, uint64(args[0].length));
        }
        for (uint i = 0; i < args.length; i++) {
            bytes[] memory arg = args[i];
            for (uint j = 0; j < arg.length; j++) {
                funcPacked = abi.encodePacked(funcPacked, arg[j]);
            }
        }
        packed = abi.encodePacked(packed, keccak256(funcPacked), txStatus);
        bytes32 hash = keccak256(packed);

        //        require(checkMultiSigns(hash, multiSignatures), "invalid MultiInterchain-multi-signature");
        return checkMultiSigns(hash, multiSignatures, validators, valThreshold);
    }

    function checkReceiptMultiSigns(string memory srcFullID,
        string memory dstFullID,
        uint64 index,
        uint64 typ,
        bytes[][] memory results,
        uint64 txStatus,
        bytes[] memory multiSignatures,
        address[] memory validators,
        uint64 valThreshold) public returns(bool) {
        bytes memory packed = abi.encodePacked(srcFullID, dstFullID, index, typ);
        bytes memory data;
        if (typ == 0) {
            string memory outServicePair = genServicePair(srcFullID, dstFullID);
            CallFunc memory callFunc = outMessages[outServicePair][index].callFunc;
            data = abi.encodePacked(data, callFunc.func);
            for (uint i = 0; i < callFunc.args.length; i++) {
                data = abi.encodePacked(data, callFunc.args[i]);
            }
        } else {
            for (uint i = 0; i < results.length; i++) {
                bytes[] memory result = results[i];
                for (uint j = 0; j < result.length; i++) {
                    data = abi.encodePacked(data, result[j]);
                }
            }
        }
        packed = abi.encodePacked(packed, keccak256(data), txStatus);
        bytes32 hash = keccak256(packed);

        return checkMultiSigns(hash, multiSignatures, validators, valThreshold);
    }

    function checkMultiSigns(bytes32 hash, bytes[] memory multiSignatures, address[] memory validators, uint64 valThreshold) private returns (bool) {
        for (uint i = 0; i < multiSignatures.length; i++) {
            bytes memory sig = multiSignatures[i];
            if (sig.length != 65) {
                continue;
            }

            (uint8 v, bytes32 r, bytes32 s) = splitSignature(sig);

            address addr = ecrecover(hash, v, r, s);

            if (addressArrayContains(validators, addr)) {
                if (addressArrayContains(bxhSigners, addr)) {
                    continue;
                }
                bxhSigners.push(addr);
                if (bxhSigners.length == valThreshold) {
                    delete bxhSigners;
                    return true;
                }
            }
        }
        delete bxhSigners;
        return false;
    }

    function addressArrayContains(address[] memory addrs, address addr) private pure returns (bool) {
        for (uint i = 0; i < addrs.length; i++) {
            if (addrs[i] == addr) {
                return true;
            }
        }

        return false;
    }

    function splitSignature(bytes memory sig) internal pure returns (uint8 v, bytes32 r, bytes32 s) {
        assembly {
        // first 32 bytes, after the length prefix
            r := mload(add(sig, 32))
        // second 32 bytes
            s := mload(add(sig, 64))
        // final byte (first byte of the next 32 bytes)
            v := byte(0, mload(add(sig, 96)))
        }

        return (v + 27, r, s);
    }

    function _getAsciiOffset(
        uint8 nibble, bool caps
    ) internal pure returns (uint8 offset) {
        // to convert to ascii characters, add 48 to 0-9, 55 to A-F, & 87 to a-f.
        if (nibble < 10) {
            offset = 48;
        } else if (caps) {
            offset = 55;
        } else {
            offset = 87;
        }
    }

    function addressToString(
        address account
    ) public pure returns (string memory asciiString) {
        // convert the account argument from address to bytes.
        bytes20 data = bytes20(account);

        // create an in-memory fixed-size bytes array.
        bytes memory asciiBytes = new bytes(40);

        // declare variable types.
        uint8 b;
        uint8 leftNibble;
        uint8 rightNibble;
        bool leftCaps;
        bool rightCaps;
        uint8 asciiOffset;

        // get the capitalized characters in the actual checksum.
        bool[40] memory caps = _toChecksumCapsFlags(account);

        // iterate over bytes, processing left and right nibble in each iteration.
        for (uint256 i = 0; i < data.length; i++) {
            // locate the byte and extract each nibble.
            b = uint8(uint160(data) / (2 ** (8 * (19 - i))));
            leftNibble = b / 16;
            rightNibble = b - 16 * leftNibble;

            // locate and extract each capitalization status.
            leftCaps = caps[2 * i];
            rightCaps = caps[2 * i + 1];

            // get the offset from nibble value to ascii character for left nibble.
            asciiOffset = _getAsciiOffset(leftNibble, leftCaps);

            // add the converted character to the byte array.
            asciiBytes[2 * i] = bytes1(leftNibble + asciiOffset);

            // get the offset from nibble value to ascii character for right nibble.
            asciiOffset = _getAsciiOffset(rightNibble, rightCaps);

            // add the converted character to the byte array.
            asciiBytes[2 * i + 1] = bytes1(rightNibble + asciiOffset);
        }


        return string(abi.encodePacked("0x", asciiBytes));
    }

    function _toChecksumCapsFlags(address account) internal pure returns (
        bool[40] memory characterCapitalized
    ) {
        // convert the address to bytes.
        bytes20 a = bytes20(account);

        // hash the address (used to calculate checksum).
        bytes32 b = keccak256(abi.encodePacked(_toAsciiString(a)));

        // declare variable types.
        uint8 leftNibbleAddress;
        uint8 rightNibbleAddress;
        uint8 leftNibbleHash;
        uint8 rightNibbleHash;

        // iterate over bytes, processing left and right nibble in each iteration.
        for (uint256 i; i < a.length; i++) {
            // locate the byte and extract each nibble for the address and the hash.
            rightNibbleAddress = uint8(a[i]) % 16;
            leftNibbleAddress = (uint8(a[i]) - rightNibbleAddress) / 16;
            rightNibbleHash = uint8(b[i]) % 16;
            leftNibbleHash = (uint8(b[i]) - rightNibbleHash) / 16;

            characterCapitalized[2 * i] = (
            leftNibbleAddress > 9 &&
            leftNibbleHash > 7
            );
            characterCapitalized[2 * i + 1] = (
            rightNibbleAddress > 9 &&
            rightNibbleHash > 7
            );
        }
    }

    // based on https://ethereum.stackexchange.com/a/56499/48410
    function _toAsciiString(
        bytes20 data
    ) internal pure returns (string memory asciiString) {
        // create an in-memory fixed-size bytes array.
        bytes memory asciiBytes = new bytes(40);

        // declare variable types.
        uint8 b;
        uint8 leftNibble;
        uint8 rightNibble;

        // iterate over bytes, processing left and right nibble in each iteration.
        for (uint256 i = 0; i < data.length; i++) {
            // locate the byte and extract each nibble.
            b = uint8(uint160(data) / (2 ** (8 * (19 - i))));
            leftNibble = b / 16;
            rightNibble = b - 16 * leftNibble;

            // to convert to ascii characters, add 48 to 0-9 and 87 to a-f.
            asciiBytes[2 * i] = bytes1(leftNibble + (leftNibble < 10 ? 48 : 87));
            asciiBytes[2 * i + 1] = bytes1(rightNibble + (rightNibble < 10 ? 48 : 87));
        }

        return string(asciiBytes);
    }

    function stringToAddress(string memory _address) public pure returns (address) {
        bytes memory temp = bytes(_address);
        if(temp.length != 42) {
            revert(string(abi.encodePacked(_address, " is not a valid address")));
        }

        uint160 result = 0;
        uint160 b1;
        uint160 b2;
        for (uint256 i = 2; i < 2 + 2 * 20; i += 2) {
            result *= 256;
            b1 = uint160(uint8(temp[i]));
            b2 = uint160(uint8(temp[i + 1]));
            if ((b1 >= 97) && (b1 <= 102)) {
                b1 -= 87;
            } else if ((b1 >= 65) && (b1 <= 70)) {
                b1 -= 55;
            } else if ((b1 >= 48) && (b1 <= 57)) {
                b1 -= 48;
            }

            if ((b2 >= 97) && (b2 <= 102)) {
                b2 -= 87;
            } else if ((b2 >= 65) && (b2 <= 70)) {
                b2 -= 55;
            } else if ((b2 >= 48) && (b2 <= 57)) {
                b2 -= 48;
            }
            result += (b1 * 16 + b2);
        }
        return address(result);
    }

    function checkAppchainIdContains (string memory appchainId, string memory destFullService) public pure returns(bool) {
        bytes memory whatBytes = bytes (appchainId);
        bytes memory whereBytes = bytes (destFullService);

        if (whereBytes.length >= whatBytes.length) {
            return false;
        }

        bool found = false;
        for (uint i = 0; i <= whereBytes.length - whatBytes.length; i++) {
            bool flag = true;
            for (uint j = 0; j < whatBytes.length; j++)
                if (whereBytes [i + j] != whatBytes [j]) {
                    flag = false;
                    break;
                }
            if (flag) {
                found = true;
                break;
            }
        }
        return found;
    }
}