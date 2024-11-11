# tks-batch

[![Go Report Card](https://goreportcard.com/badge/github.com/openinfradev/tks-batch?style=flat-square)](https://goreportcard.com/report/github.com/openinfradev/tks-batch)
[![Go Reference](https://pkg.go.dev/badge/github.com/openinfradev/tks-batch.svg)](https://pkg.go.dev/github.com/openinfradev/tks-batch)
[![Release](https://img.shields.io/github/release/sktelecom/tks-batch.svg?style=flat-square)](https://github.com/openinfradev/tks-batch/releases/latest)

TKS는 TACO Kubernetes Service의 약자로, SK Telecom이 만든 GitOps기반의 서비스 시스템을 의미합니다. 그 중 tks-batch는 클러스터 및 서비스의 상태를 관리하기 위한 batch job 컴포넌트입니다.
tks-cluster-lcm 에서 저장한 argo workflow id 를 사용하여, 주기적으로 argo workflow 를 체크 후 DB 와 상태를 동기화합니다.


## Quick Start

### Prerequisite
* docker 20.x 설치

### 서비스 구동 
#### For go developers
```
  $ go build -o bin/tks-batch ./cmd/server/
  $ bin/tks-batch -port 9110
```

#### For docker users
```
  $ docker pull sktcloud/tks-batch:latest
  $ docker run --name tks-batch -p 9110:9110 -d \
   sktcloud/tks-batch:latest server -port 9110 
```

### 서비스 Build & Deploy
```
  $ docker build -t tks-batch:latest -f Dockerfile .


```

