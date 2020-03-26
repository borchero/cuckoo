# syntax=docker/dockerfile:experimental

#################
### TERRAFORM ###
#################
FROM busybox:1.31 AS terraform

ARG TERRAFORM_VERSION=0.12.24

RUN wget https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip \
	&& unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip

##############
### CUCKOO ###
##############
FROM golang:1.14-alpine AS cuckoo

ENV CGO_ENABLED=0 GOOS=linux

# 1) Get source
WORKDIR /app
COPY source /app

# 2) Get dependencies
RUN go get github.com/markbates/pkger/cmd/pkger \
    && go get -v \
    && pkger

# 3) Run tests
RUN go test ./...

# 4) Compile production binary 
RUN go build -v \
    -o main \
    -tags netgo \
    -ldflags '-extldflags "-static"'

#############
### FINAL ###
#############
FROM alpine:3.11

# 1) Install Git and OpenSSH
RUN apk add --no-cache git openssh-client

# 2) Install Terraform
COPY --from=terraform /terraform /usr/bin/terraform

# 3) Add Cuckoo
COPY --from=cuckoo /app/main /usr/bin/cuckoo

# 4) Set entrypoint such that both Docker and Kubernetes executor can use the image as is
ENTRYPOINT [""]
CMD ["/bin/sh"]
