# Stage 1: Test stage
FROM golang:alpine AS test-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go test ./... -coverprofile=coverage.out
# Run tests and generate coverage report

# Stage 2: Build stage (final image)
FROM golang:alpine

WORKDIR /app

COPY --from=test-stage /app/coverage.out /app/coverage.out
# Copy coverage report if needed
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

EXPOSE 3000

CMD ["./main"]
