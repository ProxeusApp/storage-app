pragma solidity ^0.4.23;

contract FileListmgr{
    function fileInsert(bytes32 hash, uint32 timestamp) public returns(uint);
    function fileList(uint32 timestamp, uint amount) public view returns (bytes32[], uint32[]);
    function Remove(uint index) public ;
}

contract ProviderListMgr {
    function spAdd(address strProv, bytes32 urlPrefix, uint[] price) public returns (bool);
    function spUpdate(address strPrv, bytes32 urlPrefix) public;
    function spInfo(address strPrv) public view returns (bytes32, uint[]);
    function spList() public view returns(address[]);
    function spHas(address prv) public view returns (bool);
}

contract VMap {
    struct FObject {
        address owner;
        uint index;
    }
    mapping(bytes32 => mapping(uint256 => File)) internal bigStore;
    mapping(address => FileListmgr) internal ownerStore; // including owner & read & write
    ProviderListMgr internal whitelistedSP;

    constructor() public {
    }

    struct File {
        uint    id;
        address owner;
        uint fileType;//1=thumbnail(no-use), 2=file with undefined signers, 3=file with defined signers
        uint32 timestamp;
        mapping(address => address) definedsignerStore;
        mapping(address => address) signerStore;
        mapping(address => address) readAccessStore;
        mapping(address => address) writeAccessStore;
        address lastreadAccess;
        uint readAccessCount;
        address lastwriteAccess;
        uint writeAccessCount;
        mapping(address => address) strgPrvStore;
        mapping(address => uint) strgPrvPriceStore;
        uint strgPrvCount;
        address laststrgPrv;
        uint definedsignerCount;
        address lastdefinedsigner;
        uint signersCount;
        address lastsigner;
        bool removed;
        uint expiry;// https://ethereum.stackexchange.com/questions/32173/how-to-handle-dates-in-solidity-and-web3
        mapping(bytes32 => bytes32) replacesFileStore;
    }

    function createFileInit(uint fileType, bytes32 hash, uint32 timestamp, address owner, uint expiry, bytes32 replacesFile, address[] definedSigners) public returns (bool) {
        File storage f = bigStore[hash][timestamp];
        if (f.fileType > 0) revert();
        f.fileType = fileType;
        f.timestamp= timestamp;
        f.owner = owner;
        f.expiry = expiry;
        f.replacesFileStore[bytes32(0)] = replacesFile;
        f.id = ownerStore[owner].fileInsert(hash, timestamp);
        for (uint j=0;j<definedSigners.length;j++){
            fileAddDefinedSigner(hash, timestamp, definedSigners[j]);
            fileSetPerm(hash, timestamp, definedSigners[j], false);
        }
        return true;
    }
    function createFileEnd(bytes32 hash, uint32 timestamp, address[] prvs, uint[] prices) public returns (bool) {
        require(prvs.length==prices.length);
        for (uint i=0;i<prvs.length;i++) {
            if (whitelistedSP.spHas(prvs[i]))
                fileAddSP(hash, timestamp, prvs[i], prices[i]);
        }
        return true;
    }

    function fileList(address addr, uint32 timestamp, uint amount) public view returns (bytes32[], uint32[]) {
        return ownerStore[addr].fileList(timestamp, amount);
    }
    function fileGetOwner(bytes32 hash, uint32 timestamp) public view returns(address)
    {
        return bigStore[hash][timestamp].owner;
    }
    function fileNewOwner(bytes32 hash, uint32 timestamp, address newOwner) public {
        File storage f = bigStore[hash][timestamp];
        ownerStore[f.owner].Remove(f.id);
        fileRevokePerm(hash, timestamp, f.owner, true);
        fileRevokePerm(hash, timestamp, f.owner, false);
        f.owner = newOwner;
        f.id = ownerStore[f.owner].fileInsert(hash, timestamp);
    }
    function fileInfo(bytes32 hash, uint32 timestamp) public view returns (address ownr, uint fileType, bool removed, uint expiry) {
        File storage f = bigStore[hash][timestamp];
        ownr = f.owner;
        fileType  = f.fileType;
        removed  = f.removed;
        expiry  = f.expiry;
    }
    function fileAddSP(bytes32 hash, uint32 timestamp, address strPrv, uint price) public {
        File storage f = bigStore[hash][timestamp];
        f.strgPrvStore[f.laststrgPrv] = strPrv;
        f.strgPrvStore[strPrv] = address(1);
        f.laststrgPrv = strPrv;
        f.strgPrvPriceStore[strPrv] = price;
        f.strgPrvCount ++;
    }
    function fileRemoveSP(bytes32 hash, uint32 timestamp, address strPrv) public {
        File storage f = bigStore[hash][timestamp];
        f.strgPrvStore[strPrv] = 0;
        address tmp = f.strgPrvStore[0];
        uint counter=0;
        while ((f.strgPrvStore[tmp]!=strPrv)&&(counter<f.strgPrvCount)) {
            address adr = f.strgPrvStore[tmp];
            tmp = adr;
            counter++;
        }
        if (counter<f.strgPrvCount)
           f.strgPrvStore[tmp] = 0;
    }
    function fileProviders(bytes32 hash, uint32 timestamp) public view returns (address[], uint[]) {
        File storage f = bigStore[hash][timestamp];
        address[] memory storedAt_addr = new address[](0);
        uint[] memory storedAt_price = new uint[](0);
        address tmp = f.strgPrvStore[0];
        for(uint i = 0; i < f.strgPrvCount; i++) {
            address addr = f.strgPrvStore[tmp];
            storedAt_addr[i] = tmp;
            storedAt_price[i] = f.strgPrvPriceStore[tmp];
            tmp = addr;
        }
        return (storedAt_addr, storedAt_price);
    }
    function fileHasSP(bytes32 hash, uint32 timestamp, address strPrv) public view returns (bool) {
        return ( bigStore[hash][timestamp].strgPrvStore[strPrv] > 0)? true:false;
    }
    function fileRemove(bytes32 hash, uint32 timestamp) public {
        File storage f = bigStore[hash][timestamp];
        f.removed = true;
        ownerStore[f.owner].Remove(f.id);
    }
    function isFileRemoved(bytes32 hash, uint32 timestamp) public view returns (bool) {
        return bigStore[hash][timestamp].removed;
    }
    function fileAddDefinedSigner(bytes32 hash, uint32 timestamp, address signer) public returns (bool) {
        File storage f = bigStore[hash][timestamp];
        if (f.definedsignerStore[signer]!=0) return false;
        if (f.fileType == 3)
        {
            f.definedsignerStore[f.lastdefinedsigner]=signer;
            f.definedsignerStore[signer]= address(1);
        }
        f.definedsignerCount++;
        return true;
    }
    function fileDefinedSingers(bytes32 hash, uint32 timestamp) public view returns (address[]) {
        File storage f = bigStore[hash][timestamp];
        address[] memory storedAt = new address[](0);
        address tmp = f.definedsignerStore[0];
        for(uint i = 0; i < f.definedsignerCount; i++) {
            address adr = f.definedsignerStore[tmp];
            storedAt[i] = tmp;
            tmp = adr;
        }
        return storedAt;
    }
    function fileSign(bytes32 hash, uint32 timestamp, address signer) public returns (bool) {
        File storage f= bigStore[hash][timestamp];
        if (f.signerStore[signer]!=0) return false;
        if (((f.fileType == 3)&&(f.definedsignerStore[signer]!=0))||(f.fileType == 2)) {
            f.signerStore[f.lastsigner]=signer;
            f.signerStore[signer]= address(1);
            f.signersCount++;
            return true;
        }
        return false;
    }
    function fileSigners(bytes32 hash, uint32 timestamp) public view returns (address[]){
        File storage f = bigStore[hash][timestamp];
        address[] memory storedAt = new address[](0);
        address tmp = f.signerStore[0];
        for(uint i = 0; i < f.signersCount; i++) {
           address adr = f.signerStore[tmp];
           storedAt[i] = tmp;
           tmp = adr;
        }
        return storedAt;
    }
    function fileSignersCount(bytes32 hash, uint32 timestamp) public view returns (uint){
        return bigStore[hash][timestamp].signersCount;
    }
    function fileSetPerm(bytes32 hash, uint32 timestamp, address addr, bool write) public {
        File storage f = bigStore[hash][timestamp];
        if (write)
        {
            f.writeAccessStore[f.lastwriteAccess] = addr;
            f.writeAccessStore[addr] = address(1);
            f.lastwriteAccess = addr;
            f.writeAccessCount ++;
        }
        else {
            f.readAccessStore[f.lastreadAccess] = addr;
            f.readAccessStore[addr] = address(1);
            f.lastreadAccess = addr;
            f.readAccessCount ++;
        }
    }
    function fileHasWriteAccess(bytes32 hash, uint32 timestamp, address addr) public view returns(bool) {
        return ((bigStore[hash][timestamp].writeAccessStore[addr] > 0)? true:false);
    }
    function fileGetPerm(bytes32 hash, uint32 timestamp, address addr, bool write) public view returns(bool) {
        File storage f = bigStore[hash][timestamp];
        if(write){
            return msg.sender == f.owner || addr == f.owner|| ((bigStore[hash][timestamp].writeAccessStore[addr] > 0)? true:false);
        }
        return msg.sender == f.owner || addr == f.owner || ( (bigStore[hash][timestamp].readAccessStore[addr] > 0)? true:false);
    }
    function fileGetPerms(bytes32 hash, uint32 timestamp, bool write) public view returns (address[]) {
        File storage f = bigStore[hash][timestamp];
        address[] memory storedAt = new address[](0);
        if (write) {
            address tmp = f.writeAccessStore[0];
            for(uint i = 0; i < f.writeAccessCount; i++) {
                address adr = f.writeAccessStore[tmp];
                storedAt[i] = tmp;
                tmp = adr;
            }
        } else {
            address tmp_r = f.readAccessStore[0];
            for(uint j = 0; j < f.readAccessCount; j++) {
                address adr_r = f.readAccessStore[tmp_r];
                storedAt[j] = tmp_r;
                tmp_r = adr_r;
            }
        }
        return storedAt;
    }
    function fileRevokePerm(bytes32 hash, uint32 timestamp, address addr, bool write) public {
        File storage f = bigStore[hash][timestamp];
        f.readAccessStore[addr] = 0;
        if (write) {
            address tmp_w = f.writeAccessStore[0];
            uint counter_w=0;
            while ((f.writeAccessStore[tmp_w]!=addr)&&(counter_w<f.writeAccessCount)) {
                address adr_w = f.writeAccessStore[tmp_w];
                tmp_w = adr_w;
                counter_w++;
            }
            if (counter_w<f.writeAccessCount)
               f.writeAccessStore[tmp_w] = 0;
        }else {
            address tmp_r = f.readAccessStore[0];
            uint counter_r=0;
            while ((f.readAccessStore[tmp_r]!=addr)&&(counter_r<f.readAccessCount)) {
                address adr_r = f.readAccessStore[tmp_r];
                tmp_r = adr_r;
                counter_r++;
            }
            if (counter_r<f.readAccessCount)
               f.readAccessStore[tmp_r] = 0;
        }
    }
    function fileExpiry(bytes32 hash, uint32 timestamp) public view returns (uint) {
        return bigStore[hash][timestamp].expiry;
    }
    function fileVerify(bytes32 hash, uint32 timestamp) public view returns(bool, address[]){
        File storage f = bigStore[hash][timestamp];
        bool ok = (f.expiry==0 || f.expiry>now) && !f.removed && /*!f.invalidated* &&*/ f.definedsignerCount == f.signersCount;
        if (!ok) return (false, fileSigners(hash, timestamp));
        if (f.fileType == 3) {
            address tmp = f.signerStore[0];
            for(uint i = 0; i < f.signersCount; i++) {
                if (f.definedsignerStore[tmp]==0) return (false, fileSigners(hash, timestamp));
                address adr = f.signerStore[tmp];
                tmp = adr;
            }
        }
        return (true, fileSigners(hash, timestamp));
    }
    function spAdd(address strProv, bytes32 urlPrefix, uint[] price) public returns (bool)
    {
        return whitelistedSP.spAdd(strProv, urlPrefix, price);
    }
    function spUpdate(address strPrv, bytes32 urlPrefix) public;
    function spInfo(address strPrv) public view returns (bytes32, uint[]);
    function spList() public view returns(address[]);
    function spHas(address prv) public view returns (bool);
}
