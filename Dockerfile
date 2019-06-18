FROM ubuntu:18.04

RUN apt update && apt install -yq golang-go curl git
RUN echo "deb [arch=amd64] http://storage.googleapis.com/bazel-apt testing jdk1.8" | tee /etc/apt/sources.list.d/bazel.list
RUN curl https://bazel.build/bazel-release.pub.gpg | apt-key add -
RUN apt update && apt -yq install bazel

WORKDIR /src
RUN git clone https://github.com/CodethinkLabs/wago.git
RUN cd wago && bazel build //cmd/server:main

RUN cp ./wago/bazel-bin/cmd/server/linux_amd64_stripped/main /usr/local/bin/wago

ENTRYPOINT ["/usr/local/bin/wago"]