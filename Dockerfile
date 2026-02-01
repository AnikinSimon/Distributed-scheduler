FROM golang:1.25.0 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN --mount=type=cache,target="./root/.cache/go-build" CGO_ENABLED=0 GOOS=linux go build -o /out/scheduler ./scheduler/cmd
RUN --mount=type=cache,target="./root/.cache/go-build" CGO_ENABLED=0 GOOS=linux go build -o /out/worker ./worker/cmd

FROM gcr.io/distroless/static:nonroot AS scheduler
WORKDIR /app
COPY --from=builder /out/scheduler .
COPY --from=builder /app/scheduler/config/ ./config/
USER nonroot:nonroot
ENTRYPOINT [ "./scheduler" ]

FROM gcr.io/distroless/static:nonroot AS worker
WORKDIR /app
COPY --from=builder /out/worker .
COPY --from=builder /app/worker/config/ ./config/
USER nonroot:nonroot
ENTRYPOINT [ "./worker" ]