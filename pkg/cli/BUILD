load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "auth.go",
        "bank.go",
        "cli.go",
        "command.go",
        "delete.go",
        "new.go",
        "send.go",
        "status.go",
    ],
    importpath = "github.com/CodethinkLabs/wago/pkg/cli",
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/raft:go_default_library",
        "//pkg/util:go_default_library",
        "//pkg/wallet:go_default_library",
        "@com_github_c_bata_go_prompt//:go_default_library",
        "@org_golang_x_crypto//ed25519:go_default_library",
    ],
)
