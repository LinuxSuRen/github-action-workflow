[![codecov](https://codecov.io/gh/LinuxSuRen/github-action-workflow/branch/master/graph/badge.svg?token=mnFyeD2IQ7)](https://codecov.io/gh/LinuxSuRen/github-action-workflow)

# github-action-workflow
GitHub Actions compatible workflows

## Feature
* Convert GitHub Workflows to Argo Workflows
* Argo CD Config Management Plugin (CMP)

## Usage
You can use it as a CLI:

```shell
gaw convert .github/workflows/pull-request.yaml
```

you can install it via [hd](https://github.com/LinuxSuRen/http-downloader):

```shell
hd i gaw
```

## As CMP
This repository could be [Config Management Plugin](https://argo-cd.readthedocs.io/en/stable/user-guide/config-management-plugins/#option-2-configure-plugin-via-sidecar) as well.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: argocd-repo-server
spec:
  template:
    spec:
      containers:
      - args:
        - --loglevel
        - debug
        command:
        - /var/run/argocd/argocd-cmp-server
        image: ghcr.io/linuxsuren/github-action-workflow:master
        imagePullPolicy: IfNotPresent
        name: tool
        resources: {}
        securityContext:
          runAsNonRoot: true
          runAsUser: 999
        volumeMounts:
        - mountPath: /var/run/argocd
          name: var-files
        - mountPath: /home/argocd/cmp-server/plugins
          name: plugins
```
