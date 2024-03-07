FROM golang:latest as builder
WORKDIR /build
ADD . .
RUN go get -a && go build -trimpath -ldflags "-s -w" -o ast

FROM curlimages/curl:latest as argocd
WORKDIR /tmp
RUN curl -Lk https://github.com/argoproj/argo-cd/releases/download/v2.9.7/argocd-linux-amd64 -o argocd \
    && chmod +x argocd

FROM gcr.io/distroless/base-debian12:nonroot
LABEL org.opencontainers.image.source=https://github.com/pcanilho/argocd-sync-timeout
WORKDIR /bin
COPY --from=builder /build/ast .
COPY --from=argocd /tmp/argocd .
ENTRYPOINT ["/bin/ast"]
