# --- build stage
FROM golang:1.23 AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/bosun ./cmd/bosun

# --- runtime stage
FROM scratch
COPY --from=build /out/bosun /bosun
ENTRYPOINT ["/bosun"]
