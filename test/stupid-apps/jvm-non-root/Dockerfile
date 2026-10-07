FROM eclipse-temurin:21-jdk

WORKDIR /opt/app

# Create non-root user and group with UID/GID 65532 matching integration test specs
RUN groupadd -g 65532 appgroup && \
    useradd -u 65532 -g appgroup -d /opt/app -s /bin/false appuser && \
    chown -R 65532:65532 /opt/app

COPY --chown=65532:65532 ./Main.java ./Main.java

RUN javac Main.java

USER 65532:65532

CMD ["java", "-cp", "/opt/app", "Main"]
