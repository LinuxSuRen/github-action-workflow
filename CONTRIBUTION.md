## Setup environment

```shell
curl https://linuxsuren.github.io/tools/install.sh|bash

hd i k3d
k3d cluster create
```

### Argo workflows
```shell

kubectl create ns argo
docker run -it --rm -v $HOME/.kube/:/root/.kube --network host ghcr.io/linuxsuren/argo-workflows-guide:master
kubectl patch svc -n argo argo-server -p '{"spec":{"type":"NodePort"}}'

port=$(kubectl get svc -n argo argo-server -ojsonpath={.spec.ports[0].nodePort})
k3d node edit --port-add ${port}:${port} k3d-k3s-default-serverlb

kubectl patch deployment \
  argo-server \
  --namespace argo \
  --type='json' \
  -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/args", "value": [
  "server",
  "--auth-mode=server"
]}]'
```

### Argo CD
```shell
kubectl create ns argocd
docker run -it --rm -v /root/.kube/:/root/.kube --network host ghcr.io/linuxsuren/argo-cd-guide:master

kubectl patch svc -n argocd argocd-server -p '{"spec":{"type":"NodePort"}}'

port=$(kubectl get svc -n argocd argocd-server -ojsonpath={.spec.ports[0].nodePort})
k3d node edit --port-add ${port}:${port} k3d-k3s-default-serverlb

kubectl -n argocd get secret argocd-initial-admin-secret -ojsonpath={.data.password} | base64 -d
```

## Install

```shell
kubectl patch deployment \
  argocd-repo-server \
  --namespace argocd \
  --type='json' \
  -p='[{"op": "add", "path": "/spec/template/spec/containers/0", "value": {
    "args": ["--loglevel", "debug"],
    "command": ["/var/run/argocd/argocd-cmp-server"],
    "image": "ghcr.io/linuxsuren/github-action-workflow:master",
    "name": "tool",
    "securityContext": {
        "runAsNonRoot": true,
        "runAsUser": 999
    },
    "volumeMounts": [{
        "mountPath": "/var/run/argocd",
        "name": "var-files"
    }, {
        "mountPath": "/home/argocd/cmp-server/plugins",
        "name": "plugins"
    }]
}}]'
```

## Samples
A kustomization sample:
```shell
cat <<EOF | kubectl apply -n argocd -f -
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: learn-pipeline-go
spec:
  destination:
    namespace: default
    server: https://kubernetes.default.svc
  project: default
  source:
    repoURL: https://gitee.com/devops-ws/learn-pipeline-go
    path: kustomize
    targetRevision: HEAD
  syncPolicy:
    automated:
      selfHeal: true
EOF
```

A plugin sample:
```shell
cat <<EOF | kubectl apply -f -
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: yaml-readme
  namespace: argocd
spec:
  destination:
    namespace: default
    server: https://kubernetes.default.svc
  project: default
  source:
    path: .github/workflows/
    repoURL: https://gitee.com/linuxsuren/yaml-readme
    targetRevision: HEAD
  syncPolicy:
    automated:
      selfHeal: true
EOF
```
