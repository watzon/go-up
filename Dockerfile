# Copyright (c) 2024 Christopher Watson
# 
# This software is released under the MIT License.
# https://opensource.org/licenses/MIT

FROM golang:1.23.2-alpine

WORKDIR /app

COPY . .

ENV CGO_ENABLED=1
ENV GOOS=linux

RUN apk add --no-cache gcc g++

RUN go build -o go-up cmd/go-up/main.go

EXPOSE 1234

CMD ["./go-up", "daemon"]
