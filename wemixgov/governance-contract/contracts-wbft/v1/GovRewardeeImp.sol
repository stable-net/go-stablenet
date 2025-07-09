// SPDX-License-Identifier: MIT

pragma solidity ^0.8.14;

import "./IFeeRecipient.sol";

contract GovRewardeeImp {
    address public govStaking;

    //***********************************************************************
    //* Caution for Upgrading
    //* - If you add new state variables, please add them after this comment
    //* - Never modify existing state variables
    //***********************************************************************

    event RewardPaid(address indexed recipient, uint256 amount);
    event FeePaid(address indexed recipient, uint256 amount);

    receive() external payable {}

    modifier onlyGovStaking() {
        require(msg.sender == address(govStaking), "GovRewardee: caller is not the GovStaking contract");
        _;
    }

    function initialize(address _govStaking) external {
        require(govStaking == address(0), "GovRewardee: already initialized");
        govStaking = _govStaking;
    }

    function sendRewardTo(address payable recipient, uint256 amount) external onlyGovStaking {
        require(recipient != address(0), "GovRewardee: recipient is the zero address");
        require(amount > 0, "GovRewardee: amount is zero");
        require(amount <= address(this).balance, "GovRewardee: insufficient balance");

        (bool success, ) = recipient.call{ value: amount }(""); // don't use transfer to call receive logic of recipient
        require(success, "GovRewardee: reward transfer failed");
        emit RewardPaid(recipient, amount);
    }

    function sendFeeTo(address payable recipient, uint256 amount) external onlyGovStaking {
        require(recipient != address(0), "GovRewardee: recipient is the zero address");
        require(amount > 0, "GovRewardee: amount is zero");
        require(amount <= address(this).balance, "GovRewardee: insufficient balance");

        uint256 size;
        assembly {
            size := extcodesize(recipient)
        }
        if (size > 0) {
            // if it is for sending fee to recipient contract, try to call receiveFee()
            try IERC165(recipient).supportsInterface(type(IFeeRecipient).interfaceId) returns (bool supported) {
                if (supported) {
                    // IFeeRecipient is implemented
                    try IFeeRecipient(recipient).receiveFee{ value: amount }(amount) {
                        emit FeePaid(recipient, amount);
                        return;
                    } catch {
                        revert("GovRewardee: fee recipient contract reverted");
                    }
                } else {
                    (bool success, ) = recipient.call{ value: amount }("");
                    require(success, "GovRewardee: fee transfer failed");
                }
            } catch {
                // if receiveFee is not implemented, transfer ether directly
                (bool success, ) = recipient.call{ value: amount }("");
                require(success, "GovRewardee: fee transfer failed");
            }
        } else {
            // if it is for sending reward or recipient is EOA, transfer ether directly
            (bool success, ) = recipient.call{ value: amount }("");
            require(success, "GovRewardee: fee transfer failed");
        }
        emit FeePaid(recipient, amount);
    }
}
