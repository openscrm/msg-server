
FROM golang:latest
WORKDIR /data/xjyk
ADD . .
ENV GOPROXY="https://goproxy.cn"
ENV CGO_ENBLED 1
RUN go mod download
ENV LD_LIBRARY_PATH ${LD_LIBRARY_PATH}:/data/xjyk/lib

RUN go build  -ldflags="-s -w" -installsuffix cgo  -o msg-arch .
CMD ["./msg-arch"]
