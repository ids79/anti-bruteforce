# Собираем в гошке
FROM golang:1.19 as build

ENV BIN_FILE /opt/brutforce
ENV CODE_DIR /go/src/
WORKDIR ${CODE_DIR}

# Кэшируем слои с модулями
COPY ./ ${CODE_DIR}

# Собираем статический бинарник Go (без зависимостей на Си API),
# иначе он не будет работать в alpine образе.
ARG LDFLAGS
RUN CGO_ENABLED=0 GOOS=linux go build \
        -ldflags "$LDFLAGS" \
        -o ${BIN_FILE} ${CODE_DIR}cmd/bruteforce/*

# На выходе тонкий образ
FROM alpine:3.9

LABEL ORGANIZATION="ids"
LABEL SERVICE="anty-bruteforce"
LABEL MAINTAINERS="ids79@otus.ru"

ENV BIN_FILE "/opt/brutforce"
COPY --from=build ${BIN_FILE} ${BIN_FILE}

ENV CONFIG_FILE /etc/brutforce/config.toml
COPY ./configs/config.toml ${CONFIG_FILE}

CMD ${BIN_FILE} -config ${CONFIG_FILE}
