FROM public.ecr.aws/docker/library/golang:1.23 AS build

WORKDIR /go/src/actioneer-runner
COPY go.mod .

RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 go build -o /go/bin/actioneer-runner main.go

FROM gcr.io/distroless/static-debian12:latest
COPY --from=build /go/bin/actioneer-runner /actioneer-runner

# expose http server
EXPOSE 8001
# internal stuff such as metrics
EXPOSE 2112

CMD ["./actioneer-runner"]


# // docker build --platform=linux/amd64 -t runner .
# // kind load docker-image runner:latest
# // smee -u https://smee.io/1eOP0ZmjbtoBY9XS -p 8080 -t http://127.0.0.1:8080/webhook
# // docker tag runner:latest
# // docker push
