ARG GO_VERSION=1.24.2
FROM golang:${GO_VERSION}-bookworm as builder

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .

RUN go install github.com/a-h/templ/cmd/templ@latest
RUN templ generate
RUN go mod tidy
# Ensure static linking for linux amd64
RUN CGO_ENABLED=1 go build -o /run-app cmd/server/main.go


FROM debian:bookworm
ENV APP_HOME=/app

# Create and switch into the app directory
WORKDIR $APP_HOME

# Ensure the /mnt/database directory exists
RUN mkdir -p /mnt/database

# Copy in the executable, static files, and migrations
COPY --from=builder /run-app       $APP_HOME/run-app
COPY --from=builder /usr/src/app/static     $APP_HOME/static
COPY --from=builder /usr/src/app/migrations $APP_HOME/migrations

# Ensure the binary is executable
RUN chmod +x $APP_HOME/run-app

# Default to using local SQLite if DATABASE_URL isn't set
ENV DATABASE_URL=chores.db

# Run the binary from the current dir
CMD ["./run-app"]
