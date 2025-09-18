// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

/**
 * @title Counter
 * @dev 一个简单的计数器智能合约
 */
contract Counter {
    // 存储计数器的值
    uint256 private count;
    
    // 合约所有者
    address public owner;
    
    // 事件：当计数器值改变时触发
    event CountChanged(uint256 oldValue, uint256 newValue, address indexed changer);
    
    // 事件：当计数器重置时触发
    event CountReset(address indexed resetter);
    
    /**
     * @dev 构造函数，初始化计数器
     * @param _initialValue 初始计数值
     */
    constructor(uint256 _initialValue) {
        count = _initialValue;
        owner = msg.sender;
        emit CountChanged(0, _initialValue, msg.sender);
    }
    
    /**
     * @dev 获取当前计数器的值
     * @return 当前计数值
     */
    function getCount() public view returns (uint256) {
        return count;
    }
    
    /**
     * @dev 增加计数器的值
     */
    function increment() public {
        uint256 oldValue = count;
        count += 1;
        emit CountChanged(oldValue, count, msg.sender);
    }
    
    /**
     * @dev 减少计数器的值
     */
    function decrement() public {
        require(count > 0, "Counter: cannot decrement below zero");
        uint256 oldValue = count;
        count -= 1;
        emit CountChanged(oldValue, count, msg.sender);
    }
    
    /**
     * @dev 增加指定数量到计数器
     * @param _value 要增加的数量
     */
    function add(uint256 _value) public {
        uint256 oldValue = count;
        count += _value;
        emit CountChanged(oldValue, count, msg.sender);
    }
    
    /**
     * @dev 从计数器减去指定数量
     * @param _value 要减去的数量
     */
    function subtract(uint256 _value) public {
        require(count >= _value, "Counter: insufficient count to subtract");
        uint256 oldValue = count;
        count -= _value;
        emit CountChanged(oldValue, count, msg.sender);
    }
    
    /**
     * @dev 重置计数器为0 (仅所有者可调用)
     */
    function reset() public {
        require(msg.sender == owner, "Counter: only owner can reset");
        count = 0;
        emit CountReset(msg.sender);
        emit CountChanged(count, 0, msg.sender);
    }
    
    /**
     * @dev 设置计数器为指定值 (仅所有者可调用)
     * @param _value 新的计数值
     */
    function setCount(uint256 _value) public {
        require(msg.sender == owner, "Counter: only owner can set count");
        uint256 oldValue = count;
        count = _value;
        emit CountChanged(oldValue, count, msg.sender);
    }
    
    /**
     * @dev 获取合约的基本信息
     * @return _count 当前计数值
     * @return _owner 合约所有者地址
     */
    function getInfo() public view returns (uint256 _count, address _owner) {
        return (count, owner);
    }
}