FROM golang:1.19 as builder

WORKDIR /workspace
COPY . .
RUN go mod download
RUN CGO_ENABLE=0 go build -ldflags "-w -s" -o gaw

FROM alpine:3.10

LABEL "repository"="https://github.com/linuxsuren/github-action-workflow"
LABEL "homepage"="https://github.com/linuxsuren/github-action-workflow"

COPY --from=builder /workspace/gaw /usr/local/bin/gaw

CMD ["gaw"]