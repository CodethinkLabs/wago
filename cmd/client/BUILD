load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "github.com/CodethinkLabs/wago/cmd/client",
    visibility = ["//visibility:private"],
    deps = [
        "//pkg/cli:go_default_library",
        "//pkg/cli/client:go_default_library",
        "//pkg/cli/common:go_default_library",
        "//pkg/proto:go_default_library",
        "@org_golang_google_grpc//:go_default_library",
    ],
)

go_binary(
    name = "main",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)
