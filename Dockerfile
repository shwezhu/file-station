FROM golang:alpine

# All the following command will treated as inside this folder of docker
WORKDIR /app
# copy all the files of our project into the /app folder of docker
COPY ./ ./
RUN go mod download

# Optional:
# To bind to a TCP port, runtime parameters must be supplied to the docker command.
# But we can document in the Dockerfile what ports
# the application is going to listen on by default.
# https://docs.docker.com/engine/reference/builder/#expose
EXPOSE 80

# Install gcc to compile cgo
RUN apk add --no-cache --update go gcc g++

# GOOS=linux: set target os to linux
# 'go build -o /server .': 'go build' is a command, '-o /server' output,
# '.' tells the compiler to build the Go program located in the current directory.
RUN go build -o /server .

CMD ["/server"]

# $ docker build --platform linux/amd64 -t shwezhu/file-station:v2 .