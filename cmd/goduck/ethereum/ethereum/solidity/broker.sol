// SPDX-License-Identifier: MIT
pragma solidity >=0.5.6;

contract Broker {
    // 白名单内的合约才能调用Broker进行跨链操作
    mapping(address => int64) whiteList;
    address[] contracts;
    address[] admins;

    event throwEvent(uint64 index, address to, address fid, string tid, string func, string args, string callback);
    event LogInterchainData(bool status, string data);
    event LogInterchainStatus(bool status);

    address[] outChains;
    address[] inChains;
    address[] callbackChains;

    mapping(address => uint64) outCounter; // mapping from contract address to out event last index
    mapping(address => mapping(uint64 => uint)) outMessages;
    mapping(address => uint64) inCounter;
    mapping(address => mapping(uint64 => uint)) inMessages;
    mapping(address => uint64) callbackCounter;

    // 权限控制，业务合约需要进行注册
    modifier onlyWhiteList {
        require(whiteList[msg.sender] == 1, "Invoker are not in white list");
        _;
    }

    // 权限控制，只有在合约部署时候定下的管理员才能进行业务合约的审核
    modifier onlyAdmin {
        bool flag = false;
        for (uint i = 0; i < admins.length; i++) {
            if (msg.sender == admins[i]) {flag = true;}
        }
        if (flag) {revert();}
        _;
    }

    function initialize() public {
        for (uint i = 0; i < inChains.length; i++) {
            inCounter[inChains[i]] = 0;
        }
        for (uint i = 0; i < outChains.length; i++) {
            outCounter[outChains[i]] = 0;
        }
        for (uint i = 0; i < callbackChains.length; i++) {
            callbackCounter[callbackChains[i]] = 0;
        }
        for (uint i = 0; i < contracts.length; i++) {
            whiteList[contracts[i]] = 0;
        }
        delete outChains;
        delete inChains;
        delete callbackChains;
    }

    // 0 标识正在审核，1表示审核通过，-1表示拒绝注册
    function register(address addr) public {
        whiteList[addr] = 0;
    }

    function audit(address addr, int64 status) public returns (bool) {
        if (status != - 1 && status != 0 && status != 1) {return false;}
        whiteList[addr] = status;
        // 只有审核通过的合约，才记录下来
        if (status == 1) {
            contracts.push(addr);
        }
        return true;
    }

    function InterchainTransferInvoke(
        address destChainID,
        string memory destAddr,
        string memory args)
    public onlyWhiteList returns (bool) {
        // 发起跨链请求
        return invokeInterchain(destChainID, msg.sender, destAddr, "interchainCharge", args, "interchainConfirm");
    }

    function InterchainDataSwapInvoke(
        address destChainID,
        string memory destAddr,
        string memory key)
    public onlyWhiteList returns (bool) {
        return invokeInterchain(destChainID, msg.sender, destAddr, "interchainGet", key, "interchainSet");
    }

    function invokeInterchain(
        address destChainID,
        address sourceAddr,
        string memory destAddr,
        string memory func,
        string memory args,
        string memory callback)
    internal returns (bool) {
        // 记录各个合约已经进行的跨链合约的序号信息
        outCounter[destChainID]++;
        if (outCounter[destChainID] == 1) {
            outChains.push(destChainID);
        }

        outMessages[destChainID][outCounter[destChainID]] = block.number;

        // 抛出跨链事件，便于Plugin进行监听
        emit throwEvent(outCounter[destChainID], destChainID, sourceAddr, destAddr, func, args, callback);
        return true;
    }

    function interchainGet(address sourceChainID, uint64 index, address destAddr, string memory key) public returns (bool, string memory) {
        DataSwapper dataGetter = DataSwapper(destAddr);
        markInCounter(sourceChainID);
        if (whiteList[destAddr] != 1) {return (false, "");}
        string memory data;
        bool status;
        (status, data) = dataGetter.interchainGet(key);
        emit LogInterchainData(status, data);
        return (status, data);
    }

    function interchainSet(address sourceChainID, uint64 index, address destAddr, string memory key, string memory value) public returns (bool) {
        if (callbackCounter[sourceChainID] + 1 != index) {
            emit LogInterchainStatus(false);
            return false;
        }
        DataSwapper dataSetter = DataSwapper(destAddr);
        markCallbackCounter(sourceChainID, index);
        dataSetter.interchainSet(key, value);
        emit LogInterchainStatus(true);
        return true;
    }

    function interchainCharge(address sourceChainID, uint64 index, address destAddr, string memory sender, string memory receiver, uint64 amount) public returns (bool) {
        // 检查序号是否正确，防止replay attack
        if (inCounter[sourceChainID] + 1 != index) {
            emit LogInterchainStatus(false);
            return false;
        }

        markInCounter(sourceChainID);
        if (whiteList[destAddr] != 1) {
            emit LogInterchainStatus(false);
            return false;
        }

        Transfer exchanger = Transfer(destAddr);
        bool status = exchanger.interchainCharge(sender, receiver, amount);
        emit LogInterchainStatus(status);
        return status;
    }

    function interchainConfirm(address sourceChainID, uint64 index, address destAddr, bool status, string memory sender, uint64 amount) public returns (bool) {
        if (callbackCounter[sourceChainID] + 1 != index) {
            emit LogInterchainStatus(false);
            return false;
        }

        markCallbackCounter(sourceChainID, index);
        if (whiteList[destAddr] != 1) {
            emit LogInterchainStatus(false);
            return false;
        }
        // if status is ok, no need to rollback
        if (status) {
            emit LogInterchainStatus(true);
            return true;
        }

        Transfer exchanger = Transfer(destAddr);
        bool status = exchanger.interchainRollback(sender, amount);
        emit LogInterchainStatus(status);
        return status;
    }

    // 帮助记录Meta信息的辅助函数
    function markCallbackCounter(address from, uint64 index) private {
        if (callbackCounter[from] == 0) {
            callbackChains.push(from);
        }
        callbackCounter[from] = index;
        inMessages[from][callbackCounter[from]] = block.number;
    }

    function markInCounter(address from) private {
        inCounter[from]++;
        if (inCounter[from] == 1) {
            inChains.push(from);
        }

        inMessages[from][inCounter[from]] = block.number;
    }

    // 提供Plugin进行查询的辅助函数
    function getOuterMeta() public view returns (address[] memory, uint64[] memory) {
        uint64[] memory indices = new uint64[](outChains.length);
        for (uint64 i = 0; i < outChains.length; i++) {
            indices[i] = outCounter[outChains[i]];
        }

        return (outChains, indices);
    }

    function getOutMessage(address to, uint64 idx) public view returns (uint) {
        return outMessages[to][idx];
    }

    function getInMessage(address from, uint64 idx) public view returns (uint)  {
        return inMessages[from][idx];
    }

    function getInnerMeta() public view returns (address[] memory, uint64[] memory) {
        uint64[] memory indices = new uint64[](inChains.length);
        for (uint i = 0; i < inChains.length; i++) {
            indices[i] = inCounter[inChains[i]];
        }

        return (inChains, indices);
    }

    function getCallbackMeta() public view returns (address[] memory, uint64[] memory) {
        uint64[] memory indices = new uint64[](callbackChains.length);
        for (uint64 i = 0; i < callbackChains.length; i++) {
            indices[i] = callbackCounter[callbackChains[i]];
        }

        return (callbackChains, indices);
    }

    // some string utils
    function toString(uint _base) internal pure returns (string memory) {
        bytes memory _tmp = new bytes(32);
        uint i;
        for (i = 0; _base > 0; i++) {
            _tmp[i] = byte(uint8((_base % 10) + 48));
            _base /= 10;
        }
        bytes memory _real = new bytes(i--);
        for (uint j = 0; j < _real.length; j++) {
            _real[j] = _tmp[i--];
        }
        return string(_real);
    }

    function split(string memory _base, string memory _delimiter) internal pure returns (string[] memory splitArr) {
        bytes memory _baseBytes = bytes(_base);

        uint _offset = 0;
        uint _splitsCount = 1;
        while (_offset < _baseBytes.length - 1) {
            int _limit = _indexOf(_base, _delimiter, _offset);
            if (_limit == - 1)
                break;
            else {
                _splitsCount++;
                _offset = uint(_limit) + 1;
            }
        }

        splitArr = new string[](_splitsCount);

        _offset = 0;
        _splitsCount = 0;
        while (_offset < _baseBytes.length - 1) {

            int _limit = _indexOf(_base, _delimiter, _offset);
            if (_limit == - 1) {
                _limit = int(_baseBytes.length);
            }

            string memory _tmp = new string(uint(_limit) - _offset);
            bytes memory _tmpBytes = bytes(_tmp);

            uint j = 0;
            for (uint i = _offset; i < uint(_limit); i++) {
                _tmpBytes[j++] = _baseBytes[i];
            }
            _offset = uint(_limit) + 1;
            splitArr[_splitsCount++] = string(_tmpBytes);
        }
        return splitArr;
    }

    function _indexOf(string memory _base, string memory _value, uint _offset) internal pure returns (int) {
        bytes memory _baseBytes = bytes(_base);
        bytes memory _valueBytes = bytes(_value);

        assert(_valueBytes.length == 1);

        for (uint i = _offset; i < _baseBytes.length; i++) {
            if (_baseBytes[i] == _valueBytes[0]) {
                return int(i);
            }
        }

        return - 1;
    }
}

abstract contract Transfer {
    function interchainRollback(string memory sender, uint64 val) virtual public returns (bool);

    function interchainCharge(string memory sender, string memory receiver, uint64 val) virtual  public returns (bool);
}

abstract contract DataSwapper {
    function interchainGet(string memory key) virtual public view returns (bool, string memory);

    function interchainSet(string memory key, string memory value) virtual public;
}