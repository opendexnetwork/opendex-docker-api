FROM golang:1.15-alpine3.12 AS builder
RUN apk --no-cache add make
WORKDIR /src
ADD go.mod .
ADD go.sum .
RUN go mod download
ADD . .
RUN make

FROM node:14-alpine3.12 AS ui-builder
RUN apk add --no-cache curl tar gzip
ARG UI_COMMIT=9b7db28e00a1494267773b3b20c0c1afdd7bf316
RUN curl -sL https://github.com/opendexnetwork/opendex-ui/archive/${UI_COMMIT}.tar.gz --output src.tar.gz
WORKDIR /src
RUN tar -xzvf /src.tar.gz --strip-components 1
RUN yarn install
RUN yarn build

FROM alpine:3.12
# we need bash here becuase we launcher opendex-console inside
RUN apk add --no-cache openssl docker-cli bash
COPY --from=builder /src/proxy /usr/local/bin/proxy
COPY --from=ui-builder /src/build ./ui
ENTRYPOINT ["proxy"]
