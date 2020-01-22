pragma solidity ^0.4.23;

contract EternalStorageInterface {

    // *** bigStore Methods ***
    function getFiletype(bytes32 _hash) external view returns(uint) {}
    function getFileIspublic(bytes32 _hash) external view returns(bool) {}
    function getFileParent(bytes32 _hash) external view returns(bytes32) {}
    function getFileStorageprovider(bytes32 _hash, address _stprv) external view returns(bool) {}
    function getFileThumbhash(bytes32 _hash) external view returns(bytes32) {}
    function getFileReplacesFile(bytes32 _hash) external view returns(bytes32) {}
    function getFileOwner(bytes32 _hash) external view returns(address) {}
    function getFileData(bytes32 _hash) external view returns(bytes32) {}
    function getFileDefinedSigners(bytes32 _hash) external view returns(address[]) {}
    function getFileDefinedSignersLength(bytes32 _hash) external view returns(uint) {}
    function getFileSignersCount(bytes32 _hash) external view returns(uint) {}
    function getFileSignersLength(bytes32 _hash) external view returns(uint) {}
    function getFileExpiry(bytes32 _hash) external view returns(uint) {}
    function getFileSigners(bytes32 _hash) external view returns(address[]) {}
    function getFileHasReadAccess(bytes32 _hash, address _addr) external view returns(bool) {}
    function getFileReadAccessValues(bytes32 _hash) external view returns(address[]) {}
    function getFileRemoved(bytes32 _hash) external view returns(bool) {}


    function setFiletype(bytes32 _hash, uint _type) external {}
    function setFileIsPublic(bytes32 _hash, bool _flag) external {}
    function setFileParent(bytes32 _hash, bytes32 _parent) external {}
    function setFileStorageProvider(bytes32 _hash, address _stprv, bool _flag) external {}
    function setFileThumbhash(bytes32 _hash, bytes32 _thash) external {}
    function setFileReplacesFile(bytes32 _hash, bytes32 _filehash) external {}
    function setFileOwner(bytes32 _hash, address _owner) external {}
    function setFileData(bytes32 _hash, bytes32 _data) external {}
    function insertFileReadAccess(bytes32 _hash, address _accesshash) external {}
    function removeFileAccess(bytes32 _hash, address _accesshash) external {}
    function setFileDefinedSigners(bytes32 _hash, uint _pos, address _address) external {}
    function setFileDefinedSigners(bytes32 _hash, address[] _addresses) external {}
    function setFileDefinedSignersLength(bytes32 _hash, uint _length) external {}
    function setFileSigners(bytes32 _hash, uint _pos, address _address) external {}
    function setFileSignersCount(bytes32 _hash, uint _count) external {}
    function setFileExpiry(bytes32 _hash, uint _exp) external {}
    function setFileSignersLength(bytes32 _hash, uint _length) external {}
    function setFileRemoved(bytes32 _hash,bool _flag) external {}


    // *** FS Methods ***
    function getFSValues(address _fs) external view returns(bytes32[]) {}
    function insertFS(address _fs, bytes32 _hash) external {}
    function removeFS(address _fs,bytes32 _hash) external {}

}