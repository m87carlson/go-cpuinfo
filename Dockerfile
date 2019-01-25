FROM  golang:latest AS builder
#FROM centos:latest AS builder

# copy source tree in
COPY *.go /build/

#RUN yum install -y golang bash

# create a self-contained build structure
WORKDIR /build
RUN go get github.com/json-iterator/go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-extldflags "-static"' -a -installsuffix nocgo -o /main 

FROM scratch
COPY --from=builder /main /
EXPOSE 8080
ENTRYPOINT ["/main"]
