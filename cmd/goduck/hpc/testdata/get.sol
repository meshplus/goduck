pragma solidity >=0.4.21;

contract Broker {
    Hasher hasher = Hasher(0x00000000000000000000000000000000000000fa);

    function get() public returns (int32) {
        return 3;
    }
}

contract Hasher {
    function getHash() public returns (bytes32);
}
