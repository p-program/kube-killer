FROM golang:1.14.0-alpine3.11 as build-env
ADD . /src/github.com/p-program/kube-killer
# UFS cache
RUN apk update
WORKDIR /src/github.com/p-program/kube-killer
RUN GOARCH=amd64 CGO_ENABLED=0 go build -mod vendor -o kube-killer  github.com/p-program/kube-killer 

# use the second image to minimize the size of the docker image
FROM alpine:3.11
COPY --from=build-env /src/github.com/p-program/kube-killer/kube-killer /app/kube-killer
WORKDIR /app
ENTRYPOINT ["/app/kube-killer"]