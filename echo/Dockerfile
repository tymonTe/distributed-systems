FROM ubuntu:latest

RUN apt update && apt install -y \
    openjdk-17-jdk \
    graphviz \
    gnuplot \
    wget \
    git

RUN wget https://go.dev/dl/go1.21.5.linux-arm64.tar.gz && \
    tar -C /usr/local -xzf go1.21.5.linux-arm64.tar.gz && \
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile && \
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /root/.bashrc