load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "file.go",
        "store.go",
        "util.go",
    ],
    importpath = "github.com/CodethinkLabs/wago/pkg/wallet",
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/util:go_default_library",
        "@in_gopkg_yaml_v2//:go_default_library",
        "@io_etcd_go_etcd//etcdserver/api/snap:go_default_library",
        "@org_golang_x_crypto//ed25519:go_default_library",
        "@org_golang_x_crypto//pbkdf2:go_default_library",
        "@org_golang_x_crypto//ssh/terminal:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "store_test.go",
        "util_test.go",
    ],
    embed = [":go_default_library"],
    deps = [
        "@io_etcd_go_etcd//etcdserver/api/snap:go_default_library",
        "@org_golang_x_crypto//ed25519:go_default_library",
    ],
)
