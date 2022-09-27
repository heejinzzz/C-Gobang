FROM heejinzzz/mycentos:v0.0.1

ADD src /C-Gobang/logCollector

WORKDIR /C-Gobang/logCollector

RUN go mod tidy

ENTRYPOINT ["go", "run", "main.go"]