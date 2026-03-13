FROM php:8.2-cli-alpine

RUN mkdir /app
WORKDIR /app
COPY ./test/stupid-apps/php/main.php /app/main.php

CMD ["php", "/app/main.php"]
