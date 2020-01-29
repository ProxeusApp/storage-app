pragma solidity ^0.4.23;

import "./VMap.sol";

contract EternalStorage {

    address owner = msg.sender;

    using VMap for VMap.bytes32Set;
    using VMap for VMap.addressSet;

    struct File {
        address owner;
        uint fileType;//1=thumbnail, 2=file with undefined signers, 3=file with defined signers
        uint signersCount;
        address[] definedSigners;
        address[] signers;

        VMap.addressSet readAccess;

        mapping(address => bool) storageProviders;

        bool removed;
        uint expiry;// https://ethereum.stackexchange.com/questions/32173/how-to-handle-dates-in-solidity-and-web3
        bytes32 replacesFile;
        bool isPublic;
        bytes32 thumbnailHash;
        bytes32 parent; //for thumbnails
        bytes32 data;
    }

    mapping(bytes32 => File) internal bigStore;
    mapping(address => VMap.bytes32Set) internal fs;

    mapping(address => bool) internal permittedAddresses;

    modifier onlyPermitted() {
        require(permittedAddresses[msg.sender] == true);
        _;
    }


    // *** Access Management Methods ***
    function permitAddress(address _address) external {
        require(msg.sender == owner);
        permittedAddresses[_address] = true;
    }

    function denyAddress(address _address) external {
        require(msg.sender == owner);
        permittedAddresses[_address] = false;
    }

    // *** bigStore Methods ***
    function getFiletype(bytes32 _hash) onlyPermitted external view returns(uint) {
        return bigStore[_hash].fileType;
    }
    function getFileIspublic(bytes32 _hash) onlyPermitted external view returns(bool) {
        return bigStore[_hash].isPublic;
    }
    function getFileParent(bytes32 _hash) onlyPermitted external view returns(bytes32) {
        return bigStore[_hash].parent;
    }
    function getFileStorageprovider(bytes32 _hash, address _stprv) onlyPermitted external view returns(bool) {
        return bigStore[_hash].storageProviders[_stprv];
    }
    function getFileThumbhash(bytes32 _hash) onlyPermitted external view returns(bytes32) {
        return bigStore[_hash].thumbnailHash;
    }
    function getFileReplacesFile(bytes32 _hash) onlyPermitted external view returns(bytes32) {
        return bigStore[_hash].replacesFile;
    }
    function getFileOwner(bytes32 _hash) onlyPermitted external view returns(address) {
        return bigStore[_hash].owner;
    }
    function getFileData(bytes32 _hash) onlyPermitted external view returns(bytes32) {
        return bigStore[_hash].data;
    }
    function getFileDefinedSigners(bytes32 _hash) onlyPermitted external view returns(address[]) {
        return bigStore[_hash].definedSigners;
    }
    function getFileDefinedSignersLength(bytes32 _hash) onlyPermitted external view returns(uint) {
        return bigStore[_hash].definedSigners.length;
    }
    function getFileSignersCount(bytes32 _hash) onlyPermitted external view returns(uint) {
        return bigStore[_hash].signersCount;
    }
    function getFileSignersLength(bytes32 _hash) onlyPermitted external view returns(uint) {
        return bigStore[_hash].signers.length;
    }
    function getFileExpiry(bytes32 _hash) onlyPermitted external view returns(uint) {
        return bigStore[_hash].expiry;
    }
    function getFileSigners(bytes32 _hash) onlyPermitted external view returns(address[]) {
        return bigStore[_hash].signers;
    }
    function getFileHasReadAccess(bytes32 _hash, address _addr) onlyPermitted external view returns(bool) {
        return bigStore[_hash].readAccess.Has(_addr);
    }
    function getFileReadAccessValues(bytes32 _hash) onlyPermitted external view returns(address[]) {
        return bigStore[_hash].readAccess.Values();
    }
    function getFileRemoved(bytes32 _hash) onlyPermitted external view returns(bool) {
        return bigStore[_hash].removed;
    }


    function setFiletype(bytes32 _hash, uint _type) onlyPermitted external {
        bigStore[_hash].fileType=_type;
    }
    function setFileIsPublic(bytes32 _hash, bool _flag) onlyPermitted external {
        bigStore[_hash].isPublic=_flag;
    }
    function setFileParent(bytes32 _hash, bytes32 _parent) onlyPermitted external {
        bigStore[_hash].parent=_parent;
    }
    function setFileStorageProvider(bytes32 _hash, address _stprv, bool _flag) onlyPermitted {
        bigStore[_hash].storageProviders[_stprv]=_flag;
    }
    function setFileThumbhash(bytes32 _hash, bytes32 _thash) onlyPermitted external {
        bigStore[_hash].thumbnailHash=_thash;
    }
    function setFileReplacesFile(bytes32 _hash, bytes32 _filehash) onlyPermitted external {
        bigStore[_hash].replacesFile=_filehash;
    }
    function setFileOwner(bytes32 _hash, address _owner) onlyPermitted external  {
        bigStore[_hash].owner=_owner;
    }
    function setFileData(bytes32 _hash, bytes32 _data) onlyPermitted external  {
        bigStore[_hash].data=_data;
    }
    function insertFileReadAccess(bytes32 _hash, address _accesshash) onlyPermitted external {
        bigStore[_hash].readAccess.Insert(_accesshash);
    }
    function removeFileAccess(bytes32 _hash, address _accesshash) onlyPermitted external {
        bigStore[_hash].readAccess.Remove(_accesshash);
    }
    function setFileDefinedSigners(bytes32 _hash, uint _pos, address _address) onlyPermitted external {
        bigStore[_hash].definedSigners[_pos]=_address;
    }
    function setFileDefinedSigners(bytes32 _hash, address[] _addresses) onlyPermitted external {
        bigStore[_hash].definedSigners=_addresses;
    }
    function setFileDefinedSignersLength(bytes32 _hash, uint _length) onlyPermitted external {
        bigStore[_hash].definedSigners.length=_length;
    }
    function setFileSigners(bytes32 _hash, uint _pos, address _address) onlyPermitted external {
        bigStore[_hash].signers[_pos]=_address;
    }
    function setFileSignersCount(bytes32 _hash, uint _count) onlyPermitted external {
        bigStore[_hash].signersCount=_count;
    }
    function setFileExpiry(bytes32 _hash, uint _exp) onlyPermitted external  {
        bigStore[_hash].expiry=_exp;
    }
    function setFileSignersLength(bytes32 _hash, uint _length) onlyPermitted external {
        bigStore[_hash].signers.length=_length;
    }
    function setFileRemoved(bytes32 _hash,bool _flag) onlyPermitted external {
        bigStore[_hash].removed=_flag;
    }


    // *** FS Methods ***
    function getFSValues(address _fs) onlyPermitted external view returns(bytes32[]) {
        return fs[_fs].Values();
    }

    function insertFS(address _fs, bytes32 _hash) onlyPermitted external {
        fs[_fs].Insert(_hash);
    }

    function removeFS(address _fs,bytes32 _hash) onlyPermitted external {
        fs[_fs].Remove(_hash);
    }

}
