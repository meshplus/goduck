pragma solidity >=0.4.21;

contract get{
    function multiply(uint64 input1, uint64 input2) public pure returns (uint64 res1, uint64 res2){
        return (input1 * input2, input1 + input2);
    }
}