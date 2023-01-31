pragma solidity >=0.6.9 <=0.7.6;
pragma experimental ABIEncoderV2;

contract Broker {
    struct Proposal {
        uint64 approve;
        uint64 reject;
        address[] votedAdmins;
        bool ordered;
        bool exist;
    }

    // Only the contract in the whitelist can invoke the Broker for interchain operations.
    mapping(address => bool) localWhiteList;
    address[] localServices;
    mapping(address => Proposal) localServiceProposal;
    address[] proposalList;
    mapping(address => bool) serviceOrdered;

    string bitxhubID;
    string appchainID;
    address[] public validators;
    uint64 public valThreshold;
    address[] public admins;
    uint64 public adminThreshold;

    event throwInterchainEvent(uint64 index, string dstFullID, string srcFullID, string func, bytes[] args, bytes32 hash, string[] group);
    event throwReceiptEvent(uint64 index, string dstFullID, string srcFullID, uint64 typ, bytes[][] results, bytes32 hash, bool[] multiStatus);
    event throwReceiptStatus(bool);

    address dataAddr;

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
        uint64 _adminThreshold,
        address _dataAddr) {
        bitxhubID = _bitxhubID;
        appchainID = _appchainID;
        validators = _validators;
        valThreshold = _valThreshold;
        admins = _admins;
        adminThreshold = _adminThreshold;
        dataAddr = _dataAddr;
        BrokerData(_dataAddr).register();
    }

    function setAdmins(address[] memory _admins, uint64 _adminThreshold) public onlyAdmin {
        admins = _admins;
        adminThreshold = _adminThreshold;
    }

    function setValidators(address[] memory _validators, uint64 _valThreshold) public onlyAdmin {
        validators = _validators;
        valThreshold = _valThreshold;
    }


    function initialize() public onlyAdmin {
        for (uint n = 0; n < localServices.length; n++) {
            localWhiteList[localServices[n]] = false;
        }
        for (uint x = 0; x < proposalList.length; x++) {
            delete localServiceProposal[proposalList[x]];
        }
        delete localServices;

        BrokerData(dataAddr).initialize();
    }

    // register local service to Broker
    function register(bool ordered) public {
        require(tx.origin != msg.sender, "register not by contract");
        if (localWhiteList[msg.sender] || localServiceProposal[msg.sender].exist) {
            return;
        }

        localServiceProposal[msg.sender] = Proposal(0, 0, new address[](admins.length), ordered, true);
    }

    function audit(address addr, int64 status) public onlyAdmin returns (bool) {
        uint result = vote(addr, status);

        if (result == 0) {
            return false;
        }

        if (result == 1) {
            bool ordered = localServiceProposal[addr].ordered;
            delete localServiceProposal[addr];
            localWhiteList[addr] = true;
            serviceOrdered[addr] = ordered;
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

    // get the registered local service list
    function getLocalServiceList() public view returns (string[] memory) {
        string[] memory fullServiceIDList = new string[](localServices.length);
        for (uint i = 0; i < localServices.length; i++) {
            fullServiceIDList[i] = genFullServiceID(BrokerData(dataAddr).addressToString(localServices[i]));
        }

        return fullServiceIDList;
    }

    // get the registered counterparty service list
    function getLocalWhiteList(address addr) public view returns (bool) {
        return localWhiteList[addr];
    }

    function invokeInterchains(
        string[] memory srcFullID,
        string[] memory destAddr,
        uint64[] memory index,
        uint64[] memory typ,
        string[] memory callFunc,
        bytes[][] memory args,
        uint64[] memory txStatus,
        bytes[][] memory signatures,
        bool[] memory isEncrypt) payable external {
        for (uint8 i = 0; i <  srcFullID.length; ++i) {
            if (serviceOrdered[BrokerData(dataAddr).stringToAddress(destAddr[i])] == true) {
                string memory dstFullID = genFullServiceID(destAddr[i]);
                invokeIndexUpdateWithError(srcFullID[i], dstFullID, index[i], txStatus[i], isEncrypt[i], "dst service is not ordered", uint64(1));
                continue;
            }
            invokeInterchain(srcFullID[i], destAddr[i], index[i], typ[i], callFunc[i], args[i], txStatus[i], signatures[i], isEncrypt[i]);
        }
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
        bool isEncrypt) payable public {
        string memory dstFullID = genFullServiceID(destAddr);
        string memory servicePair = genServicePair(srcFullID, dstFullID);
        {
            bool ok = BrokerData(dataAddr).checkInterchainMultiSigns(srcFullID, dstFullID, index, typ, callFunc, args, txStatus, signatures, validators, valThreshold);
            if (!ok) {
                invokeIndexUpdateWithError(srcFullID, dstFullID, index, txStatus, isEncrypt, "invalid multi-signature", uint64(1));
                return;
            }

            if (localWhiteList[BrokerData(dataAddr).stringToAddress(destAddr)] == false) {
                invokeIndexUpdateWithError(srcFullID, dstFullID, index, txStatus, isEncrypt, "dest address is not in local white list", uint64(1));
                return;
            }
        }

        bool[] memory status = new bool[](1);
        status[0] = true;
        bytes[][] memory results = new bytes[][](1);
        if (txStatus == 0) {
            // INTERCHAIN && BEGIN
            if (BrokerData(dataAddr).getInCounter(servicePair) < index) {
                (status[0], results[0]) = callService(BrokerData(dataAddr).stringToAddress(destAddr), callFunc, args, false);
            }
            require(BrokerData(dataAddr).invokeIndexUpdate(srcFullID, dstFullID, index, 0));
            if (status[0]) {
                typ = 1;
            } else {
                typ = 2;
            }
        } else {
            // INTERCHAIN && FAILURE || INTERCHAIN && ROLLBACK, only happened in relay mode
            if (BrokerData(dataAddr).getInCounter(servicePair) >= index) {
                (status[0], results[0]) = callService(BrokerData(dataAddr).stringToAddress(destAddr), callFunc, args, true);
            }
            require(BrokerData(dataAddr).invokeIndexUpdate(srcFullID, dstFullID, index, 2));
            if (txStatus == 1) {
                typ = 2;
            } else {
                typ = 3;
            }
        }

        BrokerData(dataAddr).setReceiptMessage(servicePair, index, isEncrypt, typ, results, status);

        if (isEncrypt) {
            emit throwReceiptEvent(index, dstFullID, srcFullID, typ, new bytes[][](0), computeHash(results), status);
        } else {
            emit throwReceiptEvent(index, dstFullID, srcFullID, typ, results, computeHash(results), status);
        }
    }

    function computeHash(bytes[][] memory args) internal pure returns (bytes32) {
        bytes memory packed;
        for (uint i = 0; i < args.length; i++) {
            bytes[] memory arg = args[i];
            for (uint j = 0; j < arg.length; j++) {
                packed = abi.encodePacked(packed, arg[j]);
            }
        }

        return keccak256(packed);
    }

    // called on dest chain
    function invokeMultiInterchain(
        string memory srcFullID,
    // 地址变为string格式，这样多签不会有问题，在验证多签之前使用checksum之前的合约地址
        string memory destAddr,
        uint64 index,
        uint64 typ,
        string memory callFunc,
        bytes[][] memory args,
        uint64 txStatus,
        bytes[] memory signatures,
        bool isEncrypt) payable public {
        string memory dstFullID = genFullServiceID(destAddr);
        string memory servicePair = genServicePair(srcFullID, dstFullID);
        {
            bool ok = BrokerData(dataAddr).checkMultiInterchainMultiSigns(srcFullID, dstFullID, index, typ, callFunc, args, txStatus, signatures, validators, valThreshold);
            if (!ok) {
                invokeIndexUpdateWithError(srcFullID, dstFullID, index, txStatus, isEncrypt, "invalid multiInterchain-multi-signature", uint64(args.length));
                return;
            }

            if (localWhiteList[BrokerData(dataAddr).stringToAddress(destAddr)] == false) {
                invokeIndexUpdateWithError(srcFullID, dstFullID, index, txStatus, isEncrypt, "dest address is not in local white list", uint64(args.length));
                return;
            }
        }

        bytes[][] memory results = new bytes[][](args.length);
        bool[] memory multiStatus = new bool[](args.length);
        typ = 1;
        if (txStatus == 0) {
            // INTERCHAIN && BEGIN
            if (BrokerData(dataAddr).getInCounter(servicePair) < index) {
                (multiStatus, results) = callMultiService(BrokerData(dataAddr).stringToAddress(destAddr), callFunc, args, false);
                for (uint i = 0; i < multiStatus.length; i++){
                    if(!multiStatus[i]){
                        typ = 2;
                        break;
                    }
                }
            }
            require(BrokerData(dataAddr).invokeIndexUpdate(srcFullID, dstFullID, index, 0));
        } else {
            // INTERCHAIN && FAILURE || INTERCHAIN && ROLLBACK, only happened in relay mode
            if (BrokerData(dataAddr).getInCounter(servicePair) >= index) {
                (multiStatus, results) = callMultiService(BrokerData(dataAddr).stringToAddress(destAddr), callFunc, args, true);
            }
            require(BrokerData(dataAddr).invokeIndexUpdate(srcFullID, dstFullID, index, 2));
            if (txStatus == 1) {
                typ = 2;
            } else {
                typ = 3;
            }
        }


        BrokerData(dataAddr).setReceiptMessage(servicePair, index, isEncrypt, typ, results, multiStatus);

        if (isEncrypt) {
            emit throwReceiptEvent(index, dstFullID, srcFullID, typ, new bytes[][](0), computeHash(results), multiStatus);
        } else {
            emit throwReceiptEvent(index, dstFullID, srcFullID, typ, results, computeHash(results), multiStatus);
        }
    }

    function callMultiService(address destAddr, string memory callFunc, bytes[][] memory args, bool isRollback) private returns (bool[] memory, bytes[][] memory) {
        bool status = true;
        bytes[][] memory Results;
        bool[] memory MultiStatus;

        if (keccak256(abi.encodePacked(callFunc)) != keccak256(abi.encodePacked(""))) {
            (bool ok, bytes memory data) = address(destAddr).call(abi.encodeWithSignature(string(abi.encodePacked(callFunc, "(bytes[][],bool)")), args, isRollback));
            status = ok;
            if (status) {
                (Results, MultiStatus) = abi.decode(data, (bytes[][],bool[]));
            }
        }

        return (MultiStatus, Results);
    }

    function callService(address destAddr, string memory callFunc, bytes[] memory args, bool isRollback) private returns (bool, bytes[] memory) {
        bool status = true;
        bytes[] memory result;

        if (keccak256(abi.encodePacked(callFunc)) != keccak256(abi.encodePacked(""))) {
            (bool ok, bytes memory data) = address(destAddr).call(abi.encodeWithSignature(string(abi.encodePacked(callFunc, "(bytes[],bool)")), args, isRollback));
            status = ok;
            if (status) {
                result = abi.decode(data, (bytes[]));
            }
        }

        return (status, result);
    }

    // called on src chain
    function invokeReceipt(
        string memory srcAddr,
        string memory dstFullID,
        uint64 index,
        uint64 typ,
        bytes[][] memory results,
        uint64 txStatus,
        bytes[] memory signatures) payable external {
        string memory srcFullID = genFullServiceID(srcAddr);
        bool isRollback = false;
        if (txStatus != 0 && txStatus != 3) {
            isRollback = true;
        }

        require(BrokerData(dataAddr).invokeIndexUpdate(srcFullID, dstFullID, index, 1));

        require(BrokerData(dataAddr).checkReceiptMultiSigns(srcFullID, dstFullID, index, typ, results, txStatus, signatures, validators, valThreshold), "invalid Receipt-multi-signature");

        string memory outServicePair = genServicePair(srcFullID, dstFullID);

        receiptCall(outServicePair, index, isRollback, srcAddr, results);
    }

    function receiptCall(string memory servicePair, uint64 index, bool isRollback, string memory srcAddr, bytes[][] memory results) private {
        string memory callFunc;
        bytes[] memory callArgs;
        bytes[] memory args;
        if (isRollback) {
            (callFunc, callArgs) = BrokerData(dataAddr).getRollbackMessage(servicePair, index);
            args = new bytes[](callArgs.length);
        } else {
            (callFunc, callArgs) = BrokerData(dataAddr).getCallbackMessage(servicePair, index);
            args = new bytes[](callArgs.length + results[0].length);
        }

        for (uint i = 0; i < callArgs.length; i++) {
            args[i] = callArgs[i];
        }

        if (!isRollback) {
            for (uint i = 0; i < results[0].length; i++) {
                args[callArgs.length + i] = results[0][i];
            }
        }

        if (keccak256(abi.encodePacked(callFunc)) != keccak256(abi.encodePacked(""))) {
            string memory method = string(abi.encodePacked(callFunc, "(bytes[])"));
            (bool ok,) = address(BrokerData(dataAddr).stringToAddress(srcAddr)).call(abi.encodeWithSignature(method, args));
            emit throwReceiptStatus(ok);
            return;
        }

        emit throwReceiptStatus(true);
    }

    // called on src chain
    function invokeMultiReceipt(
        string memory srcAddr,
        string memory dstFullID,
        uint64 index,
        uint64 typ,
        bytes[][] memory results,
        bool[] memory multiStatus,
        uint64 txStatus,
        bytes[] memory signatures) payable external {
        string memory srcFullID = genFullServiceID(srcAddr);
        bool isRollback = false;

        if (txStatus != 0 && txStatus != 3) {
            isRollback = true;
        }
        {
            require(BrokerData(dataAddr).invokeIndexUpdate(srcFullID, dstFullID, index, 1));
            require(BrokerData(dataAddr).checkReceiptMultiSigns(srcFullID, dstFullID, index, typ, results, txStatus, signatures, validators, valThreshold));
        }

        string memory outServicePair = genServicePair(srcFullID, dstFullID);

        multiReceiptCall(outServicePair, index, isRollback, srcAddr, results, multiStatus);
    }

    function multiReceiptCall(string memory servicePair, uint64 index, bool isRollback, string memory srcAddr, bytes[][] memory results, bool[] memory multiStatus) private {
        string memory callFunc;
        bytes[] memory callArgs;
        bytes[] memory args;
        if (isRollback) {
            (callFunc, callArgs) = BrokerData(dataAddr).getRollbackMessage(servicePair, index);
            args = new bytes[](callArgs.length);
            for (uint i = 0; i < callArgs.length; i++) {
                args[i] = callArgs[i];
            }
            if (keccak256(abi.encodePacked(callFunc)) != keccak256(abi.encodePacked(""))) {
                (bool ok,) = address(BrokerData(dataAddr).stringToAddress(srcAddr)).call(abi.encodeWithSignature(string(abi.encodePacked(callFunc, "(bytes[],bool[])")), args, multiStatus));
                if (!ok){
                    emit throwReceiptStatus(false);
                    return;
                }
            }
        }

        bool flag = false;
        for (uint i = 0; i < multiStatus.length; i++) {
            if (multiStatus[i] == true){
                flag = true;
                break;
            }
        }

        if (flag) {
            (callFunc, callArgs) = BrokerData(dataAddr).getCallbackMessage(servicePair, index);
            args = new bytes[](callArgs.length);
            for (uint i = 0; i < callArgs.length; i++) {
                args[i] = callArgs[i];
            }
            if (keccak256(abi.encodePacked(callFunc)) != keccak256(abi.encodePacked(""))) {
                (bool ok,) = address(BrokerData(dataAddr).stringToAddress(srcAddr)).call(abi.encodeWithSignature(string(abi.encodePacked(callFunc, "(bytes[],bool[],bytes[][])")), args, multiStatus, results));
                if (!ok) {
                    emit throwReceiptStatus(false);
                    return;
                }
            }
        }
        emit throwReceiptStatus(true);
    }

    function invokeIndexUpdateWithError(string memory srcFullID, string memory dstFullID, uint64 index, uint64 txStatus, bool isEncrypt, string memory errorMsg, uint64 resultsSize) private {
        string memory servicePair = genServicePair(srcFullID, dstFullID);
        uint64 typ;
        bytes[][] memory results = new bytes[][](resultsSize);
        bytes[] memory result = new bytes[](1);
        for (uint64 i = 0; i < resultsSize; i++) {
            result[0] = bytes(errorMsg);
            results[i] = result;
        }

        if(txStatus == 0) {
            require(BrokerData(dataAddr).invokeIndexUpdate(srcFullID, dstFullID, index, 0));
            typ = 2;
        } else {
            require(BrokerData(dataAddr).invokeIndexUpdate(srcFullID, dstFullID, index, 2));
            if(txStatus == 1) {
                typ = 2;
            } else {
                typ = 3;
            }
        }

        bool[] memory multiStatus = new bool[](resultsSize);
        for (uint64 i = 0; i < resultsSize; i++) {
            multiStatus[i] = false;
        }

        BrokerData(dataAddr).setReceiptMessage(servicePair, index, isEncrypt, typ, results, multiStatus);

        if (isEncrypt) {
            emit throwReceiptEvent(index, dstFullID, srcFullID, typ, new bytes[][](0), computeHash(results), multiStatus);
        } else {
            emit throwReceiptEvent(index, dstFullID, srcFullID, typ, results, computeHash(results), multiStatus);
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
        bool isEncrypt,
        string[] memory group)
    public onlyWhiteList {
        // 不允许同broker服务自跨链
        require(!BrokerData(dataAddr).checkAppchainIdContains(appchainID, destFullServiceID), "dest service is belong to current broker!");
        string memory curFullID = genFullServiceID(BrokerData(dataAddr).addressToString(msg.sender));
        string memory outServicePair = genServicePair(curFullID, destFullServiceID);

        // Record the order of interchain contract which has been started.
        uint64 currentOutCounter = BrokerData(dataAddr).markOutCounter(outServicePair);


        BrokerData(dataAddr).setOutMessage(outServicePair, isEncrypt, group, funcCall, args, funcCb, argsCb, funcRb, argsRb);

        bytes32 hash = computeInvokeHash(funcCall, args);

        if (isEncrypt) {
            funcCall = "";
            args = new bytes[](0);
        }

        // Throw interchain event for listening of plugin.
        emit throwInterchainEvent(currentOutCounter, destFullServiceID, curFullID, funcCall, args, hash, group);
    }

    function computeInvokeHash(string memory funcCall, bytes[] memory args) private pure returns(bytes32) {
        bytes memory packed = abi.encodePacked(funcCall);
        for (uint i = 0; i < args.length; i++) {
            packed = abi.encodePacked(packed, args[i]);
        }
        return keccak256(packed);
    }

    // The helper functions that help plugin query.
    function getOuterMeta() public view returns (string[] memory, uint64[] memory) {
        return BrokerData(dataAddr).getOuterMeta();
    }

    function getOutMessage(string memory outServicePair, uint64 idx) public view returns (string memory, bytes[] memory, bool, string[] memory) {
        return BrokerData(dataAddr).getOutMessage(outServicePair, idx);
    }

    function getReceiptMessage(string memory inServicePair, uint64 idx) public view returns (bytes[][] memory, uint64, bool, bool[] memory)  {
        return BrokerData(dataAddr).getReceiptMessage(inServicePair, idx);
    }

    function getInnerMeta() public view returns (string[] memory, uint64[] memory) {
        return BrokerData(dataAddr).getInnerMeta();
    }

    function getCallbackMeta() public view returns (string[] memory, uint64[] memory) {
        return BrokerData(dataAddr).getCallbackMeta();
    }

    function getDstRollbackMeta() public view returns (string[] memory, uint64[] memory) {
        return BrokerData(dataAddr).getDstRollbackMeta();
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
}

abstract contract BrokerData {
    function register() public virtual;

    function initialize() public virtual;

    function checkInterchainMultiSigns(string memory srcFullID,
        string memory dstFullID,
        uint64 index,
        uint64 typ,
        string memory callFunc,
        bytes[] memory args,
        uint64 txStatus,
        bytes[] memory multiSignatures,
        address[] memory validators,
        uint64 valThreshold) public virtual returns(bool);

    function checkMultiInterchainMultiSigns(string memory srcFullID,
        string memory dstFullID,
        uint64 index,
        uint64 typ,
        string memory callFunc,
        bytes[][] memory args,
        uint64 txStatus,
        bytes[] memory multiSignatures,
        address[] memory validators,
        uint64 valThreshold) public virtual returns (bool);

    function checkReceiptMultiSigns(string memory srcFullID,
        string memory dstFullID,
        uint64 index,
        uint64 typ,
        bytes[][] memory result,
        uint64 txStatus,
        bytes[] memory multiSignatures,
        address[] memory validators,
        uint64 valThreshold) public virtual returns(bool);

    function setOutMessage(string memory servicePair,
        bool isEncrypt,
        string[] memory group,
        string memory funcCall,
        bytes[] memory args,
        string memory funcCb,
        bytes[] memory argsCb,
        string memory funcRb,
        bytes[] memory argsRb) public virtual;

    function invokeIndexUpdate(string memory srcFullID, string memory dstFullID, uint64 index, uint64 reqType) public virtual returns(bool);

    function getInCounter(string memory servicePair) public view virtual returns(uint64);

    function getCallbackMessage(string memory servicePair, uint64 index) public view virtual returns(string memory, bytes[] memory);

    function getRollbackMessage(string memory servicePair, uint64 index) public view virtual returns(string memory, bytes[] memory);

    function setReceiptMessage(string memory servicePair, uint64 index, bool isEncrypt, uint64 typ, bytes[][] memory results, bool[] memory multiStatus) public virtual;

    function markOutCounter(string memory servicePair) public virtual returns(uint64);

    function stringToAddress(string memory _address) public pure virtual returns (address);

    function addressToString(address account) public pure virtual returns (string memory asciiString);

    function checkAppchainIdContains (string memory appchainId, string memory destFullService) public pure virtual returns(bool);

    function getOuterMeta() public view virtual returns (string[] memory, uint64[] memory);

    function getOutMessage(string memory outServicePair, uint64 idx) public view virtual returns (string memory, bytes[] memory, bool, string[] memory);

    function getReceiptMessage(string memory inServicePair, uint64 idx) public view virtual returns (bytes[][] memory, uint64, bool, bool[] memory);

    function getInnerMeta() public view virtual returns (string[] memory, uint64[] memory);

    function getCallbackMeta() public view virtual returns (string[] memory, uint64[] memory);

    function getDstRollbackMeta() public view virtual returns (string[] memory, uint64[] memory);
}