pragma solidity >=0.5.6;
pragma experimental ABIEncoderV2;

contract Transfer {
    mapping(string => uint64) accountM; // map for accounts
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

    // contract for asset
    function transfer(string memory destChainServiceID, string memory sender, string memory receiver, uint64 amount) public {
        require(accountM[sender] >= amount);
        accountM[sender] -= amount;

        bytes[] memory args = new bytes[](3);
        args[0] = abi.encodePacked(sender);
        args[1] =abi.encodePacked(receiver);
        args[2] = abi.encodePacked(amount);

        bytes[] memory argsRb = new bytes[](2);
        argsRb[0] = abi.encodePacked(sender);
        argsRb[1] = abi.encodePacked(amount);

        Broker(BrokerAddr).emitInterchainEvent(destChainServiceID, "interchainCharge", args, "", new bytes[](0), "interchainRollback", argsRb, false);
    }

    function interchainRollback(bytes[] memory args) public onlyBroker {
        require(args.length == 2, "interchainRollback args' length is not correct, expect 2");
        string memory sender = string(args[0]);
        uint64 amount = bytesToUint64(args[1]);
        accountM[sender] += amount;
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
