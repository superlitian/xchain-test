### xchain-test

> **开源版本测试系统**

### 账号
> account.yaml
1. 拥有一个管理员账号， 负责普通账户utxo足够，以及部署合约。
2. 支持普通账户可配置数量，生成之后保存，需要人为销毁

### 节点
> xchain.yaml
1. 支持单节点和多节点
2. 支持配置是否给管理员账户转账
3. 支持配置合约账户
4. 支持配置是否创建合约账户

### 合约
> contract.yaml
1. 支持单合约测试
2. 支持配置此次测试的合约类型
3. 支持配置此次测试的合约名称
4. 支持配置此次测试的合约初始化参数
5. 支持配置此次测试的合约方法
6. 支持配置此次测试的合约方法调用参数
7. 支持配置自增、时间戳等类型参数

### 程序配置
> service.yaml
1. 支持配置程序运行时间，超时退出
2. 支持配置秒级并发数

