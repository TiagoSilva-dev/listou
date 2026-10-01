FROM golang:1.26-alpine AS build
WORKDIR /src
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
COPY apps/api/ ./
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o /out/api ./cmd/api \
 && CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/migrate ./cmd/migrate

FROM gcr.io/distroless/static-debian12:nonroot AS runtime
WORKDIR /app
COPY --from=build /out/api /out/migrate /app/
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/api"]
