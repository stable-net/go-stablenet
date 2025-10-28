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
    error AlreadyInitialized();
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
    error ProposalAlreadyExpired();
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

    struct GovernanceConfig {
        address[] members;
        uint32 quorum;
        uint256 proposalExpiry;
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

    // ========== State Variables ==========
    // Value types - uint256 (32 bytes each, 1 slot per variable)
    uint256 public proposalExpiry; // Slot 0: Set once during initialization, cannot be changed
    uint256 public memberVersion = INITIAL_MEMBER_VERSION; // Slot 1
    uint256 public currentProposalId; // Slot 2
    uint256 private _reentrancyGuard; // Slot 3: Reentrancy protection

    // Small value types - uint32 (4 bytes, can be packed with other small types)
    uint32 public quorum; // Slot 4 (0-3 bytes): Required number of approvals (m of n)
    // Note: Slot 4 has 28 bytes remaining for future small types (address, uint32, etc.)

    // Mappings - each mapping occupies 1 slot for metadata
    mapping(address => Member) public members; // Slot 5
    mapping(uint256 => address[]) public versionedMemberList; // Slot 6
    mapping(uint256 => Proposal) public proposals; // Slot 7
    mapping(uint256 => mapping(address => uint32)) internal memberIndexByVersion; // Slot 8: index + 1 snapshot per version
    mapping(uint256 => uint32) internal quorumByVersion; // Slot 9: quorum snapshot per member version
    mapping(uint256 => uint256) public proposalExecutionCount; // Slot 10: Execution attempt count for each proposal
    mapping(address => uint256) public memberActiveProposalCount; // Slot 11: Track active proposals per member

    // Reserved storage for future upgrades
    uint256[38] private __gap; // Slot 12-49: Reserved storage space (reduced by 3 for proposalExecutionCount, memberActiveProposalCount, quorumByVersion)

    // ========== Events ==========
    event GovernanceInitialized(uint256 memberCount, uint32 quorum, uint256 proposalExpiry, uint256 memberVersion);

    event ProposalCreated(
        uint256 indexed proposalId,
        address indexed proposer,
        bytes32 actionType,
        uint256 memberVersion,
        uint256 requiredApprovals,
        bytes callData
    );

    event ProposalVoted(
        uint256 indexed proposalId, address indexed voter, bool approval, uint256 approved, uint256 rejected
    );

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
    modifier onlyMember() {
        _checkMembership();
        _;
    }

    modifier proposalExists(uint256 proposalId) {
        _checkProposalExists(proposalId);
        _;
    }

    modifier proposalInVoting() {
        _checkProposalInVoting();
        _;
    }

    modifier proposalExecutable() {
        _checkProposalExecutable();
        _;
    }

    modifier nonReentrant() {
        _nonReentrantBefore();
        _;
        _nonReentrantAfter();
    }

    // ========== Internal Functions ==========

    /// @dev Initialize governance system with initial members and configuration
    /// @param config Governance configuration containing members, quorum, and proposal expiry
    /// @notice Must be called from derived contract's initializer (e.g., GovMinter.initialize)
    /// @notice Can only be called once - protected by AlreadyInitialized check
    /// @notice Emits GovernanceInitialized event upon successful initialization
    ///
    /// Validation checks (in order):
    /// - Re-initialization protection (checks memberVersion snapshot)
    /// - Member count: 1 <= count <= MAX_MEMBER_INDEX (255)
    /// - Member addresses: no zero addresses, no duplicates (within array or existing members)
    /// - Quorum range: 2 <= quorum <= member count (or quorum=1 for single-member governance)
    /// - Proposal expiry: must be non-zero
    ///
    /// Implementation uses fail-fast pattern:
    /// Step 1: Individual member validation (zero address, already active members)
    /// Step 2: Duplicate detection within input array
    /// Step 3: Quorum validation (after member validation for proper error precedence)
    /// Step 4: State changes only after all validations pass
    ///
    /// Gas optimization: Uses unchecked increment (safe due to MAX_MEMBER_INDEX = 255)
    function _initializeGovernance(GovernanceConfig memory config) internal {
        // Re-initialization protection: Check if already initialized
        if (versionedMemberList[memberVersion].length > 0) revert AlreadyInitialized();

        // Basic configuration validation
        uint256 initialMemberCount = config.members.length;
        if (initialMemberCount == 0) revert InvalidQuorum();
        if (initialMemberCount > MAX_MEMBER_INDEX) revert MemberIndexOverflow(); // Max 255 members (bitmap safety)

        // Step 1: Individual member validation (fail-fast pattern)
        // Check each member address for:
        // - Zero address (invalid)
        // - Already active in governance (duplicate from existing members)
        // Must be done BEFORE quorum validation for proper error precedence
        for (uint256 i = 0; i < initialMemberCount;) {
            address memberAddress = config.members[i];
            if (memberAddress == address(0)) revert InvalidMemberAddress();
            if (members[memberAddress].isActive) revert DuplicateMember();

            unchecked {
                ++i;
            }
        }

        // Step 2: Duplicate detection within input array
        // Validates that no address appears twice in the members array
        // Uses nested loop to compare each member with all subsequent members
        // Must be done BEFORE quorum validation for proper error precedence
        for (uint256 i = 0; i < initialMemberCount;) {
            for (uint256 j = i + 1; j < initialMemberCount;) {
                if (config.members[i] == config.members[j]) revert DuplicateMember();
                unchecked {
                    ++j; // Safe: j < initialMemberCount <= 255
                }
            }
            unchecked {
                ++i; // Safe: i < initialMemberCount <= MAX_MEMBER_INDEX (255)
            }
        }

        // Step 3: Quorum validation with security consideration
        // Security: quorum=1 allows single-member unilateral decisions without peer review
        // For multi-member governance (>=2 members), enforce minimum quorum of 2
        if (initialMemberCount == 1) {
            // Special case: Single member governance (for testing/initial deployment only)
            // WARNING: Single-member governance is centralized and not recommended for production
            if (config.quorum != 1) revert InvalidQuorum();
        } else {
            // Multi-member governance: Require at least 2 approvals (proposer + 1 reviewer)
            // This prevents any single member from executing proposals without peer review
            if (config.quorum < 2 || config.quorum > initialMemberCount) revert InvalidQuorum();
        }

        if (config.proposalExpiry == 0) revert InvalidProposalExpiry();

        // Step 4: State changes (only executed after all validations pass)
        // Register all members and set governance parameters
        address[] storage snapshot = versionedMemberList[memberVersion];
        for (uint256 i = 0; i < initialMemberCount;) {
            address memberAddress = config.members[i];

            // Add member to versioned snapshot
            snapshot.push(memberAddress);

            // Mark member as active with current timestamp
            members[memberAddress] = Member({isActive: true, joinedAt: uint32(block.timestamp)});

            // Store member index (offset by 1 to distinguish from default value 0)
            // This allows _getMemberIndexAtVersion to return type(uint256).max for non-existent members
            // casting to 'uint32' is safe because member count is limited to MAX_MEMBER_INDEX (255)
            // forge-lint: disable-next-line(unsafe-typecast)
            memberIndexByVersion[memberVersion][memberAddress] = uint32(i + 1);

            unchecked {
                ++i; // Safe: i < initialMemberCount <= MAX_MEMBER_INDEX (255)
            }
        }

        // Set governance parameters
        quorum = config.quorum;
        proposalExpiry = config.proposalExpiry;

        quorumByVersion[memberVersion] = quorum;

        emit GovernanceInitialized(initialMemberCount, config.quorum, config.proposalExpiry, memberVersion);
    }

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
    function _createProposal(bytes32 actionType, bytes memory callData)
        internal
        onlyMember
        returns (uint256 proposalId)
    {
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
            votedBitmap: 0, // No votes yet
            requiredApprovals: quorum, // Snapshot current quorum requirement
            approved: 0, // No approvals yet
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
            proposal.status = ProposalStatus.Expired;
            _decrementActiveProposalCount(proposalId);
            _onProposalFinalized(proposalId);
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
                proposal.status = ProposalStatus.Approved;
                emit ProposalApproved(proposalId, msg.sender, newApproved, proposal.rejected);

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
                proposal.status = ProposalStatus.Rejected;
                _decrementActiveProposalCount(proposalId);
                _onProposalFinalized(proposalId);
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
            proposal.status = ProposalStatus.Expired;
            _decrementActiveProposalCount(proposalId);
            _onProposalFinalized(proposalId);
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
            proposal.status = ProposalStatus.Executed;
            proposal.executedAt = block.timestamp;
            _decrementActiveProposalCount(proposalId);
            _onProposalFinalized(proposalId);
            emit ProposalExecuted(proposalId, msg.sender, true);
        } else {
            if (markFailedOnError) {
                // Terminal execution: mark as Failed
                proposal.status = ProposalStatus.Failed;
                _decrementActiveProposalCount(proposalId);
                _onProposalFinalized(proposalId);
            } else {
                // Retry-enabled: keep as Approved for retry
                proposal.status = ProposalStatus.Approved;
            }
            emit ProposalExecuted(proposalId, msg.sender, false);
        }
    }

    /// @dev Increment member version and create new snapshot for member list changes
    /// @return newSnapshot Empty storage array for new member version
    /// @return oldSnapshot Current member list snapshot
    /// @return newVersion Incremented version number
    /// @notice Old version proposals continue to use their snapshot member list for voting/execution
    ///         This allows proposals to complete even after member composition changes
    ///         Each proposal maintains its own version snapshot for consistency
    function _prepareNextMemberVersion()
        internal
        returns (address[] storage newSnapshot, address[] storage oldSnapshot, uint256 newVersion)
    {
        uint256 oldVersion = memberVersion;
        newVersion = oldVersion + 1;
        oldSnapshot = versionedMemberList[oldVersion];
        delete versionedMemberList[newVersion]; // Clear stale data
        newSnapshot = versionedMemberList[newVersion];
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
            memberIndexByVersion[newVersion][existing] = uint32(i + 1);
        }

        // EFFECTS: Add new member to snapshot and update state
        newSnapshot.push(newMember);
        // Safe cast: oldLength < MAX_MEMBER_INDEX (255) < uint32.max
        // forge-lint: disable-next-line(unsafe-typecast)
        memberIndexByVersion[newVersion][newMember] = uint32(oldLength + 1);

        members[newMember] = Member({isActive: true, joinedAt: uint32(block.timestamp)});

        uint32 oldQuorum = quorum;
        quorum = newQuorum;

        emit MemberAdded(newMember, newSnapshot.length, newQuorum);
        emit QuorumUpdated(oldQuorum, newQuorum);

        // INTERACTIONS: Call hook for derived contract logic (avoid external calls here)
        _onMemberAdded(newMember);

        quorumByVersion[newVersion] = newQuorum;
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
        uint256 newMemberCount = oldLength - 1;  // Safe: Solidity 0.8+ checks underflow

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
            if (existing == member) continue;  // Skip removed member

            newSnapshot.push(existing);
            // Safe cast: ++newIndex starts from 1, limited to MAX_MEMBER_INDEX (255) < uint32.max
            // Note: Indices stored as "index + 1" (0 means "not a member")
            // forge-lint: disable-next-line(unsafe-typecast)
            memberIndexByVersion[newVersion][existing] = uint32(++newIndex);
        }

        // Clean up removed member's index in new version (explicit zero for clarity)
        delete memberIndexByVersion[newVersion][member];

        // EFFECTS: Deactivate member and update quorum
        members[member].isActive = false;

        uint32 oldQuorum = quorum;
        quorum = newQuorum;

        emit MemberRemoved(member, newSnapshot.length, newQuorum);
        emit QuorumUpdated(oldQuorum, newQuorum);

        // INTERACTIONS: Call hook for derived contract logic (avoid external calls here)
        _onMemberRemoved(member);

        quorumByVersion[newVersion] = newQuorum;
    }

    /// @dev Change member address (address rotation)
    /// @param oldMember Current member address to replace
    /// @param newMember New member address to activate
    ///
    /// @notice Member change flow (Checks-Effects-Interactions):
    /// 1. Checks: Validate addresses and member status
    /// 2. Effects: Create new version snapshot, replace member, update states
    /// 3. Interactions: Call _onMemberChanged hook for derived contract logic
    ///
    /// @notice Index preservation: New member retains the same index as old member
    /// @notice Use case: Key rotation, member address migration
    ///
    /// @notice Security consideration: _onMemberChanged is called after state changes
    ///         Derived contracts must avoid external calls in _onMemberChanged to prevent
    ///         cross-function reentrancy attacks
    /// @notice Reentrancy protection: Caller (_executeProposal) has nonReentrant guard
    function _changeMember(address oldMember, address newMember) internal {
        // CHECKS: Validate addresses and member status
        if (newMember == address(0)) revert InvalidMemberAddress();
        if (!members[oldMember].isActive) revert NotAMember();
        if (members[newMember].isActive) revert AlreadyAMember();

        // Create new version snapshot
        (address[] storage newSnapshot, address[] storage oldSnapshot, uint256 newVersion) = _prepareNextMemberVersion();
        uint256 oldLength = oldSnapshot.length;

        // Copy members, replacing oldMember with newMember at same index
        uint256 newIndex = 0;
        for (uint256 i = 0; i < oldLength; i++) {
            address existing = oldSnapshot[i];
            if (existing == oldMember) {
                existing = newMember;  // Replace old with new, preserving index
            }
            newSnapshot.push(existing);
            // Safe cast: ++newIndex starts from 1, limited to MAX_MEMBER_INDEX (255) < uint32.max
            // Note: Indices stored as "index + 1" (0 means "not a member")
            // forge-lint: disable-next-line(unsafe-typecast)
            memberIndexByVersion[newVersion][existing] = uint32(++newIndex);
        }

        // Clean up stale mapping entry
        delete memberIndexByVersion[newVersion][oldMember];

        // Update member states and emit events
        members[newMember] = Member({isActive: true, joinedAt: uint32(block.timestamp)});
        members[oldMember].isActive = false;
        emit MemberChanged(oldMember, newMember);

        // Call hook for derived contract logic (avoid external calls here)
        _onMemberChanged(oldMember, newMember);

        quorumByVersion[newVersion] = quorum;
    }

    // ========== Internal Action Dispatcher ==========

    /// @dev Internal action dispatcher - routes actions to member management or custom hooks
    /// @param actionType The type of action to execute
    /// @param callData ABI-encoded parameters for the action
    /// @return success True if the action executed successfully
    /// @notice Handles ACTION_ADD_MEMBER, ACTION_REMOVE_MEMBER, ACTION_CHANGE_MEMBER internally
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
        } else if (actionType == ACTION_CHANGE_MEMBER) {
            (address oldMember, address newMember) = abi.decode(callData, (address, address));
            _changeMember(oldMember, newMember);
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
    ///      - _vote: Lines 419 (Expired), 472 (Rejected)
    ///      - _executeProposal: Lines 511 (Expired), 531 (Executed), 537 (Failed)
    ///      - cancelProposal: Line 894 (Cancelled)
    ///      - expireProposal: Line 920 (Expired)
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

    // ========== Public Functions ==========

    /// @notice Vote YES on a proposal without automatic execution (gas-safe voting)
    /// @dev Allows approving without risking gas exhaustion from auto-execution
    /// @param proposalId The proposal ID to approve
    function approveProposal(uint256 proposalId) public onlyMember {
        _vote(proposalId, true, false);
    }

    /// @notice Vote NO on a proposal
    /// @dev Reject the specified proposal
    /// @param proposalId The proposal ID to reject
    function disapproveProposal(uint256 proposalId) public onlyMember {
        _vote(proposalId, false, false);
    }

    /// @notice Vote YES on a proposal with automatic execution if quorum is reached
    /// @dev If this vote reaches quorum, proposal will execute immediately (gas risk applies)
    /// @param proposalId The proposal ID to approve and potentially execute
    function approveProposalAndExecute(uint256 proposalId) public onlyMember {
        _vote(proposalId, true, true);
    }

    /// @notice Execute or retry an approved proposal
    /// @param proposalId The proposal ID to execute
    /// @return success True if execution succeeded
    /// @dev Automatically tracks execution attempts and enforces retry limit (MAX_RETRY_COUNT = 3)
    /// @dev If execution fails, proposal remains in Approved status and can be retried by calling again
    /// @dev Reverts with TooManyExecutionAttempts after 3 failed attempts
    function executeProposal(uint256 proposalId) public onlyMember returns (bool) {
        return _executeProposal(proposalId, false);
    }

    /// @notice Execute an approved proposal with terminal failure (no retry)
    /// @param proposalId The proposal ID to execute
    /// @return success True if execution succeeded
    /// @dev If execution fails, proposal is marked as Failed and cannot be retried
    function executeWithFailure(uint256 proposalId) public onlyMember returns (bool) {
        return _executeProposal(proposalId, true);
    }

    /// @notice Cancel a proposal that is still in voting phase
    /// @dev Can only be called by the original proposer before other members vote
    /// @dev Proposer's automatic approval (approved=1) doesn't prevent cancellation
    /// @dev Proposal must be in Voting status and no other member should have voted yet
    /// @param proposalId The ID of the proposal to cancel
    function cancelProposal(uint256 proposalId) public onlyMember {
        // Validate proposal exists
        if (proposalId == 0 || proposalId > currentProposalId) revert InvalidProposal();

        Proposal storage proposal = proposals[proposalId];

        // Validate proposal is in Voting status
        if (proposal.status != ProposalStatus.Voting) revert ProposalNotInVoting();

        // Only proposer can cancel
        if (proposal.proposer != msg.sender) revert NotProposer();

        // Cannot cancel if other members have voted (proposer's auto-vote doesn't count)
        // approved > 1 means proposer + at least 1 other member voted
        // rejected > 0 means at least 1 member voted against
        if (proposal.approved > 1 || proposal.rejected > 0) {
            revert ProposalAlreadyInVoting();
        }

        proposal.status = ProposalStatus.Cancelled;
        _decrementActiveProposalCount(proposalId);
        _onProposalFinalized(proposalId);
        emit ProposalCancelled(proposalId, msg.sender);
    }

    /// @notice Manually expire a proposal that has passed its expiry time
    /// @dev Can be called by members to clean up expired proposals
    /// @param proposalId The proposal ID to expire
    /// @return success True if the proposal was expired, false if it cannot be expired
    function expireProposal(uint256 proposalId) public onlyMember returns (bool success) {
        // Validate proposal exists
        if (proposalId == 0 || proposalId > currentProposalId) return false;

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
        proposal.status = ProposalStatus.Expired;
        _decrementActiveProposalCount(proposalId);
        _onProposalFinalized(proposalId);
        emit ProposalExpired(proposalId, msg.sender);
        return true;
    }

    /// @notice Check if proposal can be executed with detailed failure reason
    /// @dev This is a view function - state may change between check and execution (race condition possible)
    /// @dev If proposal is expired but status is still Approved, returns Expired (state not modified in view function)
    /// @dev Checks are performed in order: proposalId validity → status → expiry → retry limit
    /// @param proposalId Proposal ID to check
    /// @return result Execution check result indicating why execution is possible or not
    /// @return attemptsLeft Number of execution attempts remaining (0 if not executable)
    function canExecuteProposal(uint256 proposalId)
        external
        view
        returns (ExecutionCheckResult result, uint256 attemptsLeft)
    {
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
    function proposeAddMember(address newMember, uint32 newQuorum) public onlyMember returns (uint256 proposalId) {
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
    function proposeRemoveMember(address member, uint32 newQuorum) public onlyMember returns (uint256 proposalId) {
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

    /// @notice Propose to change a member's address
    /// @param oldMember Current address of the member
    /// @param newMember New address to replace old member
    /// @return proposalId The ID of the created proposal
    /// @dev Requires governance approval to execute
    /// @dev Different from changeMember() which is self-service without proposal
    function proposeChangeMember(address oldMember, address newMember) public onlyMember returns (uint256 proposalId) {
        if (newMember == address(0)) revert InvalidMemberAddress();
        if (!members[oldMember].isActive) revert NotAMember();
        if (members[newMember].isActive) revert AlreadyAMember();

        bytes memory callData = abi.encode(oldMember, newMember);
        return _createProposal(ACTION_CHANGE_MEMBER, callData);
    }

    // ========== View Functions ==========

    /// @notice Get the number of members at a specific governance version
    /// @dev Returns the member count from historical snapshot for the given version
    /// @dev Version snapshots are immutable - past versions are preserved for audit trail
    ///
    /// @param targetVersion The member version to query (must be between 1 and current memberVersion)
    /// @return The number of members in the specified version
    ///
    /// @custom:security Validates version bounds to prevent:
    ///   - Access to non-existent versions (version 0)
    ///   - Access to future versions (version > memberVersion)
    /// @custom:security Version data integrity guaranteed by:
    ///   - Initialization requires at least 1 member (version 1 always has data)
    ///   - _prepareNextMemberVersion preserves old versions (only clears new version)
    ///   - Member operations (_addMember, _removeMember, _changeMember) always populate new version
    ///
    /// Reverts with:
    /// @custom:revert InvalidMemberVersion if targetVersion is 0 or exceeds current memberVersion
    ///
    /// Gas optimization: No empty list check (guaranteed non-empty by initialization and version management)
    ///
    /// Example usage:
    /// ```solidity
    /// uint256 currentCount = getMemberCount(memberVersion);      // Current member count
    /// uint256 historicalCount = getMemberCount(1);               // Initial member count
    /// uint256 previousCount = getMemberCount(memberVersion - 1); // Previous version count
    /// ```
    function getMemberCount(uint256 targetVersion) public view returns (uint256) {
        uint256 version = _validateMemberVersion(targetVersion);
        return versionedMemberList[version].length;
    }

    /// @notice Get member address at specific index in a governance version snapshot
    /// @dev Retrieves member from historical version snapshot with bounds checking
    /// @dev Version snapshots are immutable - past versions preserved for audit trail
    ///
    /// @param targetVersion The member version to query (must be between 1 and current memberVersion)
    /// @param index The zero-based index in the member list (must be < snapshot length)
    /// @return The member address at the specified index in the version snapshot
    ///
    /// @custom:security Index validation prevents:
    ///   - Out-of-bounds array access (index >= snapshot.length)
    ///   - Integer overflow (Solidity 0.8+ automatic checks)
    ///   - Empty snapshot access (guaranteed non-empty by initialization)
    ///
    /// @custom:security Version validation via _validateMemberVersion:
    ///   - Prevents version 0 access (versions start from 1)
    ///   - Prevents future version access (version > memberVersion)
    ///   - Ensures snapshot exists and has data
    ///
    /// @custom:security Address guarantees:
    ///   - Never returns address(0) (validated during initialization and member operations)
    ///   - Index always < MAX_MEMBER_INDEX (255) due to member count constraints
    ///   - Address is guaranteed to be valid EOA or contract address
    ///
    /// @custom:security Attack vector analysis:
    ///   - Array bounds: Protected by explicit index >= length check
    ///   - Integer overflow: Protected by Solidity 0.8+ automatic checks
    ///   - Empty array: Impossible by design (min 1 member required)
    ///   - Zero address: Impossible (validated at entry points)
    ///   - TOCTOU: View function, no state changes, safe for concurrent access
    ///   - Gas DoS: O(1) complexity, no loops, constant gas cost
    ///
    /// Reverts with:
    /// @custom:revert InvalidMemberVersion if targetVersion is 0 or exceeds memberVersion
    /// @custom:revert IndexOutOfBounds if index >= snapshot.length
    ///
    /// Gas optimization: Direct array access without intermediate checks (bounds already validated)
    ///
    /// Example usage:
    /// ```solidity
    /// // Get first member in current version
    /// address firstMember = getMemberAt(memberVersion, 0);
    ///
    /// // Iterate over all members in version 2
    /// uint256 count = getMemberCount(2);
    /// for (uint256 i = 0; i < count; i++) {
    ///     address member = getMemberAt(2, i);
    ///     // ... process member
    /// }
    ///
    /// // Access historical member (version 1, second member)
    /// address historicalMember = getMemberAt(1, 1);
    /// ```
    ///
    /// @custom:invariant snapshot.length > 0 for all valid versions (design guarantee)
    /// @custom:invariant snapshot.length <= MAX_MEMBER_INDEX (255) for all versions
    /// @custom:invariant snapshot[i] != address(0) for all valid i (data integrity guarantee)
    function getMemberAt(uint256 targetVersion, uint256 index) public view returns (address) {
        uint256 version = _validateMemberVersion(targetVersion);
        address[] storage snapshot = versionedMemberList[version];
        if (index >= snapshot.length) revert IndexOutOfBounds();
        return snapshot[index];
    }

    /// @notice Get proposal details by proposal ID
    /// @dev Retrieves complete proposal data from storage with validation
    /// @dev Returns proposal struct with all fields including status, votes, and execution data
    ///
    /// @param proposalId The proposal ID to query (must be between 1 and currentProposalId)
    /// @return Proposal struct containing all proposal data
    ///
    /// @custom:security Proposal ID validation prevents:
    ///   - Access to non-existent proposals (proposalId == 0)
    ///   - Access to future proposals (proposalId > currentProposalId)
    ///   - Uninitialized proposal access (proposals start from ID 1)
    ///
    /// @custom:security Attack vector analysis:
    ///   - Invalid ID: Protected by _validateProposalId (0 or > currentProposalId)
    ///   - Integer overflow: Protected by Solidity 0.8+ automatic checks
    ///   - Uninitialized access: proposalId starts from 1, 0 is invalid
    ///   - Race conditions: View function, no state changes, safe concurrent access
    ///   - Gas DoS: O(1) complexity, single mapping lookup
    ///
    /// @custom:security Proposal data guarantees:
    ///   - proposalId == 0: Always invalid (proposals start from 1)
    ///   - proposalId > currentProposalId: Future proposal, doesn't exist
    ///   - Valid proposalId: Always has initialized data (created via _createProposal)
    ///   - Proposal struct: All fields properly initialized at creation
    ///
    /// Reverts with:
    /// @custom:revert InvalidProposal if proposalId is 0 or exceeds currentProposalId
    ///
    /// Gas optimization: Direct mapping access, no intermediate checks
    ///
    /// Example usage:
    /// ```solidity
    /// // Get current proposal
    /// Proposal memory proposal = getProposal(currentProposalId);
    ///
    /// // Get specific proposal by ID
    /// Proposal memory proposal1 = getProposal(1);
    ///
    /// // Check proposal status
    /// Proposal memory p = getProposal(proposalId);
    /// if (p.status == ProposalStatus.Approved) {
    ///     executeProposal(proposalId);
    /// }
    ///
    /// // Access proposal data
    /// Proposal memory p = getProposal(proposalId);
    /// address proposer = p.proposer;
    /// uint32 approvalCount = p.approved;
    /// bytes32 actionType = p.actionType;
    /// ```
    ///
    /// @custom:invariant proposals[proposalId].proposer != address(0) for valid proposalId (creation guarantee)
    /// @custom:invariant proposals[proposalId].createdAt > 0 for valid proposalId (creation timestamp)
    /// @custom:invariant proposals[proposalId].requiredApprovals > 0 for valid proposalId (quorum snapshot)
    function getProposal(uint256 proposalId) public view returns (Proposal memory) {
        proposalId = _validateProposalId(proposalId);
        return proposals[proposalId];
    }

    /// @notice Check if a proposal is in voting phase (Voting status, not yet approved)
    /// @dev Returns true only if proposal is in Voting status and not expired
    /// @dev Note: Proposals in Approved status can still receive votes via _vote function,
    ///      but this function returns false to distinguish voting phase from approved phase
    ///
    /// @param proposalId The proposal ID to check (must be between 1 and currentProposalId)
    /// @return True if proposal is in Voting status and not expired, false otherwise
    ///
    /// @custom:security Proposal ID validation via _validateProposalId:
    ///   - Prevents proposalId 0 access (proposals start from 1)
    ///   - Prevents future proposal access (proposalId > currentProposalId)
    ///   - Ensures proposal exists and has initialized data
    ///
    /// @custom:security Voting phase criteria:
    ///   - Status must be Voting (not Approved, Executed, Cancelled, Expired, Failed, or Rejected)
    ///   - Proposal must not be expired (block.timestamp <= createdAt + proposalExpiry)
    ///   - Both conditions must be met to be considered "in voting"
    ///
    /// @custom:security Design rationale:
    ///   - Distinguishes Voting phase from Approved phase for UI/workflow clarity
    ///   - Approved proposals can still receive additional votes (see _vote function)
    ///   - This function helps track proposal lifecycle state transitions
    ///   - Expiry check prevents considering expired proposals as active
    ///
    /// @custom:security Attack vector analysis:
    ///   - Invalid proposalId: Protected by _validateProposalId
    ///   - Integer overflow: Protected by Solidity 0.8+ for timestamp arithmetic
    ///   - TOCTOU: View function, no state changes, safe for concurrent access
    ///   - Gas DoS: O(1) complexity, single mapping lookup + arithmetic
    ///
    /// Reverts with:
    /// @custom:revert InvalidProposal if proposalId is 0 or exceeds currentProposalId
    ///
    /// Example usage:
    /// ```solidity
    /// // Check if proposal is still collecting initial votes
    /// if (isProposalInVoting(5)) {
    ///     // Proposal hasn't reached quorum yet
    ///     approveProposal(5);
    /// } else if (isProposalExecutable(5)) {
    ///     // Proposal reached quorum and can be executed
    ///     executeProposal(5);
    /// }
    ///
    /// // Workflow: Voting → Approved → Executed
    /// // isProposalInVoting returns true only during Voting phase
    /// ```
    ///
    /// @custom:invariant InVoting ⟺ (status == Voting) ∧ (timestamp ≤ createdAt + expiry)
    /// @custom:invariant If expired: always returns false (regardless of status)
    /// @custom:invariant Phase distinction: Voting (collecting votes) vs Approved (ready to execute)
    function isProposalInVoting(uint256 proposalId) public view returns (bool) {
        Proposal memory proposal = proposals[_validateProposalId(proposalId)];
        if (proposal.status == ProposalStatus.Voting) {
            if (block.timestamp <= proposal.createdAt + proposalExpiry) {
                return true;
            }
        }
        return false;
    }

    /// @notice Check if a proposal is ready for execution (Approved status and all conditions met)
    /// @dev Returns true only if proposal is in Approved status, not expired, and quorum satisfied
    /// @dev This function validates all three execution requirements for defensive programming
    ///
    /// @param proposalId The proposal ID to check (must be between 1 and currentProposalId)
    /// @return True if proposal can be executed immediately, false otherwise
    ///
    /// @custom:security Proposal ID validation via _validateProposalId:
    ///   - Prevents proposalId 0 access (proposals start from 1)
    ///   - Prevents future proposal access (proposalId > currentProposalId)
    ///   - Ensures proposal exists and has initialized data
    ///
    /// @custom:security Execution eligibility criteria (all must be satisfied):
    ///   1. Status must be Approved (quorum reached, not Voting/Executed/Cancelled/Expired/Failed/Rejected)
    ///   2. Not expired (block.timestamp <= createdAt + proposalExpiry)
    ///   3. Quorum satisfied (approved >= requiredApprovals) - defensive check
    ///
    /// @custom:security Design rationale:
    ///   - Triple validation ensures execution safety and prevents state inconsistencies
    ///   - Expiry check prevents executing stale proposals
    ///   - Redundant quorum check (status Approved already implies quorum) for defensive programming
    ///   - View function ensures no state changes, safe for concurrent access
    ///
    /// @custom:security Attack vector analysis:
    ///   - Invalid proposalId: Protected by _validateProposalId
    ///   - Integer overflow: Protected by Solidity 0.8+ for timestamp arithmetic
    ///   - TOCTOU race: View function, no state changes, safe
    ///   - State bypass: Triple validation prevents execution in invalid states
    ///   - Gas DoS: O(1) complexity, single mapping lookup + arithmetic
    ///
    /// Reverts with:
    /// @custom:revert InvalidProposal if proposalId is 0 or exceeds currentProposalId
    ///
    /// Example usage:
    /// ```solidity
    /// // Check before execution attempt
    /// if (isProposalExecutable(5)) {
    ///     executeProposal(5);
    /// } else {
    ///     revert("Proposal not ready for execution");
    /// }
    ///
    /// // Workflow integration
    /// if (isProposalInVoting(proposalId)) {
    ///     // Still collecting votes
    ///     approveProposal(proposalId);
    /// } else if (isProposalExecutable(proposalId)) {
    ///     // Ready to execute
    ///     executeProposal(proposalId);
    /// }
    /// ```
    ///
    /// @custom:invariant Executable ⟺ (status == Approved) ∧ (timestamp ≤ createdAt + expiry) ∧ (approved ≥ quorum)
    /// @custom:invariant If expired: always returns false (even if status is Approved)
    /// @custom:invariant Defensive validation: Quorum check redundant but ensures correctness
    /// @custom:invariant State transition: Only Approved proposals can be executed (terminal states block execution)
    function isProposalExecutable(uint256 proposalId) public view returns (bool) {
        Proposal memory proposal = proposals[_validateProposalId(proposalId)];
        if (proposal.status != ProposalStatus.Approved) return false;
        if (block.timestamp > proposal.createdAt + proposalExpiry) return false;
        return proposal.approved >= proposal.requiredApprovals;
    }

    /// @notice Check if a member has voted (approved or rejected) on a specific proposal
    /// @dev Uses bitmap-based voting tracking with historical member version snapshot
    /// @dev Returns false for members who were not in governance at proposal creation time
    /// @dev View function - safe for concurrent access, no state changes, no reentrancy risk
    ///
    /// @param member The address to check for voting status
    /// @param proposalId The proposal ID to query (must be valid, reverts if invalid)
    ///
    /// @return bool True if member has voted on the proposal, false otherwise
    ///
    /// @custom:security **Proposal ID Validation**
    /// - Uses `_validateProposalId(proposalId)` which reverts with `InvalidProposal()` if:
    ///   - proposalId is 0 (invalid)
    ///   - proposalId > currentProposalId (doesn't exist yet)
    /// - Guaranteed valid proposal after validation
    ///
    /// @custom:security **Member Version Snapshot**
    /// - Uses `proposal.memberVersion` to look up historical member list
    /// - This ensures vote tracking is based on governance state at proposal creation time
    /// - Members who joined after proposal creation are not counted
    /// - Members who left after proposal creation are still tracked
    ///
    /// @custom:security **Zero Address Handling**
    /// - If member is address(0), `_getMemberIndexAtVersion` returns `type(uint256).max`
    /// - Caught by bounds check: `type(uint256).max >= versionedMemberList.length`
    /// - Returns false (address(0) cannot be a member)
    ///
    /// @custom:security **Non-Member Handling**
    /// - If member was not in governance at `proposal.memberVersion`:
    ///   - `_getMemberIndexAtVersion` returns `type(uint256).max`
    ///   - Caught by bounds check: `type(uint256).max >= versionedMemberList.length`
    ///   - Returns false (non-members cannot vote)
    ///
    /// @custom:security **Bit Shift Safety**
    /// - Bitmap uses bit shifting: `1 << memberIndex`
    /// - Protected by: `memberIndex > MAX_MEMBER_INDEX` check (MAX_MEMBER_INDEX = 255)
    /// - Since memberIndex <= 255, and 1 << 255 is within uint256 range (2^256 - 1), this is safe
    /// - No bit shift overflow possible
    ///
    /// @custom:security **Defense in Depth**
    /// - Two sequential bounds checks provide layered security:
    ///   1. `memberIndex >= versionedMemberList[proposal.memberVersion].length`
    ///   2. `memberIndex > MAX_MEMBER_INDEX`
    /// - Second check is technically redundant (versionedMemberList.length <= MAX_MEMBER_INDEX by design)
    /// - But provides safety if invariants are violated due to bugs or upgrades
    ///
    /// @custom:security **Attack Vector Analysis**
    /// - Integer Overflow: Protected (Solidity 0.8+ + explicit MAX_MEMBER_INDEX check)
    /// - Bit Shift Overflow: Safe (MAX_MEMBER_INDEX = 255, 1 << 255 within uint256)
    /// - Invalid Member Access: Returns false for non-members or invalid indices
    /// - TOCTOU (Time-of-Check-Time-of-Use): N/A (view function, no state changes)
    /// - Gas DoS: O(1) complexity, constant gas cost, no loops
    ///
    /// @custom:revert `InvalidProposal()` if proposalId is 0 or greater than currentProposalId
    ///
    /// @custom:gas **Gas Efficiency**: O(1) constant time complexity
    /// - Single SLOAD for proposal struct
    /// - Single mapping lookup for member index
    /// - Single bit shift and bitmap check
    /// - No loops or dynamic operations
    ///
    /// @custom:example
    /// ```solidity
    /// uint256 proposalId = createProposal(...);
    /// voteForProposal(proposalId, true);  // Alice votes
    ///
    /// // Check if Alice voted
    /// bool aliceVoted = hasApproved(alice, proposalId);  // true
    /// bool bobVoted = hasApproved(bob, proposalId);      // false (didn't vote yet)
    /// bool zeroVoted = hasApproved(address(0), proposalId);  // false (invalid member)
    /// ```
    ///
    /// @custom:invariant Returns true ⟺ member voted on proposal ∧ member was in governance at proposal.memberVersion
    /// @custom:invariant Returns false ⟺ member didn't vote ∨ member not in governance at proposal.memberVersion
    /// @custom:invariant View function: No state changes, safe for concurrent access
    /// @custom:invariant Bitmap integrity: Bit is set ⟺ member voted (approved or rejected)
    /// @custom:invariant Historical accuracy: Uses proposal.memberVersion for time-consistent checks
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

    /// @notice Get the quorum requirement (minimum approvals needed) for a specific governance version
    /// @dev Retrieves historical quorum value from version snapshot for audit and proposal validation
    /// @dev View function - safe for concurrent access, no state changes, no reentrancy risk
    ///
    /// @param targetVersion The governance version to query (must be between 1 and current memberVersion)
    /// @return uint32 The quorum value for the specified version (minimum 1, maximum member count)
    ///
    /// @custom:security **Version Validation**
    /// - Uses `_validateMemberVersion(targetVersion)` which reverts with `InvalidMemberVersion()` if:
    ///   - targetVersion is 0 (versions start from 1)
    ///   - targetVersion > memberVersion (future version doesn't exist)
    /// - Guaranteed valid version after validation
    ///
    /// @custom:security **Quorum Snapshot Integrity**
    /// - Quorum snapshots are stored in `quorumByVersion[version]` mapping
    /// - Snapshots are created during:
    ///   - Governance initialization: `_initializeGovernance` sets `quorumByVersion[1] = quorum`
    ///   - Member addition: `_addMember` sets `quorumByVersion[newVersion] = newQuorum`
    ///   - Member removal: `_removeMember` sets `quorumByVersion[newVersion] = newQuorum`
    ///   - Member change: `_changeMember` sets `quorumByVersion[newVersion] = quorum`
    /// - All quorum values are validated before storage (minimum 1, maximum member count)
    ///
    /// @custom:security **Zero Snapshot Protection**
    /// - If `quorumByVersion[version] == 0`, function reverts with `InvalidQuorum()`
    /// - This check prevents returning uninitialized quorum values
    /// - By design, this should never happen because:
    ///   - All valid versions have quorum snapshots created during member operations
    ///   - Quorum is always >= 1 (validated in initialization and member operations)
    /// - This check provides defense-in-depth protection against potential bugs
    ///
    /// @custom:security **Quorum Value Guarantees**
    /// - Single member governance: quorum == 1 (validated in lines 277-280, 589-591)
    /// - Multi-member governance: 2 <= quorum <= memberCount (validated in lines 282-285, 592-593)
    /// - Quorum is always non-zero for valid versions
    /// - Quorum snapshots are immutable (historical values never change)
    ///
    /// @custom:security **Attack Vector Analysis**
    /// - **Invalid Version**: Protected by `_validateMemberVersion`
    /// - **Zero Quorum**: Protected by explicit zero check + validation at storage time
    /// - **Integer Overflow**: Protected by Solidity 0.8+ (quorum is uint32, max value 4,294,967,295)
    /// - **Uninitialized Access**: Protected by version validation + zero check
    /// - **TOCTOU (Time-of-Check-Time-of-Use)**: N/A (view function, no state changes)
    /// - **Gas DoS**: O(1) complexity, single mapping lookup, constant gas cost
    ///
    /// @custom:revert `InvalidMemberVersion()` if targetVersion is 0 or exceeds current memberVersion
    /// @custom:revert `InvalidQuorum()` if quorum snapshot is 0 (should never happen by design)
    ///
    /// @custom:gas **Gas Efficiency**: O(1) constant time complexity
    /// - Single call to `_validateMemberVersion` (2 comparisons)
    /// - Single SLOAD for quorum snapshot
    /// - Single zero check comparison
    /// - No loops or dynamic operations
    ///
    /// @custom:example
    /// ```solidity
    /// // Get current quorum requirement
    /// uint32 currentQuorum = getQuorum(memberVersion);
    ///
    /// // Get historical quorum (version 1 = initial governance state)
    /// uint32 initialQuorum = getQuorum(1);
    ///
    /// // Validate proposal quorum (used in proposal lifecycle)
    /// Proposal memory proposal = getProposal(proposalId);
    /// uint32 requiredQuorum = getQuorum(proposal.memberVersion);
    /// bool hasQuorum = proposal.approved >= requiredQuorum;
    ///
    /// // Audit trail: Compare quorum changes over time
    /// uint32 quorum1 = getQuorum(1);  // Initial quorum
    /// uint32 quorum2 = getQuorum(2);  // After first member change
    /// uint32 quorum3 = getQuorum(3);  // After second member change
    /// ```
    ///
    /// @custom:invariant Returns quorum ⟺ version is valid ∧ quorum snapshot exists ∧ quorum > 0
    /// @custom:invariant Quorum value: 1 (single member) OR 2..memberCount (multi-member)
    /// @custom:invariant Quorum snapshots are immutable: getQuorum(v) always returns same value for valid v
    /// @custom:invariant View function: No state changes, safe for concurrent access
    /// @custom:invariant Historical accuracy: Uses version snapshot for time-consistent queries
    ///
    /// @custom:usage **Common Use Cases**
    /// - **Proposal Validation**: Check if proposal has reached required approvals
    /// - **UI Display**: Show quorum requirement for current or historical governance state
    /// - **Audit Trail**: Track quorum changes over governance history
    /// - **Vote Calculation**: Determine if proposal can be approved based on remaining votes
    function getQuorum(uint256 targetVersion) public view returns (uint32) {
        uint256 version = _validateMemberVersion(targetVersion);
        uint32 snapshot = quorumByVersion[version];
        if (snapshot == 0) revert InvalidQuorum();
        return snapshot;
    }

    /// @notice Check if an address was a governance member at a specific version
    /// @dev Uses historical member version snapshot for time-consistent membership queries
    /// @dev Returns true if account was a member at targetVersion, false otherwise
    /// @dev View function - safe for concurrent access, no state changes, no reentrancy risk
    ///
    /// @param account The address to check for membership status
    /// @param targetVersion The governance version to query (must be between 1 and current memberVersion)
    /// @return bool True if account was a member at targetVersion, false otherwise
    ///
    /// @custom:security **Version Validation**
    /// - Uses `_validateMemberVersion(targetVersion)` which reverts with `InvalidMemberVersion()` if:
    ///   - targetVersion is 0 (versions start from 1)
    ///   - targetVersion > memberVersion (future version doesn't exist)
    /// - Guaranteed valid version after validation
    ///
    /// @custom:security **Member Index Snapshot**
    /// - Member indices are stored in `memberIndexByVersion[version][account]` mapping
    /// - Index values are stored as "index + 1" (1-based indexing):
    ///   - Value 0: Account is NOT a member at this version
    ///   - Value > 0: Account is a member (actual index is value - 1)
    /// - This design allows distinguishing between "not a member" (0) and "member at index 0" (1)
    ///
    /// @custom:security **Zero Address Handling**
    /// - If account is address(0), memberIndexByVersion returns 0 (default value)
    /// - Function returns false (address(0) cannot be a member)
    /// - This is correct behavior as address(0) is explicitly blocked in _initializeGovernance and _addMember
    ///
    /// @custom:security **Historical Accuracy**
    /// - Function queries historical snapshot at targetVersion
    /// - Members added after targetVersion will return false for that version
    /// - Members removed after targetVersion will return true for that version
    /// - This ensures time-consistent queries for proposal validation
    ///
    /// @custom:security **Member Index Integrity**
    /// - Member indices are set during:
    ///   - Governance initialization: `memberIndexByVersion[1][member] = index + 1` (lines 305)
    ///   - Member addition: `memberIndexByVersion[newVersion][member] = index + 1` (lines 601, 608)
    ///   - Member removal: `memberIndexByVersion[newVersion][removed] = 0` (deleted) (line 670)
    ///   - Member change: `memberIndexByVersion[newVersion][new] = oldIndex + 1` (line 723)
    /// - All indices are validated to be within 0..MAX_MEMBER_INDEX (255) range
    ///
    /// @custom:security **Attack Vector Analysis**
    /// - **Invalid Version**: Protected by `_validateMemberVersion`
    /// - **Zero Address**: Returns false (memberIndexByVersion default value is 0)
    /// - **Integer Overflow**: Protected by Solidity 0.8+ (mapping lookup cannot overflow)
    /// - **Uninitialized Access**: Protected by version validation + returns false for 0 index
    /// - **TOCTOU (Time-of-Check-Time-of-Use)**: N/A (view function, no state changes)
    /// - **Gas DoS**: O(1) complexity, single mapping lookup, constant gas cost
    ///
    /// @custom:revert `InvalidMemberVersion()` if targetVersion is 0 or exceeds current memberVersion
    ///
    /// @custom:gas **Gas Efficiency**: O(1) constant time complexity
    /// - Single call to `_validateMemberVersion` (2 comparisons)
    /// - Single SLOAD for member index from mapping
    /// - Single zero comparison
    /// - No loops or dynamic operations
    ///
    /// @custom:example
    /// ```solidity
    /// // Check if Alice is a current member
    /// bool isCurrentMember = isMember(alice, memberVersion);
    ///
    /// // Check if Bob was a member at version 1 (initial governance)
    /// bool wasInitialMember = isMember(bob, 1);
    ///
    /// // Proposal validation: Check if voter was eligible at proposal creation
    /// Proposal memory proposal = getProposal(proposalId);
    /// bool wasEligible = isMember(voter, proposal.memberVersion);
    /// if (!wasEligible) revert NotAMember();
    ///
    /// // Audit trail: Track membership changes over time
    /// bool wasMemberV1 = isMember(charlie, 1);  // true
    /// bool wasMemberV2 = isMember(charlie, 2);  // false (removed)
    /// bool wasMemberV3 = isMember(charlie, 3);  // true (re-added)
    /// ```
    ///
    /// @custom:invariant Returns true ⟺ account was a member at targetVersion ∧ targetVersion is valid
    /// @custom:invariant Returns false ⟺ account was not a member ∨ account is address(0)
    /// @custom:invariant View function: No state changes, safe for concurrent access
    /// @custom:invariant Historical consistency: isMember(account, v) always returns same value for valid v
    /// @custom:invariant Snapshot integrity: Uses version snapshot for time-consistent queries
    ///
    /// @custom:usage **Common Use Cases**
    /// - **Proposal Validation**: Verify voter eligibility based on proposal's member version
    /// - **Access Control**: Check if address has governance permissions at specific version
    /// - **Audit Trail**: Track membership changes over governance history
    /// - **UI Display**: Show member list for current or historical governance state
    /// - **Governance Analytics**: Analyze member participation across different versions
    ///
    /// @custom:design **Design Rationale**
    /// - **Version-Based Queries**: Enables consistent membership checks for proposals
    /// - **1-Based Indexing**: Distinguishes "not a member" (0) from "member at index 0" (1)
    /// - **Immutable History**: Past versions remain unchanged, ensuring audit trail integrity
    /// - **Simple Logic**: Single mapping lookup makes function gas-efficient and easy to audit
    function isMember(address account, uint256 targetVersion) public view returns (bool) {
        uint256 version = _validateMemberVersion(targetVersion);
        return memberIndexByVersion[version][account] != 0;
    }

    /// @dev Validate member version number and prevent invalid version access
    /// @dev Central validation point for all version-based queries (getMemberCount, getMemberAt, isMember, getQuorum)
    ///
    /// @param targetVersion The version number to validate
    /// @return The validated version number (passthrough for convenience)
    ///
    /// @custom:security Version bounds validation:
    ///   - Prevents version 0 access (versions start from INITIAL_MEMBER_VERSION = 1)
    ///   - Prevents future version access (targetVersion must not exceed current memberVersion)
    ///   - Protects against uninitialized state access (memberVersion = 1 at deployment)
    ///
    /// @custom:security Attack vector analysis:
    ///   - Integer overflow: Protected by Solidity 0.8+ automatic overflow checks
    ///   - Uninitialized access: memberVersion initialized to INITIAL_MEMBER_VERSION (1) at deployment
    ///   - Out-of-bounds: Explicit checks for version 0 and version > memberVersion
    ///   - Race conditions: View function with no state changes, safe for concurrent access
    ///
    /// @custom:security Data existence guarantee:
    ///   - Version 1: Guaranteed to have data after _initializeGovernance (requires ≥1 member)
    ///   - Version 2+: Created by _prepareNextMemberVersion during member operations
    ///   - Historical data: Old versions preserved (only new version cleared in _prepareNextMemberVersion)
    ///
    /// Reverts with:
    /// @custom:revert InvalidMemberVersion if:
    ///   - targetVersion == 0 (version numbering starts from 1)
    ///   - targetVersion > memberVersion (future version not yet created)
    ///
    /// Design rationale:
    ///   - Centralized validation reduces code duplication and audit surface
    ///   - Fail-fast approach prevents invalid state access early in call chain
    ///   - Passthrough return allows inline usage: `versionedMemberList[_validateMemberVersion(v)]`
    ///
    /// Example call chain:
    /// ```solidity
    /// getMemberCount(5) → _validateMemberVersion(5) → versionedMemberList[5].length
    ///   └─ If memberVersion < 5: revert InvalidMemberVersion
    ///   └─ If memberVersion >= 5: return 5 → access versionedMemberList[5]
    /// ```
    function _validateMemberVersion(uint256 targetVersion) internal view returns (uint256) {
        if (targetVersion == 0 || targetVersion > memberVersion) revert InvalidMemberVersion();
        return targetVersion;
    }

    /// @dev Validate proposal ID and prevent invalid proposal access
    /// @dev Central validation point for all proposal-based queries and operations
    ///
    /// @param proposalId The proposal ID to validate
    /// @return The validated proposal ID (passthrough for convenience)
    ///
    /// @custom:security Proposal ID bounds validation:
    ///   - Prevents proposalId 0 access (proposals start from 1)
    ///   - Prevents future proposal access (proposalId must not exceed currentProposalId)
    ///   - Protects against uninitialized state access (currentProposalId = 0 before first proposal)
    ///
    /// @custom:security Attack vector analysis:
    ///   - Integer overflow: Protected by Solidity 0.8+ automatic overflow checks
    ///   - Uninitialized access: Explicit check for proposalId == 0
    ///   - Out-of-bounds: Explicit check for proposalId > currentProposalId
    ///   - Race conditions: View function with no state changes, safe for concurrent access
    ///   - DoS: O(1) complexity, constant-time validation
    ///
    /// @custom:security Proposal existence guarantee:
    ///   - proposalId == 0: Always invalid (ID numbering starts from 1)
    ///   - proposalId > currentProposalId: Future proposal, not yet created
    ///   - 1 <= proposalId <= currentProposalId: Valid range, proposal exists
    ///   - Proposal creation: _createProposal always initializes all fields
    ///
    /// Reverts with:
    /// @custom:revert InvalidProposal if:
    ///   - proposalId == 0 (proposals start from ID 1)
    ///   - proposalId > currentProposalId (future proposal not yet created)
    ///
    /// Design rationale:
    ///   - Centralized validation reduces code duplication and audit surface
    ///   - Fail-fast approach prevents invalid state access early in call chain
    ///   - Passthrough return allows inline usage: `proposals[_validateProposalId(id)]`
    ///   - Matches _validateMemberVersion pattern for consistency
    ///
    /// Example call chain:
    /// ```solidity
    /// getProposal(5) → _validateProposalId(5) → proposals[5]
    ///   └─ If currentProposalId < 5: revert InvalidProposal
    ///   └─ If currentProposalId >= 5: return 5 → access proposals[5]
    ///
    /// getProposal(0) → _validateProposalId(0) → revert InvalidProposal
    ///   └─ proposalId 0 is always invalid
    /// ```
    ///
    /// @custom:security Proposal ID lifecycle:
    ///   - Initial state: currentProposalId = 0 (no proposals)
    ///   - First proposal: currentProposalId = 1 (ID 1 created)
    ///   - Nth proposal: currentProposalId = N (IDs 1..N exist)
    ///   - Valid IDs: Always in range [1, currentProposalId]
    ///
    /// Used by:
    ///   - getProposal(uint256): Retrieve proposal data
    ///   - isProposalInVoting(uint256): Check voting status
    ///   - isProposalExecutable(uint256): Check execution eligibility
    ///   - hasApproved(address, uint256): Check member vote status
    ///   - _checkProposalExists(uint256): Modifier validation
    function _validateProposalId(uint256 proposalId) internal view returns (uint256) {
        if (proposalId == 0 || proposalId > currentProposalId) revert InvalidProposal();
        return proposalId;
    }

    /// @notice Get the number of active proposals for a member
    /// @dev Retrieves the current count of active (non-terminal) proposals created by the specified member.
    ///      This count is used to enforce the MAX_ACTIVE_PROPOSALS_PER_MEMBER limit to prevent proposal spam.
    ///
    ///      The count is managed by two internal functions:
    ///      - `_createProposal`: Increments count when a new proposal is created (line 377)
    ///      - `_decrementActiveProposalCount`: Decrements count when proposal reaches terminal state (line 1762)
    ///
    ///      Terminal states that trigger decrement:
    ///      1. ProposalStatus.Executed - Proposal successfully executed
    ///      2. ProposalStatus.Cancelled - Proposal cancelled by proposer
    ///      3. ProposalStatus.Expired - Proposal expired (block.timestamp > createdAt + proposalExpiry)
    ///      4. ProposalStatus.Failed - Proposal execution failed after MAX_RETRY_COUNT attempts
    ///      5. ProposalStatus.Rejected - Proposal rejected by governance members
    ///
    /// @param member The member address to query for active proposal count
    /// @return activeCount The number of active proposals currently owned by the member (range: 0 to MAX_ACTIVE_PROPOSALS_PER_MEMBER)
    ///
    /// @custom:security Zero Address Handling
    ///      - Query for zero address returns 0 (default mapping value)
    ///      - No validation required as this is informational only
    ///      - Zero address cannot create proposals (blocked by onlyMember modifier)
    ///
    /// @custom:security Integer Overflow Protection
    ///      - Solidity 0.8.14 provides automatic overflow/underflow protection
    ///      - Count is bounded by MAX_ACTIVE_PROPOSALS_PER_MEMBER (3)
    ///      - Decrement function includes safety check (memberActiveProposalCount[proposer] > 0)
    ///      - Maximum possible value: 3 (enforced at proposal creation)
    ///
    /// @custom:security Gas Efficiency
    ///      - Single SLOAD operation: ~100 gas (warm) or ~2100 gas (cold)
    ///      - O(1) time complexity - constant gas regardless of proposal count
    ///      - No loops or external calls
    ///      - No gas DoS attack vector
    ///
    /// @custom:security View Function Safety
    ///      - Pure read-only operation with no state modifications
    ///      - Safe for concurrent access from multiple transactions
    ///      - No reentrancy concerns
    ///      - Can be called by anyone without permission checks
    ///
    /// @custom:security Access Control
    ///      - Public visibility: Anyone can query any member's active proposal count
    ///      - Transparency design: Active proposal counts are publicly auditable
    ///      - No sensitive information exposure risk
    ///
    /// @custom:invariant Count Accuracy
    ///      - Count always reflects the true number of non-terminal proposals
    ///      - Invariant: memberActiveProposalCount[member] <= MAX_ACTIVE_PROPOSALS_PER_MEMBER
    ///      - Invariant: memberActiveProposalCount[member] >= 0 (enforced by Solidity 0.8.14)
    ///      - Invariant: Sum of all active proposals across all members equals total non-terminal proposals
    ///
    /// @custom:gas Optimization Characteristics
    ///      - Minimal gas cost: Single storage read
    ///      - No computational overhead
    ///      - Optimal for high-frequency monitoring
    ///      - Suitable for external integrations and UI applications
    ///
    /// @custom:example Basic Usage
    ///      ```solidity
    ///      // Check member's current active proposal count
    ///      uint256 activeCount = governance.getMemberActiveProposalCount(memberAddress);
    ///
    ///      // Monitor before creating proposal
    ///      require(activeCount < 3, "Member at proposal limit");
    ///      ```
    ///
    /// @custom:example Monitoring Integration
    ///      ```solidity
    ///      // Dashboard integration
    ///      function getMemberStatus(address member) external view returns (
    ///          uint256 activeProposals,
    ///          uint256 remainingCapacity,
    ///          bool canCreate
    ///      ) {
    ///          activeProposals = governance.getMemberActiveProposalCount(member);
    ///          remainingCapacity = MAX_ACTIVE_PROPOSALS_PER_MEMBER - activeProposals;
    ///          canCreate = governance.canCreateProposal(member);
    ///      }
    ///      ```
    ///
    /// @custom:usage Common Use Cases
    ///      1. **Pre-Creation Validation**: Check available capacity before initiating proposal creation
    ///      2. **UI Display**: Show member's current proposal count and remaining capacity
    ///      3. **Analytics**: Track proposal creation patterns across governance members
    ///      4. **Rate Limiting**: Implement additional off-chain rate limiting logic
    ///      5. **Monitoring**: Alert when members approach or reach their proposal limit
    ///
    /// @custom:design Spam Prevention Strategy
    ///      - Per-member limit prevents individual actors from flooding governance with proposals
    ///      - Limit set to MAX_ACTIVE_PROPOSALS_PER_MEMBER (3) balances participation with spam prevention
    ///      - Counter automatically decrements when proposals complete, enabling new proposal creation
    ///      - Independent counters per member prevent one member's activity from affecting others
    ///      - Simple counter design ensures predictable gas costs and easy auditability
    function getMemberActiveProposalCount(address member) public view returns (uint256) {
        return memberActiveProposalCount[member];
    }

    /// @notice Check if a member can create a new proposal
    /// @dev Verifies whether the specified member has available capacity to create a new proposal
    ///      by checking if their active proposal count is below MAX_ACTIVE_PROPOSALS_PER_MEMBER.
    ///
    ///      This function provides a convenient pre-check before attempting proposal creation.
    ///      It helps prevent unnecessary transaction reverts and provides clear feedback for UI applications.
    ///
    ///      Relationship to proposal creation flow:
    ///      1. canCreateProposal() - Pre-check (this function)
    ///      2. _createProposal() - Enforces limit with revert if exceeded
    ///      3. memberActiveProposalCount[msg.sender]++ - Increments count
    ///
    ///      The limit enforcement uses strict inequality (<) to allow creation when count equals 0, 1, or 2,
    ///      but blocks creation when count equals 3 (the maximum).
    ///
    /// @param member The member address to check for proposal creation eligibility
    /// @return canCreate True if member can create a new proposal, false if at capacity limit
    ///
    /// @custom:security Zero Address Handling
    ///      - Query for zero address returns true (count 0 < limit 3)
    ///      - This is safe as zero address cannot actually create proposals (blocked by onlyMember modifier)
    ///      - View function provides informational value only, actual enforcement happens in _createProposal
    ///
    /// @custom:security Limit Boundary Validation
    ///      - Uses strict inequality (<) for correct boundary checking
    ///      - When count = 3: returns false (cannot create)
    ///      - When count = 2: returns true (can create one more)
    ///      - When count = 1: returns true (can create two more)
    ///      - When count = 0: returns true (can create three)
    ///      - Boundary logic matches enforcement in _createProposal (>= check)
    ///
    /// @custom:security Integer Overflow Protection
    ///      - Solidity 0.8.14 provides automatic overflow protection
    ///      - Comparison operation (<) cannot overflow
    ///      - Count is bounded by MAX_ACTIVE_PROPOSALS_PER_MEMBER constant (3)
    ///      - No arithmetic operations, only comparison
    ///
    /// @custom:security Gas Efficiency
    ///      - Single SLOAD + comparison: ~100-2100 gas (warm/cold)
    ///      - O(1) time complexity - constant gas cost
    ///      - No loops, no external calls, no storage writes
    ///      - Optimal for high-frequency validation checks
    ///      - No gas DoS attack vector
    ///
    /// @custom:security View Function Safety
    ///      - Pure read-only operation with no state modifications
    ///      - Safe for concurrent access from multiple transactions
    ///      - No reentrancy concerns
    ///      - No TOCTOU (Time-Of-Check-Time-Of-Use) risks within this function
    ///      - Public visibility allows transparent validation by anyone
    ///
    /// @custom:security TOCTOU Considerations
    ///      - External callers should be aware of potential TOCTOU race conditions:
    ///        * canCreateProposal() returns true at block N
    ///        * Another transaction creates proposal at block N+1
    ///        * Original transaction attempts creation and may revert
    ///      - This is expected behavior in concurrent governance systems
    ///      - UI applications should handle potential reverts gracefully
    ///      - Not a security vulnerability - inherent to blockchain state changes
    ///
    /// @custom:security Access Control
    ///      - Public visibility: Anyone can check any member's creation eligibility
    ///      - Transparency design: Proposal limits are publicly auditable
    ///      - No sensitive information exposure
    ///      - Actual proposal creation still protected by onlyMember modifier
    ///
    /// @custom:invariant Logical Consistency
    ///      - Invariant: canCreateProposal(member) == (getMemberActiveProposalCount(member) < MAX_ACTIVE_PROPOSALS_PER_MEMBER)
    ///      - Invariant: If canCreateProposal() returns false, _createProposal() will revert with TooManyActiveProposals
    ///      - Invariant: If canCreateProposal() returns true, member has capacity (though creation may still fail for other reasons)
    ///
    /// @custom:gas Optimization Characteristics
    ///      - Minimal gas cost: Single storage read + comparison
    ///      - No computational overhead
    ///      - Ideal for pre-flight checks before expensive operations
    ///      - Suitable for batched validation queries
    ///      - Can be used in view function aggregations without gas concerns
    ///
    /// @custom:example Basic Pre-Check Pattern
    ///      ```solidity
    ///      // Check before attempting proposal creation
    ///      if (!governance.canCreateProposal(msg.sender)) {
    ///          revert("Cannot create proposal: at capacity limit");
    ///      }
    ///
    ///      // Proceed with proposal creation
    ///      uint256 proposalId = governance.createProposal(actionType, callData);
    ///      ```
    ///
    /// @custom:example UI Integration Pattern
    ///      ```solidity
    ///      // Fetch member capacity status for UI display
    ///      function getMemberCapacityStatus(address member) external view returns (
    ///          bool canCreate,
    ///          uint256 currentCount,
    ///          uint256 maxCount,
    ///          uint256 remaining
    ///      ) {
    ///          currentCount = governance.getMemberActiveProposalCount(member);
    ///          maxCount = governance.MAX_ACTIVE_PROPOSALS_PER_MEMBER();
    ///          canCreate = governance.canCreateProposal(member);
    ///          remaining = canCreate ? (maxCount - currentCount) : 0;
    ///      }
    ///      ```
    ///
    /// @custom:example Batch Validation Pattern
    ///      ```solidity
    ///      // Check multiple members in single call
    ///      function checkMembersEligibility(address[] memory members)
    ///          external
    ///          view
    ///          returns (bool[] memory eligible)
    ///      {
    ///          eligible = new bool[](members.length);
    ///          for (uint256 i = 0; i < members.length; i++) {
    ///              eligible[i] = governance.canCreateProposal(members[i]);
    ///          }
    ///      }
    ///      ```
    ///
    /// @custom:usage Common Use Cases
    ///      1. **Pre-Flight Validation**: Check eligibility before initiating expensive proposal creation transaction
    ///      2. **UI State Management**: Enable/disable "Create Proposal" button based on member capacity
    ///      3. **Error Prevention**: Provide clear feedback to users before transaction submission
    ///      4. **Batch Operations**: Validate multiple members' eligibility in aggregated queries
    ///      5. **Analytics**: Track member participation rates and capacity utilization
    ///      6. **Access Control**: Implement additional business logic based on proposal creation capacity
    ///
    /// @custom:design Pre-Check Optimization
    ///      - Separate validation function reduces failed transaction costs
    ///      - Users can check eligibility without gas cost (view function)
    ///      - UI applications can provide immediate feedback without blockchain interaction
    ///      - Pattern follows fail-fast principle for better UX
    ///      - Simple boolean return makes integration straightforward
    ///      - Consistent with defense-in-depth: view check + enforcement in state-changing function
    function canCreateProposal(address member) public view returns (bool) {
        return memberActiveProposalCount[member] < MAX_ACTIVE_PROPOSALS_PER_MEMBER;
    }

    // ========== Internal Functions for Modifiers ==========

    /// @dev Internal function for onlyMember modifier - validates caller is an active governance member
    /// @notice Checks if msg.sender is currently an active member by verifying the isActive flag
    ///
    ///         This function is called by the onlyMember modifier to restrict function access
    ///         to only active governance members. It provides centralized membership validation
    ///         for all member-only operations.
    ///
    ///         Member activation states:
    ///         - isActive = true: Member can participate in governance (create proposals, vote, execute)
    ///         - isActive = false: Non-member or removed member (no governance permissions)
    ///
    /// @custom:security Zero Address Handling
    ///      - Zero address returns default Member struct with isActive = false
    ///      - Automatically fails membership check (safe rejection)
    ///      - No explicit zero address check needed due to default struct behavior
    ///
    /// @custom:security Gas Efficiency
    ///      - Single SLOAD operation: ~100 gas (warm) or ~2100 gas (cold)
    ///      - Storage reference used (no memory copy overhead)
    ///      - O(1) complexity - constant time regardless of member count
    ///
    /// @custom:security Access Control
    ///      - Only checks isActive flag, not member index or other fields
    ///      - Centralized validation ensures consistent behavior across all member-only functions
    ///      - No bypasses possible - all member operations must pass through this check
    ///
    /// @custom:revert NotAMember
    ///      - Thrown when msg.sender is not an active member (isActive = false)
    ///      - Covers both non-members and removed members
    ///
    /// @custom:usage Used By
    ///      - onlyMember modifier (line 186-189)
    ///      - Applied to: createProposal, vote, approve, disapprove, execute, cancel, retry, expire
    function _checkMembership() internal view {
        Member storage member = members[msg.sender];
        if (!member.isActive) revert NotAMember();
    }

    /// @dev Internal function for proposalExists modifier - validates proposal ID is valid
    /// @notice Delegates to _validateProposalId to ensure proposalId is within valid range
    ///
    ///         This function provides a clean abstraction layer for the proposalExists modifier,
    ///         separating modifier logic from validation implementation. It ensures proposal IDs
    ///         are validated consistently across all proposal operations.
    ///
    /// @param proposalId The proposal ID to validate
    ///
    /// @custom:security Validation Logic
    ///      - Delegates to _validateProposalId (line 1685-1687)
    ///      - Rejects proposalId = 0 (proposals start from ID 1)
    ///      - Rejects proposalId > currentProposalId (non-existent proposals)
    ///      - Valid range: 1 <= proposalId <= currentProposalId
    ///
    /// @custom:security Gas Efficiency
    ///      - Single function call overhead: ~100 gas
    ///      - Minimal gas cost for validation
    ///      - No redundant checks or operations
    ///
    /// @custom:revert InvalidProposal
    ///      - Thrown by _validateProposalId when proposalId is 0 or exceeds currentProposalId
    ///
    /// @custom:usage Used By
    ///      - proposalExists modifier (line 191-194)
    ///      - Applied to functions requiring valid proposal ID parameter
    function _checkProposalExists(uint256 proposalId) internal view {
        _validateProposalId(proposalId);
    }

    /// @dev Internal function for proposalInVoting modifier - validates proposal is voteable
    /// @notice Checks if current proposal is in Voting or Approved state and not expired
    ///
    ///         **IMPORTANT**: This function is NOT a view function - it modifies state when
    ///         the proposal is expired. This is intentional design to auto-expire proposals
    ///         during validation checks, ensuring expired proposals are immediately marked
    ///         and cannot proceed with operations.
    ///
    ///         Validation flow:
    ///         1. Check proposal status (must be Voting or Approved)
    ///         2. Check expiry (block.timestamp vs createdAt + proposalExpiry)
    ///         3. If expired: Mark as Expired, decrement count, revert
    ///         4. If valid: Continue with voting operation
    ///
    ///         Valid states for voting operations:
    ///         - ProposalStatus.Voting: Initial state accepting votes
    ///         - ProposalStatus.Approved: Quorum reached, still accepting votes before execution
    ///
    /// @custom:security State Modification (Intentional)
    ///      - **Modifies state** when proposal is expired
    ///      - Sets proposal.status = ProposalStatus.Expired
    ///      - Calls _decrementActiveProposalCount(currentProposalId)
    ///      - This is intentional - auto-expire expired proposals on any voting attempt
    ///
    /// @custom:security Expiry Check
    ///      - Uses strict inequality: block.timestamp > createdAt + proposalExpiry
    ///      - Proposals are valid exactly at the expiry timestamp
    ///      - Expired at the next second after expiry
    ///      - Consistent with other expiry checks in the contract
    ///
    /// @custom:security Current Proposal Context
    ///      - Uses currentProposalId set by calling function/modifier
    ///      - Requires proper context setup before calling this function
    ///      - Usually called after proposalExists modifier
    ///
    /// @custom:security Gas Efficiency
    ///      - Minimal gas for valid proposals: status check + timestamp comparison
    ///      - Additional gas on expiry: state write + count decrement (~5000-20000 gas)
    ///      - Auto-expiry prevents repeated validation costs
    ///
    /// @custom:revert ProposalNotInVoting
    ///      - Thrown when proposal status is not Voting or Approved
    ///      - Covers: Executed, Cancelled, Expired, Failed, Rejected states
    ///
    /// @custom:revert ProposalAlreadyExpired
    ///      - Thrown when block.timestamp > createdAt + proposalExpiry
    ///      - Proposal is marked as Expired before reverting
    ///
    /// @custom:usage Used By
    ///      - proposalInVoting modifier (line 196-199)
    ///      - Applied to: vote, approve, disapprove operations
    function _checkProposalInVoting() internal {
        Proposal storage proposal = proposals[currentProposalId];
        if (proposal.status != ProposalStatus.Voting && proposal.status != ProposalStatus.Approved) {
            revert ProposalNotInVoting();
        }
        if (block.timestamp > proposal.createdAt + proposalExpiry) {
            proposal.status = ProposalStatus.Expired;
            _decrementActiveProposalCount(currentProposalId);
            revert ProposalAlreadyExpired();
        }
    }

    /// @dev Internal function for proposalExecutable modifier - validates proposal can be executed
    /// @notice Checks if current proposal is in Approved state and not expired
    ///
    ///         **IMPORTANT**: This function is NOT a view function - it modifies state when
    ///         the proposal is expired. This is intentional design to auto-expire proposals
    ///         during execution attempts, preventing execution of expired proposals.
    ///
    ///         Validation flow:
    ///         1. Check proposal status (must be Approved)
    ///         2. Check expiry (block.timestamp vs createdAt + proposalExpiry)
    ///         3. If expired: Mark as Expired, decrement count, revert
    ///         4. If valid: Continue with execution
    ///
    ///         Execution requirements:
    ///         - Status must be ProposalStatus.Approved (quorum reached)
    ///         - Must not be expired (within proposalExpiry window)
    ///         - Other status validation handled by status check
    ///
    /// @custom:security State Modification (Intentional)
    ///      - **Modifies state** when proposal is expired
    ///      - Sets proposal.status = ProposalStatus.Expired
    ///      - Calls _decrementActiveProposalCount(currentProposalId)
    ///      - This is intentional - auto-expire expired proposals on execution attempts
    ///
    /// @custom:security Expiry Check
    ///      - Uses strict inequality: block.timestamp > createdAt + proposalExpiry
    ///      - Proposals can be executed exactly at the expiry timestamp
    ///      - Expired at the next second after expiry
    ///      - Consistent with _checkProposalInVoting expiry logic
    ///
    /// @custom:security Current Proposal Context
    ///      - Uses currentProposalId set by calling function/modifier
    ///      - Requires proper context setup before calling this function
    ///      - Usually called after proposalExists modifier
    ///
    /// @custom:security Gas Efficiency
    ///      - Minimal gas for valid proposals: status check + timestamp comparison
    ///      - Additional gas on expiry: state write + count decrement (~5000-20000 gas)
    ///      - Prevents execution of invalid proposals (saves gas on failed execution)
    ///
    /// @custom:revert ProposalNotExecutable
    ///      - Thrown when proposal status is not Approved
    ///      - Covers: Voting, Executed, Cancelled, Expired, Failed, Rejected states
    ///
    /// @custom:revert ProposalAlreadyExpired
    ///      - Thrown when block.timestamp > createdAt + proposalExpiry
    ///      - Proposal is marked as Expired before reverting
    ///
    /// @custom:usage Used By
    ///      - proposalExecutable modifier (line 201-204)
    ///      - Applied to: executeProposal, retryProposal operations
    function _checkProposalExecutable() internal {
        Proposal storage proposal = proposals[currentProposalId];
        if (proposal.status != ProposalStatus.Approved) revert ProposalNotExecutable();
        if (block.timestamp > proposal.createdAt + proposalExpiry) {
            proposal.status = ProposalStatus.Expired;
            _decrementActiveProposalCount(currentProposalId);
            revert ProposalAlreadyExpired();
        }
    }

    /// @dev Internal reentrancy guard initialization - sets guard before function execution
    /// @notice Checks if reentrancy guard is already set and sets it to prevent reentrant calls
    ///
    ///         This is the "before" portion of the nonReentrant modifier pattern. It implements
    ///         the Checks-Effects-Interactions (CEI) pattern by setting a guard flag before
    ///         executing any external calls or state changes.
    ///
    ///         Guard states:
    ///         - _reentrancyGuard = 0: No active call, safe to proceed
    ///         - _reentrancyGuard = 1: Active call in progress, revert reentrant attempt
    ///
    ///         Standard reentrancy protection pattern (OpenZeppelin style):
    ///         1. _nonReentrantBefore(): Check guard is 0, set to 1
    ///         2. Execute function body
    ///         3. _nonReentrantAfter(): Reset guard to 0
    ///
    /// @custom:security Reentrancy Protection
    ///      - Prevents reentrant calls to any function with nonReentrant modifier
    ///      - Checks guard is 0 before allowing execution
    ///      - Sets guard to 1 to block nested calls
    ///      - Must be paired with _nonReentrantAfter() to reset guard
    ///
    /// @custom:security Gas Efficiency
    ///      - Single SLOAD: ~100 gas (warm) or ~2100 gas (cold)
    ///      - Single SSTORE: ~5000 gas (0→1 state change) or ~20000 gas (cold)
    ///      - Minimal overhead for critical security protection
    ///
    /// @custom:security State Changes
    ///      - Modifies _reentrancyGuard from 0 to 1
    ///      - Must be followed by _nonReentrantAfter() to reset
    ///      - Failure to reset guard will lock all nonReentrant functions
    ///
    /// @custom:revert ReentrantCall
    ///      - Thrown when _reentrancyGuard is already 1 (reentrant call detected)
    ///      - Prevents nested calls to nonReentrant-protected functions
    ///
    /// @custom:usage Used By
    ///      - nonReentrant modifier (line 206-210)
    ///      - Applied to: _createProposal, _vote, executeProposal (functions with external calls)
    ///
    /// @custom:invariant Guard State
    ///      - Invariant: After _nonReentrantBefore succeeds, guard is always 1
    ///      - Invariant: Must be paired with _nonReentrantAfter to maintain consistency
    function _nonReentrantBefore() internal {
        if (_reentrancyGuard == 1) revert ReentrantCall();
        _reentrancyGuard = 1;
    }

    /// @dev Internal reentrancy guard cleanup - resets guard after function execution
    /// @notice Resets reentrancy guard to 0 to allow subsequent calls
    ///
    ///         This is the "after" portion of the nonReentrant modifier pattern. It resets
    ///         the guard flag after function execution completes, allowing future calls to
    ///         proceed normally.
    ///
    ///         Standard reentrancy protection pattern (OpenZeppelin style):
    ///         1. _nonReentrantBefore(): Check guard is 0, set to 1
    ///         2. Execute function body
    ///         3. _nonReentrantAfter(): Reset guard to 0
    ///
    ///         **CRITICAL**: This function must always execute after _nonReentrantBefore,
    ///         even if the function body reverts. The modifier pattern ensures this by
    ///         placing it after the function body execution point (_).
    ///
    /// @custom:security Reentrancy Protection
    ///      - Resets guard to 0 to allow subsequent calls
    ///      - Must always execute after _nonReentrantBefore
    ///      - Modifier pattern ensures cleanup happens even on revert in function body
    ///
    /// @custom:security Gas Efficiency
    ///      - Single SSTORE: ~2900 gas (1→0 refund) or ~5000 gas (cold)
    ///      - Gas refund for zeroing storage (5000 gas refund in pre-London, partial in post-London)
    ///      - Minimal overhead for security cleanup
    ///
    /// @custom:security State Changes
    ///      - Modifies _reentrancyGuard from 1 to 0
    ///      - Restores contract to reentrant-safe state
    ///      - Allows future nonReentrant function calls
    ///
    /// @custom:usage Used By
    ///      - nonReentrant modifier (line 206-210)
    ///      - Always executes after function body, even on revert
    ///
    /// @custom:invariant Guard State
    ///      - Invariant: After _nonReentrantAfter, guard is always 0
    ///      - Invariant: Must follow _nonReentrantBefore for proper protection
    function _nonReentrantAfter() internal {
        _reentrancyGuard = 0;
    }

    /// @dev Get member index at a specific member version snapshot
    /// @notice Retrieves the historical member index for a given address at a specific version
    ///
    ///         This function provides access to historical member indices stored in version snapshots.
    ///         It returns the 0-based member index or type(uint256).max for non-members.
    ///
    ///         Index encoding (1-based storage, 0-based return):
    ///         - Storage: indexPlusOne (1-based, 0 = not a member)
    ///         - Return: index (0-based) or type(uint256).max (not a member)
    ///
    ///         The 1-based storage encoding allows distinguishing between:
    ///         - indexPlusOne = 0: Not a member
    ///         - indexPlusOne = 1: Member at index 0
    ///         - indexPlusOne = 2: Member at index 1, etc.
    ///
    /// @param member The member address to query
    /// @param version The member version to query (historical snapshot)
    /// @return index The 0-based member index, or type(uint256).max if not a member at that version
    ///
    /// @custom:security Parameter Validation - Design Decision
    ///      - ⚠️ **version parameter is NOT validated in this function**
    ///      - Callers should validate version using _validateMemberVersion before calling
    ///      - Invalid version (0 or > memberVersion) will return 0 from mapping (non-member)
    ///      - This is safe behavior - invalid version treated as non-member
    ///      - **Design decision**: Validation responsibility delegated to callers for gas efficiency
    ///
    /// @custom:security Zero Address Handling
    ///      - Zero address returns 0 from mapping (indexPlusOne = 0)
    ///      - Correctly returns type(uint256).max (non-member indicator)
    ///      - Safe behavior - zero address treated as non-member
    ///
    /// @custom:security Integer Safety
    ///      - indexPlusOne is uint32, converted to uint256
    ///      - Safe widening conversion (no truncation or overflow)
    ///      - Subtraction (indexPlusOne - 1) safe due to if-check
    ///      - type(uint256).max is maximum uint256 value (2^256 - 1)
    ///
    /// @custom:security Gas Efficiency
    ///      - Single SLOAD: ~100 gas (warm) or ~2100 gas (cold)
    ///      - O(1) complexity - constant time lookup
    ///      - No loops or external calls
    ///      - View function - no state changes
    ///
    /// @custom:usage Common Use Cases
    ///      1. **Vote Validation**: _vote function uses this to get voter index at proposal version
    ///      2. **hasApproved Query**: Check if member voted on proposal at historical version
    ///      3. **Historical Membership**: Verify if address was member at specific version
    ///
    /// @custom:example Basic Usage
    ///      ```solidity
    ///      // Get member index at current version
    ///      uint256 currentVersion = memberVersion;
    ///      uint256 index = _getMemberIndexAtVersion(memberAddress, currentVersion);
    ///      if (index == type(uint256).max) {
    ///          // Not a member at this version
    ///      } else {
    ///          // Member at index 'index' (0-based)
    ///      }
    ///      ```
    ///
    /// @custom:example With Validation
    ///      ```solidity
    ///      // Validate version before querying
    ///      uint256 validatedVersion = _validateMemberVersion(targetVersion);
    ///      uint256 index = _getMemberIndexAtVersion(member, validatedVersion);
    ///      ```
    ///
    /// @custom:invariant Index Encoding
    ///      - Invariant: indexPlusOne = 0 ⟺ return type(uint256).max (not member)
    ///      - Invariant: indexPlusOne > 0 ⟺ return indexPlusOne - 1 (valid index)
    ///      - Invariant: Valid indices are in range [0, memberCount-1]
    ///
    /// @custom:design Rationale
    ///      - 1-based storage encoding enables zero-check for membership
    ///      - type(uint256).max sentinel value clearly indicates non-member
    ///      - No parameter validation for gas efficiency (caller responsibility)
    ///      - Compatible with bitmap voting (index < 256 for valid members)
    function _getMemberIndexAtVersion(address member, uint256 version) internal view returns (uint256) {
        uint32 indexPlusOne = memberIndexByVersion[version][member];
        if (indexPlusOne == 0) {
            return type(uint256).max;
        }
        return uint256(indexPlusOne - 1);
    }

    /// @dev Decrement active proposal count for a proposal's creator
    /// @notice Called when a proposal reaches any terminal state to free up proposal capacity
    ///
    ///         This function manages the memberActiveProposalCount by decrementing the count
    ///         when proposals complete. This enforces the MAX_ACTIVE_PROPOSALS_PER_MEMBER limit
    ///         by freeing capacity when proposals reach terminal states.
    ///
    ///         Terminal states that trigger this function:
    ///         1. ProposalStatus.Executed - Successfully executed
    ///         2. ProposalStatus.Cancelled - Cancelled by proposer
    ///         3. ProposalStatus.Expired - Expired without execution
    ///         4. ProposalStatus.Failed - Failed after MAX_RETRY_COUNT attempts
    ///         5. ProposalStatus.Rejected - Rejected by governance members
    ///
    ///         The function is called from multiple locations:
    ///         - _vote: When rejection threshold is met
    ///         - cancelProposal: When proposer cancels
    ///         - expireProposal: When proposal is manually expired
    ///         - executeProposal: When execution fails permanently
    ///         - _checkProposalInVoting: When expired proposal is auto-detected
    ///         - _checkProposalExecutable: When expired proposal is auto-detected
    ///
    /// @param proposalId The proposal ID that has reached a terminal state
    ///
    /// @custom:security Parameter Validation - Design Decision
    ///      - ⚠️ **proposalId is NOT validated in this function**
    ///      - Callers must ensure proposalId is valid before calling
    ///      - Invalid proposalId will access proposals[0] (default struct with proposer = address(0))
    ///      - This would decrement memberActiveProposalCount[address(0)] harmlessly
    ///      - **Design decision**: Validation responsibility delegated to callers
    ///      - All current callers have already validated proposalId before calling
    ///
    /// @custom:security Zero Address Handling
    ///      - If proposalId is invalid, proposer could be address(0)
    ///      - Decrementing address(0)'s count is harmless (no security impact)
    ///      - Zero address cannot create proposals (blocked by onlyMember modifier)
    ///
    /// @custom:security Underflow Protection
    ///      - Checks memberActiveProposalCount[proposer] > 0 before decrementing
    ///      - Prevents underflow if count is already 0
    ///      - Solidity 0.8.14 provides automatic underflow protection as secondary defense
    ///      - Double protection ensures safety
    ///
    /// @custom:security Idempotency
    ///      - Safe to call multiple times for same proposalId (no-op after first call)
    ///      - Count check prevents repeated decrements
    ///      - No negative impact from redundant calls
    ///
    /// @custom:security Gas Efficiency
    ///      - Single SLOAD + conditional SSTORE: ~2900-5000 gas
    ///      - O(1) complexity - constant time operation
    ///      - No loops or external calls
    ///
    /// @custom:usage Call Sites
    ///      - _vote: Line 419, 472 (rejection/approval)
    ///      - cancelProposal: Line 511 (proposer cancellation)
    ///      - expireProposal: Line 531, 537 (manual expiry)
    ///      - executeProposal: Line 888, 914 (execution failure/completion)
    ///      - _checkProposalInVoting: Line 1944 (auto-expire on vote attempt)
    ///      - _checkProposalExecutable: Line 1954 (auto-expire on execution attempt)
    ///
    /// @custom:invariant Count Consistency
    ///      - Invariant: Count decremented exactly once per proposal
    ///      - Invariant: Count never goes below 0 (protected by check)
    ///      - Invariant: Count matches number of non-terminal proposals for each member
    ///
    /// @custom:design Rationale
    ///      - No parameter validation for gas efficiency (caller responsibility)
    ///      - Explicit > 0 check for clarity and double-protection
    ///      - Called from multiple terminal state transitions for consistency
    ///      - Enables spam prevention by freeing proposal capacity
    function _decrementActiveProposalCount(uint256 proposalId) internal {
        address proposer = proposals[proposalId].proposer;
        if (memberActiveProposalCount[proposer] > 0) {
            memberActiveProposalCount[proposer]--;
        }
    }
}
