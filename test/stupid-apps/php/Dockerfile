# Utiliza la imagen oficial de PHP
FROM php:8.3.2-bullseye

WORKDIR /app

COPY ./test/stupid-apps/php/program.php /app

# Comando por defecto al ejecutar el contenedor
CMD ["php", "/app/program.php"]
