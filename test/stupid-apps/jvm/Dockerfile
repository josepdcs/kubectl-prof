FROM openjdk:17-bullseye

WORKDIR /opt/app

COPY ./test/stupid-apps/jvm/Main.java ./Main.java
COPY ./test/stupid-apps/jvm/Run.sh ./Run.sh
COPY ./test/stupid-apps/jvm/Run2.sh ./Run2.sh
RUN chmod +x /opt/app/Run.sh
RUN chmod +x /opt/app/Run2.sh

RUN javac Main.java

CMD ["/bin/bash", "/opt/app/Run.sh"]
