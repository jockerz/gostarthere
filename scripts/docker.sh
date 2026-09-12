#!/bin/bash

function show_usage() {
    echo "Usage: $0 <mode>"
    echo " mode:"
    echo "  - dev :     run docker in dev mode (with mailhog)"
    echo "  - dev_down: stop docker in dev mode"
    echo "  - prod:     run docker in prod mode (without mailhog)"
    echo "  - prod_down: stop docker in prod mode"
}

if [ $# != 1 ]
then
    show_usage
    exit 1
fi


case "$1" in
    "dev")
        docker compose -f docker-compose.base.yml \
            -f docker-compose.dev.yml up
    ;;
    "dev_down")
        docker compose -f docker-compose.base.yml \
            -f docker-compose.dev.yml down
    ;;
    "prod")
        docker compose -f docker-compose.base.yml \
            -f docker-compose.prod.yml up
    ;;
    "prod_down")
        docker compose -f docker-compose.base.yml \
            -f docker-compose.prod.yml down
    ;;
    *)
        echo "Error: Invalid mode"
        show_usage
    ;;
esac
