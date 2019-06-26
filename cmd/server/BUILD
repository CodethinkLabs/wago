load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "github.com/CodethinkLabs/wago/cmd/server",
    visibility = ["//visibility:private"],
    deps = [
        "//pkg/cli:go_default_library",
        "//pkg/cli/common:go_default_library",
        "//pkg/cli/server:go_default_library",
        "//pkg/raft:go_default_library",
        "//pkg/wallet:go_default_library",
        "@io_etcd_go_etcd//raft:go_default_library",
        "@io_etcd_go_etcd//raft/raftpb:go_default_library",
        "@io_etcd_go_etcd//vendor/go.uber.org/zap:go_default_library",
    ],
)

go_binary(
    name = "main",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)
