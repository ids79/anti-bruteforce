FROM golang:1.19 as build

ENV CODE_DIR /go/src/
WORKDIR ${CODE_DIR}

# Кэшируем слои с модулями
COPY ./ ${CODE_DIR}

CMD go test -race ./tests/...

