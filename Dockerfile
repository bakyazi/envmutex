FROM golang:1.21 AS compiler
WORKDIR /src/app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o ./a.out ./cmd/main.go

FROM gcr.io/distroless/static
COPY --from=compiler /src/app/a.out /server
ENTRYPOINT ["/server"]