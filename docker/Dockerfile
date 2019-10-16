# Build gqlc in a stock Go builder container
FROM qlcchain/go-qlc-builder:1.13.1 as builder

ARG BUILD_ACT=build

COPY . /qlcchain/go-qlc
RUN cd /qlcchain/go-qlc && make clean ${BUILD_ACT}

# Pull gqlc into a second stage deploy alpine container
FROM alpine:3.9.3

COPY --from=builder /qlcchain/go-qlc/build/gqlc .

# EXPOSE 9734 9735 9736

ENTRYPOINT [ "/gqlc"]
