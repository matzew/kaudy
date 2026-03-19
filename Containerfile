FROM registry.fedoraproject.org/fedora:latest
RUN dnf install -y git nodejs npm which findutils && \
    dnf group install -y development-tools && \
    dnf clean all
RUN npm install -g @anthropic-ai/claude-code
ARG USER_NAME=kaudy
RUN groupadd -g 1000 ${USER_NAME} && \
    useradd -u 1000 -g 1000 -m ${USER_NAME} && \
    mkdir -p /var/home && \
    ln -sf /home/${USER_NAME} /var/home/${USER_NAME}
USER ${USER_NAME}
WORKDIR /workspace
ENTRYPOINT ["claude"]
CMD ["--dangerously-skip-permissions"]
