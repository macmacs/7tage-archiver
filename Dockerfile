# builder image
FROM golang:alpine as builder
RUN mkdir /build
WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

ADD cmd/*.go .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o 7tage-archiver .


# generate clean, final image for end users
FROM alpine:latest
COPY --from=builder /build/7tage-archiver .

VOLUME /music

# executable
ENTRYPOINT [ "./7tage-archiver" ]