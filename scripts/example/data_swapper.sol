pragma solidity >=0.5.6;
pragma experimental ABIEncoderV2;

contract DataSwapper {
    mapping(string => string) dataM; // map for accounts
    // change the address of Broker accordingly
    address BrokerAddr;

    // AccessControl
    modifier onlyBroker {
        require(msg.sender == BrokerAddr, "Invoker are not the Broker");
        _;
    }

    constructor(address _brokerAddr) public {
        BrokerAddr = _brokerAddr;
        Broker(BrokerAddr).register();
    }

    function register() public {
        Broker(BrokerAddr).register();
    }

    // contract for data exchange
    function getData(string memory key) public view returns(string memory) {
        return dataM[key];
    }

    function get(string memory destChainServiceID, string memory key) public {
        bytes[] memory args = new bytes[](1);
        args[0] = abi.encodePacked(key);

        bytes[] memory argsCb = new bytes[](1);
        argsCb[0] = abi.encodePacked(key);

        Broker(BrokerAddr).emitInterchainEvent(destChainServiceID, "interchainGet", args, "interchainSet", argsCb, "", new bytes[](0), false);
    }

    function set(string memory key, string memory value) public {
        dataM[key] = value;
    }

    function interchainSet(bytes[] memory args) public onlyBroker {
        require(args.length == 2, "interchainSet args' length is not correct, expect 2");
        string memory key = string(args[0]);
        string memory value = string(args[1]);
        set(key, value);
    }

    function interchainGet(bytes[] memory args, bool isRollback) public onlyBroker returns(bytes[] memory) {
        require(args.length == 1, "interchainGet args' length is not correct, expect 1");
        string memory key = string(args[0]);

        bytes[] memory result = new bytes[](1);
        result[0] = abi.encodePacked(dataM[key]);

        return result;
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
        bool isEncrypt) public virtual;

    function register() public virtual;
}