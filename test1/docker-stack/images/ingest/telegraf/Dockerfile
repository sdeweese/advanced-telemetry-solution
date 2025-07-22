# Build field_prefix binary
FROM golang:1.22-alpine AS build_field_prefix
WORKDIR /app
COPY field_prefix/go.mod field_prefix/go.sum ./
RUN go mod download
COPY field_prefix/main.go .
RUN CGO_ENABLED=0 go build -o field_prefix ./main.go

# Build split_metrics binary
FROM golang:1.22-alpine AS build_split_metrics
WORKDIR /app
COPY split_metrics/go.mod split_metrics/go.sum ./
RUN go mod download
COPY split_metrics/main.go .
RUN CGO_ENABLED=0 go build -o split_metrics ./main.go

# Copy the binaries to the telegraf image
FROM telegraf:1.33-alpine
COPY --from=build_field_prefix /app/field_prefix /usr/bin
COPY --from=build_split_metrics /app/split_metrics /usr/bin
