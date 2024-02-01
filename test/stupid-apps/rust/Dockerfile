FROM rust:1.75-alpine3.19

RUN mkdir /app
WORKDIR /app

COPY ./test/stupid-apps/rust /app
RUN rustc main.rs -o main

CMD ["/app/main"]
