FROM golang:1.26.6-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api \
	&& CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/agent ./cmd/agent

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata

COPY --from=build /out/api /api
COPY --from=build /out/agent /agent

USER nobody

EXPOSE 8080

ENTRYPOINT ["/api"]
