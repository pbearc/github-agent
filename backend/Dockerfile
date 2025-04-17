# Use AWS Public ECR mirror instead of Docker Hub
FROM public.ecr.aws/docker/library/golang:1.24-alpine AS build
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN go build -o server ./cmd/server

# Use AWS-hosted alpine instead of Docker Hub
FROM public.ecr.aws/docker/library/alpine:latest
WORKDIR /app
COPY --from=build /app/server /app/server
EXPOSE 8080
CMD ["/app/server"]
