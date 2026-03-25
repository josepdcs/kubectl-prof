FROM python:3.11.2-slim-bullseye

RUN mkdir /app
WORKDIR /app
COPY ./one.py /app/one.py
COPY ./two.py /app/two.py
COPY ./run.sh /app/run.sh
RUN chmod +x /app/one.py
RUN chmod +x /app/two.py
RUN chmod +x /app/run.sh

CMD /app/run.sh
