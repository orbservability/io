FROM golang:1.21

# Set the working directory outside $GOPATH to enable the support for modules.
WORKDIR /usr/src/app

# Copy the Go Modules manifests and download the dependencies.
# This is done before copying the code to leverage Docker cache layers.
COPY go.* ./
RUN go mod download

# Copy the source code from the current directory to the working directory inside the container.
COPY . .
