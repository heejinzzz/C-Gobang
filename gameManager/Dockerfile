FROM heejinzzz/mycentos:v0.0.1

ADD src /C-Gobang/gameManager

WORKDIR /C-Gobang/gameManager

RUN go mod tidy

ENTRYPOINT ["go", "run", "main.go"]