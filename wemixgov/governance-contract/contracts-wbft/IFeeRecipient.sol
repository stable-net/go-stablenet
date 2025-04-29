// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.0;

interface IERC165 {
    function supportsInterface(bytes4 interfaceId) external view returns (bool);
}

interface IFeeRecipient is IERC165 {
    /* ========== FUNCTION ========== */
    function receiveFee(uint256 _amount) external payable;
    function withdrawFee(address _to, uint256 _amount) external;

    /* ========== EVENTS ========== */
    event ReceivedFee(address indexed from, uint256 amount);
    event SentFeeAmount(address indexed to, uint256 amount);
}
