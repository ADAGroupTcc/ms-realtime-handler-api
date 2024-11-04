FROM golang:1.22.4-alpine3.19 AS build
WORKDIR /src

COPY . .

RUN apk add -U --no-cache gcc g++ openssh

RUN go mod download \
  && CGO_ENABLED=0 go build -ldflags='-s -w -extldflags "-static"' -o bin/api cmd/api/main.go

FROM alpine:3.18.2
WORKDIR /home/adda-tcc/app

COPY --from=build /src/docker-entrypoint.sh /src/bin/api ./
COPY --from=build /src/mongodb.pem ./
RUN chmod +x docker-entrypoint.sh

EXPOSE 8080

HEALTHCHECK --interval=5s --timeout=3s CMD wget --no-verbose --tries=3 --spider http://localhost:7000/health || exit 1
ENTRYPOINT ["/home/adda-tcc/app/docker-entrypoint.sh"]
CMD ["/home/adda-tcc/app/api"]
