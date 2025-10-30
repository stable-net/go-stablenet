// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright 2025 The go-stablenet Authors
// This file is part of the go-stablenet library.
//
// The go-stablenet library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-stablenet is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-stablenet library. If not, see <http://www.gnu.org/licenses/>.

pragma solidity ^0.8.14;
/**
 * @title GovBase
 * @dev Abstract base contract for governance functionality
 * Provides common member management, proposal system, and approval tracking
 * Supports both EOA and CA as members
 */
abstract contract GovBase {
    // ========== Custom Errors ==========
    error AlreadyInitialized();
    error InvalidQuorum();
    error InvalidMemberAddress();
    error InvalidCall();
    error NotAMember();
    error InvalidProposal();
    error ProposalNotInVoting();
    error ProposalNotExecutable();
    error ProposalAlreadyExpired();
    error ProposalAlreadyInVoting();
    error ReentrantCall();
    error AnotherProposalIsActive();
    error MemberNotFound();
    error AlreadyApproved();
    error InsufficientApprovals();
    error NotProposer();
    error AlreadyAMember();
    error IndexOutOfBounds();

    // ========== Structs ==========
    struct Member {
        bool isActive;
        uint32 joinedAt;
    }

    struct Proposal {
        bytes32 actionType;
        ProposalStatus status;
        address proposer;
        uint256 memberVersion;
        uint256 requiredApprovals;
        uint256 votedBitmap;
        uint256 approved;
        uint256 rejected;
        uint256 createdAt;
        uint256 executedAt;
        bytes callData;
    }

    enum ProposalStatus {
        None,
        Voting, // approval, disapproval allowed
        Approved, // approval, disapproval disallowed, but execution allowed
        Executed, // all operation disallowed
        Cancelled, // all operation disallowed
        Expired, // all operation disallowed
        Failed, // all operation disallowed
        Rejected // all operation disallowed
    }

    // ========== State Variables ==========
    uint32 public quorum; // Required number of approvals (m of n)
    uint256 public proposalExpiry;
    mapping(address => Member) public members;
    mapping(uint256 => address[]) public versionedMemberList;
    uint256 public memberVersion;
    mapping(uint256 => Proposal) public proposals;
    uint256 public currentProposalId;
    uint256 private __reentrancyGuard; // Reentrancy protection
    uint256[42] private __gap; // Reserved storage space

    // ========== Events ==========
    event ProposalCreated(
        uint256 indexed proposalId,
        address indexed proposer,
        bytes32 actionType,
        uint256 memberVersion,
        uint256 requiredApprovals,
        bytes callData
    );

    event ProposalVoted(uint256 indexed proposalId, address indexed voter, bool approval, uint256 approved, uint256 rejected);

    event ProposalApproved(uint256 indexed proposalId, address indexed approver, uint256 approved, uint256 rejected);

    event ProposalRejected(uint256 indexed proposalId, address indexed rejector, uint256 approved, uint256 rejected);

    event ProposalExecuted(uint256 indexed proposalId, address indexed executor, bool success);

    event ExecutionFailed(uint256 indexed proposalId, address indexed executor, bytes errorData);

    event ProposalExpired(uint256 indexed proposalId, address indexed executor);

    event ProposalCancelled(uint256 indexed proposalId, address indexed canceller);

    event MemberAdded(address indexed member, uint256 totalMembers, uint32 newQuorum);

    event MemberRemoved(address indexed member, uint256 totalMembers, uint32 newQuorum);

    event MemberChanged(address indexed oldMember, address indexed newMember);

    event QuorumUpdated(uint32 oldQuorum, uint32 newQuorum);

    // ========== Modifiers ==========
    modifier onlyMe() {
        if (msg.sender != address(this)) revert InvalidCall();
        _;
    }

    modifier onlyMember() {
        if (!members[msg.sender].isActive) revert NotAMember();
        _;
    }

    modifier proposalExists() {
        if (currentProposalId == 0) revert InvalidProposal();
        _;
    }

    modifier proposalInVoting() {
        Proposal storage proposal = proposals[currentProposalId];
        if (proposal.status != ProposalStatus.Voting && proposal.status != ProposalStatus.Approved) {
            revert ProposalNotInVoting();
        }
        if (block.timestamp > proposal.createdAt + proposalExpiry) {
            // Update status before reverting (reentrancy protection)
            proposal.status = ProposalStatus.Expired;
            revert ProposalAlreadyExpired();
        }
        _;
    }

    modifier proposalExecutable() {
        Proposal storage proposal = proposals[currentProposalId];
        if (proposal.status != ProposalStatus.Approved) {
            revert ProposalNotExecutable();
        }
        if (block.timestamp > proposal.createdAt + proposalExpiry) {
            // Update status before reverting (reentrancy protection)
            proposal.status = ProposalStatus.Expired;
            revert ProposalAlreadyExpired();
        }
        _;
    }

    modifier nonReentrant() {
        if (__reentrancyGuard == 1) revert ReentrantCall();
        __reentrancyGuard = 1;
        _;
        __reentrancyGuard = 0;
    }

    modifier noActiveProposal() {
        Proposal storage proposal = proposals[currentProposalId];
        if (proposal.status == ProposalStatus.Voting || proposal.status == ProposalStatus.Approved) {
            if (block.timestamp > proposal.createdAt + proposalExpiry) {
                proposal.status = ProposalStatus.Expired;
            } else {
                revert AnotherProposalIsActive();
            }
        }
        _;
    }

    // ========== Internal Functions ==========
    function _createProposal(bytes32 actionType, bytes memory callData) internal onlyMember noActiveProposal returns (uint256 proposalId) {
        proposalId = ++currentProposalId;
        proposals[proposalId] = Proposal({
            actionType: actionType,
            status: ProposalStatus.Voting,
            proposer: msg.sender,
            memberVersion: memberVersion,
            votedBitmap: 0,
            requiredApprovals: quorum,
            approved: 0,
            rejected: 0,
            createdAt: uint32(block.timestamp),
            executedAt: 0,
            callData: callData
        });

        emit ProposalCreated(proposalId, msg.sender, actionType, memberVersion, quorum, callData);

        // Auto-approve by proposer
        _vote(true);
    }

    function _vote(bool approved) internal {
        Proposal storage proposal = proposals[currentProposalId];
        uint256 memberIndex = _getMemberIndex(msg.sender);
        if (memberIndex >= versionedMemberList[memberVersion].length) revert MemberNotFound();
        uint256 bit = 1 << memberIndex;
        if (proposal.votedBitmap & bit != 0) revert AlreadyApproved();

        // Effects (state changes first for reentrancy protection)
        proposal.votedBitmap |= bit;
        if (approved) {
            proposal.approved++;

            emit ProposalVoted(currentProposalId, msg.sender, true, proposal.approved, proposal.rejected);

            if (proposal.approved >= proposal.requiredApprovals) {
                proposal.status = ProposalStatus.Approved;
                emit ProposalApproved(currentProposalId, msg.sender, proposal.approved, proposal.rejected);
                _executeProposal(false);
            }
        } else {
            proposal.rejected++;

            emit ProposalVoted(currentProposalId, msg.sender, false, proposal.approved, proposal.rejected);

            // If rejected, mark as Rejected
            if (proposal.rejected > (versionedMemberList[proposal.memberVersion].length - proposal.requiredApprovals)) {
                proposal.status = ProposalStatus.Rejected;
                emit ProposalRejected(currentProposalId, msg.sender, proposal.approved, proposal.rejected);
            }
        }
    }

    function _executeProposal(bool setFailed) internal nonReentrant returns (bool success) {
        Proposal storage proposal = proposals[currentProposalId];

        // Checks
        if (proposal.status != ProposalStatus.Approved) revert ProposalNotExecutable();
        if (proposal.approved < proposal.requiredApprovals) revert InsufficientApprovals();
        // Check for expiry
        if (block.timestamp > proposal.createdAt + proposalExpiry) {
            proposal.status = ProposalStatus.Expired;
            emit ProposalExpired(currentProposalId, msg.sender);
            return false;
        }

        // Effects (state changes first for reentrancy protection)
        proposal.executedAt = uint32(block.timestamp);

        // Execute the call
        bytes memory err;
        if (proposal.callData.length > 0) {
            (success, err) = address(this).call(proposal.callData);
        } else {
            // For internal operations, derived contracts will handle execution
            success = _executeInternalAction(proposal.actionType);
        }

        // Update status based on result
        if (success) {
            proposal.status = ProposalStatus.Executed;
        } else {
            if (setFailed) {
                proposal.status = ProposalStatus.Failed;
            } else {
                proposal.status = ProposalStatus.Approved;
            }
            emit ExecutionFailed(currentProposalId, msg.sender, err);
        }

        // If success, status remains Executed
        // Interactions
        emit ProposalExecuted(currentProposalId, msg.sender, success);
    }

    // To be implemented by derived contracts for their specific actions
    function _executeInternalAction(bytes32 actionType) internal virtual returns (bool);

    function _getMemberIndex(address member) internal view returns (uint256) {
        for (uint256 i = 0; i < versionedMemberList[memberVersion].length; i++) {
            if (versionedMemberList[memberVersion][i] == member) {
                return i;
            }
        }
        return type(uint256).max;
    }

    function addMember(address _newMember, uint32 _newQuorum) external onlyMe {
        if (members[_newMember].isActive) revert AlreadyAMember();
        if (_newMember == address(0)) revert InvalidMemberAddress();
        if (_newQuorum == 0 || _newQuorum > versionedMemberList[memberVersion].length + 1) revert InvalidQuorum();
        members[_newMember] = Member({ isActive: true, joinedAt: uint32(block.timestamp) });

        _newVersionedMemberList();

        // Update quorum
        uint32 _oldQuorum = quorum;
        quorum = _newQuorum;

        emit MemberAdded(_newMember, versionedMemberList[memberVersion].length, _newQuorum);
        emit QuorumUpdated(_oldQuorum, _newQuorum);

        // Call hook for derived contracts
        _onMemberAdded(_newMember);
    }

    function _newVersionedMemberList() internal {
        uint256 newVersion = memberVersion + 1;
        for (uint256 i = 0; i < versionedMemberList[memberVersion].length; i++) {
            versionedMemberList[newVersion].push(versionedMemberList[memberVersion][i]);
        }
        memberVersion = newVersion;
    }

    function _newVersionedMemberListWithout(address removal) internal {
        uint256 newVersion = memberVersion + 1;
        for (uint256 i = 0; i < versionedMemberList[memberVersion].length; i++) {
            if (versionedMemberList[memberVersion][i] == removal) continue;
            versionedMemberList[newVersion].push(versionedMemberList[memberVersion][i]);
        }
        memberVersion = newVersion;
    }

    function _newVersionedMemberListChanging(address oldMember, address newMember) internal {
        uint256 newVersion = memberVersion + 1;
        for (uint256 i = 0; i < versionedMemberList[memberVersion].length; i++) {
            address member = versionedMemberList[memberVersion][i];
            if (member == oldMember) {
                member = newMember;
            }
            versionedMemberList[newVersion].push(member);
        }
        memberVersion = newVersion;
    }

    function removeMember(address member, uint32 newQuorum) external onlyMe {
        if (!members[member].isActive) revert NotAMember();
        if (newQuorum == 0 || newQuorum > versionedMemberList[memberVersion].length - 1) revert InvalidQuorum();
        members[member].isActive = false;

        // Remove from memberList
        _newVersionedMemberListWithout(member);

        // Update quorum
        uint32 oldQuorum = quorum;
        quorum = newQuorum;

        emit MemberRemoved(member, versionedMemberList[memberVersion].length, newQuorum);
        emit QuorumUpdated(oldQuorum, newQuorum);

        // Call hook for derived contracts
        _onMemberRemoved(member);
    }

    // ========== Hooks for Derived Contracts ==========
    // These hooks are called when members are added or removed
    // Derived contracts can override these to implement custom logic
    function _onMemberAdded(address member) internal virtual {}
    function _onMemberRemoved(address member) internal virtual {}
    function _onMemberChanged(address oldMember, address newMember) internal virtual {}

    // ========== Public Functions ==========
    function approveProposal() public onlyMember proposalExists proposalInVoting {
        _vote(true);
    }

    function disapproveProposal() public onlyMember proposalExists proposalInVoting {
        _vote(false);
    }

    function executeProposal() public onlyMember proposalExists proposalExecutable returns (bool) {
        return _executeProposal(false);
    }

    function executeWithFailure() public onlyMember proposalExists proposalExecutable returns (bool) {
        return _executeProposal(true);
    }

    function cancelProposal() public onlyMember proposalExists proposalInVoting {
        if (proposals[currentProposalId].proposer != msg.sender) revert NotProposer();
        if (proposals[currentProposalId].approved > 1 || proposals[currentProposalId].rejected > 0) revert ProposalAlreadyInVoting();

        proposals[currentProposalId].status = ProposalStatus.Cancelled;
        emit ProposalCancelled(currentProposalId, msg.sender);
    }

    // ========== Member Management ==========
    function proposeAddMember(address _newMember, uint32 _newQuorum) public onlyMember noActiveProposal returns (uint256) {
        bytes4 _selector = this.addMember.selector;
        bytes memory _encodedParams = abi.encode(_newMember, _newQuorum);
        return _createProposal(keccak256("ADD_MEMBER"), abi.encodePacked(_selector, _encodedParams));
    }

    function proposeRemoveMember(address _member, uint32 _newQuorum) public onlyMember noActiveProposal returns (uint256) {
        bytes4 _selector = this.removeMember.selector;
        bytes memory _encodedParams = abi.encode(_member, _newQuorum);
        return _createProposal(keccak256("REMOVE_MEMBER"), abi.encodePacked(_selector, _encodedParams));
    }

    function changeMember(address _newMember) public onlyMember noActiveProposal {
        if (members[_newMember].isActive) revert AlreadyAMember();

        members[_newMember] = members[msg.sender];
        delete members[msg.sender];

        _newVersionedMemberListChanging(msg.sender, _newMember);

        emit MemberChanged(msg.sender, _newMember);

        _onMemberChanged(msg.sender, _newMember);
    }

    // ========== View Functions ==========
    function getMemberCount() public view returns (uint256) {
        return versionedMemberList[memberVersion].length;
    }

    function getMemberAt(uint256 index) public view returns (address) {
        if (index >= versionedMemberList[memberVersion].length) revert IndexOutOfBounds();
        return versionedMemberList[memberVersion][index];
    }

    function getProposal() public view proposalExists returns (Proposal memory) {
        return proposals[currentProposalId];
    }

    function isProposalInVoting() public view returns (bool) {
        Proposal memory proposal = proposals[currentProposalId];
        if (proposal.status == ProposalStatus.Voting) {
            if (block.timestamp <= proposal.createdAt + proposalExpiry) {
                return true;
            }
        }
        return false;
    }

    function isProposalExecutable() public view returns (bool) {
        if (currentProposalId == 0) return false;
        Proposal memory proposal = proposals[currentProposalId];
        if (proposal.status != ProposalStatus.Approved) return false;
        if (block.timestamp > proposal.createdAt + proposalExpiry) return false;
        return proposal.approved >= proposal.requiredApprovals;
    }

    function hasApproved(address member) public view returns (bool) {
        uint256 memberIndex = _getMemberIndex(member);
        if (memberIndex >= versionedMemberList[memberVersion].length) return false;
        uint256 bit = 1 << memberIndex;
        return proposals[currentProposalId].votedBitmap & bit != 0;
    }
}
