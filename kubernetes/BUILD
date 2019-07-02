load("@io_bazel_rules_k8s//k8s:objects.bzl", "k8s_objects")
load("@io_bazel_rules_k8s//k8s:object.bzl", "k8s_object")

k8s_object(
    name = "deploy-server",
    cluster = "eks_k8-cluster-alex",
    images = {"index.docker.io/arlyon/wago:latest": "//cmd/server:image"},
    kind = "deployment",
    template = "//kubernetes:wago-server.yaml",
)

k8s_object(
    name = "deploy-namespace",
    cluster = "eks_k8-cluster-alex",
    kind = "namespace",
    template = "//kubernetes:wago-namespace.yaml",
)

k8s_object(
    name = "deploy-service",
    cluster = "eks_k8-cluster-alex",
    kind = "service",
    template = "//kubernetes:wago-service.yaml",
)

k8s_objects(
    name = "deploy",
    objects = [
        ":deploy-namespace",
        ":deploy-server",
        ":deploy-service",
    ],
)
