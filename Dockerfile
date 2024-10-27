# Copyright (c) 2024 Christopher Watson
# 
# This software is released under the MIT License.
# https://opensource.org/licenses/MIT

FROM golang:1.23.2-alpine

WORKDIR /app

COPY . .

RUN go build -o go-up cmd/go-up/main.go

EXPOSE 1234

CMD ["./go-up daemon"]