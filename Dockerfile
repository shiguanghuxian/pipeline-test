FROM alpine:latest
# 解决go 时区和https请求证书错误问题
RUN  apk update \
  && apk add ca-certificates \
  && update-ca-certificates \
  && apk add tzdata
COPY ["./bin", "/app/"]
EXPOSE 11260
WORKDIR /app
CMD ["./pipeline-test"]
