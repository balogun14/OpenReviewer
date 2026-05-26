FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal

RUN go build -o /out/openreview ./cmd/openreview

FROM alpine:3.22

RUN adduser -D -H openreview
USER openreview

COPY --from=build /out/openreview /usr/local/bin/openreview

EXPOSE 8080
ENTRYPOINT ["openreview"]
