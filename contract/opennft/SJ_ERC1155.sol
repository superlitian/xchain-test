// SPDX-License-Identifier: MIT
// Shengjian Contracts v1.0.0 (token/ERC1155/SJ_ERC1155.sol)
pragma solidity ^0.8.0;

// 引入 openzeppelin-contracts
import "@openzeppelin/contracts/token/ERC1155/ERC1155.sol";
import "./StringUtils.sol";

contract TestContract is ERC1155 {

    // store the toke data
    mapping(uint256 => bytes) _stores;

    // token transaction datetime
    mapping(address => mapping(uint256 => uint256)) _tokenDateTime;

    // single token authorization mapping key is owner , value key is be approved address
    mapping(address => mapping(uint256 => address)) _singleTokenApproved;

    mapping(address => uint256[]) _hasApprovedTokenIds;

    // token transfer protect time default 7 days
    uint256 _tokenExpireTime = 1;

    // localdatime （because xuperchain unsupport block.timestamp. so regular manual update）
    uint256 _timeStamp;

    // store the nft token transation protect time, mint token set
    mapping(uint256 =>  uint256) _tokenPrtectTime;

    address _manger;

    constructor(string memory uri_, uint256 _time, address admin) ERC1155(uri_) {
        _timeStamp = _time;
        _manger = admin;
    }

    /**
        constraint condition
     */
    modifier ownerOnly {
        require (msg.sender == _manger, "ERC1155: not contract owner,can't operate");
        _;
    }

    /**
       Convert bytes to address    
     */
    function bytesToAddress(bytes memory bys) public pure returns (address addr) {
        assembly {
            addr := mload(add(bys,20))
        }
    }

    // Convert an hexadecimal character to their value
    function fromHexChar(uint8 c) public pure returns (uint8) {
        if (bytes1(c) >= bytes1('0') && bytes1(c) <= bytes1('9')) {
            return c - uint8(bytes1('0'));
        }
        if (bytes1(c) >= bytes1('a') && bytes1(c) <= bytes1('f')) {
            return 10 + c - uint8(bytes1('a'));
        }
        if (bytes1(c) >= bytes1('A') && bytes1(c) <= bytes1('F')) {
            return 10 + c - uint8(bytes1('A'));
        }
    }

    // Convert an hexadecimal string to raw bytes
    function fromHex(string memory s) public pure returns (bytes memory) {
        bytes memory ss = bytes(s);
        require(ss.length%2 == 0); // length must be even
        bytes memory r = new bytes(ss.length/2);
        for (uint i=0; i<ss.length/2; ++i) {
            r[i] = bytes1(fromHexChar(uint8(ss[2*i])) * 16 +
                fromHexChar(uint8(ss[2*i+1])));
        }
        return r;
    }

    /**
     * @dev See {IERC1155-safeTransferFrom}.
     */
    function safeTransferFrom (
        address from,
        address to,
        uint256 id,
        uint256 amount,
        bytes memory data
    ) public virtual override {
        require( _tokenDateTime[from][id] == 0 || _tokenDateTime[from][id] + _tokenPrtectTime[id] <= _timeStamp,
            "ERC1155: token protect time is unexpired"
        );
        require(
            from == msg.sender || msg.sender == _singleTokenApproved[from][id],
            "ERC1155: caller is not owner nor approved"
        );
        _safeTransferFrom(from, to, id, amount, data);
        _tokenDateTime[to][id] = _timeStamp;
    }

    /**
        @dev single token approve other address operator. only token owner can approve;
     */
    function approveForOne(address from, uint256 id, address to) public {
        require(from != address(0) && to != address(0), "ERC1155: from and to address are zero address");
        require(from == msg.sender, "ERC1155: caller is not owner");
        require(_singleTokenApproved[from][id] == address(0),"ERC1155: this address token id is approved");
        _singleTokenApproved[from][id] = to;
        _hasApprovedTokenIds[to].push(id);
    }

    function getApproveOne() public view returns (uint256[] memory) {
        return _hasApprovedTokenIds[msg.sender];
    }

    /**
        @dev override balanceOfBatch because not support string[], clien input address should connect with ","
     */
    function balanceOfBatch(string memory accounts, uint256[] memory ids)
    public
    view
    returns (uint256[] memory){
        Strings.Slice memory str = Strings.toSlice(accounts);
        Strings.Slice memory delim = Strings.toSlice(",");
        string[] memory parts = new string[](Strings.count(str, delim) + 1);
        for (uint256 i = 0; i < parts.length; i++) {
            parts[i] = Strings.toString(Strings.split(str, delim));
        }
        require(parts.length == ids.length, "ERC1155: accounts and ids length mismatch");
        uint256[] memory batchBalances = new uint256[](parts.length);

        for (uint256 i = 0; i < parts.length; ++i) {
            batchBalances[i] = balanceOf(bytesToAddress(fromHex(parts[i])), ids[i]);
        }
        return batchBalances;
    }

    /**
        @dev set token expire base timestamp, only contract send can set
      
     */
    function setTimeStamp(uint256 _time) public ownerOnly  {
        _timeStamp = _time;
    }

    /**
        @dev get token expire base timestamp
     */
    function getTimeStamp() public view returns (uint256) {
        return _timeStamp;
    }

    /**
        @dev get token expire time, return sender's {id} token expire time
     */
    function getTokenExpireTime(uint256 _id) public view returns (uint256) {
        return _tokenDateTime[msg.sender][_id];
    }
    /**
        @dev seturi, only contract sender can set
     */
    function setUri(string memory _uri) public ownerOnly {
        _setURI(_uri);
    }

    /**
        @dev update Expire Time, only contract sender can set,this time influence token transfer 
     */
    function updateExpireTime(uint256 _id, uint256 _expireTime) public ownerOnly {
        require(_expireTime >= _tokenExpireTime, "ERC1155: token protect time is unless 7 days");
        _tokenPrtectTime[_id] = _expireTime;
    }

    /**
        @dev get Expire Time, only contract sender can set,this time influence token transfer 
     */
    function getExpireTime(uint256 _id) public view returns (uint256) {
        return _tokenPrtectTime[_id];
    }

    /**
        @dev mint token , require token id is not exist, require token data is not empty 
     */
    function addNewToken(uint256 _id, uint256 _initialSupply, bytes memory _data, uint256 tokenTime) public {
        require(tokenTime >= _tokenExpireTime, "ERC1155: token protect time is unless 7 days");
        require(keccak256(_data) != keccak256(""), "ERC1155: token data can't empty");
        require(keccak256(_stores[_id]) == keccak256(""), "ERC1155: token id is exist");
        _mint(msg.sender, _id, _initialSupply, "");
        _stores[_id] = _data;
        _tokenDateTime[msg.sender][_id] = 0;
        _tokenPrtectTime[_id] = tokenTime;
    }

    /**
        @dev get token data
     */
    function getTokenBytes(uint256 _id) public view returns (bytes memory _response) {
        return _stores[_id];
    }

    /**
        @dev mint token bathch
     */
    function addNewTokenBatch( uint256[] memory _ids, uint256[] memory _amounts, string memory _data) public {
        Strings.Slice memory str = Strings.toSlice(_data);
        Strings.Slice memory delim = Strings.toSlice(",,,");
        string[] memory parts = new string[](Strings.count(str, delim) + 1);
        for (uint256 i = 0; i < _ids.length; i++) {
            parts[i] = Strings.toString(Strings.split(str, delim));
            require(keccak256(bytes(parts[i])) != keccak256(""), "ERC1155: Every token data can't empty");
            require(keccak256(_stores[_ids[i]]) == keccak256(""), "ERC1155: This token id is exist");
        }
        _mintBatch(msg.sender, _ids, _amounts, "");
        for (uint256 i = 0; i < _ids.length; i++) {
            _stores[_ids[i]] = bytes(parts[i]);
            _tokenDateTime[msg.sender][_ids[i]] = 0;
        }
        // require (false, "ERC1155: mint token batch is not support");
    }

    /**
        @dev burn Token
     */
    function burnToken(uint256 _id, uint256  _initialSupply) public {
        _burn(msg.sender, _id, _initialSupply);
    }
    /**
        @dev burn Token Batch
     */
    function burnTokenBatch(uint256[] memory _ids, uint256[] memory _initialSupply) public {
        _burnBatch(msg.sender, _ids, _initialSupply);
    }
}
