load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "a82a352bffae6bee4e95f68a8d80a70e87f42c4741e6a448bec11998fcc82329",
    urls = ["https://github.com/bazelbuild/rules_go/releases/download/0.18.5/rules_go-0.18.5.tar.gz"],
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "3c681998538231a2d24d0c07ed5a7658cb72bfb5fd4bf9911157c0e9ac6a2687",
    urls = ["https://github.com/bazelbuild/bazel-gazelle/releases/download/0.17.0/bazel-gazelle-0.17.0.tar.gz"],
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains()

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")

gazelle_dependencies()

go_repository(
    name = "io_etcd_go_etcd",
    commit = "b1812a410fbca6fb77bf95b496408c7b75d0a370",
    importpath = "github.com/etcd-io/etcd",
)

go_repository(
    name = "com_github_c_bata_go_prompt",
    importpath = "github.com/c-bata/go-prompt",
    tag = "v0.2.3",
)

go_repository(
    name = "org_golang_x_crypto",
    importpath = "golang.org/x/crypto",
    type = "zip",
)

go_repository(
    name = "org_uber_go_zap",
    importpath = "go.uber.org/zap",
)

go_repository(
    name = "com_github_alecaivazis_survey",
    importpath = "github.com/alecaivazis/survey",
    tag = "v2.0.0",
)
