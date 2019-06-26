load("@io_bazel_rules_go//go:def.bzl", "go_library")
load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")

proto_library(
    name = "wallet_proto",
    srcs = ["wallet.proto"],
    visibility = ["//visibility:public"],
    deps = ["@com_google_protobuf//:timestamp_proto"],
)

go_proto_library(
    name = "wallet_proto_go",
    compilers = ["@io_bazel_rules_go//proto:go_grpc"],
    importpath = "github.com/CodethinkLabs/wago/pkg/proto",
    proto = ":proto_proto",
    visibility = ["//visibility:public"],
)

proto_library(
    name = "proto_proto",
    srcs = ["wallet.proto"],
    visibility = ["//visibility:public"],
    deps = ["@com_google_protobuf//:timestamp_proto"],
)

go_library(
    name = "go_default_library",
    embed = [":wallet_proto_go"],
    importpath = "github.com/CodethinkLabs/wago/pkg/proto",
    visibility = ["//visibility:public"],
)
