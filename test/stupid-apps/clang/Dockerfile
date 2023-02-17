FROM bitnami/minideb

WORKDIR /app

RUN apt-get update && apt-get install -y gcc

COPY ./test/stupid-apps/clang/main.c /app/main.c
RUN gcc -pthread main.c -o program

CMD ["/app/program"]
