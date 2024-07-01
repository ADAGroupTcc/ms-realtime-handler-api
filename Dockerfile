FROM 289208114389.dkr.ecr.us-east-1.amazonaws.com/golang:1.22.4-alpine3.19 as build
ARG GITHUB_TOKEN

WORKDIR /src

COPY . .

RUN echo machine github.com login picpay-devex password "$GITHUB_TOKEN" > ~/.netrc \
    ; GOPRIVATE=github.com/PicPay go mod download \
    ; CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/api cmd/api/main.go

FROM 289208114389.dkr.ecr.us-east-1.amazonaws.com/alpine:3.18.2

RUN addgroup -S picpay && adduser -S picpay -G picpay
WORKDIR /home/picpay/app

COPY --from=build /src/docker-entrypoint.sh /src/bin/api ./
RUN chmod +x docker-entrypoint.sh

USER picpay
EXPOSE 8080

HEALTHCHECK --interval=5s --timeout=3s CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1
ENTRYPOINT ["/home/picpay/app/docker-entrypoint.sh"]
CMD ["/home/picpay/app/api"]
