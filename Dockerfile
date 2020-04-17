FROM golang:1.12.0 as build

ENV GOPROXY https://goproxy.io
ENV CGO_ENABLED 0
ENV GOOS linux

RUN go get github.com/fsnotify/fsnotify
RUN go get github.com/shirou/gopsutil/process
RUN mkdir -p /go/src/app
ADD main.go /go/src/app/
WORKDIR /go/src/app

RUN  go build -a -o nginx-reloader .

##main image
FROM nginx:1.12.1-alpine

COPY --from=build /go/src/app/nginx-reloader /nginx-reloader

CMD ["/nginx-reloader"]
