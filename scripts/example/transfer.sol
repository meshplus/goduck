pragma solidity >=0.5.6;
pragma experimental ABIEncoderV2;

contract Transfer {
    mapping(string => uint64) accountM; // map for accounts
    // change the address of Broker accordingly
    address BrokerAddr;

    string bitxhubID;
    string appchainID;
    string curFullID;

    // AccessControl
    modifier onlyBroker {
        require(msg.sender == BrokerAddr, "Invoker are not the Broker");
        _;
    }

    constructor(address _brokerAddr, bool _ordered) {
        BrokerAddr = _brokerAddr;
        Broker(BrokerAddr).register(_ordered);
        (bitxhubID, appchainID) = Broker(BrokerAddr).getChainID();
        curFullID = genFullServiceID(addressToString(getAddress()));
    }

    function getCurFullID() public view returns (string memory) {
        return curFullID;
    }

    // get local contract address
    function getAddress() internal view returns (address) {
        return address(this);
    }

    function register(bool _ordered) public {
        Broker(BrokerAddr).register(_ordered);
    }

    // contract for asset
    function transfer(string memory destChainServiceID, string memory sender, string memory receiver, uint64 amount) public {
        require(accountM[sender] >= amount);
        accountM[sender] -= amount;

        bytes[] memory args = new bytes[](4);
        args[0] = abi.encodePacked(uint64(0));
        args[1] = abi.encodePacked(sender);
        args[2] = abi.encodePacked(receiver);
        args[3] = abi.encodePacked(amount);

        bytes[] memory argsRb = new bytes[](2);
        argsRb[0] = abi.encodePacked(sender);
        argsRb[1] = abi.encodePacked(amount);
        string[] memory func = new string[](3);
        func[0] = "interchainCharge";
        func[1] = "";
        func[2] = "interchainRollback";
        Broker(BrokerAddr).emitInterchainEvent(destChainServiceID, "interchainCharge", args, "", new bytes[](0), "interchainRollback", argsRb, false, new string[](0));
    }


    function multiTransfer(string memory destChainServiceID, string[] memory sender, string[] memory receiver, uint64[] memory amount) public {
        uint len = sender.length;
        require(len == receiver.length && len == amount.length);
        for (uint i = 0; i < len; i++) {
            require(accountM[sender[i]] >= amount[i]);
            accountM[sender[i]] -= amount[i];
        }
        bytes[] memory args = new bytes[](3 * len + 2);
        args[0] = abi.encodePacked(uint64(1));
        args[1] = abi.encodePacked(uint64(3));
        for (uint i = 0; i < len; i++) {
            args[i * 3 + 2] = abi.encodePacked(sender[i]);
            args[i * 3 + 3] = abi.encodePacked(receiver[i]);
            args[i * 3 + 4] = abi.encodePacked(amount[i]);
        }
        bytes[] memory argsRb = new bytes[](2 * len + 1);
        argsRb[0] = abi.encodePacked(uint64(2));
        for (uint i = 0; i < len; i++) {
            argsRb[i * 2 + 1] = abi.encodePacked(sender[i]);
            argsRb[i * 2 + 2] = abi.encodePacked(amount[i]);
        }

        Broker(BrokerAddr).emitInterchainEvent(destChainServiceID, "interchainMultiCharge", args, "", new bytes[](0), "interchainMultiRollback", argsRb, false, new string[](0));
    }

    // contract for asset
    function transferOne2Multi(string[] memory destChainServiceIDs, string[] memory senders, string[] memory receivers, uint64[] memory amounts) public {
        uint len = destChainServiceIDs.length;

        require(senders.length == len && receivers.length == len && amounts.length == len);

        string[] memory brokerOuts;
        uint64[] memory outCounters;
        string[] memory group;
        (brokerOuts, outCounters) = Broker(BrokerAddr).getOuterMeta();
        for (uint i = 0; i < len; i++)
        {
            group = fillMultiGroup(destChainServiceIDs, brokerOuts, outCounters);
        }

        for (uint i = 0; i < len; i++)
        {
            require(accountM[senders[i]] >= amounts[i]);
            accountM[senders[i]] -= amounts[i];

            bytes[] memory args = new bytes[](4);
            args[0] = abi.encodePacked(uint64(0));
            args[1] = abi.encodePacked(senders[i]);
            args[2] =abi.encodePacked(receivers[i]);
            args[3] = abi.encodePacked(amounts[i]);

            bytes[] memory argsRb = new bytes[](2);
            argsRb[0] = abi.encodePacked(senders[i]);
            argsRb[1] = abi.encodePacked(amounts[i]);
            Broker(BrokerAddr).emitInterchainEvent(destChainServiceIDs[i], "interchainCharge", args, "", new bytes[](0), "interchainRollback", argsRb, false, group);
        }
    }

    function fillMultiGroup(string[] memory destChainServiceIDs, string[] memory brokerOuts, uint64[] memory outCounters) private view returns (string[] memory) {
        string[] memory multiGroup = new string[](destChainServiceIDs.length);
        for (uint i = 0; i < destChainServiceIDs.length; i++)
        {
            uint index = brokerOuts.length;
            for (uint j = 0; j < brokerOuts.length; j++)
            {
                string memory outServicePair = genServicePair(curFullID, destChainServiceIDs[i]);
                if (keccak256(abi.encodePacked(brokerOuts[j])) == keccak256(abi.encodePacked(outServicePair))) {
                    index = j;
                }
            }

            uint groupValue;
            // not found in out meta, indicate index equal 1
            if (index == brokerOuts.length) {
                groupValue = 1;
            } else {
                groupValue = outCounters[index]+1;
            }
            multiGroup[i] = genMultiGroup(destChainServiceIDs[i], uint2str(groupValue));
        }
        return multiGroup;
    }


    function interchainRollback(bytes[] memory args) public onlyBroker {
        require(args.length == 2, "interchainRollback args' length is not correct, expect 2");
        string memory sender = string(args[0]);
        uint64 amount = bytesToUint64(args[1]);
        accountM[sender] += amount;
    }

    function interchainMultiRollback(bytes[] memory args, bool[] memory multiStatus) public onlyBroker {
        uint64 arglenth = bytesToUint64(args[0]);
        string memory sender;
        uint64 amount;
        if (multiStatus.length == 0){
            multiStatus = new bool[]((args.length-1)/arglenth);
            for (uint i = 0; i < (args.length-1)/arglenth; i++){
                multiStatus[i] = false;
            }
        }
        for (uint i = 0; i < multiStatus.length; i++){
            if (!multiStatus[i]){
                sender = string(args[i * arglenth + 1]);
                amount = bytesToUint64(args[i * arglenth + 2]);
                accountM[sender] += amount;
            }
        }
    }

    function interchainCharge(bytes[] memory args, bool isRollback) public onlyBroker returns (bytes[] memory) {
        require(args.length == 3, "interchainCharge args' length is not correct, expect 3");
        string memory receiver = string(args[1]);
        uint64 amount = bytesToUint64(args[2]);

        if (!isRollback) {
            accountM[receiver] += amount;
        } else {
            accountM[receiver] -= amount;
        }

        return new bytes[](0);
    }

    function interchainMultiCharge(bytes[][] memory args, bool isRollback) public onlyBroker returns (bytes[][] memory results, bool[] memory multiStatus ) {
        for(uint i = 0; i< args.length; i++){
            require(args[i].length == 3, "interchainMultiCharge args' length is not correct, expect 3");
        }

        bool[] memory multiStatus = new bool[](args.length);

        for (uint i = 0; i < args.length; i++){
            string memory receiver = string(args[i][1]);
            uint64 amount = bytesToUint64(args[i][2]);
            if (!isRollback) {
                accountM[receiver] += amount;
            } else {
                accountM[receiver] -= amount;
            }
            multiStatus[i] = true;
        }

        return (new bytes[][](0), multiStatus);
    }

    function getBalance(string memory id) public view returns(uint64) {
        return accountM[id];
    }

    function setBalance(string memory id, uint64 amount) public {
        accountM[id] = amount;
    }

    function bytesToUint64(bytes memory b) public pure returns (uint64){
        uint64 number;
        for(uint i=0;i<b.length;i++){
            number = uint64(number + uint8(b[i])*(2**(8*(b.length-(i+1)))));
        }
        return number;
    }

    function genFullServiceID(string memory serviceID) private view returns (string memory) {
        return string(abi.encodePacked(bitxhubID, ":", appchainID, ":", serviceID));
    }

    function genServicePair(string memory from, string memory to) private pure returns (string memory) {
        return string(abi.encodePacked(from, "-", to));
    }

    function genMultiGroup(string memory groupKey, string memory groupValue) private pure returns (string memory) {
        return string(abi.encodePacked(groupKey, "-", groupValue));
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

    function uint2str(uint _i) internal pure returns (string memory _uintAsString) {
        if (_i == 0) {
            return "0";
        }
        uint j = _i;
        uint len;
        while (j != 0) {
            len++;
            j /= 10;
        }
        bytes memory bstr = new bytes(len);
        uint k = len;
        while (_i != 0) {
            k = k-1;
            uint8 temp = (48 + uint8(_i - _i / 10 * 10));
            bytes1 b1 = bytes1(temp);
            bstr[k] = b1;
            _i /= 10;
        }
        return string(bstr);
    }
}

abstract contract Broker {
    function emitInterchainEvent(
        string memory destFullServiceID,
        string memory func,
        bytes[] memory args,
        string memory funcCb,
        bytes[] memory argsCb,
        string memory funcRb,
        bytes[] memory argsRb,
        bool isEncrypt,
        string[] memory group
    ) public virtual;

    function register(bool ordered) public virtual;

    function getOuterMeta() public virtual view returns (string[] memory, uint64[] memory);

    function getChainID() public virtual view returns (string memory, string memory);
}