pragma solidity ^0.4.23;

contract VMap {
    function createFileInit(uint fileType, bytes32 hash, uint32 timestamp, address owner, uint expiry, bytes32 replacesFile, address[] definedSigners) public returns (bool);
    function createFileEnd(bytes32 hash, uint32 timestamp, address[] prvs, uint[] prices) public returns (bool);
    function fileAddSinger(bytes32 hash, uint32 timestamp, address singer) public returns (bool) ;
    //function fileHasSinger(bytes32 hash, address singer) public returns (bool);
    //function fileRemoveSinger(bytes32 hash, address singer) public returns (bool);
    function fileSign(bytes32 hash, uint32 timestamp, address signer) public returns (bool);
    function fileAddSP(bytes32 hash, uint32 timestamp, address strPrv, uint price) public ;
    //function fileHasSP(bytes32 hash, address addr) public view returns (bool) ;
    function fileRemoveSP(bytes32 hash, uint32 timestamp, address addr) public view returns (bool);
    //function fileSetFilesize(bytes32 hash, uint filesize) public ;
    //function fileGetFileInfo(bytes32 hash) public ;
    function fileRemove(bytes32 hash, uint32 timestamp) public ;
    //function isFileRemoved(bytes32 hash) public view returns (bool) ;
    function fileNewOwner(bytes32 hash, uint32 timestamp, address newOwner) public ;
    function fileSetPerm(bytes32 hash, uint32 timestamp, address addr) public ;
    function fileGetPerm(bytes32 hash, uint32 timestamp, address addr, bool write) public view returns (bool);
    function fileRevokePerm(bytes32 hash, uint32 timestamp, address addr) public returns (bool);
    //function fileExpiry(bytes32 hash) public view returns (uint) ;
    //function fileVerify(bytes32 hash) public view returns(bool, address[]);
    function fileGetOwner(bytes32 hash, uint32 timestamp) public view returns(address);
    function fileList(address owner) public view returns (bytes32[]);
    function fileHasWriteAccess(bytes32 hash, uint32 timestamp, address addr) public view returns(bool);
    function fileInfo(bytes32 hash, uint32 timestamp) public view returns (address ownr, uint fileType, bool removed, uint expiry);

    function fileProviders(bytes32 hash, uint32 timestamp) public view returns (address[], uint[]);
    function fileDefinedSingers(bytes32 hash, uint32 timestamp) public view returns (address[]);
    function fileSigners(bytes32 hash, uint32 timestamp) public view returns (address[]);
    function fileGetPerms(bytes32 hash, uint32 timestamp, bool write) public view returns (address[]);

    function spAdd(address strProv, bytes32 urlPrefix, uint[] price) public returns (bool);
    function spUpdate(address strPrv, bytes32 urlPrefix) public;
    function spInfo(address strPrv) public view returns (bytes32, uint[]);
    function spList() public view returns(address[]);
    function spHas(address prv) public view returns (bool);
}

contract ERC20IF {
    function transfer(address to, uint value) public returns (bool ok);
    function transferFrom(address from, address to, uint value) public returns (bool ok);
    function spUpdate(address strPrv, bytes32 urlPrefix) public;
    function allowance(address tokenOwner, address spender) public constant returns (uint remaining);
}
contract ProxeusFS {
    uint internal proxeusPrice  = 0.1 ether;
    address XES_ADDRESS;// = 0x84E0b37e8f5B4B86d5d299b0B0e33686405A3919;//ropsten
    uint public dappVersion = 0;

    address internal owner;
    address internal issuer;

    //address public sender_remix;
    //address public spender_remix;
    //ProviderListMgr internal whitelistedSP;

    VMap internal storagemgr;
    constructor(address ownr, address tokenAddr) public {
        owner = ownr;
        issuer = msg.sender;
        XES_ADDRESS = tokenAddr;
    }

    event Deleted(bytes32 indexed hash);
    event UpdatedFileInfo(bytes32 hash, uint32 timestamp);
    event RequestSign(bytes32 hash, uint32 timestamp, address indexed to);
    event NotifySign(bytes32 indexed hash, uint32 timestamp, address indexed who);
    event OwnerChanged(bytes32 indexed hash, uint32 timestamp, address oldOwner, address newOwner);
    event RequestAccess(bytes32 hash, uint32 timestamp, address who);
    event UpdatedPerm(bytes32 hash, uint32 timestamp);
    event UpdatedRevoke(bytes32 hash, uint32 timestamp);

    function setDappVersion(uint version) public {
        require(msg.sender == owner);
        dappVersion = version;
    }
    function XESAmountPerFile(address prvs, uint filesize, uint price) public view returns (uint) {
        bytes32 url;
        uint[] memory prices ;
        (url,prices) = storagemgr.spInfo(prvs);
        return safeAdd(proxeusPrice, safeMul(prices[price], filesize));
    }
    function XESAllowence(address sendr) public view returns (uint sum) {
        if(sendr != address(0)){
            sum = ERC20IF(XES_ADDRESS).allowance(sendr, this);
        }else if(msg.sender != address(0)){
            sum = ERC20IF(XES_ADDRESS).allowance(msg.sender, this);
        }
    }
    function spAdd(address strProv, bytes32 urlPrefix, uint[] prices) public {
        require(msg.sender == owner, "denied!");    // needed to be confirmed
        storagemgr.spAdd(strProv, urlPrefix, prices);
    }
    function spUpdate(address strPrv, bytes32 urlPrefix) public {
        require(msg.sender == owner, "denied!");
        storagemgr.spUpdate(strPrv, urlPrefix);
    }
    function spList() public view returns(address[]) {
        return storagemgr.spList();
    }
    function createFileUndefinedSignersInit(bytes32 hash, uint32 timestamp, uint mandatorySigners, uint expiry, bytes32 replacesFile) public {
        _createFileInit(2, hash, timestamp, new address[](mandatorySigners), expiry, replacesFile);
    }
    function createFileDefinedSignersInit(bytes32 hash, uint32 timestamp, address[] definedSigners, uint expiry, bytes32 replacesFile) public {
        _createFileInit(3, hash, timestamp, definedSigners, expiry, replacesFile);
    }
    function _createFileInit(uint fileType, bytes32 hash, uint32 timestamp, address[] definedSigners, uint expiry, bytes32 replacesFile) internal {
        if (!storagemgr.createFileInit(fileType, hash, timestamp, msg.sender, expiry, replacesFile, definedSigners)) revert();
        for (uint j=0;j<definedSigners.length;j++)
            emit RequestSign(hash, timestamp, definedSigners[j]);
    }
    function createFilePayment(bytes32 hash, uint32 timestamp, uint32 filesize, address[] prvs, uint[] prices) public {
        if (!storagemgr.createFileEnd(hash, timestamp,  prvs, prices)) revert();
        uint sum = 0;
        bytes32 url;
        uint[] memory storageProvPrices ;
        for (uint i = 0; i < prvs.length; i++){
            (url, storageProvPrices)  =  storagemgr.spInfo(prvs[i]);
            uint storageProvPrice = safeMul(filesize, storageProvPrices[prices[i]]);
            sum += safeAdd(proxeusPrice, storageProvPrice);
            require(ERC20IF(XES_ADDRESS).transfer(prvs[i], storageProvPrice));
        }

        require(ERC20IF(XES_ADDRESS).allowance(msg.sender, this) >= sum);
        require(ERC20IF(XES_ADDRESS).transferFrom(msg.sender, this, sum));
        require(ERC20IF(XES_ADDRESS).transfer(owner, proxeusPrice));
    }
    function createFileEnd(bytes32 hash, uint32 timestamp) public {
        emit UpdatedFileInfo(hash, timestamp);
    }
    function fileInfo(bytes32 hash, uint32 timestamp) public view returns (address ownr, uint fileType, bool removed, uint expiry) {
        return storagemgr.fileInfo(hash, timestamp);
    }
    function fileInfo2(bytes32 hash, uint32 timestamp) public view returns (address[] strPrv, uint[] prices, address[] readAccess, address[] writeAccess, address[] definedSigners, address[] signers) {
        (strPrv, prices) = storagemgr.fileProviders(hash, timestamp);
        readAccess = storagemgr.fileGetPerms(hash, timestamp, false);
        writeAccess = storagemgr.fileGetPerms(hash, timestamp, true);
        definedSigners = storagemgr.fileDefinedSingers(hash, timestamp);
        signers = storagemgr.fileSigners(hash, timestamp);
        return (strPrv, prices, readAccess, writeAccess, definedSigners, signers);
    }
    function fileAdmsdSinger(bytes32 hash, uint32 timestamp, address singer) public returns (bool) {
        bool ret = storagemgr.fileAddSinger(hash, timestamp, singer);
        emit RequestSign(hash, timestamp, singer);
        return ret;
    }
    function fileRequestSign(bytes32 hash, uint32 timestamp, address signer) public {
        require(msg.sender == storagemgr.fileGetOwner(hash, timestamp), "denied!");
        emit RequestSign(hash, timestamp, signer);
    }
    function fileSign(bytes32 hash, uint32 timestamp) public {
        bool bret = storagemgr.fileSign(hash, timestamp, msg.sender);
        if (!bret) revert();
        emit NotifySign(hash, timestamp, msg.sender);
    }
    function fileAddSP(bytes32 hash, uint32 timestamp, address strPrv, uint price) public {
        require(msg.sender == storagemgr.fileGetOwner(hash, timestamp));
        require(!storagemgr.spHas(strPrv));
        storagemgr.fileAddSP(hash, timestamp, strPrv, price);
    }
    //function fileList() public view returns (bytes32[]) {
    //     return storagemgr.fileList(msg.sender);
    //}
    function fileRemoveSP(bytes32 hash, uint32 timestamp, address addr) public view returns (bool) {
        require(msg.sender == storagemgr.fileGetOwner(hash, timestamp), "denied!");
        return storagemgr.fileRemoveSP(hash, timestamp, addr);
    }
    function fileRemove(bytes32 hash, uint32 timestamp) public {
        require(msg.sender == storagemgr.fileGetOwner(hash, timestamp), "denied!");
        storagemgr.fileRemove(hash, timestamp);
        emit Deleted(hash);
    }
    function fileNewOwner(bytes32 hash, uint32 timestamp, address newOwner) public {
        address own = storagemgr.fileGetOwner(hash, timestamp);
        require(msg.sender == own, "denied!");
        storagemgr.fileNewOwner(hash, timestamp, newOwner);
        emit OwnerChanged(hash, timestamp, msg.sender, newOwner);
    }
    function fileSetPerm(bytes32 hash, uint32 timestamp, address addr, bool write) public {
        address own = storagemgr.fileGetOwner(hash, timestamp);
        if (write) require(msg.sender == own, "denied!");
        else require(storagemgr.fileHasWriteAccess(hash, timestamp, msg.sender));
        storagemgr.fileSetPerm(hash, timestamp, addr);
        emit UpdatedPerm(hash, timestamp);
    }
    function fileGetPerm(bytes32 hash, uint32 timestamp, address addr, bool write) public view returns (bool) {
        address own = storagemgr.fileGetOwner(hash, timestamp);
        require(msg.sender == own, "denied!");
        return storagemgr.fileGetPerm(hash, timestamp, addr, write);
    }
    function fileRevokePerm(bytes32 hash, uint32 timestamp,  address addr) public returns (bool){
       address own = storagemgr.fileGetOwner(hash, timestamp);
       require(msg.sender == own, "denied!");
       storagemgr.fileRevokePerm(hash, timestamp, addr);
       emit UpdatedRevoke(hash, timestamp);
    }
    function fileRequestAccess(bytes32 hash, uint32 timestamp) public {
        emit RequestAccess(hash, timestamp, msg.sender);
    }
    function safeAdd(uint a, uint b) pure internal returns (uint) {
        uint c = a + b;
        assert(c >= a && c >= b);
        return c;
    }
    function safeMul(uint a, uint b) internal pure returns (uint c) {
        if (a == 0) {
            return 0;
        }
        c = a * b;
        assert(c / a == b);
        return c;
    }
}
