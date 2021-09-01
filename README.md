## 开始使用
### 1.使用默认配置 
```
// Get
result, err := xhttp.Get("url").Result()

// Post
result, err := xhttp.Post("url", strings.NewReader("body")).Result()

// PostForm
result, err := xhttp.PostForm("url", strings.NewReader("body")).Result()

// PostJSON
result, err := xhttp.PostJSON("url", strings.NewReader("body")).Result()
```
### 2.使用自定义配置
```
var cli xhttp.Client
cli = xhttp.NewClient(http.Client{
    Timeout: 500 * time.Millisecond,
})
```

## 接收返回结果
```
// 获得原始返回
result, err := cli.Get("url").Result()
fmt.Printf("%v,%v\n", string(result), err)

// 绑定对象，支持json和xml，通过response Content-Type判断，保底使用json解析
obj := &struct{}
err := cli.Get("url").Bind(obj)

// 指定json解析
err = cli.Get("url").BindJSON(obj)

// 指定xml解析
err = cli.Get("url").BindXML(obj)

// 自定义解析器，解析器只需实现binding.Binding接口
type xxxBinding struct{}
func (xxxBinding) Name() {
    return "xxx"
}
func (xxxBinding) Bind([]byte, obj interface{}) error{
    // 解析代码
    return nil
}
err = cli.Get("url").BindWith(obj, xxxBinding{})
```

## 监听请求完成事件
```
// 可用于日志打印、第三方监控平台上报等操作
func httpLog(method, url string, body io.Reader, resp *xhttp.Response) {
    // 打印日志
    xlog.Infof("http_request: method=%v url=%v duration=%v\n", method, url, resp.Duration)
    
    // 上报promethus
}
xhttp.Listen(httpLog)
```

## 测试用例请求
```
import (
	"github.com/gin-gonic/gin"
	"github.com/teixie-go/xhttp/httptest"
)

func Serve() httptest.Client {
	// gin
	e := gin.Default()
	routes.Setup(e) // 设置路由
	return httptest.NewClient(e)
	
	// beego
	// return httptest.NewClient(beego.BeeApp.Handlers)
}

func TestCreateXXXOk(t *testing.T) {
	resp := Serve().Post("/xxx/create", strings.NewReader(`{
		"name": "test",
	}`))
	assert.Equal(t, http.StatusOK, resp.RawResponse.StatusCode)
	
	result := &struct {
	    Code int    `json:"code"`
	    Msg  string `json:"msg"`
   	}{}
   	_ = resp.Bind(result)
   	assert.Equal(t, result.Code, 200)
   	assert.Equal(t, result.Msg, "success")
}
```