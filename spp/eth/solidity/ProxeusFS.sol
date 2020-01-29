pragma solidity ^0.4.23;

import "./VMap.sol";
import "./EternalStorageInterface.sol";

contract ERC20IF {
    function transfer(address to, uint value) public returns (bool ok);
    function transferFrom(address from, address to, uint value) public returns (bool ok);
    function approve(address spender, uint value) public returns (bool ok);
    function allowance(address tokenOwner, address spender) public constant returns (uint remaining);
}
contract ProxeusFS {

    using VMap for VMap.addressSetProviderStore;

    uint internal proxeusPrice  = 0.1 ether;
    uint internal spPrice = 0.1 ether;
    address internal XES_ADDRESS;// = 0x84E0b37e8f5B4B86d5d299b0B0e33686405A3919;//ropsten
    bytes32 public dappVersion;

    address internal owner;
    address internal issuer;
    EternalStorageInterface internal eternalstorage;

    VMap.addressSetProviderStore internal whitelistedSP;

    constructor(address ownr, address tokenAddr) public {
        owner = ownr;
        issuer = msg.sender;
        XES_ADDRESS = tokenAddr;
    }

    event Deleted(bytes32 indexed hash);
    event UpdatedEvent(bytes32 indexed oldHash, bytes32 indexed newHash);
    event RequestSign(bytes32 hash, address indexed to);
    event NotifySign(bytes32 indexed hash, address indexed who);
    event OwnerChanged(bytes32 indexed hash, address oldOwner, address newOwner);
    event RequestAccess(bytes32 hash, address who);
    event PaymentReceived(bytes32 hash, uint xesAmount, address storageProvider);


    function setDappVersion(bytes32 version) public {
        require(msg.sender == owner);
        dappVersion = version;
    }

    function setEternalStorage(address _eternalstorage) public {
        require(msg.sender == owner);
        eternalstorage = EternalStorageInterface(_eternalstorage);
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
    }

    function spRemove(address strProv) public {
        require(msg.sender == owner, "denied!");
        whitelistedSP.Remove(strProv);
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
        uint fileType = getFiletype(hash);
        if(fileType > 0){
            address fileOwner = getFileOwner(hash);
            require(msg.sender == fileOwner);
            eternalstorage.setFileStorageProvider(hash,strPrv,true);
            return;
        }
        revert();
    }

    function fileHasSP(bytes32 hash, address addr) public view returns (bool) {
        uint fileType = getFiletype(hash);
        if(fileType > 0){
            if(fileType == 1) {
                bytes32 fileParent = getFileParent(hash);
                return getFileStorageprovider(fileParent, addr);
            }else {
                return getFileStorageprovider(hash, addr);
            }
        }
        return false;
    }

    function createFileThumbnail(bytes32 hash, bytes32 pParent, bool pPublic) public {
        uint fileTypeParent = getFiletype(pParent);
        if(fileTypeParent > 1){
            uint fileType = getFiletype(hash);
            if(fileType == 0){
                address fileParentOwner = getFileOwner(pParent);
                require(msg.sender == fileParentOwner, "denied!");
                bytes32 fileParentThumbHash = eternalstorage.getFileThumbhash(pParent);
                if(fileParentThumbHash != bytes32(0)){
                    uint oldfileThumbhashFileType = getFiletype(fileParentThumbHash);
                    if(oldfileThumbhashFileType == 1){
                        eternalstorage.setFileReplacesFile(hash,fileParentThumbHash);
                    }
                }

                eternalstorage.setFiletype(hash,1);


                eternalstorage.setFileIsPublic(hash,pPublic);

                eternalstorage.setFileParent(hash,pParent);
                eternalstorage.setFileThumbhash(pParent,hash);

                return;
            }
        }
        revert();
    }

    function createFileUndefinedSigners(bytes32 hash, uint mandatorySigners, uint expiry, bytes32 replacesFile, address[] prvs, uint xesAmount) public {
        _createFile(2, hash, new address[](mandatorySigners), expiry, replacesFile, prvs, xesAmount);
    }

    function createFileShared(bytes32 hash, uint mandatorySigners, uint expiry, bytes32 replacesFile, address[] prvs, address[] readers, uint xesAmount) public {
        _createFile(2, hash, new address[](mandatorySigners), expiry, replacesFile, prvs, xesAmount);
        fileSetPerm(hash,readers);
    }

    function createFileDefinedSigners(bytes32 hash, address[] definedSigners, uint expiry, bytes32 replacesFile, address[] prvs, uint xesAmount) public {
        _createFile(3, hash, definedSigners, expiry, replacesFile, prvs, xesAmount);
    }

    function _createFile(uint fileType, bytes32 hash, address[] definedSigners, uint expiry, bytes32 replacesFile, address[] prvs, uint xesAmount) internal {
        uint lfileType = getFiletype(hash);
        if(lfileType == 0){
            require(ERC20IF(XES_ADDRESS).allowance(msg.sender, this) >= xesAmount);

            for (uint ii = 0; ii < prvs.length; ii++){
                eternalstorage.setFileStorageProvider(hash,prvs[ii],true);
                require(ERC20IF(XES_ADDRESS).transferFrom(msg.sender, prvs[ii], xesAmount));

                emit PaymentReceived(hash, xesAmount, prvs[ii]);
                break; //for now only one storage provider is supported
            }

            if(getFiletype(replacesFile) > 1){
                eternalstorage.setFileReplacesFile(hash,replacesFile);
            }
            eternalstorage.setFileOwner(hash,msg.sender);

            if(fileType==3){
                eternalstorage.setFileDefinedSigners(hash,definedSigners);
            }else{
                eternalstorage.setFileDefinedSignersLength(hash,0);
            }
            eternalstorage.setFiletype(hash,fileType);
            setFileSignersCount(hash,0);

            eternalstorage.setFileExpiry(hash,expiry);
            eternalstorage.setFileSignersLength(hash,definedSigners.length);

            uint defSigLength = eternalstorage.getFileDefinedSignersLength(hash);
            address[] memory fdefinedSigners=getFileDefinedSigners(hash);
            for (uint d = 0; d < defSigLength; d++){
                eternalstorage.insertFileReadAccess(hash,fdefinedSigners[d]);
                insertFS(fdefinedSigners[d],hash);
                emit RequestSign(hash, fdefinedSigners[d]);
            }
            insertFS(msg.sender,hash);
            emit UpdatedEvent(hash, hash);
        }else{
            revert();
        }
    }

    function fileList() public view returns (bytes32[]) {
        return getFSValues(msg.sender);
    }

    //1=thumbnail, 2=file with undefined signers, 3=file with defined signers
    function fileInfo(bytes32 hash) public view returns (bytes32 id, address ownr, uint fileType, bool removed, uint expiry, bool isPublic, bytes32 thumbnailHash, bytes32 fparent, bytes32 replacesFile, address[] readAccess, address[] definedSigners) {
        uint ffileType = getFiletype(hash);
        if(ffileType > 0){

            id = hash;
            if(ffileType==1){
                bytes32 pParent = getFileParent(hash);
                if(getFiletype(pParent)>1){
                    ownr = getFileOwner(pParent);
                }
            }else{
                ownr = getFileOwner(hash);
            }

            fileType =ffileType;
            removed = getFileRemoved(hash);
            expiry = getFileExpiry(hash);
            isPublic = getFileIspublic(hash);
            thumbnailHash = eternalstorage.getFileThumbhash(hash);
            fparent = getFileParent(hash);
            replacesFile = eternalstorage.getFileReplacesFile(hash);
            readAccess = eternalstorage.getFileReadAccessValues(hash);
            definedSigners = getFileDefinedSigners(hash);
        }
    }

    function fileSigners(bytes32 hash) public view returns(address[]){
        return getFileSigners(hash);
    }

    function fileSignersCount(bytes32 hash) public view returns(uint){
        return getFileSignersCount(hash);

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

    function fileRemove(bytes32 hash) public {
        uint fileType = getFiletype(hash);
        if(fileType > 0){
            if(fileType > 1){
                address fileOwner = getFileOwner(hash);
                require(msg.sender == fileOwner, "denied!");
            }else{
                bytes32 fileParent = getFileParent(hash);
                uint fileParentType = getFiletype(fileParent);
                if(fileParentType>1){
                    address fileParentOwner = getFileOwner(fileParent);
                    require(msg.sender == fileParentOwner, "denied!");
                }
            }
            eternalstorage.setFileRemoved(hash,true);
            emit Deleted(hash);
        } else {
            revert();
        }
    }
    function isFileRemoved(bytes32 hash) public view returns (bool) {
        return getFileRemoved(hash);
    }
    function fileSetPerm(bytes32 hash, address[] addr/*, bool write*/) public {
        uint fileType = getFiletype(hash);
        if(fileType > 0){
            if(getFileIspublic(hash)){
                revert();
                return;
            }
            address fileOwner = getFileOwner(hash);
            require(msg.sender == fileOwner, "denied!");
            bool updated=false;
            for (uint i = 0; i < addr.length; i++){
                if(/*!f.writeAccess[addr] && */fileOwner != addr[i]){
                    updated=true;
                    eternalstorage.insertFileReadAccess(hash,addr[i]);
                    insertFS(addr[i],hash);
                }
            }
            if(updated==true){
                emit UpdatedEvent(hash, hash);
                return;
            }
        }
        revert();
    }

    function fileGetPerm(bytes32 hash, address addr, bool write) public view returns(bool) {
        uint fileType = getFiletype(hash);
        bool isPublic = getFileIspublic(hash);
        if(fileType == 1){
            bytes32 fileParent = getFileParent(hash);
            uint fileParentType = getFiletype(fileParent);
            if(fileParentType>1){
                if(write){
                    address fileParentOwner = getFileOwner(fileParent);
                    return msg.sender == fileParentOwner || addr == fileParentOwner;
                }
                return isPublic || msg.sender == fileParentOwner || addr == fileParentOwner || getFileHasReadAccess(fileParent,addr);
            }
        }
        address fileOwner = getFileOwner(hash);
        if(write){
            return msg.sender == fileOwner || addr == fileOwner;
        }
        return isPublic || msg.sender == fileOwner || addr == fileOwner || getFileHasReadAccess(hash,addr);
    }
    function fileRevokePerm(bytes32 hash, address[] addr/*, bool write*/) public returns(bool){
        uint fileType = getFiletype(hash);
        if(fileType > 0){
            bool isPublic = getFileIspublic(hash);
            if(isPublic){
                revert();
            }
            address fileOwner = getFileOwner(hash);
            require(msg.sender == fileOwner, "denied!");
            for (uint i = 0; i < addr.length; i++){
                if(getFileHasReadAccess(hash,addr[i])){
                    eternalstorage.removeFileAccess(hash,addr[i]);
                    removeFS(addr[i],hash);
                }
            }
            emit UpdatedEvent(hash, hash);
        }
    }

    function fileExpiry(bytes32 hash) public view returns (uint) {
        return getFileExpiry(hash);
    }

    function fileVerify(bytes32 hash) public view returns(bool, address[]){
        uint fileType = getFiletype(hash);
        address[] memory signers = getFileSigners(hash);
        if(fileType > 1){
            uint fileexpiry=getFileExpiry(hash);
            bool fileremoved=getFileRemoved(hash);
            uint signersCount=getFileSignersCount(hash);
            uint signersLength = eternalstorage.getFileSignersLength(hash);
            if(fileType > 2){
                address[] memory definedSigners = getFileDefinedSigners(hash);
                bool ok = (fileexpiry==0 || fileexpiry>now) && !fileremoved && signersLength == signersCount;
                for (uint i = 0; i < signersLength; i++){
                    if(signers[i] == address(0) || definedSigners[i] != signers[i]){
                        return (false, signers);
                    }
                }
                return (ok, signers);
            }else{
                return ((fileexpiry==0 || fileexpiry>now) && !fileremoved && signersLength == signersCount, signers);
            }
        }
        return (false, signers);
    }

    //1=thumbnail, 2=file with undefined signers, 3=file with defined signers
    function fileSign(bytes32 hash) public {
        uint fileType = getFiletype(hash);
        if(fileType > 1){
            uint signersLength = eternalstorage.getFileSignersLength(hash);
            uint signersCount=getFileSignersCount(hash);
            address[] memory signers = getFileSigners(hash);
            if(fileType > 2){
                uint definedSignersLength = eternalstorage.getFileDefinedSignersLength(hash);
                address[] memory definedSigners = getFileDefinedSigners(hash);
                require(definedSignersLength>0);
                require(signersLength == definedSignersLength);
                for (uint ii = 0; ii < signersLength; ii++){
                    if(signers[ii]==msg.sender){
                        revert();
                        return;
                    } else if(definedSigners[ii] == msg.sender){
                        eternalstorage.setFileSigners(hash,ii,definedSigners[ii]);
                        signersCount++;
                        setFileSignersCount(hash,signersCount);
                        insertFS(msg.sender,hash);
                        emit NotifySign(hash, msg.sender);
                        return;
                    }
                }
            }else{
                require(signersLength > signersCount);
                uint freeIndex = 0;
                for (uint i = 0; i < signersLength; i++){
                    if(signers[i]==msg.sender){
                        revert();
                        return;
                    }else if(signers[i] == address(0)){
                        freeIndex = i;
                    }
                }
                eternalstorage.setFileSigners(hash,freeIndex,msg.sender);
                signersCount++;
                setFileSignersCount(hash,signersCount);
                insertFS(msg.sender,hash);
                emit NotifySign(hash, msg.sender);
                return;
            }
        }
        revert();
    }

    function fileRequestAccess(bytes32 hash) public {
        emit RequestAccess(hash, msg.sender);
    }
    function fileRequestSign(bytes32 hash, address[] signer) public {
        uint fileType = getFiletype(hash);
        if(fileType > 1){
            address fileOwner = getFileOwner(hash);
            require(msg.sender == fileOwner, "denied!");
            for (uint i = 0; i < signer.length; i++){
                emit RequestSign(hash, signer[i]);
            }
        }
    }
    function getFiletype(bytes32 hash) internal view returns(uint){
        return eternalstorage.getFiletype(hash);
    }
    function getFileOwner(bytes32 hash) internal view returns(address){
        return eternalstorage.getFileOwner(hash);
    }
    function getFileIspublic(bytes32 _hash) internal view returns(bool) {
        return eternalstorage.getFileIspublic(_hash);
    }
    function getFileParent(bytes32 _hash) internal view returns(bytes32) {
        return eternalstorage.getFileParent(_hash);
    }
    function getFileDefinedSigners(bytes32 _hash) internal view returns(address[]) {
        return eternalstorage.getFileDefinedSigners(_hash);
    }
    function getFileExpiry(bytes32 _hash) internal view returns(uint) {
        return eternalstorage.getFileExpiry(_hash);
    }
    function getFileRemoved(bytes32 _hash) internal view returns(bool) {
        return eternalstorage.getFileRemoved(_hash);
    }
    function getFileSigners(bytes32 _hash) internal view returns(address[]) {
        return eternalstorage.getFileSigners(_hash);
    }
    function getFileSignersCount(bytes32 _hash) internal view returns(uint) {
        return eternalstorage.getFileSignersCount(_hash);
    }
    function getFileStorageprovider(bytes32 _hash, address _addr) internal view returns(bool) {
        return eternalstorage.getFileStorageprovider(_hash,_addr);
    }
    function getFileHasReadAccess(bytes32 _hash, address _addr) internal view returns(bool) {
        return eternalstorage.getFileHasReadAccess(_hash,_addr);
    }

    function setFileSignersCount(bytes32 _hash, uint _count) internal {
        eternalstorage.setFileSignersCount(_hash, _count);
    }

    function insertFS(address fsaddr, bytes32 fshash) internal {
        eternalstorage.insertFS(fsaddr, fshash);
    }
    function getFSValues(address fsaddr) internal view returns(bytes32[]){
        return eternalstorage.getFSValues(fsaddr);
    }
    function removeFS(address fsaddr, bytes32 fshash) internal {
        eternalstorage.removeFS(fsaddr, fshash);
    }
}
