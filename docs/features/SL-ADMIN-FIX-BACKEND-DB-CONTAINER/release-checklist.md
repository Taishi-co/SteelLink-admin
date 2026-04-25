# Backend Container MySQL Config Guard Release Checklist

## 发布前检查

- [ ] `DB_HOST` 已在生产后端容器运行命令中显式注入。
- [ ] `DB_PORT`、`DB_USER`、`DB_PASSWORD`、`DB_NAME` 已在生产后端容器运行命令中显式注入。
- [ ] 生产服务器部署脚本没有依赖容器内默认 `localhost`。
- [ ] 后端镜像已重新构建并使用新镜像重启容器。

## 验证命令

```bash
./harness/branch-guard.sh local
./harness/verify.sh
./harness/review.sh docs/features/SL-ADMIN-FIX-BACKEND-DB-CONTAINER
./harness/release.sh docs/features/SL-ADMIN-FIX-BACKEND-DB-CONTAINER
```

## 人工验证

1. 在生产服务器执行后端部署脚本前，确认容器启动命令包含 `-e DB_HOST=...` 或 `--env-file ...`。
2. 部署后执行 `docker logs -f steellink-admin-backend`，确认不再出现连接 `[::1]:3306`。
3. 访问后端健康检查或管理端登录页，确认后端可连通数据库。

## 回滚

1. 回滚本次代码变更并重新构建后端镜像。
2. 或在部署脚本中立即补齐 `DB_HOST` 后重新启动当前镜像。
