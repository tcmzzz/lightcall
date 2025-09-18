# 生产环境部署指南

## 设置环境变量

在启动服务之前，需要设置管理员邮箱环境变量：

```bash
export INIT_ADMIN_EMAIL=admin@yourdomain.com
```

## 启动服务

使用docker compose启动生产环境服务：

```bash
docker compose up -d
```

## 获取管理员密码

服务启动后，系统会自动生成管理员账户和密码。可以通过以下命令查看：

```bash
docker compose logs | grep "Superuser created successfully"
```

或者：

```bash
docker compose logs backend | grep "Superuser created successfully"
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