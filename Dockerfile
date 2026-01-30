FROM ubuntu:20.04
# 把编译后的程序打包进镜像
COPY webook /app/webook
# 设定工作目录
WORKDIR /app
# 最佳
ENTRYPOINT ["/app/webook"]