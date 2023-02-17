FROM python:3.11.2-slim-bullseye

RUN mkdir /app
WORKDIR /app
COPY ./test/stupid-apps/python/main.py /app/main.py

CMD ["python", "/app/main.py"]
