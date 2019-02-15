# Name of this service/application
SERVICE_NAME := resilience-demo

# Shell to use for running scripts
SHELL := $(shell which bash)

# Get docker path or an empty string
DOCKER := $(shell command -v docker)

# Get docker-compose path or an empty string
DOCKER_COMPOSE := $(shell command -v docker-compose)

# Get the main unix group for the user running make (to be used by docker-compose later)
GID := $(shell id -g)

# Get the unix user id for the user running make (to be used by docker-compose later)
UID := $(shell id -u)


# Dev direcotry has all the required dev files.
DEV_DIR := ./docker/dev

# cmds
DEPS_CMD := GO111MODULE=on go mod tidy && GO111MODULE=on go mod vendor

# environment dirs
DEV_DIR := docker/dev

BUILD_CMD := go build -o ./bin/server ./cmd

EXP_CMD := docker run --rm -it -v ${PWD}:/src --network=host --memory="50m" --cpus="0.1"  golang:1.11 /src/bin/server --experiment 

# The default action of this Makefile is to build the development docker image
.PHONY: default
default: build

# Test if the dependencies we need to run this Makefile are installed
.PHONY: deps-development
deps-development:
ifndef DOCKER
	@echo "Docker is not available. Please install docker"
	@exit 1
endif
ifndef DOCKER_COMPOSE
	@echo "docker-compose is not available. Please install docker-compose"
	@exit 1
endif


# run the development stack.
.PHONY: deps-development
stack: deps-development
	cd $(DEV_DIR) && \
    ( docker-compose -p $(SERVICE_NAME) up --build; \
      docker-compose -p $(SERVICE_NAME) stop; \
      docker-compose -p $(SERVICE_NAME) rm -f; )

.PHONY: deps
deps:
	$(DEPS_CMD)

.PHONY: dev
dev:
	$(DEV_CMD)

.PHONY: build
build:
	$(BUILD_CMD)

.PHONY: exp1
exp1:
	$(EXP_CMD) 1

.PHONY: exp2
exp2:
	$(EXP_CMD) 2

.PHONY: exp3
exp3:
	$(EXP_CMD) 3

.PHONY: exp4
exp4:
	$(EXP_CMD) 4

.PHONY: exp5
exp5:
	$(EXP_CMD) 5