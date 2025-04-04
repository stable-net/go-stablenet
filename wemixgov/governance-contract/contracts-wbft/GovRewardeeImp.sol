// SPDX-License-Identifier: MIT

pragma solidity 0.8.14;

contract GovRewardeeImp {
    address public constant GOV_STAKING = address(0x1001);

    event RewardPaid(address indexed recipient, uint256 amount);

    receive() external payable {}

    modifier onlyGovStaking() {
        require(msg.sender == address(GOV_STAKING), "GovRewardee: caller is not the GovStaking contract");
        _;
    }

    function sendRewardTo(address payable recipient, uint256 amount) onlyGovStaking external {
        require(recipient != address(0), "GovRewardee: recipient is the zero address");
        require(amount > 0, "GovRewardee: amount is zero");
        require(amount <= address(this).balance, "GovRewardee: insufficient balance");

        (bool success, ) = recipient.call{value: amount}(""); // don't use transfer to call receive logic of recipient
        require(success, "failed to transfer");

        emit RewardPaid(recipient, amount);
    }
}
