FROM heejinzzz/mycentos:v0.0.1

ADD src /C-Gobang/userManager

WORKDIR /C-Gobang/userManager

RUN go mod tidy

ENTRYPOINT ["go", "run", "main.go"]