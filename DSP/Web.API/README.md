# Advertiser Management API

基于 Go 标准库的广告配置 CRUD API，包含登入登出 session API。当前数据存储在内存中，服务重启后数据会清空。

## MVC 架构

- `controller`：处理 HTTP request/response 与路由参数。
- `service`：处理认证、session 与广告业务规则；`AuthService.Middleware` 是 Auth Filter。
- `repository`：资料沟通层，提供 memory 与 MySQL login 实作；后续可替换为 DB 或 Redis 实作而不影响 Controller。
- `model`：API request、response 与领域资料结构。

## 启动

```powershell
go run .
```

默认监听 `http://localhost:8080`，健康检查：

```powershell
curl http://localhost:8080/healthz
```

## 接口

| 方法   | 路径                             | 行为                                  |
| ------ | -------------------------------- | ------------------------------------- |
| POST   | `/api/v1/membership/login`       | 使用帐号密码登入并建立 session cookie |
| POST   | `/api/v1/membership/logout`      | 注销当前 session 并清除 cookie        |
| POST   | `/api/v1/advertiser/offers`      | 创建广告配置                          |
| GET    | `/api/v1/advertiser/offers/{id}` | 查询广告配置                          |
| PATCH  | `/api/v1/advertiser/offers/{id}` | 局部更新配置                          |
| DELETE | `/api/v1/advertiser/offers/{id}` | 归档广告，不物理删除                  |

广告 CRUD API 需要先登入；`/healthz`、membership 登入与登出 API 不需要登入。

创建示例：

```powershell
curl -X POST http://localhost:8080/api/v1/advertiser/offers `
  -H "Content-Type: application/json" `
  -d '{"name":"Ecommerce App CPI Campaign","advertiser_id":"ADV_202609_001","status":"active","payout":{"type":"CPI","amount":2.5,"currency":"USD"},"targeting":{"countries":["US","CA"],"os":"android","min_os_version":"10.0"},"caps":{"daily_cap":1000},"landing_page_url":"https://store.example.com/app","tracking_url_template":"https://trk.example.com/click?offer_id={offer_id}&click_id={click_id}"}'
```

所有业务响应使用 `{ "code": 0, "message": "...", "data": ... }` 格式；错误响应使用 HTTP 错误状态码和相同外层格式。

登入默认使用 memory repository，凭证是 `admin` / `admin123`，可通过 `API_USERNAME` 与 `API_PASSWORD` 环境变量覆盖。

若设置 `MYSQL_DSN`，登入会改由 MySQL `LoginRepository` 检查资料库。预设读取 `users` 表的 `username` 与 `password` 栏位，也可用 `MYSQL_LOGIN_TABLE` 指定表名。密码目前按照资料表中的原值比对，正式环境建议改为 hash 验证。

```sql
CREATE TABLE users (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  username VARCHAR(100) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL
);
```

PowerShell 环境设定示例：

```powershell
$env:MYSQL_DSN = "app_user:app_password@tcp(127.0.0.1:3306)/advertiser?parseTime=true"
$env:MYSQL_LOGIN_TABLE = "users"
go run .
```

登入示例：

```powershell
curl -i -c cookies.txt -X POST http://localhost:8080/api/v1/membership/login `
  -H "Content-Type: application/json" `
  -d '{"username":"admin","password":"admin123"}'
curl -i -b cookies.txt -X POST http://localhost:8080/api/v1/membership/logout
```
