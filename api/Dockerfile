FROM golang:latest
RUN mkdir /app
WORKDIR /app
ADD . /app
RUN go build ./main.go
CMD ["./main"]