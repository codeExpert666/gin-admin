# 基础镜像:使用微软提供的Go开发容器镜像
# 版本为1.22,基于Debian bookworm版本
FROM mcr.microsoft.com/devcontainers/go:1.22-bookworm

# 定义构建参数APP,设置默认值为ginadmin
# 这个参数可以在构建时通过--build-arg修改
ARG APP=ginadmin

# 设置CGO_CFLAGS环境变量以启用大文件支持
# 这允许程序处理大于2GB的文件
ENV CGO_CFLAGS "-D_LARGEFILE64_SOURCE"

# 安装开发所需的工具:
# 1. wire: Google的依赖注入工具,用于自动生成依赖注入代码
# 2. swag: 用于自动生成Swagger API文档的工具
# 3. gin-admin-cli: Gin Admin框架的命令行工具
# 最后将/go目录的所有权更改为vscode用户,确保开发环境权限正确
RUN go install github.com/google/wire/cmd/wire@latest \
    && go install github.com/swaggo/swag/cmd/swag@latest \
    && go install github.com/gin-admin/gin-admin-cli/v10@latest \
    && chown -R vscode /go



