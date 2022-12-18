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

First, please patch `argocd-repo-server` with the following snippet:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: argocd-repo-server
  namespace: argocd
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

then, create an Application on the Argo CD UI or CLI:

```yaml
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
    path: .github/workflows/                            # It will generate multiple Argo CD application manifests 
                                                        # base on YAML files from this directory.
                                                        # Please make sure the path ends with slash.
    plugin: {}                                          # Argo CD will choose the corresponding CMP automatically
    repoURL: https://gitee.com/linuxsuren/yaml-readme   # a sample project for discovering manifests
    targetRevision: HEAD
  syncPolicy:
    automated:
      selfHeal: true
```

## Compatible
Considering [GitHub Workflows](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstepsuses) 
has a complex syntax. Currently, we support the following ones:

* [Event filter](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#on)
  * Support `on.push` and `on.merge_request`
* keyword `uses`
  * support `actions/checkout`, `actions/setup-go`, `goreleaser/goreleaser-action` and `docker://`
* keyword `run`
* keyword `env`
* [concurrency](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#concurrency)
  * Not support `cancel-in-progress` yet

There are some limitations. For example, only the first job could be recognized in each file.
