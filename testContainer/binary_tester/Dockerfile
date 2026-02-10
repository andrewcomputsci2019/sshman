FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        openssh-client \
        bash \
        coreutils \
        util-linux \
        sshpass \
        tini


WORKDIR /test

COPY build/bin/ssh-man /usr/local/bin/ssh-man

copy testContainer/binary_tester/run-tests.sh /test/run-tests.sh
copy testContainer/binary_tester/custom_ssh_config /test/ssh_config

RUN chmod +x /test/run-tests.sh

ENTRYPOINT ["/usr/bin/tini", "--"]
CMD ["/test/run-tests.sh"]