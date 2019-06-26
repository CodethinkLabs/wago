load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "cli.go",
        "command.go",
    ],
    importpath = "github.com/CodethinkLabs/wago/pkg/cli",
    visibility = ["//visibility:public"],
    deps = ["@com_github_c_bata_go_prompt//:go_default_library"],
)
