FROM node:16-bullseye-slim

WORKDIR /usr/src/app

COPY ./package*.json ./

RUN npm install --ignore-scripts  \
    && apt-get update \
    && apt-get install -y procps \
    && apt-get clean

# Bundle app source
COPY ./app2.js ./app2.js
COPY ./subprocess1.js ./subprocess1.js
COPY ./subprocess2.js ./subprocess2.js

CMD [ "node", "--perf-basic-prof", "--heapsnapshot-signal=SIGUSR2", "app2.js" ]
