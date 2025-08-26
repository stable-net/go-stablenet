// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright 2025 The go-wemix-wbft Authors
// This file is part of the go-wemix-wbft library.
//
// The go-wemix-wbft library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-wemix-wbft library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-wemix-wbft library. If not, see <http://www.gnu.org/licenses/>.

pragma solidity ^0.8.14;

import "@openzeppelin/contracts/utils/structs/EnumerableSet.sol";
import { IGovCouncil } from "./IGovCouncil.sol";

contract GovNCP is IGovCouncil {
    using EnumerableSet for EnumerableSet.AddressSet;

    enum Decision {
        None,
        Accept,
        Reject
    }

    enum ProposalState {
        None,
        Voting,
        Accepted,
        Rejected,
        Canceled
    }

    enum ProposalType {
        None,
        NCPAdd,
        NCPRemoval,
        EmergencyMode,
        ReleaseEmergencyMode
    }

    struct Proposal {
        ProposalState state;
        uint256 startTime;
        uint256 endTime;
        uint256 proposer;
        address newNCP;
        uint256 removalNCP;
        ProposalType proposalType;
        uint256[] voters;
        uint256 accepts;
        uint256 rejects;
    }

    uint256 public constant VOTING_PERIOD = 1 weeks;

    EnumerableSet.AddressSet private __ncpList; // 0x0, 0x1; staker operator addresses

    uint256 public lastNCPID; // 0x2
    mapping(uint256 => address) public ncpIDToAddress; // 0x3; mapping from NCP ID to NCP address
    mapping(address => uint256) public addressToNCPID; // 0x4; mapping from NCP address to NCP ID

    uint256 public currentProposalID;
    mapping(uint256 => Proposal) public proposals;
    mapping(uint256 => mapping(uint256 => Decision)) public decisions;

    bool public emergencyMode;

    //***********************************************************************
    //* Caution for Upgrading
    //* - If you add new state variables, please add them after this comment
    //* - Never modify existing state variables
    //***********************************************************************

    event NewProposal(uint256 indexed id, uint256 proposalType, address ncp, address proposer, uint256 time, uint256 endtime);

    event Vote(uint256 indexed proposalID, address voter, bool accept);
    event ProposalFinalized(uint256 indexed proposalID, bool accepted);
    event ProposalCanceled(uint256 indexed proposalID);

    event NCPAdded(address indexed ncp);
    event NCPRemoved(address indexed ncp);
    event NCPChanged(address indexed oldNCP, address indexed newNCP);

    modifier onlyNCP() {
        require(__ncpList.contains(msg.sender), "msg.sender is not ncp");
        _;
    }

    function inspectOperation(bytes4 selector, address _sender, bytes memory arguments) external view override returns (bool) {
        // in NCP governance, we do not restrict any operations if not in emergency mode
        return !emergencyMode;
    }

    function isNCP(address _ncp) external view returns (bool) {
        return __ncpList.contains(_ncp);
    }

    function ncpList() external view returns (address[] memory) {
        return __ncpList.values();
    }

    function ncpCount() external view returns (uint256) {
        return __ncpList.length();
    }

    function newProposalToAddNCP(address _newNCP) external onlyNCP {
        require(!__ncpList.contains(_newNCP), "ncp exists");
        _newProposal(_newNCP, ProposalType.NCPAdd);
    }

    function newProposalToRemoveNCP(address _ncp) external onlyNCP {
        require(__ncpList.contains(_ncp), "invalid ncp");
        _newProposal(_ncp, ProposalType.NCPRemoval);

        if (msg.sender == _ncp) {
            Proposal storage _proposal = _getVotingProposal(currentProposalID);
            _finalizeProposal(_proposal, true);
        }
    }

    function newProposalEmergencyMode(bool toMode) external onlyNCP {
        require(emergencyMode != toMode, "already in the mode");

        if (toMode) {
            _newProposal(address(0), ProposalType.EmergencyMode);
        } else {
            _newProposal(address(0), ProposalType.ReleaseEmergencyMode);
        }
    }

    function changeNCP(address _ncp) external onlyNCP {
        require(!__ncpList.contains(_ncp), "ncp already exists");

        if (
            (proposals[currentProposalID].state != ProposalState.None) &&
            (proposals[currentProposalID].proposalType == ProposalType.NCPAdd) &&
            (proposals[currentProposalID].endTime >= block.timestamp)
        ) {
            require(proposals[currentProposalID].newNCP != _ncp, "cannot change the ncp to an address that is proposed as the new ncp");
        }
        __ncpList.remove(msg.sender);
        __ncpList.add(_ncp);
        addressToNCPID[_ncp] = addressToNCPID[msg.sender];
        ncpIDToAddress[addressToNCPID[msg.sender]] = _ncp;
        addressToNCPID[msg.sender] = 0; // remove old ncp id

        emit NCPChanged(msg.sender, _ncp);
    }

    function vote(uint256 _proposalID, bool _accept) external onlyNCP {
        Proposal storage _proposal = _getVotingProposal(_proposalID);
        require(block.timestamp <= _proposal.endTime, "already closed vote");
        require(decisions[_proposalID][addressToNCPID[msg.sender]] == Decision.None, "already voted");

        Decision _decision;
        if (_accept) {
            _decision = Decision.Accept;
            _proposal.accepts++;
        } else {
            _decision = Decision.Reject;
            _proposal.rejects++;
        }
        _proposal.voters.push(addressToNCPID[msg.sender]);
        decisions[_proposalID][addressToNCPID[msg.sender]] = _decision;

        emit Vote(_proposalID, msg.sender, _accept);
        uint256 _threshold = __ncpList.length();
        if (_proposal.accepts * 2 > _threshold || _proposal.rejects * 2 >= _threshold) {
            _finalizeProposal(_proposal, _proposal.accepts > _proposal.rejects);
        }
    }

    function cancelProposal(uint256 _proposalID) external onlyNCP {
        Proposal storage _proposal = _getVotingProposal(_proposalID);
        if (block.timestamp <= _proposal.endTime) {
            require(_proposal.proposer == addressToNCPID[msg.sender], "non-proposer cannot cancel before timeout");
            require(_proposal.voters.length == 0, "cannot cancel after vote");
        }
        _cancelProposal(_proposal);
    }

    function _newProposal(address _targetNCP, ProposalType _proposalType) private {
        Proposal storage _proposal = proposals[currentProposalID];
        if (_proposal.state != ProposalState.None) {
            if (_proposal.endTime >= block.timestamp) {
                revert("previous vote is in progress");
            } else {
                _cancelProposal(_proposal);
                _proposal = proposals[currentProposalID];
            }
        }
        if (_proposalType == ProposalType.NCPAdd) {
            _proposal.newNCP = _targetNCP;
        } else if (_proposalType == ProposalType.NCPRemoval) {
            _proposal.removalNCP = addressToNCPID[_targetNCP];
        }
        _proposal.proposer = addressToNCPID[msg.sender];
        _proposal.startTime = block.timestamp;
        _proposal.endTime = block.timestamp + VOTING_PERIOD;
        _proposal.proposalType = _proposalType;
        _proposal.state = ProposalState.Voting;

        emit NewProposal(currentProposalID, uint(_proposalType), _targetNCP, msg.sender, block.timestamp, _proposal.endTime);
    }

    function _getVotingProposal(uint256 _proposalID) private view returns (Proposal storage _proposal) {
        require(_proposalID == currentProposalID, "invalid proposal id");
        _proposal = proposals[_proposalID];
        require(_proposal.state == ProposalState.Voting, "not in voting");
    }

    function _finalizeProposal(Proposal storage _proposal, bool _accepted) private {
        if (_accepted) {
            if (_proposal.proposalType == ProposalType.NCPAdd) {
                __ncpList.add(_proposal.newNCP);
                lastNCPID++;
                addressToNCPID[_proposal.newNCP] = lastNCPID;
                ncpIDToAddress[lastNCPID] = _proposal.newNCP;
                emit NCPAdded(_proposal.newNCP);
            } else if (_proposal.proposalType == ProposalType.NCPRemoval) {
                address _oldNCP = ncpIDToAddress[_proposal.removalNCP];
                __ncpList.remove(_oldNCP);
                ncpIDToAddress[_proposal.removalNCP] = address(0);
                addressToNCPID[_oldNCP] = 0;
                emit NCPRemoved(_oldNCP);
            } else if (_proposal.proposalType == ProposalType.EmergencyMode) {
                emergencyMode = true;
            } else if (_proposal.proposalType == ProposalType.ReleaseEmergencyMode) {
                emergencyMode = false;
            }
            _proposal.state = ProposalState.Accepted;
        } else {
            _proposal.state = ProposalState.Rejected;
        }

        emit ProposalFinalized(currentProposalID, _accepted);

        currentProposalID++;
    }

    function _cancelProposal(Proposal storage _proposal) private {
        _proposal.state = ProposalState.Canceled;
        emit ProposalCanceled(currentProposalID);

        currentProposalID++;
    }
}
