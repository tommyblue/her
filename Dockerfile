# Build the Go Binary.
FROM golang:1.13 as her
ENV CGO_ENABLED 0
ARG VCS_REF
ARG PACKAGE_NAME
ARG PACKAGE_PREFIX

# Create a location in the container for the source code. Using the
# default GOPATH location.
RUN mkdir -p /go/src/github.com/tommyblue/her

# Copy the module files first and then download the dependencies. If this
# doesn't change, we won't need to do this again in future builds.
COPY go.* /go/src/github.com/tommyblue/her/
WORKDIR /go/src/github.com/tommyblue/her
RUN go mod download

# Copy the source code into the container.
COPY cmd cmd
# COPY internal internal

# Build the service binary. We are doing this last since this will be different
# every time we run through this process.
WORKDIR /go/src/github.com/tommyblue/her/cmd/${PACKAGE_PREFIX}${PACKAGE_NAME}
RUN go build -mod=readonly -ldflags "-X main.build=${VCS_REF}"


# Run the Go Binary in Alpine.
FROM alpine:3.16
ARG BUILD_DATE
ARG VCS_REF
ARG PACKAGE_NAME
ARG PACKAGE_PREFIX
ARG TELEGRAM_BOT_TOKEN
ENV TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN}
COPY --from=her /go/src/github.com/tommyblue/her/cmd/${PACKAGE_PREFIX}${PACKAGE_NAME}/${PACKAGE_NAME} /app/main
WORKDIR /app
CMD /app/main

LABEL org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.title="${PACKAGE_NAME}" \
      org.opencontainers.image.authors="Tommaso Visconti<tommaso.visconti@gmail.com>" \
      org.opencontainers.image.source="https://github.com/tommyblue/her/cmd/${PACKAGE_PREFIX}${PACKAGE_NAME}" \
      org.opencontainers.image.revision="${VCS_REF}" \
      org.opencontainers.image.vendor="TommyBlue"
