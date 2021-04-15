pragma solidity >=0.5.6;

contract Transfer {
    mapping(string => uint64) accountM; // map for accounts
    // change the address of Broker accordingly
    address BrokerAddr = 0x97135d4d2578dd2347FF5382db77553bE50bff3f;
    Broker broker = Broker(BrokerAddr);

    // AccessControl
    modifier onlyBroker {
        require(msg.sender == BrokerAddr, "Invoker are not the Broker");
        _;
    }

    function setBroker(address newBroker) public {
        BrokerAddr = newBroker;
        broker = Broker(newBroker);
    }

    function getBroker() public view returns(address) {
        return BrokerAddr;
    }

    // contract for asset
    function transfer(address destChainID, string memory destAddr, string memory sender, string memory receiver, string memory amount) public {
        uint64 am = uint64(parseInt(amount));
        require(accountM[sender] >= am);
        accountM[sender] -= am;

        // stitch parameters
        string memory args = concat(toSlice(sender), toSlice(","));
        args = concat(toSlice(args), toSlice(receiver));
        args = concat(toSlice(args), toSlice(","));
        args = concat(toSlice(args), toSlice(amount));

        string memory argsRb = concat(toSlice(sender), toSlice(","));
        argsRb = concat(toSlice(argsRb), toSlice(amount));

        broker.emitInterchainEvent(destChainID, destAddr, "interchainCharge,,interchainRollback", args, "", argsRb);
    }

    function interchainRollback(string memory sender, uint64 val) public onlyBroker {
        accountM[sender] += val;
    }

    function interchainCharge(string memory sender, string memory receiver, uint64 val) public onlyBroker returns(bool) {
        accountM[receiver] += val;
        return true;
    }

    function getBalance(string memory id) public returns(uint64) {
        return accountM[id];
    }

    function setBalance(string memory id, uint64 amount) public {
        accountM[id] = amount;
    }

    function parseInt(string memory self) public pure returns (uint _ret) {
        bytes memory _bytesValue = bytes(self);
        uint j = 1;
        for(uint i = _bytesValue.length-1; i >= 0 && i < _bytesValue.length; i--) {
            assert(uint8(_bytesValue[i]) >= 48 && uint8(_bytesValue[i]) <= 57);
            _ret += (uint8(_bytesValue[i]) - 48)*j;
            j*=10;
        }
    }

    struct slice {
        uint _len;
        uint _ptr;
    }

    function toSlice(string memory self) internal pure returns (slice memory) {
        uint ptr;
        assembly {
            ptr := add(self, 0x20)
        }
        return slice(bytes(self).length, ptr);
    }

    function concat(slice memory self, slice memory other) internal pure returns (string memory) {
        string memory ret = new string(self._len + other._len);
        uint retptr;
        assembly { retptr := add(ret, 32) }
        memcpy(retptr, self._ptr, self._len);
        memcpy(retptr + self._len, other._ptr, other._len);
        return ret;
    }

    function memcpy(uint dest, uint src, uint len) private pure {
        // Copy word-length chunks while possible
        for(; len >= 32; len -= 32) {
            assembly {
                mstore(dest, mload(src))
            }
            dest += 32;
            src += 32;
        }

        // Copy remaining bytes
        uint mask = 256 ** (32 - len) - 1;
        assembly {
            let srcpart := and(mload(src), not(mask))
            let destpart := and(mload(dest), mask)
            mstore(dest, or(destpart, srcpart))
        }
    }
}

abstract contract Broker {
    function emitInterchainEvent(
        address destChainID,
        string memory destAddr,
        string memory funcs,
        string memory args,
        string memory argsCb,
        string memory argsRb) virtual public;
}