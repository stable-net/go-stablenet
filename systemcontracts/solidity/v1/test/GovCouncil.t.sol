// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright 2025 The go-stablenet Authors
pragma solidity ^0.8.14;

import "forge-std/Test.sol";
import "../GovCouncil.sol";

/**
 * @title GovCouncilTest
 * @notice Comprehensive test suite for GovCouncil contract
 * @dev Tests blacklist and authorized account management with governance proposals
 */
contract GovCouncilTest is Test {
    GovCouncil public govCouncil;

    // Test accounts
    address public member1;
    address public member2;
    address public member3;
    address public nonMember;
    address public testAccount1;
    address public testAccount2;
    address public testAccount3;

    // Events for testing
    event AddressBlacklisted(address indexed account, uint256 indexed proposalId);
    event AddressUnblacklisted(address indexed account, uint256 indexed proposalId);
    event AuthorizedAccountAdded(address indexed account, uint256 indexed proposalId);
    event AuthorizedAccountRemoved(address indexed account, uint256 indexed proposalId);
    event ProposalCreated(
        uint256 indexed proposalId,
        address indexed proposer,
        bytes32 actionType,
        uint256 memberVersion,
        uint256 requiredApprovals,
        bytes callData
    );
    event ProposalApproved(uint256 indexed proposalId, address indexed approver, uint256 approved, uint256 rejected);
    event ProposalExecuted(uint256 indexed proposalId, address indexed executor, bool success);

    function setUp() public {
        // Create test accounts
        member1 = makeAddr("member1");
        member2 = makeAddr("member2");
        member3 = makeAddr("member3");
        nonMember = makeAddr("nonMember");
        testAccount1 = makeAddr("testAccount1");
        testAccount2 = makeAddr("testAccount2");
        testAccount3 = makeAddr("testAccount3");

        // Deploy GovCouncil
        govCouncil = new GovCouncil();

        // Initialize with 3 members, quorum of 2
        address[] memory members = new address[](3);
        members[0] = member1;
        members[1] = member2;
        members[2] = member3;

        govCouncil.initialize(members, 2);
    }

    // ========================================
    // Initialization Tests
    // ========================================

    function test_Initialization() public view {
        // Check member count
        assertEq(govCouncil.getMemberCount(1), 3);

        // Check quorum
        assertEq(govCouncil.quorum(), 2);

        // Check members
        assertTrue(govCouncil.members(member1).isActive);
        assertTrue(govCouncil.members(member2).isActive);
        assertTrue(govCouncil.members(member3).isActive);
        assertFalse(govCouncil.members(nonMember).isActive);

        // Check empty lists
        assertEq(govCouncil.getBlacklistCount(), 0);
        assertEq(govCouncil.getAuthorizedAccountCount(), 0);
    }

    // ========================================
    // Blacklist: Single Address Tests
    // ========================================

    function test_ProposeAddBlacklist() public {
        vm.startPrank(member1);

        vm.expectEmit(true, true, true, true);
        emit ProposalCreated(1, member1, govCouncil.ACTION_ADD_BLACKLIST(), 1, 2, abi.encode(testAccount1));

        uint256 proposalId = govCouncil.proposeAddBlacklist(testAccount1);
        assertEq(proposalId, 1);

        // Check proposal details
        GovBase.Proposal memory proposal = govCouncil.getProposal(proposalId);
        assertEq(proposal.actionType, govCouncil.ACTION_ADD_BLACKLIST());
        assertEq(proposal.proposer, member1);
        assertEq(proposal.approved, 1); // Proposer auto-approves
        assertEq(uint8(proposal.status), uint8(GovBase.ProposalStatus.Voting));

        vm.stopPrank();
    }

    function test_AddBlacklist_ExecuteAfterApproval() public {
        // Member1 proposes
        vm.prank(member1);
        uint256 proposalId = govCouncil.proposeAddBlacklist(testAccount1);

        // Check not blacklisted yet
        assertFalse(govCouncil.isBlacklisted(testAccount1));

        // Member2 approves and auto-executes (quorum reached)
        vm.prank(member2);
        vm.expectEmit(true, true, true, true);
        emit AddressBlacklisted(testAccount1, proposalId);
        govCouncil.approveProposal(proposalId);

        // Check proposal executed
        GovBase.Proposal memory proposal = govCouncil.getProposal(proposalId);
        assertEq(uint8(proposal.status), uint8(GovBase.ProposalStatus.Executed));

        // Check blacklisted
        assertTrue(govCouncil.isBlacklisted(testAccount1));
        assertEq(govCouncil.getBlacklistCount(), 1);
        assertEq(govCouncil.getBlacklistedAddress(0), testAccount1);
    }

    function test_ProposeRemoveBlacklist() public {
        // First, add to blacklist
        vm.prank(member1);
        uint256 addProposalId = govCouncil.proposeAddBlacklist(testAccount1);
        vm.prank(member2);
        govCouncil.approveProposal(addProposalId);

        assertTrue(govCouncil.isBlacklisted(testAccount1));

        // Now propose removal
        vm.prank(member1);
        uint256 removeProposalId = govCouncil.proposeRemoveBlacklist(testAccount1);

        // Member2 approves
        vm.prank(member2);
        vm.expectEmit(true, true, true, true);
        emit AddressUnblacklisted(testAccount1, removeProposalId);
        govCouncil.approveProposal(removeProposalId);

        // Check removed
        assertFalse(govCouncil.isBlacklisted(testAccount1));
        assertEq(govCouncil.getBlacklistCount(), 0);
    }

    function testFail_ProposeAddBlacklist_ZeroAddress() public {
        vm.prank(member1);
        govCouncil.proposeAddBlacklist(address(0));
    }

    function testFail_ProposeAddBlacklist_AlreadyInList() public {
        // Add first time
        vm.prank(member1);
        uint256 proposalId = govCouncil.proposeAddBlacklist(testAccount1);
        vm.prank(member2);
        govCouncil.approveProposal(proposalId);

        // Try to add again (should fail)
        vm.prank(member1);
        govCouncil.proposeAddBlacklist(testAccount1);
    }

    function testFail_ProposeRemoveBlacklist_NotInList() public {
        vm.prank(member1);
        govCouncil.proposeRemoveBlacklist(testAccount1);
    }

    function testFail_ProposeAddBlacklist_NotMember() public {
        vm.prank(nonMember);
        govCouncil.proposeAddBlacklist(testAccount1);
    }

    // ========================================
    // Blacklist: Batch Operations Tests
    // ========================================

    function test_ProposeAddBlacklistBatch() public {
        address[] memory accounts = new address[](3);
        accounts[0] = testAccount1;
        accounts[1] = testAccount2;
        accounts[2] = testAccount3;

        vm.prank(member1);
        uint256 proposalId = govCouncil.proposeAddBlacklistBatch(accounts);

        // Approve and execute
        vm.prank(member2);
        govCouncil.approveProposal(proposalId);

        // Check all blacklisted
        assertTrue(govCouncil.isBlacklisted(testAccount1));
        assertTrue(govCouncil.isBlacklisted(testAccount2));
        assertTrue(govCouncil.isBlacklisted(testAccount3));
        assertEq(govCouncil.getBlacklistCount(), 3);
    }

    function test_ProposeRemoveBlacklistBatch() public {
        // First add all
        address[] memory accounts = new address[](3);
        accounts[0] = testAccount1;
        accounts[1] = testAccount2;
        accounts[2] = testAccount3;

        vm.prank(member1);
        uint256 addProposalId = govCouncil.proposeAddBlacklistBatch(accounts);
        vm.prank(member2);
        govCouncil.approveProposal(addProposalId);

        // Now remove all
        vm.prank(member1);
        uint256 removeProposalId = govCouncil.proposeRemoveBlacklistBatch(accounts);
        vm.prank(member2);
        govCouncil.approveProposal(removeProposalId);

        // Check all removed
        assertFalse(govCouncil.isBlacklisted(testAccount1));
        assertFalse(govCouncil.isBlacklisted(testAccount2));
        assertFalse(govCouncil.isBlacklisted(testAccount3));
        assertEq(govCouncil.getBlacklistCount(), 0);
    }

    function testFail_ProposeAddBlacklistBatch_ContainsZeroAddress() public {
        address[] memory accounts = new address[](3);
        accounts[0] = testAccount1;
        accounts[1] = address(0); // Zero address
        accounts[2] = testAccount3;

        vm.prank(member1);
        govCouncil.proposeAddBlacklistBatch(accounts);
    }

    function testFail_ProposeAddBlacklistBatch_ContainsDuplicate() public {
        // Add first address
        vm.prank(member1);
        uint256 proposalId = govCouncil.proposeAddBlacklist(testAccount1);
        vm.prank(member2);
        govCouncil.approveProposal(proposalId);

        // Try batch with duplicate
        address[] memory accounts = new address[](2);
        accounts[0] = testAccount1; // Already in list
        accounts[1] = testAccount2;

        vm.prank(member1);
        govCouncil.proposeAddBlacklistBatch(accounts);
    }

    // ========================================
    // Authorized Account: Single Address Tests
    // ========================================

    function test_ProposeAddAuthorizedAccount() public {
        vm.startPrank(member1);

        vm.expectEmit(true, true, true, true);
        emit ProposalCreated(1, member1, govCouncil.ACTION_ADD_AUTHORIZED_ACCOUNT(), 1, 2, abi.encode(testAccount1));

        uint256 proposalId = govCouncil.proposeAddAuthorizedAccount(testAccount1);
        assertEq(proposalId, 1);

        vm.stopPrank();
    }

    function test_AddAuthorizedAccount_ExecuteAfterApproval() public {
        // Member1 proposes
        vm.prank(member1);
        uint256 proposalId = govCouncil.proposeAddAuthorizedAccount(testAccount1);

        // Check not authorized yet
        assertFalse(govCouncil.isAuthorizedAccount(testAccount1));

        // Member2 approves and auto-executes
        vm.prank(member2);
        vm.expectEmit(true, true, true, true);
        emit AuthorizedAccountAdded(testAccount1, proposalId);
        govCouncil.approveProposal(proposalId);

        // Check authorized
        assertTrue(govCouncil.isAuthorizedAccount(testAccount1));
        assertEq(govCouncil.getAuthorizedAccountCount(), 1);
        assertEq(govCouncil.getAuthorizedAccountAddress(0), testAccount1);
    }

    function test_ProposeRemoveAuthorizedAccount() public {
        // First, add authorized account
        vm.prank(member1);
        uint256 addProposalId = govCouncil.proposeAddAuthorizedAccount(testAccount1);
        vm.prank(member2);
        govCouncil.approveProposal(addProposalId);

        assertTrue(govCouncil.isAuthorizedAccount(testAccount1));

        // Now propose removal
        vm.prank(member1);
        uint256 removeProposalId = govCouncil.proposeRemoveAuthorizedAccount(testAccount1);

        // Member2 approves
        vm.prank(member2);
        vm.expectEmit(true, true, true, true);
        emit AuthorizedAccountRemoved(testAccount1, removeProposalId);
        govCouncil.approveProposal(removeProposalId);

        // Check removed
        assertFalse(govCouncil.isAuthorizedAccount(testAccount1));
        assertEq(govCouncil.getAuthorizedAccountCount(), 0);
    }

    function testFail_ProposeAddAuthorizedAccount_ZeroAddress() public {
        vm.prank(member1);
        govCouncil.proposeAddAuthorizedAccount(address(0));
    }

    function testFail_ProposeAddAuthorizedAccount_AlreadyInList() public {
        // Add first time
        vm.prank(member1);
        uint256 proposalId = govCouncil.proposeAddAuthorizedAccount(testAccount1);
        vm.prank(member2);
        govCouncil.approveProposal(proposalId);

        // Try to add again
        vm.prank(member1);
        govCouncil.proposeAddAuthorizedAccount(testAccount1);
    }

    function testFail_ProposeRemoveAuthorizedAccount_NotInList() public {
        vm.prank(member1);
        govCouncil.proposeRemoveAuthorizedAccount(testAccount1);
    }

    // ========================================
    // Authorized Account: Batch Operations Tests
    // ========================================

    function test_ProposeAddAuthorizedAccountBatch() public {
        address[] memory accounts = new address[](3);
        accounts[0] = testAccount1;
        accounts[1] = testAccount2;
        accounts[2] = testAccount3;

        vm.prank(member1);
        uint256 proposalId = govCouncil.proposeAddAuthorizedAccountBatch(accounts);

        // Approve and execute
        vm.prank(member2);
        govCouncil.approveProposal(proposalId);

        // Check all authorized
        assertTrue(govCouncil.isAuthorizedAccount(testAccount1));
        assertTrue(govCouncil.isAuthorizedAccount(testAccount2));
        assertTrue(govCouncil.isAuthorizedAccount(testAccount3));
        assertEq(govCouncil.getAuthorizedAccountCount(), 3);
    }

    function test_ProposeRemoveAuthorizedAccountBatch() public {
        // First add all
        address[] memory accounts = new address[](3);
        accounts[0] = testAccount1;
        accounts[1] = testAccount2;
        accounts[2] = testAccount3;

        vm.prank(member1);
        uint256 addProposalId = govCouncil.proposeAddAuthorizedAccountBatch(accounts);
        vm.prank(member2);
        govCouncil.approveProposal(addProposalId);

        // Now remove all
        vm.prank(member1);
        uint256 removeProposalId = govCouncil.proposeRemoveAuthorizedAccountBatch(accounts);
        vm.prank(member2);
        govCouncil.approveProposal(removeProposalId);

        // Check all removed
        assertFalse(govCouncil.isAuthorizedAccount(testAccount1));
        assertFalse(govCouncil.isAuthorizedAccount(testAccount2));
        assertFalse(govCouncil.isAuthorizedAccount(testAccount3));
        assertEq(govCouncil.getAuthorizedAccountCount(), 0);
    }

    // ========================================
    // Query Functions Tests
    // ========================================

    function test_GetBlacklistRange() public {
        // Add multiple addresses
        address[] memory accounts = new address[](5);
        for (uint256 i = 0; i < 5; i++) {
            accounts[i] = address(uint160(1000 + i));
        }

        vm.prank(member1);
        uint256 proposalId = govCouncil.proposeAddBlacklistBatch(accounts);
        vm.prank(member2);
        govCouncil.approveProposal(proposalId);

        // Get range [1, 3] (indices 1, 2, 3)
        address[] memory range = govCouncil.getBlacklistRange(1, 3);
        assertEq(range.length, 3);
        assertEq(range[0], accounts[1]);
        assertEq(range[1], accounts[2]);
        assertEq(range[2], accounts[3]);
    }

    function test_GetAllBlacklisted() public {
        // Add multiple addresses
        address[] memory accounts = new address[](3);
        accounts[0] = testAccount1;
        accounts[1] = testAccount2;
        accounts[2] = testAccount3;

        vm.prank(member1);
        uint256 proposalId = govCouncil.proposeAddBlacklistBatch(accounts);
        vm.prank(member2);
        govCouncil.approveProposal(proposalId);

        // Get all
        address[] memory all = govCouncil.getAllBlacklisted();
        assertEq(all.length, 3);
        assertEq(all[0], testAccount1);
        assertEq(all[1], testAccount2);
        assertEq(all[2], testAccount3);
    }

    function test_GetAuthorizedAccountRange() public {
        // Add multiple addresses
        address[] memory accounts = new address[](5);
        for (uint256 i = 0; i < 5; i++) {
            accounts[i] = address(uint160(2000 + i));
        }

        vm.prank(member1);
        uint256 proposalId = govCouncil.proposeAddAuthorizedAccountBatch(accounts);
        vm.prank(member2);
        govCouncil.approveProposal(proposalId);

        // Get range [0, 2]
        address[] memory range = govCouncil.getAuthorizedAccountRange(0, 2);
        assertEq(range.length, 3);
        assertEq(range[0], accounts[0]);
        assertEq(range[1], accounts[1]);
        assertEq(range[2], accounts[2]);
    }

    function test_GetAllAuthorizedAccounts() public {
        // Add multiple addresses
        address[] memory accounts = new address[](3);
        accounts[0] = testAccount1;
        accounts[1] = testAccount2;
        accounts[2] = testAccount3;

        vm.prank(member1);
        uint256 proposalId = govCouncil.proposeAddAuthorizedAccountBatch(accounts);
        vm.prank(member2);
        govCouncil.approveProposal(proposalId);

        // Get all
        address[] memory all = govCouncil.getAllAuthorizedAccounts();
        assertEq(all.length, 3);
        assertEq(all[0], testAccount1);
        assertEq(all[1], testAccount2);
        assertEq(all[2], testAccount3);
    }

    // ========================================
    // Complex Scenarios
    // ========================================

    function test_BlacklistAndAuthorizedAccount_Independent() public {
        // Add to both lists
        vm.prank(member1);
        uint256 blacklistProposalId = govCouncil.proposeAddBlacklist(testAccount1);
        vm.prank(member2);
        govCouncil.approveProposal(blacklistProposalId);

        vm.prank(member1);
        uint256 authorizedProposalId = govCouncil.proposeAddAuthorizedAccount(testAccount1);
        vm.prank(member2);
        govCouncil.approveProposal(authorizedProposalId);

        // Check both lists independently
        assertTrue(govCouncil.isBlacklisted(testAccount1));
        assertTrue(govCouncil.isAuthorizedAccount(testAccount1));
        assertEq(govCouncil.getBlacklistCount(), 1);
        assertEq(govCouncil.getAuthorizedAccountCount(), 1);
    }

    function test_MultipleProposalsSequential() public {
        // Proposal 1: Add to blacklist
        vm.prank(member1);
        uint256 proposal1 = govCouncil.proposeAddBlacklist(testAccount1);
        vm.prank(member2);
        govCouncil.approveProposal(proposal1);

        // Proposal 2: Add to authorized accounts
        vm.prank(member1);
        uint256 proposal2 = govCouncil.proposeAddAuthorizedAccount(testAccount2);
        vm.prank(member2);
        govCouncil.approveProposal(proposal2);

        // Proposal 3: Remove from blacklist
        vm.prank(member1);
        uint256 proposal3 = govCouncil.proposeRemoveBlacklist(testAccount1);
        vm.prank(member2);
        govCouncil.approveProposal(proposal3);

        // Check final state
        assertFalse(govCouncil.isBlacklisted(testAccount1));
        assertTrue(govCouncil.isAuthorizedAccount(testAccount2));
        assertEq(govCouncil.getBlacklistCount(), 0);
        assertEq(govCouncil.getAuthorizedAccountCount(), 1);
    }

    function test_ProposalRejection() public {
        // Proposal to add to blacklist
        vm.prank(member1);
        uint256 proposalId = govCouncil.proposeAddBlacklist(testAccount1);

        // Member2 and member3 reject (2 rejections > max allowed)
        vm.prank(member2);
        govCouncil.disapproveProposal(proposalId);
        vm.prank(member3);
        govCouncil.disapproveProposal(proposalId);

        // Check proposal rejected
        GovBase.Proposal memory proposal = govCouncil.getProposal(proposalId);
        assertEq(uint8(proposal.status), uint8(GovBase.ProposalStatus.Rejected));

        // Check address not blacklisted
        assertFalse(govCouncil.isBlacklisted(testAccount1));
    }

    // ========================================
    // Stress Tests
    // ========================================

    function test_AddRemoveMultipleTimes() public {
        for (uint256 i = 0; i < 3; i++) {
            // Add
            vm.prank(member1);
            uint256 addProposalId = govCouncil.proposeAddBlacklist(testAccount1);
            vm.prank(member2);
            govCouncil.approveProposal(addProposalId);
            assertTrue(govCouncil.isBlacklisted(testAccount1));

            // Remove
            vm.prank(member1);
            uint256 removeProposalId = govCouncil.proposeRemoveBlacklist(testAccount1);
            vm.prank(member2);
            govCouncil.approveProposal(removeProposalId);
            assertFalse(govCouncil.isBlacklisted(testAccount1));
        }
    }

    function test_LargeBlacklistAndAuthorizedAccountLists() public {
        // Add 10 addresses to each list
        address[] memory blacklistAccounts = new address[](10);
        address[] memory authorizedAccounts = new address[](10);

        for (uint256 i = 0; i < 10; i++) {
            blacklistAccounts[i] = address(uint160(3000 + i));
            authorizedAccounts[i] = address(uint160(4000 + i));
        }

        // Add to blacklist
        vm.prank(member1);
        uint256 blacklistProposalId = govCouncil.proposeAddBlacklistBatch(blacklistAccounts);
        vm.prank(member2);
        govCouncil.approveProposal(blacklistProposalId);

        // Add to authorized accounts
        vm.prank(member1);
        uint256 authorizedProposalId = govCouncil.proposeAddAuthorizedAccountBatch(authorizedAccounts);
        vm.prank(member2);
        govCouncil.approveProposal(authorizedProposalId);

        // Verify counts
        assertEq(govCouncil.getBlacklistCount(), 10);
        assertEq(govCouncil.getAuthorizedAccountCount(), 10);

        // Verify all addresses
        for (uint256 i = 0; i < 10; i++) {
            assertTrue(govCouncil.isBlacklisted(blacklistAccounts[i]));
            assertTrue(govCouncil.isAuthorizedAccount(authorizedAccounts[i]));
        }
    }
}
