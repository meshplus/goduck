pragma solidity >=0.6.9 <=0.7.6;
pragma experimental ABIEncoderV2;

contract Broker {
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
        CallFunc callFunc;
        CallFunc callback;
        CallFunc rollback;
    }

    struct Receipt {
        bool encrypt;
        uint64 typ;
        bytes[] result;
    }

    // Only the contract in the whitelist can invoke the Broker for interchain operations.
    mapping(address => bool) localWhiteList;
    address[] localServices;
    mapping(address => Proposal) localServiceProposal;
    address[] proposalList;

    // transaction management contract in direct mode
    address directTransactionAddr;

    string bitxhubID;
    string appchainID;
    address[] validators;
    uint64 valThreshold;
    address[] admins;
    uint64 adminThreshold;

    address[] bxhSigners;

    event throwInterchainEvent(uint64 index, string dstFullID, string srcFullID, string func, bytes[] args, bytes32 hash);
    event throwReceiptEvent(uint64 index, string dstFullID, string srcFullID, uint64 typ, bool status, bytes[] result, bytes32 hash);
    event throwReceiptStatus(bool);

    string[] outServicePairs;
    string[] inServicePairs;
    string[] callbackServicePairs;

    mapping(string => uint64) outCounter;
    mapping(string => uint64) callbackCounter;
    mapping(string => uint64) inCounter;
    mapping(string => uint64) dstRollbackCounter;

    mapping(string => mapping(uint64 => InterchainInvoke)) outMessages;
    mapping(string => mapping(uint64 => Receipt)) receiptMessages;

    // Authority control. Contracts need to be registered.
    modifier onlyWhiteList {
        require(localWhiteList[msg.sender] == true, "Invoker are not in white list");
        _;
    }

    // Authority control. Only the administrator can audit the contract
    modifier onlyAdmin {
        bool flag = false;
        for (uint i = 0; i < admins.length; i++) {
            if (msg.sender == admins[i]) {flag = true;}
        }

        require(flag == true, "Invoker are not in admin list");
        _;
    }

    constructor(string memory _bitxhubID,
        string memory _appchainID,
        address[] memory _validators,
        uint64 _valThreshold,
        address[] memory _admins,
        uint64 _adminThreshold) public {
        bitxhubID = _bitxhubID;
        appchainID = _appchainID;
        validators = _validators;
        valThreshold = _valThreshold;
        admins = _admins;
        adminThreshold = _adminThreshold;
    }


    function initialize() public {
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
        for (uint n = 0; n < localServices.length; n++) {
            localWhiteList[localServices[n]] = false;
        }
        for (uint x = 0; x < proposalList.length; x++) {
            delete localServiceProposal[proposalList[x]];
        }
        delete outServicePairs;
        delete inServicePairs;
        delete callbackServicePairs;
        delete localServices;
        if (valThreshold == 0) {
            // direct mode
            Transaction(directTransactionAddr).initialize();
        }
    }

    // register transaction management contract address in direct mode
    // invoke by transaction management contract
    function registerDirectTransaction() public {
        require(tx.origin != msg.sender, "register not by contract");
        directTransactionAddr = msg.sender;
    }

    // register local service to Broker
    function register() public {
        require(tx.origin != msg.sender, "register not by contract");
        if (localWhiteList[msg.sender] || localServiceProposal[msg.sender].exist) {
            return;
        }

        localServiceProposal[msg.sender] = Proposal(0, 0, new address[](admins.length), true);
    }

    function audit(address addr, int64 status) public onlyAdmin returns (bool) {
        uint result = vote(addr, status);

        if (result == 0) {
            return false;
        }

        if (result == 1) {
            delete localServiceProposal[addr];
            localWhiteList[addr] = true;
            localServices.push(addr);
        } else {
            delete localServiceProposal[addr];
        }

        return true;
    }

    // return value explain:
    // 0: vote is not finished
    // 1: approve the proposal
    // 2: reject the proposal
    function vote(address addr, int64 status) private returns (uint) {
        require(localServiceProposal[addr].exist, "the proposal does not exist");
        require(status == 0 || status == 1, "vote status should be 0 or 1");

        for (uint i = 0; i < localServiceProposal[addr].votedAdmins.length; i++) {
            require(localServiceProposal[addr].votedAdmins[i] != msg.sender, "current use has voted the proposal");
        }

        localServiceProposal[addr].votedAdmins[localServiceProposal[addr].reject + localServiceProposal[addr].approve] = msg.sender;
        if (status == 0) {
            localServiceProposal[addr].reject++;
            if (localServiceProposal[addr].reject == admins.length - adminThreshold + 1) {
                return 2;
            }
        } else {
            localServiceProposal[addr].approve++;
            if (localServiceProposal[addr].approve == adminThreshold) {
                return 1;
            }
        }

        return 0;
    }

    // register remote appchain ID in direct mode, invoked by appchain admin
    function registerAppchain(string memory chainID, string memory broker, address ruleAddr, bytes memory trustRoot) public onlyAdmin {
        Transaction(directTransactionAddr).registerAppchain(chainID, broker, ruleAddr, trustRoot);
    }

    // register service ID from counterparty appchain in direct mode, invoked by appchain admin
    // serviceID: the service from counterparty appchain which will call service on current appchain
    // whiteList：service list on current appchain which are allowed to be called by remote service
    function registerRemoteService(string memory chainID, string memory serviceID, address[] memory whiteList) public onlyAdmin {
        Transaction(directTransactionAddr).registerRemoteService(chainID, serviceID, whiteList);
    }

    function getAppchainInfo(string memory chainID) public view returns (string memory, bytes memory, address) {
        return Transaction(directTransactionAddr).getAppchainInfo(chainID);
    }

    // get the registered local service list
    function getLocalServiceList() public view returns (string[] memory) {
        string[] memory fullServiceIDList = new string[](localServices.length);
        for (uint i = 0; i < localServices.length; i++) {
            fullServiceIDList[i] = genFullServiceID(addressToString(localServices[i]));
        }

        return fullServiceIDList;
    }

    // get the registered counterparty service list
    function getRemoteServiceList() public view returns (string[] memory) {
        return Transaction(directTransactionAddr).getRemoteServiceList();
    }

    // get the registered counterparty service list
    function getRSWhiteList(string memory remoteAddr) public view returns (address[] memory) {
        return Transaction(directTransactionAddr).getRSWhiteList(remoteAddr);
    }

    // get the registered counterparty service list
    function getLocalWhiteList(address addr) public view returns (bool) {
        return localWhiteList[addr];
    }


    // called on dest chain
    function invokeInterchain(
        string memory srcFullID,
    // 地址变为string格式，这样多签不会有问题，在验证多签之前使用checksum之前的合约地址
        string memory destAddr,
        uint64 index,
        uint64 typ,
        string memory callFunc,
        bytes[] memory args,
        uint64 txStatus,
        bytes[] memory signatures,
        bool isEncrypt) payable external {
        // bool isRollback = false;
        string memory dstFullID = genFullServiceID(destAddr);
        string memory servicePair = genServicePair(srcFullID, dstFullID);

        checkInterchainMultiSigns(srcFullID, dstFullID, index, typ, callFunc, args, txStatus, signatures);
        //address _destAddr = stringToAddress(destAddr);
        bool status = true;
        bytes[] memory result;
        if (txStatus == 0) {
            // INTERCHAIN && BEGIN
            checkService(srcFullID, stringToAddress(destAddr));

            if (inCounter[servicePair] < index) {
                (status, result) = callService(stringToAddress(destAddr), callFunc, args, false);
            }
            invokeIndexUpdate(srcFullID, dstFullID, index, 0);
            if (status) {
                typ = 1;
            } else {
                typ = 2;
            }
        } else {
            // INTERCHAIN && FAILURE || INTERCHAIN && ROLLBACK, only happened in relay mode
            if (inCounter[servicePair] >= index) {
                checkService(srcFullID, stringToAddress(destAddr));
                (status, result) = callService(stringToAddress(destAddr), callFunc, args, true);
            }
            invokeIndexUpdate(srcFullID, dstFullID, index, 2);
            if (validators.length == 0) {
                // direct mode
                // ROLLBACK -> ROLLBACK_END
                typ = 4;
            } else {
                // relay mode
                if (txStatus == 1) {
                    typ = 2;
                } else {
                    typ = 3;
                }
            }
        }

        receiptMessages[servicePair][index] = Receipt(isEncrypt, typ, result);

        if (isEncrypt) {
            emit throwReceiptEvent(index, dstFullID, srcFullID, typ, status, new bytes[](0), computeHash(result));
        } else {
            emit throwReceiptEvent(index, dstFullID, srcFullID, typ, status, result, computeHash(result));
        }
    }

    function callService(address destAddr, string memory callFunc, bytes[] memory args, bool isRollback) private returns (bool, bytes[] memory) {
        bool status = true;
        bytes[] memory result;

        if (keccak256(abi.encodePacked(callFunc)) != keccak256(abi.encodePacked(""))) {
            (bool ok, bytes memory data) = address(destAddr).call(abi.encodeWithSignature(string(abi.encodePacked(callFunc, "(bytes[],bool)")), args, isRollback));
            status = ok;
            result = abi.decode(data, (bytes[]));
        }

        return (status, result);
    }

    function computeHash(bytes[] memory args) internal pure returns (bytes32) {
        bytes memory packed;
        for (uint i = 0; i < args.length; i++) {
            packed = abi.encodePacked(packed, args[i]);
        }

        return keccak256(packed);
    }

    // called on src chain
    function invokeReceipt(
        string memory srcAddr,
        string memory dstFullID,
        uint64 index,
        uint64 typ,
        bytes[] memory result,
        uint64 txStatus,
        bytes[] memory signatures) payable external {
        string memory srcFullID = genFullServiceID(srcAddr);
        bool isRollback = false;
        if (validators.length == 0) {
            // IBTP_RECEIPT_SUCCESS || IBTP_RECEIPT_FAILURE || IBTP_RECEIPT_ROLLBACK || IBTP_RECEIPT_ROLLBACK_END
            require(typ == 1 || typ == 2 || typ == 3 || typ == 4, "IBTP type is not correct in direct mode");
            if (typ == 1) {
                Transaction(directTransactionAddr).endTransactionSuccess(srcFullID, dstFullID, index);
            }
            if (typ == 2) {
                isRollback = true;
                Transaction(directTransactionAddr).endTransactionFail(srcFullID, dstFullID, index);
            }
            if (typ == 3) {
                isRollback = true;
                Transaction(directTransactionAddr).rollbackTransaction(srcFullID, dstFullID, index);
            }
            if (typ == 4) {
                Transaction(directTransactionAddr).endTransactionRollback(srcFullID, dstFullID, index);
                return;
            }
        } else {
            if (txStatus != 0 && txStatus != 3) {
                isRollback = true;
            }
        }

        invokeIndexUpdate(srcFullID, dstFullID, index, 1);

        checkReceiptMultiSigns(srcFullID, dstFullID, index, typ, result, txStatus, signatures);

        string memory outServicePair = genServicePair(srcFullID, dstFullID);
        CallFunc memory invokeFunc = outMessages[outServicePair][index].callback;
        bytes[] memory args = new bytes[](invokeFunc.args.length + result.length);

        if (isRollback) {
            invokeFunc = outMessages[outServicePair][index].rollback;
            args = new bytes[](invokeFunc.args.length);
        }

        for (uint i = 0; i < invokeFunc.args.length; i++) {
            args[i] = invokeFunc.args[i];
        }

        if (!isRollback) {
            for (uint i = 0; i < result.length; i++) {
                args[invokeFunc.args.length + i] = result[i];
            }
        }

        if (keccak256(abi.encodePacked(invokeFunc.func)) != keccak256(abi.encodePacked(""))) {

            string memory method = string(abi.encodePacked(invokeFunc.func, "(bytes[])"));
            (bool ok,) = address(stringToAddress(srcAddr)).call(abi.encodeWithSignature(method, args));
            emit throwReceiptStatus(ok);
            return;
        }

        emit throwReceiptStatus(true);
    }

    function invokeIndexUpdate(string memory srcFullID, string memory dstFullID, uint64 index, uint64 reqType) private {
        string memory servicePair = genServicePair(srcFullID, dstFullID);
        if (reqType == 0) {
            require(inCounter[servicePair] + 1 == index);
            markInCounter(servicePair);
        } else if (reqType == 1) {
            // invoke src callback or rollback
            require(callbackCounter[servicePair] + 1 == index);
            markCallbackCounter(servicePair, index);
        } else if (reqType == 2) {
            // invoke dst rollback
            require(dstRollbackCounter[servicePair] + 1 <= index);
            markDstRollbackCounter(servicePair, index);
            if (inCounter[servicePair] + 1 == index) {
                markInCounter(servicePair);
            }
        }
    }

    function emitInterchainEvent(
        string memory destFullServiceID,
        string memory funcCall,
        bytes[] memory args,
        string memory funcCb,
        bytes[] memory argsCb,
        string memory funcRb,
        bytes[] memory argsRb,
        bool isEncrypt)
    public onlyWhiteList {
        checkAppchainIdContains(appchainID, destFullServiceID);
        string memory curFullID = genFullServiceID(addressToString(msg.sender));
        string memory outServicePair = genServicePair(curFullID, destFullServiceID);

        // 直连模式下未注册的remoteService无法发出跨链交易
        if (valThreshold == 0) {
            // direct mode
            bool flag = false;
            string[] memory remoteServices = Transaction(directTransactionAddr).getRemoteServiceList();
            for (uint i = 0; i < remoteServices.length; i++) {
                if (keccak256(abi.encodePacked(destFullServiceID)) == keccak256(abi.encodePacked(remoteServices[i]))) {
                    flag = true;
                    break;
                }
            }
            require(flag == true, "remote service is not registered");
            flag = false;
            address[] memory banList = Transaction(directTransactionAddr).getRSWhiteList(destFullServiceID);
            for (uint i = 0; i < banList.length; i++) {
                if (msg.sender == banList[i]) {
                    flag = true;
                    break;
                }
            }
            require(flag == false, "remote service is not allowed to call dest address");
        }


        // Record the order of interchain contract which has been started.
        outCounter[outServicePair]++;
        if (outCounter[outServicePair] == 1) {
            outServicePairs.push(outServicePair);
        }

        outMessages[outServicePair][outCounter[outServicePair]] = InterchainInvoke(isEncrypt,
            CallFunc(funcCall, args),
            CallFunc(funcCb, argsCb),
            CallFunc(funcRb, argsRb));

        bytes memory packed = abi.encodePacked(funcCall);
        for (uint i = 0; i < args.length; i++) {
            packed = abi.encodePacked(packed, args[i]);
        }
        bytes32 hash = keccak256(packed);

        if (isEncrypt) {
            funcCall = "";
            args = new bytes[](0);
        }

        // Start transaction and record current block number in direct mode
        if (validators.length == 0) {
            Transaction(directTransactionAddr).startTransaction(curFullID, destFullServiceID, outCounter[outServicePair]);
        }

        // Throw interchain event for listening of plugin.
        emit throwInterchainEvent(outCounter[outServicePair], destFullServiceID, curFullID, funcCall, args, hash);
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

    // The helper functions that help plugin query.
    function getOuterMeta() public view returns (string[] memory, uint64[] memory) {
        uint64[] memory indices = new uint64[](outServicePairs.length);
        for (uint64 i = 0; i < outServicePairs.length; i++) {
            indices[i] = outCounter[outServicePairs[i]];
        }

        return (outServicePairs, indices);
    }

    function getOutMessage(string memory outServicePair, uint64 idx) public view returns (string memory, bytes[] memory, bool) {
        InterchainInvoke memory invoke = outMessages[outServicePair][idx];
        return (invoke.callFunc.func, invoke.callFunc.args, invoke.encrypt);
    }

    function getReceiptMessage(string memory inServicePair, uint64 idx) public view returns (bytes[] memory, uint64, bool)  {
        Receipt memory receipt = receiptMessages[inServicePair][idx];
        return (receipt.result, receipt.typ, receipt.encrypt);
    }

    function getInnerMeta() public view returns (string[] memory, uint64[] memory) {
        uint64[] memory indices = new uint64[](inServicePairs.length);
        for (uint i = 0; i < inServicePairs.length; i++) {
            indices[i] = inCounter[inServicePairs[i]];
        }

        return (inServicePairs, indices);
    }

    function getCallbackMeta() public view returns (string[] memory, uint64[] memory) {
        uint64[] memory indices = new uint64[](callbackServicePairs.length);
        for (uint64 i = 0; i < callbackServicePairs.length; i++) {
            indices[i] = callbackCounter[callbackServicePairs[i]];
        }

        return (callbackServicePairs, indices);
    }

    function getDstRollbackMeta() public view returns (string[] memory, uint64[] memory) {
        uint64[] memory indices = new uint64[](inServicePairs.length);
        for (uint i = 0; i < inServicePairs.length; i++) {
            indices[i] = dstRollbackCounter[inServicePairs[i]];
        }

        return (inServicePairs, indices);
    }

    // get transaction start timestamp and transaction status in direct mode
    function getDirectTransactionMeta(string memory id) public view returns (uint, uint64) {
        return (Transaction(directTransactionAddr).getStartTimestamp(id), Transaction(directTransactionAddr).getTransactionStatus(id));
    }

    function genFullServiceID(string memory serviceID) private view returns (string memory) {
        return string(abi.encodePacked(bitxhubID, ":", appchainID, ":", serviceID));
    }

    function genServicePair(string memory from, string memory to) private pure returns (string memory) {
        return string(abi.encodePacked(from, "-", to));
    }

    function getChainID() public view returns (string memory, string memory) {
        return (bitxhubID, appchainID);
    }

    function checkService(string memory remoteService, address destAddr) private view {
        require(localWhiteList[destAddr] == true, "dest address is not in local white list");

        if (valThreshold == 0) {
            // direct mode

            bool flag = false;
            string[] memory remoteServices = Transaction(directTransactionAddr).getRemoteServiceList();
            for (uint i = 0; i < remoteServices.length; i++) {
                if (keccak256(abi.encodePacked(remoteService)) == keccak256(abi.encodePacked(remoteServices[i]))) {
                    flag = true;
                    break;
                }
            }
            require(flag == true, "remote service is not registered");

            flag = false;
            address[] memory banList = Transaction(directTransactionAddr).getRSWhiteList(remoteService);
//            address[] memory banList = remoteWhiteList[remoteService];
            for (uint i = 0; i < banList.length; i++) {
                if (destAddr == banList[i]) {
                    flag = true;
                    break;
                }
            }
            require(flag == false, "remote service is not allowed to call dest address");
        }
    }

    function checkInterchainMultiSigns(string memory srcFullID,
        string memory dstFullID,
        uint64 index,
        uint64 typ,
        string memory callFunc,
        bytes[] memory args,
        uint64 txStatus,
        bytes[] memory multiSignatures) private {
        if (valThreshold == 0) {
            return;
        }

        bytes memory packed = abi.encodePacked(srcFullID, dstFullID, index, typ);
        bytes memory funcPacked = abi.encodePacked(callFunc);
        for (uint i = 0; i < args.length; i++) {
            funcPacked = abi.encodePacked(funcPacked, args[i]);
        }
        packed = abi.encodePacked(packed, keccak256(funcPacked), txStatus);
        bytes32 hash = keccak256(packed);

        require(checkMultiSigns(hash, multiSignatures), "invalid multi-signature");
    }

    function checkReceiptMultiSigns(string memory srcFullID,
        string memory dstFullID,
        uint64 index,
        uint64 typ,
        bytes[] memory result,
        uint64 txStatus,
        bytes[] memory multiSignatures) private {
        if (valThreshold == 0) {
            return;
        }

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
            for (uint i = 0; i < result.length; i++) {
                data = abi.encodePacked(data, result[i]);
            }
        }
        packed = abi.encodePacked(packed, keccak256(data), txStatus);
        bytes32 hash = keccak256(packed);

        require(checkMultiSigns(hash, multiSignatures), "invalid multi-signature");
    }

    function checkMultiSigns(bytes32 hash, bytes[] memory multiSignatures) private returns (bool) {
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
    ) internal pure returns (string memory asciiString) {
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

    function stringToAddress(string memory _address) internal pure returns (address) {
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


    function char(bytes1 b) internal pure returns (bytes1 c) {
        if (uint8(b) < 10) return bytes1(uint8(b) + 0x30);
        else return bytes1(uint8(b) + 0x57);
    }

    function checkAppchainIdContains (string memory appchainId, string memory destFullService) private pure {
        bytes memory whatBytes = bytes (appchainId);
        bytes memory whereBytes = bytes (destFullService);

        require(whereBytes.length >= whatBytes.length);

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
        // 不允许同broker服务自跨链
        require(!found, "dest service is belong to current broker!");
    }
}

abstract contract Transaction {
    function initialize() public virtual;

    function registerAppchain(string memory chainID, string memory broker, address ruleAddr, bytes memory trustRoot) public virtual;

    function getAppchainInfo(string memory chainID) public view virtual returns (string memory, bytes memory, address);

    function registerRemoteService(string memory chainID, string memory serviceID, address[] memory whiteList) public virtual;

    function getRSWhiteList(string memory remoteAddr) public view virtual returns (address[] memory);

    function getRemoteServiceList() public view virtual returns (string[] memory);

    function startTransaction(string memory from, string memory to, uint64 index) public virtual;

    function rollbackTransaction(string memory from, string memory to, uint64 index) public virtual;

    function endTransactionSuccess(string memory from, string memory to, uint64 index) public virtual;

    function endTransactionFail(string memory from, string memory to, uint64 index) public virtual;

    function endTransactionRollback(string memory from, string memory to, uint64 index) public virtual;

    function getTransactionStatus(string memory IBTPid) public view virtual returns (uint64);

    function getStartTimestamp(string memory IBTPid) public view virtual returns (uint);
}