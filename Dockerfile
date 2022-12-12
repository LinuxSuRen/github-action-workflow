FROM golang:1.19 as builder

WORKDIR /workspace
COPY . .
RUN go mod download
RUN CGO_ENABLE=0 go build -ldflags "-w -s" -o gaw

FROM ubuntu:kinetic

LABEL "repository"="https://github.com/linuxsuren/github-action-workflow"
LABEL "homepage"="https://github.com/linuxsuren/github-action-workflow"

RUN mkdir -p /home/argocd/cmp-server/config
COPY --from=builder /workspace/gaw /usr/local/bin/gaw
COPY --from=builder /workspace/plugin.yaml /home/argocd/cmp-server/config/plugin.yaml

CMD ["gaw"]