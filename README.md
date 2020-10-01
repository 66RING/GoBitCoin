# An Bitcoin Project Written with Go

## 技术栈

- 序列化/反序列化
    * Base58


## 使用的API

### flag包

```go
// 注册flag解析器
flagcmd := flag.NewFlagSet("argName", flag.ExitOnError)

// 绑定参数的key-value， 可以多个
sendfrom := flagcmd.String("from", "", "from who") // key value desc
sendto := flagcmd.String("to", "", "to who")
sendamount := flagcmd.Int("amount", 0, "amount")
sendmine := flagcmd.Bool("mine", false, "mine now?")

// 解析后就可使用上面绑定的数据了
flagcmd.Parse() bool {}

SomeFunc(*sendfrom, *sendto, *sendamount)
```
