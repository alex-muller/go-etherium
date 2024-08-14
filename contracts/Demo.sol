// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.0;

contract Demo {
    function getBalance() public view returns(uint) {
        return address(this).balance;
    }
}
