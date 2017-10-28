FROM google/golang

WORKDIR /gopath/src/blockchain
ADD .  /gopath/src/blockchain

CMD []
EXPOSE 9200
ENTRYPOINT ./src/cli/cli