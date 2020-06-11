FROM golang:1.12

WORKDIR /catch
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install ./...

EXPOSE 9090
ENTRYPOINT ["catch"]