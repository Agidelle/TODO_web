FROM golang:1.24-alpine3.20 AS build

WORKDIR /app

RUN apk add git

ENV APP_NAME="TODO_web"

RUN mkdir /out
COPY . /app/

RUN go build  \
    -o /out/${APP_NAME}  \
    github.com/agidelle/todo_web


FROM alpine:3.20

WORKDIR /app

ENV TODO_PORT=7540
ENV TODO_DBFILE=./scheduler.db
ENV TODO_DRIVER=sqlite
ENV TODO_PASSWORD=1111
ENV TODO_JWTSECRET=secret

COPY --from=build /out/TODO_web /app/
COPY --from=build /app/web /app/web


EXPOSE 7540
RUN apk add curl
CMD ["/app/TODO_web"]
