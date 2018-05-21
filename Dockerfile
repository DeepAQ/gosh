# Builder container
FROM golang:latest AS gobuild

COPY . /root/workspace/agent
WORKDIR /root/workspace/agent
ENV GOPATH /root/workspace/agent
RUN set -ex \
 && go get github.com/coreos/etcd/clientv3 \
 && go get github.com/valyala/fasthttp \
 && go get github.com/fatih/pool \
 && go get github.com/shirou/gopsutil \
 && go build -o gosh ./src \
 && ls -la

# Runner container
FROM registry.cn-hangzhou.aliyuncs.com/aliware2018/services AS services
FROM registry.cn-hangzhou.aliyuncs.com/aliware2018/debian-jdk8

COPY --from=services /root/workspace/services/mesh-provider/target/mesh-provider-1.0-SNAPSHOT.jar /root/dists/mesh-provider.jar
COPY --from=services /root/workspace/services/mesh-consumer/target/mesh-consumer-1.0-SNAPSHOT.jar /root/dists/mesh-consumer.jar
COPY --from=gobuild /root/workspace/agent/gosh /usr/local/bin/gosh

COPY --from=services /usr/local/bin/docker-entrypoint.sh /usr/local/bin
COPY start-agent.sh /usr/local/bin

RUN set -ex \
 && chmod a+x /usr/local/bin/start-agent.sh \
 && chmod a+x /usr/local/bin/gosh \
 && mkdir -p /root/logs

EXPOSE 8087

ENTRYPOINT ["docker-entrypoint.sh"]