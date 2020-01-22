pragma solidity ^0.4.23;

import "./VMap.sol";

contract ERC20IF {
    function transfer(address to, uint value) public returns (bool ok);
    function transferFrom(address from, address to, uint value) public returns (bool ok);
    function approve(address spender, uint value) public returns (bool ok);
    function allowance(address tokenOwner, address spender) public constant returns (uint remaining);
}
contract ProxeusFS {

    using VMap for VMap.bytes32Set;
    using VMap for VMap.addressSet;
    using VMap for VMap.addressSetProviderStore;

    uint internal proxeusPrice  = 0.1 ether;
    // uint internal proxeusPrice;// = 500000000000000000;//0.5
    uint internal spPrice = 0.1 ether;
    // uint internal spPrice;// = 500000000000000000;//0.5
    address internal XES_ADDRESS;// = 0x84E0b37e8f5B4B86d5d299b0B0e33686405A3919;//ropsten
    uint public dappVersion = 0;

    struct File {
        //bytes32 hash;
        address owner;
        uint fileType;//1=thumbnail, 2=file with undefined signers, 3=file with defined signers
        uint signersCount;
        address[] definedSigners;
        address[] signers;

        VMap.addressSet readAccess;
        //mapping(address => bool) readAccess;
        //mapping(address => bool) writeAccess;

        //VMap.addressSet storageProviders;
        mapping(address => bool) storageProviders;

        bool removed;
        bool invalidated;
        uint expiry;// https://ethereum.stackexchange.com/questions/32173/how-to-handle-dates-in-solidity-and-web3
        bytes32 replacesFile;
        bool isPublic;
        bytes32 thumbnailHash;
        bytes32 parent; //for thumbnails
    }

    address internal owner;
    address internal issuer;

    VMap.addressSetProviderStore internal whitelistedSP;

    mapping(bytes32 => File) internal bigStore;
    mapping(address => VMap.bytes32Set) internal fs;

    constructor(address ownr, address tokenAddr) public {
        owner = ownr;
        issuer = msg.sender;
        XES_ADDRESS = tokenAddr;
    }

    event Invalidated(bytes32 indexed hash);
    event Deleted(bytes32 indexed hash);
    event UpdatedEvent(bytes32 indexed oldHash, bytes32 indexed newHash);
    event RequestSign(bytes32 hash, address indexed to);
    event NotifySign(bytes32 indexed hash, address indexed who);
    event OwnerChanged(bytes32 indexed hash, address oldOwner, address newOwner);
    event RequestAccess(bytes32 hash, address who);


    function setDappVersion(uint version) public {
        require(msg.sender == owner);
        dappVersion = version;
    }

    function XESAmountPerFile(address[] prvs) public view returns (uint sum) {
        uint storageProvPrice = safeMul(spPrice, prvs.length);
        /*if(whitelistedStorageProviders.Size()>0){
            storageProvPrice = safeMul(spPrice, whitelistedStorageProviders.Size());
        }*/
        sum = safeAdd(proxeusPrice, storageProvPrice);
    }

    function XESAllowence(address sendr) public view returns (uint sum) {
        if(sendr != address(0)){
            sum = ERC20IF(XES_ADDRESS).allowance(sendr, this);
        }else if(msg.sender != address(0)){
            sum = ERC20IF(XES_ADDRESS).allowance(msg.sender, this);
        }
    }

    function spAdd(address strProv, bytes32 urlPrefix) public {
        require(msg.sender == owner, "denied!");
        if(whitelistedSP.Insert(strProv)){
            if(urlPrefix != bytes32(0)){
                whitelistedSP.SetURLPrefix(strProv, urlPrefix);
            }
            return;
        }
        //revert();
    }

    function spUpdate(address strPrv, bytes32 urlPrefix) public {
        require(msg.sender == owner, "denied!");
        whitelistedSP.SetURLPrefix(strPrv, urlPrefix);
    }

    function spList() public view returns(address[]) {
        return whitelistedSP.Values();
    }

    function spInfo(address strPrv) public view returns (bytes32 urlPrefix) {
        urlPrefix = whitelistedSP.GetURLPrefix(strPrv);
    }

    function fileAddSP(bytes32 hash, address strPrv) public {
        require(!whitelistedSP.Has(strPrv));
        File storage f = bigStore[hash];
        if(f.fileType > 0){
            require(msg.sender == f.owner);
            f.storageProviders[strPrv] = true;
            return;
        }
        revert();
    }

    function fileHasSP(bytes32 hash, address addr) public view returns (bool) {
        File storage f = bigStore[hash];
        if(f.fileType > 0) {
            if(f.fileType == 1) {
                // if thumbnail then check the providers of the parent
                File storage fp = bigStore[f.parent];
                return fp.storageProviders[addr];
            }else {
                return f.storageProviders[addr];
            }
        }
        return false;
    }

    function createFileThumbnail(bytes32 hash, bytes32 pParent, bool pPublic) public {
        File storage fparent = bigStore[pParent];
        if(fparent.fileType > 1){
            File storage f = bigStore[hash];
            if(f.fileType == 0){ // TODO payment
                require(msg.sender == fparent.owner, "denied!");
                //require(payWithTokens(fparent), "denied!");
                if(fparent.thumbnailHash != bytes32(0)){
                    File storage oldThumb = bigStore[fparent.thumbnailHash];
                    if(oldThumb.fileType == 1){
                        f.replacesFile = fparent.thumbnailHash;
                    }
                }

                //f.hash = hash;
                f.fileType = 1;//1=thumbnail

                f.isPublic = pPublic;
                f.parent = pParent;
                fparent.thumbnailHash = hash;

                //emit UpdatedEvent(hash, hash);
                return;
            }
        }
        revert();
    }

    function createFileUndefinedSigners(bytes32 hash, uint mandatorySigners, uint expiry, bytes32 replacesFile, address[] prvs) public {
        _createFile(2, hash, new address[](mandatorySigners), expiry, replacesFile, prvs);
    }

    function createFileDefinedSigners(bytes32 hash, address[] definedSigners, uint expiry, bytes32 replacesFile, address[] prvs) public {
        _createFile(3, hash, definedSigners, expiry, replacesFile, prvs);
    }

    function _createFile(uint fileType, bytes32 hash, address[] definedSigners, uint expiry, bytes32 replacesFile, address[] prvs) internal {
        require(msg.sender == owner, "denied!");
        //require(prvs.length>0, "denied!");
        File storage f = bigStore[hash];
        if(f.fileType == 0){
            //if(whitelistedStorageProviders.Contains(prvs)){

            //require(payWithTokens(f), "denied!");
            //require(_writeGrant(f, address(0)), "denied!");
            /*uint count = 0;
            for(uint i = 0; i < prvs.length; i++){
                if(whitelistedStorageProviders.Has(prvs[i])){
                    f.storageProviders[prvs[i]]= true;
                    count++;
                }
            }*/
            //require(count>0, "denied!");
            uint storageProvPrice = safeMul(spPrice, prvs.length);

            /*if(whitelistedStorageProviders.Size()>0){
                storageProvPrice = safeMul(spPrice, whitelistedStorageProviders.Size());
            }*/

            uint sum = safeAdd(proxeusPrice, storageProvPrice);

            for (uint ii = 0; ii < prvs.length; ii++){
                f.storageProviders[prvs[ii]]= true;
            }

            /*if(whitelistedStorageProviders.Size()>0){
                for (uint i = 0; i < whitelistedStorageProviders._values.length; i++){
                    if(whitelistedStorageProviders._values[i] != address(0)){
                        require(ERC20IF(XES_ADDRESS).transfer(whitelistedStorageProviders._values[i], spPrice));
                    }
                }
            }*/
            if(bigStore[replacesFile].fileType > 1){
                f.replacesFile = replacesFile;
            }
            //f.hash = hash;
            f.owner = msg.sender;

            if(fileType==3){
                f.definedSigners = definedSigners;
            }else{
                f.definedSigners.length = 0;
            }
            f.fileType = fileType;//3=file with defined signers
            //f.mandatorySigners = definedSigners.length;
            f.signersCount = 0;

            f.expiry = expiry;
            f.signers.length = definedSigners.length;
            //f.signers = new address[](mandatorySigners);

            for (uint d = 0; d < f.definedSigners.length; d++){
                f.readAccess.Insert(f.definedSigners[d]);
                fs[f.definedSigners[d]].Insert(hash);
                emit RequestSign(hash, f.definedSigners[d]);
            }
            fs[f.owner].Insert(hash);
            emit UpdatedEvent(hash, hash);
            //return;
        }else{
            revert();
        }
        //}
        //revert();
    }

    function fileList() public view returns (bytes32[]) {
        return fs[msg.sender].Values();
    }

    //1=thumbnail, 2=file with undefined signers, 3=file with defined signers
    function fileInfo(bytes32 hash) public view returns (bytes32 id, address ownr, uint fileType, bool removed, uint expiry, bool isPublic, bytes32 thumbnailHash, bytes32 fparent, bytes32 replacesFile, address[] readAccess, address[] definedSigners) {
        File storage f = bigStore[hash];
        if(f.fileType > 0){
            id = hash;
            if(f.fileType==1){
                File storage fp = bigStore[f.parent];
                if(fp.fileType>1){
                    ownr = fp.owner;
                }
            }else{
                ownr = f.owner;
            }

            fileType  = f.fileType;
            removed  = f.removed;
            expiry  = f.expiry;
            isPublic  = f.isPublic;
            thumbnailHash  = f.thumbnailHash;
            fparent  = f.parent;
            replacesFile = f.replacesFile;
            readAccess = f.readAccess.Values();
            definedSigners = f.definedSigners;
        }
    }

    function fileProviders(bytes32 hash) public view returns (address[] , bytes32[]) {
        File storage f = bigStore[hash];

        // get the storages the file is stored in
        address[] memory splist = whitelistedSP.Values();
        address[] memory storedAt = new address[](0);
        bytes32[] memory storedAtURLPrefix = new bytes32[](0);

        uint pos = 0;
        for(uint i = 0; i < splist.length; i++) {
            if(f.storageProviders[splist[i]]) {
                storedAt[pos] = splist[i];
                storedAtURLPrefix[pos] = whitelistedSP.GetURLPrefix(splist[i]);
                pos++;
            }
        }

        return (storedAt, storedAtURLPrefix);
    }

    function fileSigners(bytes32 hash) public view returns(address[]){
        return bigStore[hash].signers;
    }

    function fileSignersCount(bytes32 hash) public view returns(uint){
        return bigStore[hash].signersCount;
    }

    /*function payWithTokens(File storage f) internal returns(bool) {
        require(_writeGrant(f, address(0)), "denied!");

        uint storageProvPrice = spPrice;

        if(f.storageProviders.Size()>0){
            storageProvPrice = safeMul(spPrice, f.storageProviders.Size());
        }

        if(whitelistedStorageProviders.Size()>0){
            storageProvPrice = safeMul(spPrice, whitelistedStorageProviders.Size());
        }

        uint sum = safeAdd(proxeusPrice, storageProvPrice);

        require(ERC20IF(XES_ADDRESS).allowance(msg.sender, this) >= sum);
        require(ERC20IF(XES_ADDRESS).transferFrom(msg.sender, this, sum));
        require(ERC20IF(XES_ADDRESS).transfer(owner, proxeusPrice));

        if(f.storageProviders.Size()>0){
            for (uint i = 0; i < f.storageProviders._values.length; i++){
                if(f.storageProviders._values[i] != address(0)){
                    require(ERC20IF(XES_ADDRESS).transfer(f.storageProviders._values[i], spPrice));
                }
            }
        }
        if(whitelistedStorageProviders.Size()>0){
            for (uint i = 0; i < whitelistedStorageProviders._values.length; i++){
                if(whitelistedStorageProviders._values[i] != address(0)){
                    require(ERC20IF(XES_ADDRESS).transfer(whitelistedStorageProviders._values[i], spPrice));
                }
            }
        }
        return true;
    }*/

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

    function fileRemove(bytes32 hash) public {
        File storage f = bigStore[hash];
        if(f.fileType > 0){
            if(f.fileType > 1){
                require(msg.sender == f.owner, "denied!");
            }else{
                File storage fp = bigStore[f.parent];
                if(fp.fileType>1){
                    require(msg.sender == fp.owner, "denied!");
                }
            }
            f.removed = true;
            emit Deleted(hash);
        } else {
            revert();
        }
    }
    function isFileRemoved(bytes32 hash) public view returns (bool) {
        return bigStore[hash].removed;
    }

    function fileInvalidate(bytes32 hash) public {
        File storage f = bigStore[hash];
        if(f.fileType > 0){
            if(f.fileType > 1){
                require(msg.sender == f.owner, "denied!");
            }else{
                File storage fp = bigStore[f.parent];
                if(fp.fileType>1){
                    require(msg.sender == fp.owner, "denied!");
                }
            }
            f.invalidated = true;
            emit Invalidated(hash);
        } else {
            revert();
        }
    }
    function isFileInvalidated(bytes32 hash) public view returns (bool) {
        return bigStore[hash].invalidated;
    }

    function fileNewOwner(bytes32 hash, address newOwner) public {
        File storage f = bigStore[hash];
        if(f.fileType > 1){
            require(msg.sender == f.owner, "denied!");
            fs[f.owner].Remove(hash);
            f.owner = newOwner;
            //if(f.readAccess.Has(f.owner)){
            f.readAccess.Remove(f.owner);
            //}
            /*if(f.writeAccess[f.owner]){
                delete f.writeAccess[f.owner];
            }*/
            fs[f.owner].Insert(hash);
            emit OwnerChanged(hash, msg.sender, newOwner);
            return;
        }
        revert();
    }
    function fileSetPerm(bytes32 hash, address addr/*, bool write*/) public {
        File storage f = bigStore[hash];
        if(f.fileType > 0){
            /*if(write){
                require(msg.sender == f.owner, "denied!");
                if(f.owner != addr){
                    if(f.readAccess[f.owner]){
                        delete f.readAccess[f.owner];
                    }
                    f.writeAccess[addr] = true;
                    fs[addr].Insert(hash);
                    emit UpdatedEvent(hash, hash);
                    return;
                }
            }else{*/
            if(f.isPublic){
                revert();
                return;
            }
            require(msg.sender == f.owner, "denied!");
            if(/*!f.writeAccess[addr] && */f.owner != addr){
                f.readAccess.Insert(addr);
                fs[addr].Insert(hash);
                emit UpdatedEvent(hash, hash);
                return;
            }
            //}
        }
        revert();
    }

    function fileGetPerm(bytes32 hash, address addr, bool write) public view returns(bool) {
        File storage f = bigStore[hash];
        if(f.fileType==1){
            File storage fp = bigStore[f.parent];
            if(fp.fileType>1){
                if(write){
                    return msg.sender == fp.owner || addr == fp.owner;
                }
                return f.isPublic || msg.sender == fp.owner || addr == fp.owner || fp.readAccess.Has(addr);
            }
        }
        if(write){
            return msg.sender == f.owner || addr == f.owner;
        }
        return f.isPublic || msg.sender == f.owner || addr == f.owner || f.readAccess.Has(addr);
    }
    /*function _writeGrant(File storage f, address addr) internal view returns(bool){
        if(f.fileType > 0){
            if(addr == address(0)){
                if(msg.sender == f.owner){
                    return true;
                }
                //return f.writeAccess[msg.sender];
            }else{
                if(addr == f.owner){
                    return true;
                }
                //return f.writeAccess[addr];
            }
        }
        return false;
    }*/
    function fileRevokePerm(bytes32 hash, address addr/*, bool write*/) public returns(bool){
        File storage f = bigStore[hash];
        if(f.fileType > 0){
            /*if(write){
                require(msg.sender == f.owner, "denied!");
                if(f.writeAccess[addr]){
                    delete f.writeAccess[addr];
                    fs[addr].Remove(hash);
                }
            }else{*/
            if(f.isPublic){
                revert();
            }
            require(msg.sender == f.owner, "denied!");
            if(f.readAccess.Remove(addr)){
                //delete f.readAccess[addr];
                fs[addr].Remove(hash);
            }
            //}
            emit UpdatedEvent(hash, hash);
        }
        revert();
    }

    function fileExpiry(bytes32 hash) public view returns (uint) {
        return bigStore[hash].expiry;
    }

    function fileVerify(bytes32 hash) public view returns(bool, address[]){
        File memory f = bigStore[hash];
        if(f.fileType > 1){
            if(f.fileType > 2){
                bool ok = (f.expiry==0 || f.expiry>now) && !f.removed && !f.invalidated && f.signers.length == f.signersCount;
                for (uint i = 0; i < f.signers.length; i++){
                    if(f.signers[i] == address(0) || f.definedSigners[i] != f.signers[i]){
                        return (false, f.signers);
                    }
                }
                return (ok, f.signers);
            }else{
                return ((f.expiry==0 || f.expiry>now) && !f.removed && !f.invalidated && f.signers.length == f.signersCount, f.signers);
            }
        }
        return (false, f.signers);
    }

    //1=thumbnail, 2=file with undefined signers, 3=file with defined signers
    function fileSign(bytes32 hash) public {
        File storage f = bigStore[hash];
        if(f.fileType > 1){
            if(f.fileType > 2){
                require(f.definedSigners.length>0);
                require(f.signers.length == f.definedSigners.length);
                for (uint ii = 0; ii < f.signers.length; ii++){
                    if(f.signers[ii]==msg.sender){
                        revert();
                        return;
                    } else if(f.definedSigners[ii] == msg.sender){
                        f.signers[ii] = f.definedSigners[ii];
                        f.signersCount++;
                        fs[msg.sender].Insert(hash);
                        emit NotifySign(hash, msg.sender);
                        return;
                    }
                }
            }else{
                require(f.signers.length > f.signersCount);
                uint freeIndex = 0;
                for (uint i = 0; i < f.signers.length; i++){
                    if(f.signers[i]==msg.sender){
                        revert();
                        return;
                    }else if(f.signers[i] == address(0)){
                        freeIndex = i;
                    }
                }
                f.signers[freeIndex] = msg.sender;
                f.signersCount++;
                fs[msg.sender].Insert(hash);
                emit NotifySign(hash, msg.sender);
                return;
            }
        }
        revert();
    }

    function fileRequestAccess(bytes32 hash) public {
        emit RequestAccess(hash, msg.sender);
    }
    function fileRequestSign(bytes32 hash, address signer) public {
        File storage f = bigStore[hash];
        if(f.fileType > 2){
            require(msg.sender == f.owner, "denied!");
            emit RequestSign(hash, signer);
        }
    }
}
