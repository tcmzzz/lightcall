## 快速开始

### 1. 配置环境变量

编辑 `.env` 文件，设置您的域名和管理员邮箱：

```bash
DOMAIN=your.domain.com
INIT_ADMIN_EMAIL=admin@yourdomain.com
```

### 2. 启动服务


**演示环境启动（包含模拟运营商）：**
```bash
docker compose --profile demo up -d
```

**生产环境启动(需要调整域名和证书)：**
```bash
docker compose up -d
```

## 获取管理员密码

服务启动后，系统会自动生成管理员账户和密码。可以通过以下命令查看：

```bash
docker compose logs backend| grep "password"
```


你将看到类似以下的输出：
```
Superuser created successfully, username: admin@yourdomain.com, password: generated_password
```

请保存好这个密码，用于首次登录系统。

## 环境说明


### 演示环境
- 包含所有生产环境服务
- 额外包含模拟运营商服务（用于演示外呼功能）
- 适用于功能演示和测试
- 使用自签名证书（浏览器会提示"不安全"，点击继续访问即可）

### 生产环境
- 需要准备好域名和证书并配置
- 去掉compose文件中`demo`的相关内容

## 证书说明

- 系统使用自签名证书，无需手动配置
- 首次访问时浏览器会显示"不安全"警告，点击"继续访问"即可
- 生产环境建议配置 Let's Encrypt 证书（修改 Caddyfile 中的 `tls internal` 为 `tls` 即可）

## 常见问题

1. **无法访问网站**：检查防火墙设置和域名解析
2. **证书警告**：点击"继续访问"，这是正常的自签名证书提示
3. **服务启动失败**：检查 `.env` 文件配置是否正确
