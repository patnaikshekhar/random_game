FROM golang

RUN git clone https://github.com/edenhill/librdkafka.git && \
      cd librdkafka && \
      ./configure --prefix /usr && \
      make && \
      make install

RUN go get \
      github.com/codegangsta/gin \
      github.com/xo/dburl \ 
      github.com/lib/pq \ 
      github.com/go-redis/redis \
      github.com/google/uuid \
      github.com/gorilla/websocket \
      github.com/confluentinc/confluent-kafka-go/kafka

ADD . /go/src/github.com/patnaikshekhar/battleship/frontend
WORKDIR /go/src/github.com/patnaikshekhar/battleship/frontend
CMD gin run frontend.go
EXPOSE 3000