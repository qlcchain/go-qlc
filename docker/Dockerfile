# Build gqlc in a stock Go builder container
FROM qlcchain/go-qlc-builder:latest as builder

ARG BUILD_ACT=build

COPY . /qlcchain/go-qlc
RUN cd /qlcchain/go-qlc && make clean ${BUILD_ACT}

# Pull gqlc into a second stage deploy alpine container
FROM alpine:3.14.3
LABEL maintainer="developers@qlink.mobi"

ENV QLCHOME /qlcchain

RUN apk --no-cache add ca-certificates && \
    addgroup qlcchain && \
    adduser -S -G qlcchain qlcchain -s /bin/sh -h "$QLCHOME" && \
    chown -R qlcchain:qlcchain "$QLCHOME"

USER qlcchain

WORKDIR $QLCHOME

COPY --from=builder /qlcchain/go-qlc/build/gqlc /usr/local/bin/gqlc

EXPOSE 9734 9735 9736

ENTRYPOINT [ "gqlc"]

VOLUME [ "$QLCHOME" ]
