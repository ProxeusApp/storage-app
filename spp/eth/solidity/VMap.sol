pragma solidity ^0.4.23;

library VMap {
    struct addressMap {
        mapping(address => MObject) _data;
        address[] _values;
        uint[] _deleted;
        uint _size;
    }

    struct MObject {
        bool exists;
        uint index;
    }

    function Insert(addressMap storage self, address addr, address val) public returns(bool) {
        MObject storage mo = self._data[addr];
        if(!self._data[addr].exists){
            mo.exists = true;
            if(self._deleted.length>0){
                mo.index = self._deleted[self._deleted.length-1];
                self._values[mo.index]=val;
                self._deleted.length--;
            }else{
                mo.index = self._values.length;
                self._values.push(val);
            }
            self._size++;
            return true;
        }
        return false;
    }
    function Remove(addressMap storage self, address addr) public returns(bool) {
        if(self._data[addr].exists){
            self._deleted.push(self._data[addr].index);
            delete self._values[self._data[addr].index];
            delete self._data[addr];
            self._size--;
            return true;
        }
        return false;
    }
    function Has(addressMap storage self, address addr) public view returns(bool) {
        return self._data[addr].exists;
    }
    function Get(addressMap storage self, address addr) public view returns(address) {
        if(self._data[addr].exists){
            return self._values[self._data[addr].index];
        }
        return address(0);
    }
    function Values(addressMap storage self) public view returns(address[]) {
        address[] memory cleaned = new address[](self._size);
        uint count = 0;
        for (uint i = 0; i<self._values.length; i++){
            if(self._values[i]!=address(0)){
                cleaned[count] = (self._values[i]);
                count++;
            }
        }
        return (cleaned);
    }
    function Size(addressMap storage self) public view returns(uint) {
        return self._size;
    }
    struct addressBytes32Map {
        mapping(address => MObject) _data;
        bytes32[] _values;
        uint[] _deleted;
        uint _size;
    }

    function Insert(addressBytes32Map storage self, address addr, bytes32 val) public returns(bool) {
        MObject storage mo = self._data[addr];
        if(!self._data[addr].exists){
            mo.exists = true;
            if(self._deleted.length>0){
                mo.index = self._deleted[self._deleted.length-1];
                self._values[mo.index]=val;
                self._deleted.length--;
            }else{
                mo.index = self._values.length;
                self._values.push(val);
            }
            self._size++;
            return true;
        }
        return false;
    }
    function Remove(addressBytes32Map storage self, address addr) public returns(bool) {
        if(self._data[addr].exists){
            self._deleted.push(self._data[addr].index);
            delete self._values[self._data[addr].index];
            delete self._data[addr];
            self._size--;
            return true;
        }
        return false;
    }
    function Has(addressBytes32Map storage self, address addr) public view returns(bool) {
        return self._data[addr].exists;
    }
    function Get(addressBytes32Map storage self, address addr) public view returns(bytes32) {
        if(self._data[addr].exists){
            return self._values[self._data[addr].index];
        }
        return bytes32(0);
    }
    function Values(addressBytes32Map storage self) public view returns(bytes32[]) {
        bytes32[] memory cleaned = new bytes32[](self._size);
        uint count = 0;
        for (uint i = 0; i<self._values.length; i++){
            if(self._values[i]!=bytes32(0)){
                cleaned[count] = (self._values[i]);
                count++;
            }
        }
        return (cleaned);
    }
    struct bytes32Set {
        bool exists;
        mapping(bytes32 => MObject) _data;
        bytes32[] _values;
        uint[] _deleted;
        uint _size;
    }

    function Insert(bytes32Set storage self, bytes32 addr) public returns(bool) {
        MObject storage mo = self._data[addr];
        if(!self._data[addr].exists){
            self.exists = true;
            mo.exists = true;
            if(self._deleted.length>0){
                mo.index = self._deleted[self._deleted.length-1];
                self._values[mo.index]=addr;
                self._deleted.length--;
            }else{
                mo.index = self._values.length;
                self._values.push(addr);
            }
            self._size++;
            return true;
        }
        return false;
    }
    function Remove(bytes32Set storage self, bytes32 addr) public returns(bool) {
        if(self._data[addr].exists){
            self._deleted.push(self._data[addr].index);
            delete self._values[self._data[addr].index];
            delete self._data[addr];
            self._size--;
            return true;
        }
        return false;
    }
    function Has(bytes32Set storage self, bytes32 addr) public view returns(bool) {
        return self._data[addr].exists;
    }
    function Contains(bytes32Set storage self, bytes32[] addrs) public view returns(bool) {
        uint count = 0;
        for(uint c = 0; c < addrs.length; i++){
            for(uint i = 0; i < self._values.length; i++){
                if(self._values[i]!=bytes32(0) && self._values[i] == addrs[c]){
                   count++;
                }
            }
        }
        return count == addrs.length;
    }
    function Get(bytes32Set storage self, bytes32 addr) public view returns(bytes32) {
        if(self._data[addr].exists){
            return self._values[self._data[addr].index];
        }
        return bytes32(0);
    }
    function Values(bytes32Set storage self) public view returns(bytes32[]) {
        bytes32[] memory cleaned = new bytes32[](self._size);
        uint count = 0;
        for (uint i = 0; i<self._values.length; i++){
            if(self._values[i]!=bytes32(0)){
                cleaned[count] = (self._values[i]);
                count++;
            }
        }
        return (cleaned);
    }
    function Size(bytes32Set storage self) public view returns(uint) {
        return self._size;
    }
    function Replace(bytes32Set storage self, bytes32 addr, bytes32 addr2) public returns(bool) {
        if(self._data[addr].exists){
            self._data[addr2].exists = self._data[addr].exists;
            self._data[addr2].index = self._data[addr].index;
            self._values[self._data[addr].index] = addr2;
            delete self._data[addr];
            return true;
        }
        return false;
    }
    function Clear(bytes32Set storage self) public returns(bool) {
        for (uint i = 0; i<self._values.length; i++){
            if(self._values[i]!=bytes32(0)){
                delete self._data[self._values[i]];
            }
        }
        self._deleted.length=0;
        self._values.length=0;
        self._size = 0;
        return true;
    }
    struct addressSet {
        bool exists;
        mapping(address => MObject) _data;
        address[] _values;
        uint[] _deleted;
        uint _size;
    }

    function Insert(addressSet storage self, address addr) public returns(bool) {
        MObject storage mo = self._data[addr];
        if(!self._data[addr].exists){
            self.exists = true;
            mo.exists = true;
            if(self._deleted.length>0){
                mo.index = self._deleted[self._deleted.length-1];
                self._values[mo.index]=addr;
                self._deleted.length--;
            }else{
                mo.index = self._values.length;
                self._values.push(addr);
            }
            self._size++;
            return true;
        }
        return false;
    }
    function Remove(addressSet storage self, address addr) public returns(bool) {
        if(self._data[addr].exists){
            self._deleted.push(self._data[addr].index);
            delete self._values[self._data[addr].index];
            delete self._data[addr];
            self._size--;
            return true;
        }
        return false;
    }
    function Has(addressSet storage self, address addr) public view returns(bool) {
        return self._data[addr].exists;
    }
    function Contains(addressSet storage self, address[] addrs) public view returns(bool) {
        uint count = 0;
        for(uint c = 0; c < addrs.length; i++){
            for(uint i = 0; i < self._values.length; i++){
                if(self._values[i]!=address(0) && self._values[i] == addrs[c]){
                   count++;
                }
            }
        }
        return count == addrs.length;
    }
    function Get(addressSet storage self, address addr) public view returns(address) {
        if(self._data[addr].exists){
            return self._values[self._data[addr].index];
        }
        return address(0);
    }
    function Values(addressSet storage self) public view returns(address[]) {
        address[] memory cleaned = new address[](self._size);
        uint count = 0;
        for (uint i = 0; i<self._values.length; i++){
            if(self._values[i]!=address(0)){
                cleaned[count] = (self._values[i]);
                count++;
            }
        }
        return (cleaned);
    }
    function Size(addressSet storage self) public view returns(uint) {
        return self._size;
    }
    function Replace(addressSet storage self, address addr, address addr2) public returns(bool) {
        if(self._data[addr].exists){
            self._data[addr2].exists = self._data[addr].exists;
            self._data[addr2].index = self._data[addr].index;
            self._values[self._data[addr].index] = addr2;
            delete self._data[addr];
            return true;
        }
        return false;
    }
    function Clear(addressSet storage self) public returns(bool) {
        for (uint i = 0; i<self._values.length; i++){
            if(self._values[i]!=address(0)){
                delete self._data[self._values[i]];
            }
        }
        self._deleted.length=0;
        self._values.length=0;
        self._size = 0;
        return true;
    }
    struct addressSetProviderStore {
        bool exists;
        mapping(address => MStorageProvider) _data;
        address[] _values;
        uint[] _deleted;
        uint _size;
    }

    struct MStorageProvider {
        bool exists;
        uint index;
        bytes32 urlPrefix;
        //TODO impl. price management
    }

    function Insert(addressSetProviderStore storage self, address addr) public returns(bool) {
        MStorageProvider storage mo = self._data[addr];
        if(!self._data[addr].exists){
            self.exists = true;
            mo.exists = true;
            if(self._deleted.length>0){
                mo.index = self._deleted[self._deleted.length-1];
                self._values[mo.index]=addr;
                self._deleted.length--;
            }else{
                mo.index = self._values.length;
                self._values.push(addr);
            }
            self._size++;
            return true;
        }
        return false;
    }
    function Remove(addressSetProviderStore storage self, address addr) public returns(bool) {
        if(self._data[addr].exists){
            self._deleted.push(self._data[addr].index);
            delete self._values[self._data[addr].index];
            delete self._data[addr];
            self._size--;
            return true;
        }
        return false;
    }
    function Has(addressSetProviderStore storage self, address addr) public view returns(bool) {
        return self._data[addr].exists;
    }
    function Contains(addressSetProviderStore storage self, address[] addrs) public view returns(bool) {
        uint count = 0;
        for(uint c = 0; c < addrs.length; i++){
            for(uint i = 0; i < self._values.length; i++){
                if(self._values[i]!=address(0) && self._values[i] == addrs[c]){
                   count++;
                }
            }
        }
        return count == addrs.length;
    }
    function Get(addressSetProviderStore storage self, address addr) public view returns(address) {
        if(self._data[addr].exists){
            return self._values[self._data[addr].index];
        }
        return address(0);
    }
    function SetURLPrefix(addressSetProviderStore storage self, address addr, bytes32 urlPrefix) public returns(bool) {
        if(self._data[addr].exists){
            self._data[addr].urlPrefix = urlPrefix;
            return true;
        }
        return false;
    }
    function GetURLPrefix(addressSetProviderStore storage self, address addr) public view returns(bytes32) {
        return self._data[addr].urlPrefix;
    }
    function Values(addressSetProviderStore storage self) public view returns(address[]) {
        address[] memory cleaned = new address[](self._size);
        uint count = 0;
        for (uint i = 0; i<self._values.length; i++){
            if(self._values[i]!=address(0)){
                cleaned[count] = (self._values[i]);
                count++;
            }
        }
        return (cleaned);
    }
    function Size(addressSetProviderStore storage self) public view returns(uint) {
        return self._size;
    }
    function Replace(addressSetProviderStore storage self, address addr, address addr2) public returns(bool) {
        if(self._data[addr].exists){
            self._data[addr2].exists = self._data[addr].exists;
            self._data[addr2].index = self._data[addr].index;
            self._values[self._data[addr].index] = addr2;
            delete self._data[addr];
            return true;
        }
        return false;
    }
    function Clear(addressSetProviderStore storage self) public returns(bool) {
        for (uint i = 0; i<self._values.length; i++){
            if(self._values[i]!=address(0)){
                delete self._data[self._values[i]];
            }
        }
        self._deleted.length=0;
        self._values.length=0;
        self._size = 0;
        return true;
    }
}

/*
usage:

contract Map {
    using VMap for VMap.bytes32Set;

    mapping(address => VMap.bytes32Set) internal fs;

    function Insert(address addr, bytes32 val) public returns(bool) {
        VMap.bytes32Set storage fsf = fs[addr];
        return fsf.Insert(val);
    }
    function Remove(address addr, bytes32 val) public returns(bool) {
        return fs[addr].Remove(val);
    }
    function Has(address addr, bytes32 val) public view returns(bool) {
        return fs[addr].Has(val);
    }
    function Values(address addr) public view returns(bytes32[]) {
        return fs[addr].Values();
    }
}
*/