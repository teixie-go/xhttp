# 初始化
```
var cli xhttp.Client
cli = xhttp.NewClient(http.Client{
    Timeout: 500 * time.Millisecond,
})
```

# 获得原始返回
```
result, err := cli.Get("url").Result()
fmt.Printf("%v,%v\n", string(result), err)
```

# 返回内容绑定对象
```
obj := &struct{}
err := cli.Get("url").Bind(obj)
```
