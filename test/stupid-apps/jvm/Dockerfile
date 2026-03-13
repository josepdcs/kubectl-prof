FROM openjdk:25-ea-17-bullseye

WORKDIR /opt/app

COPY ./Main.java ./Main.java
COPY ./Run.sh ./Run.sh
COPY ./Run2.sh ./Run2.sh
RUN chmod +x /opt/app/Run.sh
RUN chmod +x /opt/app/Run2.sh

RUN javac Main.java

CMD ["/bin/bash", "/opt/app/Run.sh"]
