FROM node:16-bullseye-slim

WORKDIR /usr/src/app

COPY ./test/stupid-apps/node/package*.json ./

RUN npm install

# Bundle app source
COPY ./test/stupid-apps/node/app.js ./app.js

CMD [ "node", "--perf-basic-prof", "app.js" ]
