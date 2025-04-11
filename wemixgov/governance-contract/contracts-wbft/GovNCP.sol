// SPDX-License-Identifier: MIT

pragma solidity 0.8.14;

import "@openzeppelin/contracts/utils/structs/EnumerableSet.sol";

contract GovNCP {
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
        NCPRemoval
    }

    struct Proposal {
        ProposalState state;
        uint256 startTime;
        uint256 endTime;
        address proposer;
        address targetNCP;
        ProposalType proposalType;
        address[] voters;
        uint256 accepts;
        uint256 rejects;
        mapping(address => Decision) decisions;
    }

    uint256 public constant VOTING_PERIOD = 1 weeks;

    EnumerableSet.AddressSet private __ncpList;

    uint256 public currentProposalID;
    mapping(uint256 => Proposal) private __proposals;
    mapping(address => bool) private __lockedNCPs;

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

    function isNCP(address _ncp) external view returns (bool) {
        return __ncpList.contains(_ncp);
    }

    function ncpList() external view returns (address[] memory) {
        return __ncpList.values();
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

    function changeNCP(address _ncp) external onlyNCP {
        require(!__ncpList.contains(_ncp), "ncp already exists");
        require(!__lockedNCPs[msg.sender], "belong in an on-going proposal");

        __ncpList.remove(msg.sender);
        __ncpList.add(_ncp);

        emit NCPChanged(msg.sender, _ncp);
    }

    function vote(uint256 _proposalID, bool _accept) external onlyNCP {
        Proposal storage _proposal = _getVotingProposal(_proposalID);
        require(block.timestamp <= _proposal.endTime, "already closed vote");
        require(_proposal.decisions[msg.sender] == Decision.None, "already voted");

        Decision _decision;
        if (_accept) {
            _decision = Decision.Accept;
            _proposal.accepts++;
        } else {
            _decision = Decision.Reject;
            _proposal.rejects++;
        }
        _proposal.voters.push(msg.sender);
        _proposal.decisions[msg.sender] = _decision;

        emit Vote(_proposalID, msg.sender, _accept);
        uint256 _threshold = __ncpList.length();
        if (_proposal.accepts * 2 > _threshold || _proposal.rejects * 2 >= _threshold) {
            _finalizeProposal(_proposal, _proposal.accepts > _proposal.rejects);
        }
    }

    function cancelProposal(uint256 _proposalID) external onlyNCP {
        Proposal storage _proposal = _getVotingProposal(_proposalID);
        if (block.timestamp <= _proposal.endTime) {
            require(_proposal.proposer == msg.sender, "non-proposer cannot cancel before timeout");
            require(_proposal.voters.length == 0, "cannot cancel after vote");
        }
        _cancelProposal(_proposal);
    }

    function _newProposal(address _targetNCP, ProposalType _proposalType) private {
        Proposal storage _proposal = __proposals[currentProposalID];
        if (_proposal.state != ProposalState.None) {
            if (_proposal.endTime >= block.timestamp) {
                revert("previous vote is in progress");
            } else {
                _cancelProposal(_proposal);
                _proposal = __proposals[currentProposalID];
            }
        }
        _proposal.proposer = msg.sender;
        _proposal.startTime = block.timestamp;
        _proposal.endTime = block.timestamp + VOTING_PERIOD;
        _proposal.targetNCP = _targetNCP;
        _proposal.proposalType = _proposalType;
        _proposal.state = ProposalState.Voting;

        __lockedNCPs[msg.sender] = true;
        __lockedNCPs[_targetNCP] = true;

        emit NewProposal(currentProposalID, uint(_proposalType), _targetNCP, msg.sender, block.timestamp, _proposal.endTime);
    }

    function _getVotingProposal(uint256 _proposalID) private view returns (Proposal storage _proposal) {
        require(_proposalID == currentProposalID, "invalid proposal id");
        _proposal = __proposals[_proposalID];
        require(_proposal.state == ProposalState.Voting, "not in voting");
    }

    function _finalizeProposal(Proposal storage _proposal, bool _accepted) private {
        if (_accepted) {
            if (_proposal.proposalType == ProposalType.NCPAdd) {
                __ncpList.add(_proposal.targetNCP);
                emit NCPAdded(_proposal.targetNCP);
            } else {
                __ncpList.remove(_proposal.targetNCP);
                emit NCPRemoved(_proposal.targetNCP);
            }
            _proposal.state = ProposalState.Accepted;
        } else {
            _proposal.state = ProposalState.Rejected;
        }

        __lockedNCPs[_proposal.proposer] = false;
        __lockedNCPs[_proposal.targetNCP] = false;

        emit ProposalFinalized(currentProposalID, _accepted);

        currentProposalID++;
    }

    function _cancelProposal(Proposal storage _proposal) private {
        _proposal.state = ProposalState.Canceled;
        emit ProposalCanceled(currentProposalID);

        currentProposalID++;
    }
}
