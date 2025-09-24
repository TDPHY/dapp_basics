// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

/**
 * @title Counter
 * @dev 计数器合约示例
 * 演示状态变更和事件日志功能
 */
contract Counter {
    // 计数器值
    uint256 private count;
    
    // 合约所有者
    address public owner;
    
    // 计数器历史记录
    struct CountRecord {
        uint256 value;
        address changer;
        uint256 timestamp;
        string operation;
    }
    
    // 存储历史记录
    CountRecord[] public history;
    
    // 事件定义
    event CountChanged(
        uint256 indexed newCount, 
        address indexed changer, 
        string operation,
        uint256 timestamp
    );
    
    event OwnershipTransferred(
        address indexed previousOwner, 
        address indexed newOwner
    );
    
    // 修饰符
    modifier onlyOwner() {
        require(msg.sender == owner, "Only owner can call this function");
        _;
    }
    
    modifier validOperation(string memory operation) {
        require(
            keccak256(bytes(operation)) == keccak256(bytes("increment")) ||
            keccak256(bytes(operation)) == keccak256(bytes("decrement")) ||
            keccak256(bytes(operation)) == keccak256(bytes("reset")) ||
            keccak256(bytes(operation)) == keccak256(bytes("set")),
            "Invalid operation"
        );
        _;
    }
    
    /**
     * @dev 构造函数
     * 初始化计数器为0，设置部署者为所有者
     */
    constructor() {
        owner = msg.sender;
        count = 0;
        
        // 记录初始状态
        _recordChange(0, "initialize");
    }
    
    /**
     * @dev 增加计数器
     */
    function increment() public {
        count += 1;
        _recordChange(count, "increment");
    }
    
    /**
     * @dev 减少计数器
     */
    function decrement() public {
        require(count > 0, "Counter cannot be negative");
        count -= 1;
        _recordChange(count, "decrement");
    }
    
    /**
     * @dev 获取当前计数值
     * @return 当前计数器值
     */
    function getCount() public view returns (uint256) {
        return count;
    }
    
    /**
     * @dev 重置计数器（仅所有者）
     */
    function reset() public onlyOwner {
        count = 0;
        _recordChange(count, "reset");
    }
    
    /**
     * @dev 设置计数器为指定值（仅所有者）
     * @param newCount 新的计数值
     */
    function setCount(uint256 newCount) public onlyOwner {
        count = newCount;
        _recordChange(count, "set");
    }
    
    /**
     * @dev 批量增加计数器
     * @param times 增加次数
     */
    function incrementBatch(uint256 times) public {
        require(times > 0 && times <= 100, "Times must be between 1 and 100");
        
        for (uint256 i = 0; i < times; i++) {
            count += 1;
        }
        
        _recordChange(count, "increment_batch");
    }
    
    /**
     * @dev 获取历史记录数量
     * @return 历史记录总数
     */
    function getHistoryLength() public view returns (uint256) {
        return history.length;
    }
    
    /**
     * @dev 获取指定索引的历史记录
     * @param index 历史记录索引
     * @return value 计数值
     * @return changer 操作者地址
     * @return timestamp 操作时间戳
     * @return operation 操作类型
     */
    function getHistoryRecord(uint256 index) public view returns (
        uint256 value,
        address changer,
        uint256 timestamp,
        string memory operation
    ) {
        require(index < history.length, "Index out of bounds");
        
        CountRecord memory record = history[index];
        return (
            record.value,
            record.changer,
            record.timestamp,
            record.operation
        );
    }
    
    /**
     * @dev 获取最近N条历史记录
     * @param n 记录数量
     * @return values 计数值数组
     * @return changers 操作者地址数组
     * @return timestamps 时间戳数组
     */
    function getRecentHistory(uint256 n) public view returns (
        uint256[] memory values,
        address[] memory changers,
        uint256[] memory timestamps
    ) {
        require(n > 0, "N must be greater than 0");
        
        uint256 length = history.length;
        uint256 returnLength = n > length ? length : n;
        
        values = new uint256[](returnLength);
        changers = new address[](returnLength);
        timestamps = new uint256[](returnLength);
        
        for (uint256 i = 0; i < returnLength; i++) {
            uint256 index = length - returnLength + i;
            values[i] = history[index].value;
            changers[i] = history[index].changer;
            timestamps[i] = history[index].timestamp;
        }
        
        return (values, changers, timestamps);
    }
    
    /**
     * @dev 转移所有权
     * @param newOwner 新所有者地址
     */
    function transferOwnership(address newOwner) public onlyOwner {
        require(newOwner != address(0), "New owner cannot be zero address");
        
        address previousOwner = owner;
        owner = newOwner;
        
        emit OwnershipTransferred(previousOwner, newOwner);
    }
    
    /**
     * @dev 获取合约统计信息
     * @return currentCount 当前计数值
     * @return totalOperations 总操作次数
     * @return contractOwner 合约所有者
     * @return deploymentTime 部署时间（第一条记录的时间戳）
     */
    function getStats() public view returns (
        uint256 currentCount,
        uint256 totalOperations,
        address contractOwner,
        uint256 deploymentTime
    ) {
        return (
            count,
            history.length,
            owner,
            history.length > 0 ? history[0].timestamp : 0
        );
    }
    
    /**
     * @dev 内部函数：记录状态变更
     * @param newValue 新的计数值
     * @param operation 操作类型
     */
    function _recordChange(uint256 newValue, string memory operation) 
        internal 
        validOperation(operation) 
    {
        history.push(CountRecord({
            value: newValue,
            changer: msg.sender,
            timestamp: block.timestamp,
            operation: operation
        }));
        
        emit CountChanged(newValue, msg.sender, operation, block.timestamp);
    }
}