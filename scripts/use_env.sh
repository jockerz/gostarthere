#!/bin/bash

#!/bin/bash

function show_usage() {
    echo "Usage: $0 <local,docker>"
    echo "  docker: copy {backend,frontend}/.env.docker -> {backend,frontend}/.env"
    echo "  local : copy {backend,frontend}/.env.local -> {backend,frontend}/.env"
}

if [ $# != 1 ]
then
    show_usage
    exit 1
fi


case "$1" in
    "docker")
        cp backend/.env.docker  backend/.env
        cp frontend/.env.docker frontend/.env
    ;;
    "local")
        cp backend/.env.local  backend/.env
        cp frontend/.env.local frontend/.env
    ;;
    *)
        echo "Invalid argument"
        show_usage
    ;;
esac
