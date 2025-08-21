FROM golang:latest
ENV GO111MODULE=on

COPY . /app

WORKDIR /app

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o provider-identification .

FROM scratch

# Copy Runtime
COPY --from=0 /app/provider-identification /provider-identification
COPY --from=0 /app/public/* /public/

EXPOSE 80 443

ENTRYPOINT ["/provider-identification"]
