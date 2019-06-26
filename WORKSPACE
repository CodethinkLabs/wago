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
    build_file_proto_mode = "disable",
    commit = "b1812a410fbca6fb77bf95b496408c7b75d0a370",
    importpath = "go.etcd.io/etcd",
)

go_repository(
    name = "com_github_c_bata_go_prompt",
    importpath = "github.com/c-bata/go-prompt",
    tag = "v0.2.3",
)

go_repository(
    name = "org_golang_x_crypto",
    commit = "57b3e21c3d5606066a87e63cfe07ec6b9f0db000",
    importpath = "golang.org/x/crypto",
)

go_repository(
    name = "org_uber_go_atomic",
    importpath = "go.uber.org/atomic",
    tag = "v1.4.0",
)

go_repository(
    name = "org_uber_go_multierr",
    importpath = "go.uber.org/multierr",
    tag = "v1.1.0",
)

go_repository(
    name = "in_gopkg_yaml_v2",
    importpath = "gopkg.in/yaml.v2",
    tag = "v2.2.2",
)

go_repository(
    name = "com_github_mattn_go_runewidth",
    importpath = "github.com/mattn/go-runewidth",
    tag = "v0.0.4",
)

go_repository(
    name = "com_github_pkg_term",
    commit = "aa71e9d9e942418fbb97d80895dcea70efed297c",
    importpath = "github.com/pkg/term",
)

# add grpc
go_repository(
    name = "org_golang_google_grpc",
    importpath = "google.golang.org/grpc",  # Import path used in the .go files
)

go_repository(
    name = "com_github_golang_protobuf",
    commit = "b285ee9cfc6c881bb20c0d8dc73370ea9b9ec90f",
    importpath = "github.com/golang/protobuf",
)
