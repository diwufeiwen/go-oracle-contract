## go-oracle-contract

本文档为不可注册计算方式下的实现，计算调用本地服务。

允许用户自定义计算方式修改：

- 监听事件修改
```shell
type ComputationRequestedEvent struct {
    ID        big.Int
	Requester common.Address
	Input     *big.Int
	LogicID   *big.Int
}
```

- 修改计算逻辑，调用客户执行的`Compute Service`
