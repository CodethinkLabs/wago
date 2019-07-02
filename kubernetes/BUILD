load("@io_bazel_rules_k8s//k8s:objects.bzl", "k8s_objects")
load("@k8s_object//:defaults.bzl", "k8s_object")

k8s_object(
    name = "deploy-server",
    images = {"index.docker.io/arlyon/wago:latest": "//cmd/server:image"},
    kind = "deployment",
    template = "//kubernetes:wago-server.yaml",
)

k8s_object(
    name = "deploy-namespace",
    kind = "namespace",
    template = "//kubernetes:wago-namespace.yaml",
)

k8s_object(
    name = "deploy-service",
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
