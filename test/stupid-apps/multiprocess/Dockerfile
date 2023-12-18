FROM python:3.11.2-slim-bullseye

RUN mkdir /app
WORKDIR /app
COPY ./test/stupid-apps/multiprocess/one.py /app/one.py
COPY ./test/stupid-apps/multiprocess/two.py /app/two.py
COPY ./test/stupid-apps/multiprocess/run.sh /app/run.sh
RUN chmod +x /app/one.py
RUN chmod +x /app/two.py
RUN chmod +x /app/run.sh

CMD /app/run.sh
