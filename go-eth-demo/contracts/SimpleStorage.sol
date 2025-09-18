// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

/**
 * @title SimpleStorage
 * @dev 简单的存储合约，演示基本的智能合约功能
 */
contract SimpleStorage {
    // 状态变量
    uint256 private storedData;
    address public owner;
    mapping(address => uint256) public userValues;
    
    // 事件
    event DataStored(address indexed user, uint256 indexed value, uint256 timestamp);
    event OwnershipTransferred(address indexed previousOwner, address indexed newOwner);
    
    // 修饰符
    modifier onlyOwner() {
        require(msg.sender == owner, "Only owner can call this function");
        _;
    }
    
    // 构造函数
    constructor(uint256 _initialValue) {
        storedData = _initialValue;
        owner = msg.sender;
        emit DataStored(msg.sender, _initialValue, block.timestamp);
    }
    
    /**
     * @dev 存储一个值
     * @param _value 要存储的值
     */
    function store(uint256 _value) public {
        storedData = _value;
        userValues[msg.sender] = _value;
        emit DataStored(msg.sender, _value, block.timestamp);
    }
    
    /**
     * @dev 获取存储的值
     * @return 当前存储的值
     */
    function retrieve() public view returns (uint256) {
        return storedData;
    }
    
    /**
     * @dev 获取用户存储的值
     * @param _user 用户地址
     * @return 用户存储的值
     */
    function getUserValue(address _user) public view returns (uint256) {
        return userValues[_user];
    }
    
    /**
     * @dev 增加存储的值
     * @param _amount 要增加的数量
     */
    function increment(uint256 _amount) public {
        storedData += _amount;
        userValues[msg.sender] = storedData;
        emit DataStored(msg.sender, storedData, block.timestamp);
    }
    
    /**
     * @dev 转移合约所有权
     * @param _newOwner 新所有者地址
     */
    function transferOwnership(address _newOwner) public onlyOwner {
        require(_newOwner != address(0), "New owner cannot be zero address");
        address previousOwner = owner;
        owner = _newOwner;
        emit OwnershipTransferred(previousOwner, _newOwner);
    }
    
    /**
     * @dev 重置存储值（仅所有者）
     */
    function reset() public onlyOwner {
        storedData = 0;
        emit DataStored(msg.sender, 0, block.timestamp);
    }
    
    /**
     * @dev 获取合约信息
     */
    function getContractInfo() public view returns (
        uint256 currentValue,
        address contractOwner,
        uint256 userValue,
        uint256 blockNumber
    ) {
        return (
            storedData,
            owner,
            userValues[msg.sender],
            block.number
        );
    }
}