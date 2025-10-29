FROM golang:1.25.0 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN --mount=type=cache,target="./root/.cache/go-build" CGO_ENABLED=0 GOOS=linux go build -o /scheduler/scheduler ./scheduler/cmd

FROM alpine:3.22.2

WORKDIR /app
COPY --from=builder ./scheduler/scheduler  .
COPY --from=builder /app/scheduler/config/ ./config/

ENTRYPOINT [ "./scheduler" ]