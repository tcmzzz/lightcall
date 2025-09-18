## 设置环境变量

在启动服务之前，需要:

1. 在`docker-compose.yaml` 里设置管理员邮箱环境变量 `INIT_ADMIN_EMAIL`
2. 配置证书. 修改`Caddyfile`, 解析域名至Caddy服务器让其自动获取证书, 或者配置已有的证书
3. `docker-compose.yaml` 中 模拟运营商的配置 仅用于演示, 不需要可以去掉


## 启动服务

使用docker compose启动生产环境服务：

```bash
docker compose up -d
```

## 获取管理员密码

服务启动后，系统会自动生成管理员账户和密码。可以通过以下命令查看：

```bash
docker compose logs | grep "password"
```


你将看到类似以下的输出：
```
Superuser created successfully, username: admin@yourdomain.com, password: generated_password
```

请保存好这个密码，用于首次登录系统。

## 注意事项

1. 请确保设置的INIT_ADMIN_EMAIL环境变量是有效的邮箱地址
2. 生成的密码是随机的，请妥善保管
3. 首次登录后请立即修改默认密码
