pragma solidity >=0.4.21;

contract test{
    uint64 number;
    constructor(uint64 _number) public{
        number = _number;
    }

    function get() public view returns (uint64){
        return number;
    }

    function set(uint64 _number) public{
        number = _number;
    }
}