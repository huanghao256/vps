FROM node:24-alpine AS web
WORKDIR /src/web
COPY web/package*.json ./
RUN npm ci
COPY web ./
RUN npm run build

FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN rm -rf internal/httpapi/webdist
COPY --from=web /src/web/dist ./internal/httpapi/webdist
RUN go build -trimpath -ldflags="-s -w" -o /out/vps-inspector ./cmd/vps-inspector

FROM alpine:3.22
RUN adduser -D -H -s /sbin/nologin app
USER app
COPY --from=build /out/vps-inspector /usr/local/bin/vps-inspector
EXPOSE 8719
ENTRYPOINT ["vps-inspector"]
