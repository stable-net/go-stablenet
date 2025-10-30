// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright 2025 The stable-one Authors
// This file is part of the stable-one library.
//
// The stable-one library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The stable-one is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the stable-one library. If not, see <http://www.gnu.org/licenses/>.

pragma solidity ^0.8.14;

/**
 * @title GovBaseV2
 * @dev Abstract base contract for governance functionality (Version 2)
 * Provides common member management, proposal system, and approval tracking
 * Supports both EOA and CA as members
 *
 * @notice Derived contracts MUST implement:
 * - Action type constants (e.g., ACTION_ADD_MEMBER, ACTION_MINT) for their specific governance model
 * - _executeProposalAction() function to route and execute proposal actions
 *
 * @notice Each derived contract (GovMasterMinter, GovMinter, GovValidator) defines its own action set
 * based on its governance requirements and security model
 */
abstract contract GovBaseV2 {
    // ========== Custom Errors ==========
    error AlreadyAMember();
    error AlreadyApproved();
    error DuplicateMember();
    error IndexOutOfBounds();
    error InsufficientApprovals();
    error InvalidMemberAddress();
    error InvalidProposal();
    error InvalidProposalExpiry();
    error InvalidMemberVersion();
    error InvalidQuorum();
    error MemberIndexOverflow();
    error NotAMember();
    error NotProposer();
    error ProposalAlreadyInVoting();
    error ProposalNotExecutable();
    error ProposalNotInVoting();
    error ReentrantCall();
    error TooManyActiveProposals();
    error TooManyExecutionAttempts();

    // ========== Structs ==========
    struct Member {
        bool isActive;
        uint32 joinedAt;
    }

    /// @dev Proposal structure with optimized storage layout
    /// Fields are ordered by type and size for efficient storage packing
    ///
    /// Storage layout (7 slots + dynamic):
    /// - Slot 0: actionType (32 bytes)
    /// - Slot 1: memberVersion (32 bytes)
    /// - Slot 2: votedBitmap (32 bytes)
    /// - Slot 3: createdAt (32 bytes)
    /// - Slot 4: executedAt (32 bytes)
    /// - Slot 5: proposer (20 bytes) + requiredApprovals (4 bytes) + approved (4 bytes) + rejected (4 bytes) = 32 bytes
    /// - Slot 6: status (1 byte)
    /// - Slot 7+: callData (dynamic bytes)
    ///
    /// Field order benefits:
    /// - uint256 types grouped together for clear slot boundaries
    /// - Small types (address, uint32, enum) packed into single slot
    /// - Easy slot calculation: large types occupy 1 slot each, small types share Slot 5
    ///
    /// Gas optimization: uint32 used for vote counts (max 255 members, well within uint32 range)
    /// Storage optimization: All small types packed into Slot 5 (32 bytes exactly)
    struct Proposal {
        bytes32 actionType; // Slot 0
        uint256 memberVersion; // Slot 1
        uint256 votedBitmap; // Slot 2 - must be uint256 for bitmap operations
        uint256 createdAt; // Slot 3 - must be uint256 for timestamp
        uint256 executedAt; // Slot 4 - must be uint256 for timestamp
        address proposer; // Slot 5 (20 bytes)
        uint32 requiredApprovals; // Slot 5 (4 bytes)
        uint32 approved; // Slot 5 (4 bytes)
        uint32 rejected; // Slot 5 (4 bytes)
        ProposalStatus status; // Slot 6 (1 byte)
        bytes callData; // Slot 7+ (dynamic)
    }

    enum ProposalStatus {
        None, // No proposal exists
        Voting, // Voting in progress - voting and execution allowed
        Approved, // Quorum reached - voting and execution allowed
        Executed, // Successfully executed - all operations disallowed
        Cancelled, // Cancelled by proposer - all operations disallowed
        Expired, // Expired due to timeout - all operations disallowed
        Failed, // Execution failed - all operations disallowed
        Rejected // Rejected by votes - all operations disallowed
    }

    /// @notice Execution check result for canExecuteProposal function
    /// @dev Provides detailed reason why a proposal can or cannot be executed
    enum ExecutionCheckResult {
        Executable, // Proposal can be executed
        InvalidProposalId, // Proposal ID doesn't exist (0 or > currentProposalId)
        NotApproved, // Proposal not in Approved status (may be Voting, Executed, Cancelled, Expired, Failed, Rejected)
        Expired, // Proposal has expired (block.timestamp > createdAt + proposalExpiry)
        TooManyAttempts // Retry limit reached (executionCount >= MAX_RETRY_COUNT)
    }

    // ========== Constants ==========
    uint256 public constant MAX_MEMBER_INDEX = 255; // Maximum safe index for bit shifting (1 << 255)
    uint256 public constant MAX_RETRY_COUNT = 3; // Maximum number of execution attempts per proposal (includes initial + retries)
    uint256 public constant INITIAL_MEMBER_VERSION = 1; // Initial version number for member snapshots
    uint256 public constant MAX_ACTIVE_PROPOSALS_PER_MEMBER = 3; // Maximum concurrent active proposals per member

    // Member management action types
    bytes32 public constant ACTION_ADD_MEMBER = keccak256("ACTION_ADD_MEMBER");
    bytes32 public constant ACTION_REMOVE_MEMBER = keccak256("ACTION_REMOVE_MEMBER");
    bytes32 public constant ACTION_CHANGE_MEMBER = keccak256("ACTION_CHANGE_MEMBER");
    bytes32 public constant ACTION_CHANGE_QUORUM = keccak256("ACTION_CHANGE_QUORUM");

    // ========== State Variables ==========
    // Value types - uint256 (32 bytes each, 1 slot per variable)
    uint256 public proposalExpiry; // Slot 0: Set once during initialization, cannot be changed
    uint256 public memberVersion = INITIAL_MEMBER_VERSION; // Slot 1
    uint256 public currentProposalId; // Slot 2
    uint256 private __reentrancyGuard; // Slot 3: Reentrancy protection

    // Small value types - uint32 (4 bytes, can be packed with other small types)
    uint32 public quorum; // Slot 4 (0-3 bytes): Required number of approvals (m of n)
    // Note: Slot 4 has 28 bytes remaining for future small types (address, uint32, etc.)

    // Mappings - each mapping occupies 1 slot for metadata
    mapping(address => Member) public members; // Slot 5
    mapping(uint256 => address[]) public versionedMemberList; // Slot 6
    mapping(uint256 => Proposal) public proposals; // Slot 7
    mapping(uint256 => mapping(address => uint32)) internal _memberIndexByVersion; // Slot 8: index + 1 snapshot per version
    mapping(uint256 => uint32) internal _quorumByVersion; // Slot 9: quorum snapshot per member version
    mapping(uint256 => uint256) public proposalExecutionCount; // Slot 10: Execution attempt count for each proposal
    mapping(address => uint256) public memberActiveProposalCount; // Slot 11: Track active proposals per member

    // Reserved storage for future upgrades
    uint256[38] private __gap; // Slot 12-49: Reserved storage space (reduced by 3 for proposalExecutionCount, memberActiveProposalCount, quorumByVersion)

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
    event ProposalFailed(uint256 indexed proposalId, address indexed executor, bytes reason);
    event ProposalExpired(uint256 indexed proposalId, address indexed executor);
    event ProposalCancelled(uint256 indexed proposalId, address indexed canceller);
    event MemberAdded(address indexed member, uint256 totalMembers, uint32 newQuorum);
    event MemberRemoved(address indexed member, uint256 totalMembers, uint32 newQuorum);
    event MemberChanged(address indexed oldMember, address indexed newMember);
    event QuorumUpdated(uint32 oldQuorum, uint32 newQuorum);

    // ========== Modifiers ==========

    /// @dev Check if caller is currently active member (for proposal creation)
    modifier onlyActiveMember() {
        Member storage member = members[msg.sender];
        if (!member.isActive) revert NotAMember();
        _;
    }

    /// @dev Check if caller was a member at proposal's member version (for proposal actions)
    /// @param proposalId The proposal ID to check membership against
    modifier onlyProposalMember(uint256 proposalId) {
        _checkProposalMembership(proposalId);
        _;
    }

    modifier nonReentrant() {
        _nonReentrantBefore();
        _;
        _nonReentrantAfter();
    }

    // ============================================================
    // 1. INITIALIZATION
    // ============================================================

    /// @notice Governance initialization is handled by genesis configuration
    /// @dev Initial state (members, quorum, proposalExpiry) is set via genesis params
    /// @dev See systemcontracts/gov_base.go for genesis initialization implementation
    ///
    /// Genesis initialization process:
    /// 1. Parse genesis params (members, quorum, expiry, memberVersion)
    /// 2. Directly set storage slots:
    ///    - versionedMemberList[1]: Initial member addresses
    ///    - memberIndexByVersion[1][address]: Member index mappings
    ///    - members[address]: Member status (isActive, joinedAt)
    ///    - quorum, proposalExpiry: Governance parameters
    ///    - memberVersion: Always starts at INITIAL_MEMBER_VERSION (1)
    /// 3. No function call required - pure storage initialization

    // ============================================================
    // 2. MODIFIERS (Internal Check Functions)
    // ============================================================

    // ========== Internal Functions for Modifiers ==========

    /// @dev Internal function to validate membership at proposal's member version
    /// @param proposalId The proposal ID to check membership against
    /// @notice This allows members who were part of the governance when proposal was created
    ///         to continue interacting with it, even if they were later removed
    function _checkProposalMembership(uint256 proposalId) internal view {
        Proposal storage proposal = proposals[_validateProposalId(proposalId)];
        uint256 memberIndex = _getMemberIndexAtVersion(msg.sender, proposal.memberVersion);
        address[] storage snapshot = versionedMemberList[proposal.memberVersion];

        if (memberIndex >= snapshot.length) revert NotAMember();
        if (snapshot[memberIndex] != msg.sender) revert NotAMember();
    }

    /// @dev Internal reentrancy guard initialization (OpenZeppelin pattern)
    /// @notice Sets guard to prevent reentrant calls. Must be paired with _nonReentrantAfter()
    /// @custom:revert ReentrantCall if guard is already set (reentrant call detected)
    function _nonReentrantBefore() internal {
        if (__reentrancyGuard == 1) revert ReentrantCall();
        __reentrancyGuard = 1;
    }

    /// @dev Internal reentrancy guard cleanup (OpenZeppelin pattern)
    /// @notice Resets guard to allow subsequent calls. Executes even if function body reverts.
    function _nonReentrantAfter() internal {
        __reentrancyGuard = 0;
    }

    // ============================================================
    // 3. PUBLIC FUNCTIONS - Proposal Operations
    // ============================================================

    // ========== Public Functions ==========

    /// @notice Vote YES on a proposal
    /// @dev If this vote reaches quorum, proposal will execute automatically
    /// @dev If execution fails, use executeProposal() to retry
    /// @param proposalId The proposal ID to approve
    function approveProposal(uint256 proposalId) public onlyProposalMember(proposalId) {
        _vote(proposalId, true, true);
    }

    /// @notice Vote NO on a proposal
    /// @dev Reject the specified proposal
    /// @param proposalId The proposal ID to reject
    function disapproveProposal(uint256 proposalId) public onlyProposalMember(proposalId) {
        _vote(proposalId, false, false);
    }

    /// @notice Execute or retry an approved proposal
    /// @param proposalId The proposal ID to execute
    /// @return success True if execution succeeded
    /// @dev Automatically tracks execution attempts and enforces retry limit (MAX_RETRY_COUNT = 3)
    /// @dev If execution fails, proposal remains in Approved status and can be retried by calling again
    /// @dev Reverts with TooManyExecutionAttempts after 3 failed attempts
    function executeProposal(uint256 proposalId) public onlyProposalMember(proposalId) returns (bool) {
        return _executeProposal(proposalId, false);
    }

    /// @notice Execute an approved proposal with terminal failure (no retry)
    /// @param proposalId The proposal ID to execute
    /// @return success True if execution succeeded
    /// @dev If execution fails, proposal is marked as Failed and cannot be retried
    function executeWithFailure(uint256 proposalId) public onlyProposalMember(proposalId) returns (bool) {
        return _executeProposal(proposalId, true);
    }

    /// @notice Cancel a proposal that is still in voting phase
    /// @dev Can only be called by the original proposer before other members vote
    /// @dev Proposer's automatic approval (approved=1) doesn't prevent cancellation
    /// @dev Proposal must be in Voting status and no other member should have voted yet
    /// @param proposalId The ID of the proposal to cancel
    function cancelProposal(uint256 proposalId) public onlyProposalMember(proposalId) {
        // proposalId validation is already checked in modifier

        Proposal storage proposal = proposals[proposalId];

        // Validate proposal is in Voting status
        if (proposal.status != ProposalStatus.Voting) revert ProposalNotInVoting();

        // Only proposer can cancel
        if (proposal.proposer != msg.sender) revert NotProposer();

        // Cannot cancel if other members have voted
        // Proposer automatically approves on creation (approved=1)
        // approved > 1 means at least one other member approved
        // rejected > 0 means at least one member rejected
        if (proposal.approved > 1 || proposal.rejected > 0) {
            revert ProposalAlreadyInVoting();
        }

        _finalizeProposal(proposalId, ProposalStatus.Cancelled);
        emit ProposalCancelled(proposalId, msg.sender);
    }

    /// @notice Manually expire a proposal that has passed its expiry time
    /// @dev Can be called by members to clean up expired proposals
    /// @param proposalId The proposal ID to expire
    /// @return success True if the proposal was expired, false if it cannot be expired
    function expireProposal(uint256 proposalId) public onlyProposalMember(proposalId) returns (bool success) {
        // proposalId validation is already checked in modifier

        Proposal storage proposal = proposals[proposalId];

        // Can only expire Voting or Approved proposals
        if (proposal.status != ProposalStatus.Voting && proposal.status != ProposalStatus.Approved) {
            return false;
        }

        // Check if actually expired
        if (block.timestamp <= proposal.createdAt + proposalExpiry) {
            return false;
        }

        // Update state
        _finalizeProposal(proposalId, ProposalStatus.Expired);
        emit ProposalExpired(proposalId, msg.sender);
        return true;
    }

    // ============================================================
    // 4. PUBLIC FUNCTIONS - Member Management Proposals
    // ============================================================

    // ========== Member Management ==========
    // NOTE: Governance proposal functions moved to derived contracts
    // - GovMasterMinter: Implements self-governance + hierarchical control over GovMinter
    // - GovMinter: No member management functions (security: prevent collusion attacks)
    //
    // Derived contracts implement their own proposal functions based on their governance model:
    // - proposeAddMember() / proposeRemoveMember() / proposeSetProposalExpiry() in GovMasterMinter
    // - Operational proposal functions only in GovMinter (proposeMint, proposeBurn, etc.)

    /// @notice Propose to add a new member to governance
    /// @param newMember Address of the new member to add
    /// @param newQuorum New quorum value after adding member
    /// @return proposalId The ID of the created proposal
    /// @dev Requires governance approval to execute
    function proposeAddMember(address newMember, uint32 newQuorum) public onlyActiveMember returns (uint256 proposalId) {
        if (newMember == address(0)) revert InvalidMemberAddress();
        if (members[newMember].isActive) revert AlreadyAMember();

        uint256 memberCount = versionedMemberList[memberVersion].length;
        if (memberCount >= MAX_MEMBER_INDEX) revert MemberIndexOverflow();

        uint256 newMemberCount = memberCount + 1;
        if (newMemberCount == 1) {
            if (newQuorum != 1) revert InvalidQuorum();
        } else {
            if (newQuorum < 2 || newQuorum > newMemberCount) revert InvalidQuorum();
        }

        bytes memory callData = abi.encode(newMember, newQuorum);
        return _createProposal(ACTION_ADD_MEMBER, callData);
    }

    /// @notice Propose to remove an existing member from governance
    /// @param member Address of the member to remove
    /// @param newQuorum New quorum value after removing member
    /// @return proposalId The ID of the created proposal
    /// @dev Requires governance approval to execute
    function proposeRemoveMember(address member, uint32 newQuorum) public onlyActiveMember returns (uint256 proposalId) {
        if (!members[member].isActive) revert NotAMember();

        uint256 memberCount = versionedMemberList[memberVersion].length;
        if (memberCount <= 1) revert InvalidQuorum();

        uint256 newMemberCount = memberCount - 1;
        if (newMemberCount == 1) {
            if (newQuorum != 1) revert InvalidQuorum();
        } else {
            if (newQuorum < 2 || newQuorum > newMemberCount) revert InvalidQuorum();
        }

        bytes memory callData = abi.encode(member, newQuorum);
        return _createProposal(ACTION_REMOVE_MEMBER, callData);
    }

    /// @notice Propose to change the quorum requirement for governance proposals
    /// @param newQuorum New quorum value to set
    /// @return proposalId The ID of the created proposal
    /// @dev Requires governance approval to execute
    /// @dev Quorum must be at least 2 and at most the current number of members
    function proposeChangeQuorum(uint32 newQuorum) public onlyActiveMember returns (uint256 proposalId) {
        uint256 memberCount = versionedMemberList[memberVersion].length;

        // Validate quorum requirements
        if (memberCount == 1) {
            if (newQuorum != 1) revert InvalidQuorum();
        } else {
            if (newQuorum < 2 || newQuorum > memberCount) revert InvalidQuorum();
        }

        bytes memory callData = abi.encode(newQuorum);
        return _createProposal(ACTION_CHANGE_QUORUM, callData);
    }

    /// @notice Allow active member to change their own address (self-service key rotation)
    /// @param newMember New member address to replace msg.sender's address
    ///
    /// @dev Member change flow (Checks-Effects-Interactions):
    /// 1. Checks: Validate new address (non-zero, not same, not already active)
    /// 2. Effects:
    ///    - Find and replace msg.sender with newMember in current version's versionedMemberList
    ///    - Transfer memberIndexByVersion mapping from old to new address
    ///    - Update members mapping: activate newMember, deactivate msg.sender
    /// 3. Interactions: Call _onMemberChanged hook for derived contract logic
    ///
    /// @dev Access control: onlyActiveMember modifier ensures msg.sender is active
    /// @dev Version behavior: Does NOT increment memberVersion, modifies current version in-place
    /// @dev Index preservation: newMember inherits the exact same index position as msg.sender
    function changeMember(address newMember) public onlyActiveMember {
        address oldMember = msg.sender;

        // CHECKS: Validate new address requirements
        if (newMember == address(0)) revert InvalidMemberAddress();
        if (oldMember == newMember) revert InvalidMemberAddress();
        if (members[newMember].isActive) revert AlreadyAMember();

        // Get current version's member list (no new version created)
        uint256 currentVersion = memberVersion;
        address[] storage currentMemberList = versionedMemberList[currentVersion];
        uint256 length = currentMemberList.length;

        // Find and replace oldMember with newMember in current version's list
        bool found = false;
        for (uint256 i = 0; i < length; i++) {
            if (currentMemberList[i] == oldMember) {
                // Replace address in-place, preserving array position
                currentMemberList[i] = newMember;

                // Transfer index mapping from old to new address
                // Note: Index value remains the same (i+1), only the key changes
                uint32 memberIndex = _memberIndexByVersion[currentVersion][oldMember];
                _memberIndexByVersion[currentVersion][newMember] = memberIndex;
                delete _memberIndexByVersion[currentVersion][oldMember];
                found = true;
                break;
            }
        }

        // This should never happen if the contract state is consistent
        // (onlyActiveMember ensures msg.sender is active, and active members must exist in versionedMemberList)
        if (!found) revert NotAMember();

        // Update member states
        // Activate new member with current timestamp
        members[newMember] = Member({ isActive: true, joinedAt: uint32(block.timestamp) });
        // Deactivate old member (preserves joinedAt for historical record)
        members[oldMember].isActive = false;

        // Emit event for off-chain tracking
        emit MemberChanged(oldMember, newMember);

        // Call hook for derived contract-specific logic
        // WARNING: Derived contracts must avoid external calls to prevent reentrancy
        _onMemberChanged(oldMember, newMember);
    }

    // ============================================================
    // 5. VIEW FUNCTIONS - Proposal Queries
    // ============================================================

    /// @notice Get proposal details by proposal ID
    /// @param proposalId The proposal ID to query (must be between 1 and currentProposalId)
    /// @return Proposal struct containing all proposal data
    /// @custom:revert InvalidProposal if proposalId is 0 or exceeds currentProposalId
    function getProposal(uint256 proposalId) public view returns (Proposal memory) {
        proposalId = _validateProposalId(proposalId);
        return proposals[proposalId];
    }

    /// @notice Check if a proposal is in voting phase (Voting status, not yet approved)
    /// @dev Note: Proposals in Approved status can still receive votes, but this returns false to distinguish phases
    /// @param proposalId The proposal ID to check (must be between 1 and currentProposalId)
    /// @return True if proposal is in Voting status and not expired
    /// @custom:revert InvalidProposal if proposalId is 0 or exceeds currentProposalId
    function isProposalInVoting(uint256 proposalId) public view returns (bool) {
        Proposal memory proposal = proposals[_validateProposalId(proposalId)];
        if (proposal.status == ProposalStatus.Voting) {
            if (block.timestamp <= proposal.createdAt + proposalExpiry) {
                return true;
            }
        }
        return false;
    }

    /// @notice Check if a proposal is ready for execution
    /// @dev Validates: Approved status, not expired, and quorum satisfied
    /// @param proposalId The proposal ID to check (must be between 1 and currentProposalId)
    /// @return True if proposal can be executed immediately
    /// @custom:revert InvalidProposal if proposalId is 0 or exceeds currentProposalId
    function isProposalExecutable(uint256 proposalId) public view returns (bool) {
        Proposal memory proposal = proposals[_validateProposalId(proposalId)];
        if (proposal.status != ProposalStatus.Approved) return false;
        if (block.timestamp > proposal.createdAt + proposalExpiry) return false;
        return proposal.approved >= proposal.requiredApprovals;
    }

    /// @notice Check if proposal can be executed with detailed failure reason
    /// @dev This is a view function - state may change between check and execution (race condition possible)
    /// @dev If proposal is expired but status is still Approved, returns Expired (state not modified in view function)
    /// @dev Checks are performed in order: proposalId validity → status → expiry → retry limit
    /// @param proposalId Proposal ID to check
    /// @return result Execution check result indicating why execution is possible or not
    /// @return attemptsLeft Number of execution attempts remaining (0 if not executable)
    function canExecuteProposal(uint256 proposalId) external view returns (ExecutionCheckResult result, uint256 attemptsLeft) {
        // Check 1: Validate proposalId exists
        if (proposalId == 0 || proposalId > currentProposalId) {
            return (ExecutionCheckResult.InvalidProposalId, 0);
        }

        Proposal storage proposal = proposals[proposalId];

        // Check 2: Validate status is Approved
        if (proposal.status != ProposalStatus.Approved) {
            return (ExecutionCheckResult.NotApproved, 0);
        }

        // Check 3: Validate not expired
        if (block.timestamp > proposal.createdAt + proposalExpiry) {
            return (ExecutionCheckResult.Expired, 0);
        }

        // Check 4: Validate retry limit not exceeded
        uint256 executionCount = proposalExecutionCount[proposalId];
        if (executionCount >= MAX_RETRY_COUNT) {
            return (ExecutionCheckResult.TooManyAttempts, 0);
        }

        // All checks passed - proposal can be executed
        return (ExecutionCheckResult.Executable, MAX_RETRY_COUNT - executionCount);
    }

    /// @notice Check if a member has voted on a specific proposal
    /// @dev Uses bitmap tracking with historical member version snapshot
    /// @dev Returns false for non-members or members who joined after proposal creation
    /// @param member The address to check for voting status
    /// @param proposalId The proposal ID to query
    /// @return True if member has voted, false otherwise
    /// @custom:revert InvalidProposal if proposalId is 0 or exceeds currentProposalId
    function hasApproved(address member, uint256 proposalId) public view returns (bool) {
        proposalId = _validateProposalId(proposalId);
        Proposal storage proposal = proposals[proposalId];
        uint256 memberIndex = _getMemberIndexAtVersion(member, proposal.memberVersion);
        if (memberIndex >= versionedMemberList[proposal.memberVersion].length) return false;
        if (memberIndex > MAX_MEMBER_INDEX) return false; // Safe fallback for overflow
        // Shift operation is correct: 1 << memberIndex creates a bit mask
        // solhint-disable-next-line incorrect-shift
        // forge-lint: disable-next-line(incorrect-shift)
        uint256 bit = 1 << memberIndex;
        return proposal.votedBitmap & bit != 0;
    }

    // ============================================================
    // 6. VIEW FUNCTIONS - Member Queries
    // ============================================================

    // ========== View Functions ==========

    /// @notice Get the number of members at a specific governance version
    /// @param targetVersion The member version to query (must be between 1 and current memberVersion)
    /// @return The number of members in the specified version
    /// @custom:revert InvalidMemberVersion if targetVersion is 0 or exceeds current memberVersion
    function getMemberCount(uint256 targetVersion) public view returns (uint256) {
        uint256 version = _validateMemberVersion(targetVersion);
        return versionedMemberList[version].length;
    }

    /// @notice Get member address at specific index in a governance version snapshot
    /// @param targetVersion The member version to query (must be between 1 and current memberVersion)
    /// @param index The zero-based index in the member list (must be < snapshot length)
    /// @return The member address at the specified index
    /// @custom:revert InvalidMemberVersion if targetVersion is 0 or exceeds memberVersion
    /// @custom:revert IndexOutOfBounds if index >= snapshot.length
    function getMemberAt(uint256 targetVersion, uint256 index) public view returns (address) {
        uint256 version = _validateMemberVersion(targetVersion);
        address[] storage snapshot = versionedMemberList[version];
        if (index >= snapshot.length) revert IndexOutOfBounds();
        return snapshot[index];
    }

    /// @notice Check if an address was a governance member at a specific version
    /// @dev Uses historical member version snapshot (1-based indexing: 0 = not member, >0 = member)
    /// @param account The address to check for membership status
    /// @param targetVersion The governance version to query (must be between 1 and current memberVersion)
    /// @return True if account was a member at targetVersion
    /// @custom:revert InvalidMemberVersion if targetVersion is 0 or exceeds current memberVersion
    function isMember(address account, uint256 targetVersion) public view returns (bool) {
        uint256 version = _validateMemberVersion(targetVersion);
        return _memberIndexByVersion[version][account] != 0;
    }

    /// @notice Get the quorum requirement (minimum approvals needed) for a specific governance version
    /// @dev Retrieves historical quorum value from version snapshot for audit and proposal validation
    /// @param targetVersion The governance version to query (must be between 1 and current memberVersion)
    /// @return The quorum value for the specified version (minimum 1, maximum member count)
    /// @custom:revert InvalidMemberVersion if targetVersion is 0 or exceeds current memberVersion
    /// @custom:revert InvalidQuorum if quorum snapshot is 0 (defense-in-depth check, should never happen)
    function getQuorum(uint256 targetVersion) public view returns (uint32) {
        uint256 version = _validateMemberVersion(targetVersion);
        uint32 snapshot = _quorumByVersion[version];
        if (snapshot == 0) revert InvalidQuorum();
        return snapshot;
    }

    /// @notice Get the number of active (non-terminal) proposals for a member
    /// @dev This count enforces the MAX_ACTIVE_PROPOSALS_PER_MEMBER limit to prevent proposal spam
    /// @param member The member address to query for active proposal count
    /// @return The number of active proposals (range: 0 to MAX_ACTIVE_PROPOSALS_PER_MEMBER)
    function getMemberActiveProposalCount(address member) public view returns (uint256) {
        return memberActiveProposalCount[member];
    }

    /// @notice Check if a member can create a new proposal
    /// @dev Verifies whether the member has available capacity (count < MAX_ACTIVE_PROPOSALS_PER_MEMBER)
    /// @param member The member address to check for proposal creation eligibility
    /// @return True if member can create a new proposal, false if at capacity limit
    function canCreateProposal(address member) public view returns (bool) {
        return memberActiveProposalCount[member] < MAX_ACTIVE_PROPOSALS_PER_MEMBER;
    }

    // ============================================================
    // 7. INTERNAL FUNCTIONS - Core Workflow
    // ============================================================

    /// @dev Create a new governance proposal
    /// @param actionType The type of action to execute (e.g., ACTION_ADD_MEMBER, ACTION_MINT)
    /// @param callData ABI-encoded parameters for the action
    /// @return proposalId The unique identifier for the created proposal
    /// @notice Must be called from derived contract's proposal functions
    /// @notice Multiple proposals can be active simultaneously
    /// @notice Emits ProposalCreated event upon successful creation
    ///
    /// Proposal lifecycle:
    /// 1. Create proposal (this function) - proposer auto-approves
    /// 2. Members vote via vote(proposalId, approved) or voteAndExecute(proposalId, approved)
    /// 3. Reaches quorum → status changes to Approved
    /// 4. Execute via executeProposal(proposalId) or auto-execute on final vote
    ///
    /// Member version snapshot:
    /// - Proposal captures current memberVersion at creation time
    /// - Future member changes don't affect existing proposal voting rights
    /// - Ensures consistent voting rules throughout proposal lifecycle
    ///
    /// Auto-approval behavior:
    /// - Proposer automatically approves their own proposal
    /// - Auto-execution enabled (executes immediately if quorum = 1)
    /// - Gas consideration: If quorum = 1, proposal executes within this transaction
    function _createProposal(bytes32 actionType, bytes memory callData) internal onlyActiveMember returns (uint256 proposalId) {
        // Validate proposal expiry is configured (set during initialization)
        if (proposalExpiry == 0) revert InvalidProposalExpiry();

        // Check proposal spam limit
        if (memberActiveProposalCount[msg.sender] >= MAX_ACTIVE_PROPOSALS_PER_MEMBER) {
            revert TooManyActiveProposals();
        }

        // Generate unique proposal ID (starts from 1, 0 = non-existent)
        proposalId = ++currentProposalId;

        proposals[proposalId] = Proposal({
            actionType: actionType,
            status: ProposalStatus.Voting,
            proposer: msg.sender,
            memberVersion: memberVersion, // Snapshot current member composition
            votedBitmap: 0, // No votes recorded yet (proposer auto-approves after)
            requiredApprovals: quorum, // Snapshot current quorum requirement
            approved: 0, // Proposer approval added after creation
            rejected: 0, // No rejections yet
            createdAt: block.timestamp, // Record creation time for expiry check
            executedAt: 0, // Not executed yet
            callData: callData // Action parameters
        });

        emit ProposalCreated(proposalId, msg.sender, actionType, memberVersion, quorum, callData);

        // Increment active proposal count for proposer
        memberActiveProposalCount[msg.sender]++;

        // Auto-approve by proposer with auto-execution enabled
        _vote(proposalId, true, true);
    }

    /// @dev Internal voting function for proposal approval or rejection
    /// @param proposalId The proposal ID to vote on
    /// @param approved True to approve, false to reject the proposal
    /// @param autoExecute Whether to automatically execute if quorum is reached
    /// @notice Updates vote bitmap, counts, and emits appropriate events
    /// @notice Auto-executes proposal if approved reaches quorum (when autoExecute=true)
    /// @notice Marks proposal as Rejected if rejection threshold is exceeded
    ///
    /// Voting mechanics:
    /// - Each member can vote once per proposal (tracked via bitmap)
    /// - Approval threshold: approved >= requiredApprovals
    /// - Rejection threshold: rejected > (total members - requiredApprovals)
    ///   Example: 5 members, quorum=3 → max rejections = 2 (if >2, approval impossible)
    ///
    /// State changes (Checks-Effects-Interactions pattern):
    /// 1. Validate member eligibility and check duplicate voting
    /// 2. Update votedBitmap and vote counts (state changes first)
    /// 3. Emit events
    /// 4. Execute proposal if conditions met (external call last)
    ///
    /// Gas optimization: Caches frequently accessed values in local variables
    /// Type safety: approved/rejected are uint32 in storage, auto-converted to uint256 in events
    function _vote(uint256 proposalId, bool approved, bool autoExecute) internal {
        // Validate proposal exists
        if (proposalId == 0 || proposalId > currentProposalId) revert InvalidProposal();

        Proposal storage proposal = proposals[proposalId];

        // Validate proposal status (must be Voting or Approved)
        if (proposal.status != ProposalStatus.Voting && proposal.status != ProposalStatus.Approved) {
            revert ProposalNotInVoting();
        }

        // Check and enforce expiry
        // Note: Vote attempt on expired proposal will trigger expiry state transition
        // instead of recording the vote. This maintains state changes and cleanup.
        if (block.timestamp > proposal.createdAt + proposalExpiry) {
            _finalizeProposal(proposalId, ProposalStatus.Expired);
            emit ProposalExpired(proposalId, msg.sender);

            // Important: Return instead of revert to maintain state changes
            // - Proposal status updated to Expired (persisted)
            // - Cleanup hook executed (reservation released)
            // - ProposalExpired event emitted (off-chain monitoring)
            // - Voter's transaction succeeds but no vote recorded
            // - Subsequent operations will fail with ProposalNotInVoting
            return;
        }

        // Validate voter membership at proposal's member version snapshot
        uint256 proposalMemberVersion = proposal.memberVersion;
        uint256 memberIndex = _getMemberIndexAtVersion(msg.sender, proposalMemberVersion);
        address[] storage snapshot = versionedMemberList[proposalMemberVersion];

        // Validation checks
        if (memberIndex >= snapshot.length) revert NotAMember();
        if (memberIndex > MAX_MEMBER_INDEX) revert MemberIndexOverflow();

        // Check duplicate voting using bitmap
        // Shift operation is correct: 1 << memberIndex creates a bit mask
        // solhint-disable-next-line incorrect-shift
        // forge-lint: disable-next-line(incorrect-shift)
        uint256 bit = 1 << memberIndex;
        if (proposal.votedBitmap & bit != 0) revert AlreadyApproved();

        // Effects (state changes first for reentrancy protection)
        proposal.votedBitmap |= bit;

        if (approved) {
            uint32 newApproved = proposal.approved + 1;
            proposal.approved = newApproved;

            // Note: newApproved (uint32) is auto-converted to uint256 in event emission
            emit ProposalVoted(proposalId, msg.sender, true, newApproved, proposal.rejected);

            // Check if quorum reached
            if (newApproved >= proposal.requiredApprovals) {
                // Only update status and emit event if not already Approved
                // (handles retry votes after auto-execution failure)
                if (proposal.status != ProposalStatus.Approved) {
                    proposal.status = ProposalStatus.Approved;
                    emit ProposalApproved(proposalId, msg.sender, newApproved, proposal.rejected);
                }

                // Interactions (external calls last)
                if (autoExecute) {
                    _executeProposal(proposalId, false);
                }
            }
        } else {
            // Rejection path: Update storage and cache value for reuse
            uint32 newRejected = proposal.rejected + 1;
            proposal.rejected = newRejected;

            // Note: newRejected (uint32) is auto-converted to uint256 in event emission
            emit ProposalVoted(proposalId, msg.sender, false, proposal.approved, newRejected);

            // Calculate rejection threshold: maximum rejections before approval becomes impossible
            // Example: 5 members, quorum=3 → maxRejections=2 (if rejected>2, cannot reach quorum)
            uint32 maxRejections = uint32(snapshot.length) - proposal.requiredApprovals;
            if (newRejected > maxRejections) {
                _finalizeProposal(proposalId, ProposalStatus.Rejected);
                emit ProposalRejected(proposalId, msg.sender, proposal.approved, newRejected);
            }
        }
    }

    /// @dev Execute approved proposal with automatic retry tracking
    /// @param proposalId The proposal ID to execute
    /// @param markFailedOnError Failure handling mode:
    ///        - true: Terminal execution, proposal marked Failed on error (used by executeWithFailure)
    ///        - false: Retry-enabled, proposal stays Approved on error (used by executeProposal, auto-execute)
    /// @return success True if execution succeeded
    ///
    /// @notice Execution flow (Checks-Effects-Interactions):
    /// 1. Checks: Validate proposal state, quorum, expiry, and retry limit
    /// 2. Effects: Increment execution count before external call
    /// 3. Interactions: Execute action via _executeInternalAction
    /// 4. Effects: Update proposal status based on result
    ///
    /// @notice Retry limit: MAX_RETRY_COUNT (3 attempts total)
    /// - markFailedOnError=false: Enforces limit, reverts with TooManyExecutionAttempts
    /// - markFailedOnError=true: Bypasses limit for terminal execution
    ///
    /// @notice Reentrancy protection: nonReentrant modifier prevents same-function reentry
    /// @notice Security consideration: State changes occur after _executeInternalAction call
    ///         Derived contracts must avoid external calls in _executeCustomAction to prevent
    ///         cross-function reentrancy attacks
    function _executeProposal(uint256 proposalId, bool markFailedOnError) internal nonReentrant returns (bool success) {
        // CHECKS: Validate proposal state
        if (proposalId == 0 || proposalId > currentProposalId) revert InvalidProposal();

        Proposal storage proposal = proposals[proposalId];

        if (proposal.status != ProposalStatus.Approved) revert ProposalNotExecutable();
        if (proposal.approved < proposal.requiredApprovals) revert InsufficientApprovals();

        // Handle expiry gracefully (update state and return false)
        // Note: Execution attempt on expired proposal will trigger expiry state transition.
        // Returns false to indicate execution did not succeed, but state changes persist.
        if (block.timestamp > proposal.createdAt + proposalExpiry) {
            _finalizeProposal(proposalId, ProposalStatus.Expired);
            emit ProposalExpired(proposalId, msg.sender);
            return false;
        }

        // Check retry limit (bypassed for terminal execution)
        if (!markFailedOnError && proposalExecutionCount[proposalId] >= MAX_RETRY_COUNT) {
            revert TooManyExecutionAttempts();
        }

        // Increment execution count (rolls back on revert)
        proposalExecutionCount[proposalId]++;

        // Execute action (may call derived contract's _executeCustomAction)
        success = _executeInternalAction(proposal.actionType, proposal.callData);

        // Update proposal state based on execution result
        if (success) {
            _finalizeProposal(proposalId, ProposalStatus.Executed);
            emit ProposalExecuted(proposalId, msg.sender, true);
        } else {
            if (markFailedOnError) {
                // Terminal execution: mark as Failed
                _finalizeProposal(proposalId, ProposalStatus.Failed);
            } else {
                // Retry-enabled: keep as Approved for retry
                proposal.status = ProposalStatus.Approved;
            }
            emit ProposalExecuted(proposalId, msg.sender, false);
        }
    }

    // ============================================================
    // 8. INTERNAL FUNCTIONS - Member Management
    // ============================================================

    /// @dev Increment member version and create new snapshot for member list changes
    /// @return newSnapshot Empty storage array for new member version
    /// @return oldSnapshot Current member list snapshot
    /// @return newVersion Incremented version number
    /// @notice Old version proposals continue to use their snapshot member list for voting/execution
    ///         This allows proposals to complete even after member composition changes
    ///         Each proposal maintains its own version snapshot for consistency
    function _prepareNextMemberVersion() internal returns (address[] storage newSnapshot, address[] storage oldSnapshot, uint256 newVersion) {
        uint256 oldVersion = memberVersion;
        newVersion = oldVersion + 1;
        oldSnapshot = versionedMemberList[oldVersion];
        newSnapshot = versionedMemberList[newVersion]; // Empty array by default
        memberVersion = newVersion;
    }

    /// @dev Add new governance member with quorum update
    /// @param newMember Address of new member to add
    /// @param newQuorum New quorum requirement (1 for single member, ≥2 for multiple)
    ///
    /// @notice Member addition flow (Checks-Effects-Interactions):
    /// 1. Checks: Validate member status, address, and quorum requirements
    /// 2. Effects: Create new version snapshot, update member mappings and quorum
    /// 3. Interactions: Call _onMemberAdded hook for derived contract logic
    ///
    /// @notice Quorum rules: Single member requires quorum=1, multiple members require quorum≥2
    /// @notice Member limit: MAX_MEMBER_INDEX (255) enforced via MemberIndexOverflow
    ///
    /// @notice Security consideration: _onMemberAdded is called after state changes
    ///         Derived contracts must avoid external calls in _onMemberAdded to prevent
    ///         cross-function reentrancy attacks
    /// @notice Reentrancy protection: Caller (_executeProposal) has nonReentrant guard
    function _addMember(address newMember, uint32 newQuorum) internal {
        // CHECKS: Validate member and address
        if (members[newMember].isActive) revert AlreadyAMember();
        if (newMember == address(0)) revert InvalidMemberAddress();

        // EFFECTS: Create new version snapshot
        (address[] storage newSnapshot, address[] storage oldSnapshot, uint256 newVersion) = _prepareNextMemberVersion();
        uint256 oldLength = oldSnapshot.length;
        uint256 newMemberCount = oldLength + 1;
        if (oldLength >= MAX_MEMBER_INDEX) revert MemberIndexOverflow();

        // CHECKS: Validate quorum requirements
        if (newMemberCount == 1) {
            if (newQuorum != 1) revert InvalidQuorum();
        } else {
            if (newQuorum < 2 || newQuorum > newMemberCount) revert InvalidQuorum();
        }

        // EFFECTS: Copy existing members to new snapshot
        for (uint256 i = 0; i < oldLength; i++) {
            address existing = oldSnapshot[i];
            newSnapshot.push(existing);
            // Safe cast: member count limited to MAX_MEMBER_INDEX (255) < uint32.max
            // forge-lint: disable-next-line(unsafe-typecast)
            _memberIndexByVersion[newVersion][existing] = uint32(i + 1);
        }

        // EFFECTS: Add new member to snapshot and update state
        newSnapshot.push(newMember);
        // Safe cast: oldLength < MAX_MEMBER_INDEX (255) < uint32.max
        // forge-lint: disable-next-line(unsafe-typecast)
        _memberIndexByVersion[newVersion][newMember] = uint32(oldLength + 1);

        members[newMember] = Member({ isActive: true, joinedAt: uint32(block.timestamp) });

        uint32 oldQuorum = quorum;
        quorum = newQuorum;

        emit MemberAdded(newMember, newSnapshot.length, newQuorum);
        emit QuorumUpdated(oldQuorum, newQuorum);

        // INTERACTIONS: Call hook for derived contract logic (avoid external calls here)
        _onMemberAdded(newMember);

        _quorumByVersion[newVersion] = newQuorum;
    }

    /// @dev Change governance quorum requirement
    /// @param newQuorum New quorum value to set
    ///
    /// @notice Quorum change flow (Checks-Effects):
    /// 1. Checks: Validate quorum requirements based on current member count
    /// 2. Effects: Update quorum value and emit QuorumUpdated event
    ///
    /// @notice Quorum rules: Single member requires quorum=1, multiple members require quorum≥2
    /// @notice Quorum cannot exceed current member count
    ///
    /// @notice Reentrancy protection: Caller (_executeProposal) has nonReentrant guard
    function _changeQuorum(uint32 newQuorum) internal {
        uint256 memberCount = versionedMemberList[memberVersion].length;

        // CHECKS: Validate quorum requirements
        if (memberCount == 1) {
            if (newQuorum != 1) revert InvalidQuorum();
        } else {
            if (newQuorum < 2 || newQuorum > memberCount) revert InvalidQuorum();
        }

        // Update quorum value
        uint32 oldQuorum = quorum;
        quorum = newQuorum;

        emit QuorumUpdated(oldQuorum, newQuorum);

        // Update snapshot for current version
        _quorumByVersion[memberVersion] = newQuorum;
    }

    /// @dev Remove governance member with quorum update
    /// @param member Address of member to remove
    /// @param newQuorum New quorum requirement (1 for single member, ≥2 for multiple)
    ///
    /// @notice Member removal flow (Checks-Effects-Interactions):
    /// 1. Checks: Validate member status and quorum requirements
    /// 2. Effects: Create new version snapshot, deactivate member, update quorum
    /// 3. Interactions: Call _onMemberRemoved hook for derived contract logic
    ///
    /// @notice Quorum rules: Cannot remove last member, single member requires quorum=1
    /// @notice Index reassignment: Remaining members get sequential indices starting from 1
    ///
    /// @notice Security consideration: _onMemberRemoved is called after state changes
    ///         Derived contracts must avoid external calls in _onMemberRemoved to prevent
    ///         cross-function reentrancy attacks
    /// @notice Reentrancy protection: Caller (_executeProposal) has nonReentrant guard
    function _removeMember(address member, uint32 newQuorum) internal {
        // CHECKS: Validate member status
        if (!members[member].isActive) revert NotAMember();

        // EFFECTS: Create new version snapshot
        (address[] storage newSnapshot, address[] storage oldSnapshot, uint256 newVersion) = _prepareNextMemberVersion();
        uint256 oldLength = oldSnapshot.length;
        uint256 newMemberCount = oldLength - 1; // Safe: Solidity 0.8+ checks underflow

        // CHECKS: Validate quorum requirements
        if (newMemberCount == 0) revert InvalidQuorum(); // Cannot remove last member
        if (newMemberCount == 1) {
            if (newQuorum != 1) revert InvalidQuorum();
        } else {
            if (newQuorum < 2 || newQuorum > newMemberCount) revert InvalidQuorum();
        }

        // EFFECTS: Copy existing members except the removed one, reassign sequential indices
        uint256 newIndex = 0;
        for (uint256 i = 0; i < oldLength; i++) {
            address existing = oldSnapshot[i];
            if (existing == member) continue; // Skip removed member

            newSnapshot.push(existing);
            // Safe cast: ++newIndex starts from 1, limited to MAX_MEMBER_INDEX (255) < uint32.max
            // Note: Indices stored as "index + 1" (0 means "not a member")
            // forge-lint: disable-next-line(unsafe-typecast)
            _memberIndexByVersion[newVersion][existing] = uint32(++newIndex);
        }

        // Clean up removed member's index in new version (explicit zero for clarity)
        delete _memberIndexByVersion[newVersion][member];

        // EFFECTS: Deactivate member and update quorum
        members[member].isActive = false;

        uint32 oldQuorum = quorum;
        quorum = newQuorum;

        emit MemberRemoved(member, newSnapshot.length, newQuorum);
        emit QuorumUpdated(oldQuorum, newQuorum);

        // INTERACTIONS: Call hook for derived contract logic (avoid external calls here)
        _onMemberRemoved(member);

        _quorumByVersion[newVersion] = newQuorum;
    }

    // ============================================================
    // 9. INTERNAL FUNCTIONS - Execution Hooks (Virtual)
    // ============================================================

    // ========== Internal Action Dispatcher ==========

    /// @dev Internal action dispatcher - routes actions to member management or custom hooks
    /// @param actionType The type of action to execute
    /// @param callData ABI-encoded parameters for the action
    /// @return success True if the action executed successfully
    /// @notice Handles ACTION_ADD_MEMBER, ACTION_REMOVE_MEMBER, ACTION_CHANGE_QUORUM internally
    /// @notice Delegates unknown actions to _executeCustomAction for derived contract handling
    function _executeInternalAction(bytes32 actionType, bytes memory callData) internal virtual returns (bool success) {
        if (actionType == ACTION_ADD_MEMBER) {
            (address newMember, uint32 newQuorum) = abi.decode(callData, (address, uint32));
            _addMember(newMember, newQuorum);
            return true;
        } else if (actionType == ACTION_REMOVE_MEMBER) {
            (address member, uint32 newQuorum) = abi.decode(callData, (address, uint32));
            _removeMember(member, newQuorum);
            return true;
        } else if (actionType == ACTION_CHANGE_QUORUM) {
            uint32 newQuorum = abi.decode(callData, (uint32));
            _changeQuorum(newQuorum);
            return true;
        }

        // Delegate to derived contracts for custom actions
        return _executeCustomAction(actionType, callData);
    }

    // ========== Hooks for Derived Contracts ==========
    // Hook functions allow derived contracts to inject custom logic at specific points.
    // All hooks are virtual and must/can be overridden by derived contracts.
    //
    // Available Hooks:
    // 1. _executeCustomAction: Handle contract-specific actions (e.g., mint, burn, validator ops)
    // 2. _onMemberAdded: React to member addition (e.g., sync permissions)
    // 3. _onMemberRemoved: React to member removal (e.g., cleanup state)
    // 4. _onMemberChanged: React to member address change (e.g., migrate state)
    //
    // SECURITY WARNING FOR HOOK IMPLEMENTATIONS:
    // - DO NOT make external calls (token.transfer, externalContract.call, etc.)
    // - Hooks are called AFTER state changes, enabling cross-function reentrancy if external calls are made
    // - Safe: Storage updates, event emissions, internal logic
    // - Unsafe: External calls to tokens, contracts, or payable addresses
    // - See docs/SECURITY.md for detailed guidelines and examples

    /// @dev Execute custom actions - override in derived contracts to implement specific action sets
    /// @param actionType The type of custom action (e.g., ACTION_MINT, ACTION_BURN, ACTION_CONFIGURE_VALIDATOR)
    /// @param callData ABI-encoded parameters for the action
    /// @return success True if the action executed successfully
    ///
    /// @notice Default implementation returns false (unknown action type)
    /// @notice Override examples: GovMinter (mint/burn), GovMasterMinter (minter management), GovValidator (validator ops)
    ///
    /// @notice SECURITY: Follow CEI pattern (Checks-Effects-Interactions)
    ///         Perform all state changes BEFORE making external calls
    ///         See docs/SECURITY.md for safe implementation patterns
    function _executeCustomAction(bytes32 actionType, bytes memory callData) internal virtual returns (bool success) {
        // Default: unknown action type (parameters unused in base implementation)
        return false;
    }

    // ============================================================
    // 10. INTERNAL FUNCTIONS - Lifecycle Hooks (Virtual)
    // ============================================================

    /// @dev Override to perform contract-specific logic when member is added
    /// @param member Address of the newly added member
    /// @notice Called after member is added and quorum is updated
    /// @notice Use this to sync member state in derived contracts
    /// @notice SECURITY: No external calls - see SECURITY WARNING above and docs/SECURITY.md
    function _onMemberAdded(address member) internal virtual {}

    /// @dev Override to perform contract-specific logic when member is removed
    /// @param member Address of the removed member
    /// @notice Called after member is removed and quorum is updated
    /// @notice Use this to clean up member-specific state in derived contracts
    /// @notice SECURITY: No external calls - see SECURITY WARNING above and docs/SECURITY.md
    function _onMemberRemoved(address member) internal virtual {}

    /// @dev Override to perform contract-specific logic when member changes address
    /// @param oldMember Previous member address (now inactive)
    /// @param newMember New member address (now active)
    /// @notice Called after member change is complete
    /// @notice Use this to transfer member-specific state in derived contracts
    /// @notice SECURITY: No external calls - see SECURITY WARNING above and docs/SECURITY.md
    function _onMemberChanged(address oldMember, address newMember) internal virtual {}

    /// @dev Hook called when proposal reaches terminal state
    /// @notice Override in derived contracts to implement cleanup logic
    /// @param proposalId The proposal that reached terminal state
    ///
    /// @custom:security Called AFTER state transition (status already updated)
    /// @custom:security Derived contracts should use view/pure logic or minimal state changes
    /// @custom:security Must NOT make external calls to prevent reentrancy
    ///
    /// @custom:usage Terminal States
    ///      - Executed: Proposal successfully executed
    ///      - Failed: Proposal execution failed (markFailedOnError mode)
    ///      - Expired: Proposal expired before execution
    ///      - Cancelled: Proposal cancelled by proposer
    ///      - Rejected: Proposal rejected by members
    ///
    /// @custom:usage Call Sites
    ///      - _vote: Expired and Rejected states
    ///      - _executeProposal: Expired, Executed, and Failed states
    ///      - cancelProposal: Cancelled state
    ///      - expireProposal: Expired state
    ///
    /// @custom:design Rationale
    ///      - Template Method Pattern: Allows derived contracts to extend behavior
    ///      - Open/Closed Principle: GovBaseV2 is open for extension, closed for modification
    ///      - Single Responsibility: Lifecycle management separated from cleanup logic
    ///      - Backward Compatible: Empty default implementation (virtual)
    ///
    /// @custom:example GovMinter Cleanup
    ///      ```solidity
    ///      function _onProposalFinalized(uint256 proposalId) internal override {
    ///          _cleanupMintReservation(proposalId);
    ///      }
    ///      ```
    function _onProposalFinalized(uint256 proposalId) internal virtual {
        // Default implementation: do nothing (backward compatible)
        // Derived contracts can override to implement cleanup logic
    }

    /// @dev Finalize proposal to terminal state with atomic cleanup
    /// @notice Atomically performs three operations that must always occur together:
    ///         1. Update proposal status to terminal state
    ///         2. Decrement active proposal count (free member's slot)
    ///         3. Execute cleanup hook (_onProposalFinalized)
    ///
    /// @param proposalId The proposal ID to finalize
    /// @param finalStatus Terminal status (Executed, Failed, Expired, Cancelled, or Rejected)
    ///
    /// @custom:security Atomicity
    ///      Prevents inconsistent states by ensuring all three operations execute together.
    ///
    /// @custom:note Event Emission
    ///      Caller must emit appropriate event after calling this function.
    ///      Each terminal state has different event parameters.
    ///
    /// @custom:note Executed State
    ///      Automatically sets executedAt timestamp when finalStatus is Executed.
    function _finalizeProposal(uint256 proposalId, ProposalStatus finalStatus) internal {
        // Update status to terminal state
        proposals[proposalId].status = finalStatus;

        // Record execution timestamp for Executed state
        if (finalStatus == ProposalStatus.Executed) {
            proposals[proposalId].executedAt = block.timestamp;
        }

        // Free member's proposal slot
        _decrementActiveProposalCount(proposalId);

        // Execute cleanup hook
        _onProposalFinalized(proposalId);
    }

    // ============================================================
    // 11. INTERNAL FUNCTIONS - Validation Helpers
    // ============================================================

    /// @dev Validate member version number and prevent invalid version access
    /// @dev Central validation point for all version-based queries
    /// @param targetVersion The version number to validate
    /// @return The validated version number (passthrough for convenience)
    /// @custom:revert InvalidMemberVersion if targetVersion is 0 or exceeds current memberVersion
    function _validateMemberVersion(uint256 targetVersion) internal view returns (uint256) {
        if (targetVersion == 0 || targetVersion > memberVersion) revert InvalidMemberVersion();
        return targetVersion;
    }

    /// @dev Validate proposal ID and prevent invalid proposal access
    /// @dev Central validation point for all proposal-based queries and operations
    /// @param proposalId The proposal ID to validate
    /// @return The validated proposal ID (passthrough for convenience)
    /// @custom:revert InvalidProposal if proposalId is 0 or exceeds currentProposalId
    function _validateProposalId(uint256 proposalId) internal view returns (uint256) {
        if (proposalId == 0 || proposalId > currentProposalId) revert InvalidProposal();
        return proposalId;
    }

    // ============================================================
    // 12. INTERNAL FUNCTIONS - Utility Helpers
    // ============================================================

    /// @dev Get member index at a specific member version snapshot
    /// @notice Returns 0-based member index, or type(uint256).max if not a member at that version
    /// @notice Index encoding: Storage uses 1-based indexing (0 = not member), returns 0-based
    /// @param member The member address to query
    /// @param version The member version to query (validation is caller's responsibility)
    /// @return The 0-based member index, or type(uint256).max if not a member
    function _getMemberIndexAtVersion(address member, uint256 version) internal view returns (uint256) {
        uint32 indexPlusOne = _memberIndexByVersion[version][member];
        if (indexPlusOne == 0) {
            return type(uint256).max;
        }
        return uint256(indexPlusOne - 1);
    }

    /// @dev Decrement active proposal count for a proposal's creator
    /// @notice Called when proposal reaches terminal state (Executed, Cancelled, Expired, Failed, Rejected)
    /// @notice Frees up proposal capacity to enforce MAX_ACTIVE_PROPOSALS_PER_MEMBER limit
    /// @param proposalId The proposal ID that has reached terminal state (validation is caller's responsibility)
    function _decrementActiveProposalCount(uint256 proposalId) internal {
        address proposer = proposals[proposalId].proposer;
        if (memberActiveProposalCount[proposer] > 0) {
            memberActiveProposalCount[proposer]--;
        }
    }
}
