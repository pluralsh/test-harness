FROM debian:stretch

ARG KUBECTL_VERSION='1.22.1'
ADD https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl /usr/local/bin/kubectl

RUN chmod +x /usr/local/bin/kubectl && kubectl --help
COPY watch.sh watch.sh
RUN chmod +x watch.sh && mv watch.sh /usr/local/bin/watch.sh

ENTRYPOINT ["/usr/local/bin/watch.sh"]