pragma solidity ^0.4;

contract Testcontract{

    event TestEvent(bytes32 message);

    function Testfunction(bytes32 input) public {
        emit TestEvent(input);
    }
}