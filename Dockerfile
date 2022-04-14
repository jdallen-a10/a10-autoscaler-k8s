#
#  Dockerfile for a10-autoscaler-k8s Container
#
#  John D. Allen
#  Global Solutions Architect - Cloud, IoT, & Automation
#  A10 Networks, Inc.
#
#  April, 2022
#

FROM golang:1.17.3 AS builder

RUN mkdir /app
RUN mkdir /app/axapi
RUN mkdir /app/k8s-go
ADD ./a10-golang-axapi /app/a10-golang-axapi
ADD ./k8s-go /app/k8s-go
ADD go.* /app
ADD main.go /app
ADD ./config.yaml /app/config.yaml
ADD Dockerfile /app

WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/a10-autoscaler-k8s main.go

##
## Build Final Container
FROM alpine:latest AS production
COPY --from=builder /app .

CMD ["./a10-autoscaler-k8s"]
