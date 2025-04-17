#! /usr/bin/env bash

FILE=$1
if [ ! -f $FILE ]; then
    echo "+++ no component file provided"
    exit 1
fi

mkdir -p tmp
echo generating configs for component $FILE

# generate the configs from the provided file
go run ./cmd/hpsf -i $FILE -o tmp/refinery-rules.yaml rRules || exit 1
go run ./cmd/hpsf -i $FILE -o tmp/refinery-config.yaml rConfig || exit 1
go run ./cmd/hpsf -i $FILE -o tmp/collector-config.yaml cConfig || exit 1

# run refinery with the generated configs
cid=$(docker run -d --rm --name smoke-refinery \
    -v ./tmp/refinery-config.yaml:/etc/refinery/refinery.yaml \
    -v ./tmp/refinery-rules.yaml:/etc/refinery/rules.yaml \
    honeycombio/refinery:latest || exit 1)
sleep 1

# check if refinery container is running
if [ "$(docker inspect -f '{{.State.Running}}' $cid 2>/dev/null)" != "true" ]; then
    echo "+++ failed to start refinery with generated configs for component $FILE"
    exit 1
else
    echo "+++ started refinery successfully with generated configs for component $FILE"
    docker kill $cid > /dev/null
fi

# use yq to remove the usage processor and honeycomb extension from collector config
yq -i e \
    'del(.processors.usage) | del(.extensions) | del(.service.extensions) | del(.service.pipelines.traces.processors[] | select(. == "usage"))' \
    tmp/collector-config.yaml

# run collector with the generated config
cid=$(docker run -d --rm --name smoke-collector \
    --entrypoint /otelcol-contrib \
    -v ./tmp/collector-config.yaml:/etc/otelcol-contrib/config.yaml \
    honeycombio/supervised-collector:latest \
    --config /etc/otelcol-contrib/config.yaml)
sleep 1

# check if collector container is running
if [ "$(docker inspect -f '{{.State.Running}}' $cid 2>/dev/null)" != "true" ]; then
    echo "+++ failed to start collector with generated configs for component $FILE"
    exit 1
else
    echo "+++ started collector successfully with generated configs for component $FILE"
    docker kill $cid > /dev/null
fi
