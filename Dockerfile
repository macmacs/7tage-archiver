# builder image
FROM golang:alpine as builder
RUN mkdir /build
WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

ADD cmd/*.go .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o fm4-archiver .


# generate clean, final image for end users
FROM alpine:latest
COPY --from=builder /build/fm4-archiver .

VOLUME /music

# executable
ENTRYPOINT [ "./fm4-archiver" ]