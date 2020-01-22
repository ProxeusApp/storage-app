pragma solidity ^0.4.23;

contract ProviderListMgr {
    mapping(address => SDLinkedList) _data;
    SDLinkedList[] _values;  // keep record of file hash
    uint _size;           // valid file number

    struct StorageProvider {
        address prv;
        bytes32 urlPrefix;
        uint[] prices;
    }

    struct SDLinkedList {
        StorageProvider provider;
        int prev;
        int next;
        bool exists;
    }

    constructor() public {
    }

    function spAdd(address prv, bytes32 urlPrefix, uint[] prices) public returns (bool) {
        SDLinkedList storage item = _data[prv];
        if (item.exists) return false;
        int length = (int)(_values.length);
        if (_size > 0)
        {
           SDLinkedList storage prv_item = _values[(uint)(length - 1)];
           item.prev = length - 1;
           item.next = prv_item.next;
           prv_item.next = length;
        }
        item.provider.prv = prv;
        item.provider.urlPrefix = urlPrefix;
        for (uint i;i<prices.length;i++)
            item.provider.prices[i] = prices[i];
        _values.push(item);
        _size++;
        return true;
    }
    function Remove(address prv) public returns (bool) {
        SDLinkedList storage item = _data[prv];
        if (!item.exists) return false;
        SDLinkedList storage prv_item = _values[(uint)(item.prev)];
        prv_item.next = item.next;
        SDLinkedList storage next_item = _values[(uint)(item.next)];
        next_item.prev = item.prev;
        _size--;
        _data[prv].exists = false;
        return true;
    }
    function spList() public view returns(address[]) {
        address[] memory cleaned = new address[](_values.length);
        uint count = 0;
        for (uint i = 0; i<_values.length; i++){
            cleaned[count] = _values[i].provider.prv;
            count++;
        }
        return (cleaned);
    }
    function spUpdate(address prv, bytes32 urlPrefix) public {
        SDLinkedList storage item = _data[prv];
        item.provider.urlPrefix = urlPrefix;
    }
    function spInfo(address strPrv) public view returns (bytes32, uint[]) {
        return (GetURLPrefix(strPrv), GetPrices(strPrv));
    }
    function GetURLPrefix(address prv) public view returns (bytes32) {
        SDLinkedList storage item = _data[prv];
        if (!item.exists) return bytes32(0);
        return item.provider.urlPrefix;
    }
    function GetPrices(address prv) public view returns (uint[]) {
        SDLinkedList storage item = _data[prv];
        return item.provider.prices;
    }
    function spHas(address prv) public view returns (bool) {
        SDLinkedList storage item = _data[prv];
        return item.exists;
    }
}

contract FileListmgr{
        FDLinkedList[] _values;  // keep record of file hash
        uint _size;           // valid file number
        mapping(uint => uint32) _timetable;
        uint TIME_WINDOW_SIZE = 10;
        uint _timetableSize;

    struct FDLinkedList {
        bytes32 hash;
        uint32 timestamp;
        uint prev;
        uint next;
    }

    constructor() public {
    }

    function fileInsert(bytes32 hash, uint32 timestamp) public returns(uint) {
        FDLinkedList memory item;
        item.hash = hash;
        item.timestamp = timestamp;

        uint length = _values.length;
        if (_size > 0)
        {
            uint counter = length - 1;
            while (_values[counter].timestamp > timestamp) {
                counter--;
            }
            FDLinkedList storage prv_item = _values[counter];
            FDLinkedList storage next_item =  _values[prv_item.next];
            item.prev = counter;
            item.next = prv_item.next;
            prv_item.next = length;
            next_item.prev = length;

        }else {
            _timetable[0] = timestamp;
        }
        _values.push(item);
        _size++;

        _timetable[1] = timestamp;
        if ((_values.length % TIME_WINDOW_SIZE)==0) {
           _timetable[ _values.length / TIME_WINDOW_SIZE] = (uint32)(length);
           _timetableSize++;
        }
        return (uint)(length);
    }
    function fileList(uint32 timestamp, uint amount) public view returns (bytes32[], uint32[]) {
        uint length = (amount > _values.length)?amount:_values.length;
        uint current = _values.length - 1;
        if (timestamp > 0) {
            uint guess_timeslice = 0;  //guess block location in TIME_WINDOW_SIZE
            if (_timetableSize > 1)
            {
                guess_timeslice = (timestamp-_timetable[0]) * (TIME_WINDOW_SIZE*_timetableSize)/((_timetable[1]-_timetable[0])*TIME_WINDOW_SIZE);
                if (!((_timetable[guess_timeslice-TIME_WINDOW_SIZE] < timestamp)&&(_timetable[guess_timeslice] > timestamp))) //guess wrong
                {
                   guess_timeslice = 0;
                   while ((_timetable[guess_timeslice]<timestamp)&&(guess_timeslice<_values.length)) {
                        guess_timeslice+=TIME_WINDOW_SIZE;
                    }
                }
            }
            current = guess_timeslice;  // find the precise one
            uint counter=0;
            while ((counter<_values.length)&&(_values[current].timestamp<timestamp)) {
                current = (uint)(_values[current].next);
                counter ++;
            }
            if (counter>=_values.length) length = 0;
        }

        bytes32[] memory cleaned_hash = new bytes32[](length);
        uint32[] memory cleaned_timestamp = new uint32[](length);

        for (uint i = 0; i<length; i++){
            cleaned_hash[i] = _values[current].hash;
            cleaned_timestamp[i] = _values[current].timestamp;
            current = (timestamp > 0) ? _values[current].next : _values[current].prev;
        }
        return (cleaned_hash, cleaned_timestamp);
    }
    function Remove(uint index) public {
        FDLinkedList storage item = _values[index];
        FDLinkedList storage prv_item = _values[item.prev];
        prv_item.next = item.next;
        FDLinkedList storage next_item = _values[item.next];
        next_item.prev = item.prev;
        item.hash = bytes32(0);
        _size--;
    }
}
