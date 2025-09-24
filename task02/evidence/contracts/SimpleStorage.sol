// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

/**
 * @title SimpleStorage
 * @dev 简单的存储合约示例
 * 用于演示基本的智能合约功能
 */
contract SimpleStorage {
    // 存储的数据
    uint256 private storedData;
    
    // 合约所有者
    address public owner;
    
    // 事件定义
    event DataStored(uint256 indexed newValue, address indexed setter, uint256 timestamp);
    event OwnershipTransferred(address indexed previousOwner, address indexed newOwner);
    
    // 修饰符：只有所有者可以执行
    modifier onlyOwner() {
        require(msg.sender == owner, "Only owner can call this function");
        _;
    }
    
    /**
     * @dev 构造函数
     * 设置合约部署者为所有者，初始值为0
     */
    constructor() {
        owner = msg.sender;
        storedData = 0;
        emit DataStored(0, msg.sender, block.timestamp);
    }
    
    /**
     * @dev 设置存储的数据
     * @param x 要存储的新值
     */
    function set(uint256 x) public {
        storedData = x;
        emit DataStored(x, msg.sender, block.timestamp);
    }
    
    /**
     * @dev 获取存储的数据
     * @return 当前存储的值
     */
    function get() public view returns (uint256) {
        return storedData;
    }
    
    /**
     * @dev 获取合约所有者地址
     * @return 所有者地址
     */
    function getOwner() public view returns (address) {
        return owner;
    }
    
    /**
     * @dev 增加存储的数据
     * @param increment 要增加的值
     */
    function increment(uint256 increment) public {
        storedData += increment;
        emit DataStored(storedData, msg.sender, block.timestamp);
    }
    
    /**
     * @dev 减少存储的数据
     * @param decrement 要减少的值
     */
    function decrement(uint256 decrement) public {
        require(storedData >= decrement, "Cannot decrement below zero");
        storedData -= decrement;
        emit DataStored(storedData, msg.sender, block.timestamp);
    }
    
    /**
     * @dev 重置存储的数据为0（仅所有者）
     */
    function reset() public onlyOwner {
        storedData = 0;
        emit DataStored(0, msg.sender, block.timestamp);
    }
    
    /**
     * @dev 转移合约所有权（仅所有者）
     * @param newOwner 新所有者地址
     */
    function transferOwnership(address newOwner) public onlyOwner {
        require(newOwner != address(0), "New owner cannot be zero address");
        address previousOwner = owner;
        owner = newOwner;
        emit OwnershipTransferred(previousOwner, newOwner);
    }
    
    /**
     * @dev 获取合约的基本信息
     * @return value 当前存储值
     * @return contractOwner 合约所有者
     * @return blockNumber 当前区块号
     * @return blockTimestamp 当前区块时间戳
     */
    function getInfo() public view returns (
        uint256 value,
        address contractOwner,
        uint256 blockNumber,
        uint256 blockTimestamp
    ) {
        return (
            storedData,
            owner,
            block.number,
            block.timestamp
        );
    }
}