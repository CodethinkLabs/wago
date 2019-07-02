load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")
load("@io_bazel_rules_docker//go:image.bzl", "go_image")
load("@io_bazel_rules_docker//container:container.bzl", "container_push")

go_library(
    name = "go_default_library",
    srcs = [
        "http.go",
        "main.go",
    ],
    importpath = "github.com/CodethinkLabs/wago/cmd/server",
    visibility = ["//visibility:private"],
    deps = [
        "//pkg/cli:go_default_library",
        "//pkg/cli/common:go_default_library",
        "//pkg/cli/server:go_default_library",
        "//pkg/proto:go_default_library",
        "//pkg/raft:go_default_library",
        "//pkg/wallet:go_default_library",
        "@io_etcd_go_etcd//raft:go_default_library",
        "@io_etcd_go_etcd//raft/raftpb:go_default_library",
        "@io_etcd_go_etcd//vendor/go.uber.org/zap:go_default_library",
        "@org_golang_google_grpc//:go_default_library",
        "@org_golang_google_grpc//codes:go_default_library",
        "@org_golang_google_grpc//status:go_default_library",
        "@org_golang_x_crypto//ed25519:go_default_library",
    ],
)

go_binary(
    name = "binary",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)

go_image(
    name = "image",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)

container_push(
    name = "push",
    format = "Docker",
    image = ":image",
    registry = "index.docker.io",
    repository = "arlyon/wago",
    tag = "latest",
)
