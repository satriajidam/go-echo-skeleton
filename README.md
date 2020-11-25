# go-echo-skeleton

A skeleton project for building RESTful API with Go &amp; Echo using Clean Architecture.

This project use [golang-standards/project-layout](https://github.com/golang-standards/project-layout) to structure its layout.

## Dependencies

* Web Framework: [labstack/echo](https://github.com/labstack/echo)
* REST Client: [go-resty/resty](https://github.com/go-resty/resty)
* Configuration: [kelseyhightower/envconfig](https://github.com/kelseyhightower/envconfig)
* Test Framework: [stretchr/testify](https://github.com/stretchr/testify)
* SQL Database ORM: [jinzhu/gorm](https://github.com/jinzhu/gorm)
* Redis Client: [go-redis/redis](https://github.com/go-redis/redis)
* Logging: [rs/zerolog](https://github.com/rs/zerolog)
* Prometheus: [prometheus/client_golang](https://github.com/prometheus/client_golang)
* Go HTTP Metrics: [slok/go-http-metrics](https://github.com/slok/go-http-metrics)
* Validator: [go-playground/validator](github.com/go-playground/validator)

## Docker

Use [docker-makefile](https://github.com/mvanholsteijn/docker-makefile) to build docker image with semantic versioning as its tag.

The Makefile has the following targets:

```shell
make patch-release    increments the patch release level, build and push to registry
make minor-release    increments the minor release level, build and push to registry
make major-release    increments the major release level, build and push to registry
make release          build the current release and push the image to the registry
make build            builds a new version of your Docker image and tags it
make snapshot         build from the current (dirty) workspace and pushes the image to the registry
make check-status     will check whether there are outstanding changes
make check-release    will check whether the current directory matches the tagged release in git
make showver          will show the current release tag based on the directory content
```
