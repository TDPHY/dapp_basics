// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

/**
 * @title Token
 * @dev ERC20兼容的代币合约示例
 * 演示完整的代币功能实现
 */
contract Token {
    // 代币基本信息
    string public name;
    string public symbol;
    uint8 public decimals;
    uint256 public totalSupply;
    
    // 账户余额映射
    mapping(address => uint256) public balanceOf;
    
    // 授权额度映射 (owner => spender => amount)
    mapping(address => mapping(address => uint256)) public allowance;
    
    // 合约所有者
    address public owner;
    
    // 是否暂停转账
    bool public paused;
    
    // 黑名单
    mapping(address => bool) public blacklist;
    
    // 事件定义
    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);
    event Mint(address indexed to, uint256 value);
    event Burn(address indexed from, uint256 value);
    event Pause();
    event Unpause();
    event BlacklistUpdated(address indexed account, bool isBlacklisted);
    event OwnershipTransferred(address indexed previousOwner, address indexed newOwner);
    
    // 修饰符
    modifier onlyOwner() {
        require(msg.sender == owner, "Only owner can call this function");
        _;
    }
    
    modifier whenNotPaused() {
        require(!paused, "Contract is paused");
        _;
    }
    
    modifier notBlacklisted(address account) {
        require(!blacklist[account], "Account is blacklisted");
        _;
    }
    
    /**
     * @dev 构造函数
     * @param _name 代币名称
     * @param _symbol 代币符号
     * @param _decimals 小数位数
     * @param _totalSupply 总供应量
     */
    constructor(
        string memory _name,
        string memory _symbol,
        uint8 _decimals,
        uint256 _totalSupply
    ) {
        name = _name;
        symbol = _symbol;
        decimals = _decimals;
        totalSupply = _totalSupply * 10**_decimals;
        
        owner = msg.sender;
        balanceOf[msg.sender] = totalSupply;
        paused = false;
        
        emit Transfer(address(0), msg.sender, totalSupply);
    }
    
    /**
     * @dev 转账功能
     * @param to 接收者地址
     * @param value 转账金额
     * @return 是否成功
     */
    function transfer(address to, uint256 value) 
        public 
        whenNotPaused 
        notBlacklisted(msg.sender) 
        notBlacklisted(to) 
        returns (bool) 
    {
        require(to != address(0), "Cannot transfer to zero address");
        require(balanceOf[msg.sender] >= value, "Insufficient balance");
        
        balanceOf[msg.sender] -= value;
        balanceOf[to] += value;
        
        emit Transfer(msg.sender, to, value);
        return true;
    }
    
    /**
     * @dev 授权功能
     * @param spender 被授权者地址
     * @param value 授权金额
     * @return 是否成功
     */
    function approve(address spender, uint256 value) 
        public 
        whenNotPaused 
        notBlacklisted(msg.sender) 
        notBlacklisted(spender) 
        returns (bool) 
    {
        require(spender != address(0), "Cannot approve zero address");
        
        allowance[msg.sender][spender] = value;
        
        emit Approval(msg.sender, spender, value);
        return true;
    }
    
    /**
     * @dev 代理转账功能
     * @param from 发送者地址
     * @param to 接收者地址
     * @param value 转账金额
     * @return 是否成功
     */
    function transferFrom(address from, address to, uint256 value) 
        public 
        whenNotPaused 
        notBlacklisted(msg.sender) 
        notBlacklisted(from) 
        notBlacklisted(to) 
        returns (bool) 
    {
        require(to != address(0), "Cannot transfer to zero address");
        require(balanceOf[from] >= value, "Insufficient balance");
        require(allowance[from][msg.sender] >= value, "Insufficient allowance");
        
        balanceOf[from] -= value;
        balanceOf[to] += value;
        allowance[from][msg.sender] -= value;
        
        emit Transfer(from, to, value);
        return true;
    }
    
    /**
     * @dev 增加授权额度
     * @param spender 被授权者地址
     * @param addedValue 增加的金额
     * @return 是否成功
     */
    function increaseAllowance(address spender, uint256 addedValue) 
        public 
        whenNotPaused 
        returns (bool) 
    {
        approve(spender, allowance[msg.sender][spender] + addedValue);
        return true;
    }
    
    /**
     * @dev 减少授权额度
     * @param spender 被授权者地址
     * @param subtractedValue 减少的金额
     * @return 是否成功
     */
    function decreaseAllowance(address spender, uint256 subtractedValue) 
        public 
        whenNotPaused 
        returns (bool) 
    {
        uint256 currentAllowance = allowance[msg.sender][spender];
        require(currentAllowance >= subtractedValue, "Decreased allowance below zero");
        
        approve(spender, currentAllowance - subtractedValue);
        return true;
    }
    
    /**
     * @dev 铸造代币（仅所有者）
     * @param to 接收者地址
     * @param value 铸造数量
     */
    function mint(address to, uint256 value) 
        public 
        onlyOwner 
        whenNotPaused 
        notBlacklisted(to) 
    {
        require(to != address(0), "Cannot mint to zero address");
        
        totalSupply += value;
        balanceOf[to] += value;
        
        emit Transfer(address(0), to, value);
        emit Mint(to, value);
    }
    
    /**
     * @dev 销毁代币
     * @param value 销毁数量
     */
    function burn(uint256 value) 
        public 
        whenNotPaused 
        notBlacklisted(msg.sender) 
    {
        require(balanceOf[msg.sender] >= value, "Insufficient balance to burn");
        
        balanceOf[msg.sender] -= value;
        totalSupply -= value;
        
        emit Transfer(msg.sender, address(0), value);
        emit Burn(msg.sender, value);
    }
    
    /**
     * @dev 暂停合约（仅所有者）
     */
    function pause() public onlyOwner {
        require(!paused, "Contract is already paused");
        paused = true;
        emit Pause();
    }
    
    /**
     * @dev 恢复合约（仅所有者）
     */
    function unpause() public onlyOwner {
        require(paused, "Contract is not paused");
        paused = false;
        emit Unpause();
    }
    
    /**
     * @dev 更新黑名单（仅所有者）
     * @param account 账户地址
     * @param isBlacklisted 是否加入黑名单
     */
    function updateBlacklist(address account, bool isBlacklisted) 
        public 
        onlyOwner 
    {
        require(account != owner, "Cannot blacklist owner");
        blacklist[account] = isBlacklisted;
        emit BlacklistUpdated(account, isBlacklisted);
    }
    
    /**
     * @dev 批量转账（仅所有者）
     * @param recipients 接收者地址数组
     * @param values 转账金额数组
     */
    function batchTransfer(address[] memory recipients, uint256[] memory values) 
        public 
        onlyOwner 
        whenNotPaused 
    {
        require(recipients.length == values.length, "Arrays length mismatch");
        require(recipients.length <= 100, "Too many recipients");
        
        for (uint256 i = 0; i < recipients.length; i++) {
            require(recipients[i] != address(0), "Cannot transfer to zero address");
            require(!blacklist[recipients[i]], "Recipient is blacklisted");
            require(balanceOf[msg.sender] >= values[i], "Insufficient balance");
            
            balanceOf[msg.sender] -= values[i];
            balanceOf[recipients[i]] += values[i];
            
            emit Transfer(msg.sender, recipients[i], values[i]);
        }
    }
    
    /**
     * @dev 转移所有权
     * @param newOwner 新所有者地址
     */
    function transferOwnership(address newOwner) public onlyOwner {
        require(newOwner != address(0), "New owner cannot be zero address");
        require(!blacklist[newOwner], "New owner cannot be blacklisted");
        
        address previousOwner = owner;
        owner = newOwner;
        
        emit OwnershipTransferred(previousOwner, newOwner);
    }
    
    /**
     * @dev 获取代币信息
     * @return tokenName 代币名称
     * @return tokenSymbol 代币符号
     * @return tokenDecimals 小数位数
     * @return tokenTotalSupply 总供应量
     * @return contractOwner 合约所有者
     * @return isPaused 是否暂停
     */
    function getTokenInfo() public view returns (
        string memory tokenName,
        string memory tokenSymbol,
        uint8 tokenDecimals,
        uint256 tokenTotalSupply,
        address contractOwner,
        bool isPaused
    ) {
        return (
            name,
            symbol,
            decimals,
            totalSupply,
            owner,
            paused
        );
    }
}