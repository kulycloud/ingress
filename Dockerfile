FROM golang:1.15.3-alpine AS builder

ADD ingress/go.mod ingress/go.sum /build/ingress/
ADD protocol/go.mod protocol/go.sum /build/protocol/
ADD common/go.mod common/go.sum /build/common/

ENV CGO_ENABLED=0

WORKDIR /build/ingress
RUN go mod download

COPY ingress/ /build/ingress/
COPY protocol/ /build/protocol
COPY common/ /build/common
RUN go build -o /build/kuly .

FROM scratch

COPY --from=builder /build/kuly /

CMD ["/kuly"]
