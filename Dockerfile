FROM golang:1.20

WORKDIR /app
COPY . .

RUN mkdir build
RUN go mod tidy
RUN go build -o build/app cmd/main.go
RUN chmod +x build/app

EXPOSE 80

ENTRYPOINT [ "./build/app" ]