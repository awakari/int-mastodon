#!/bin/bash

export SLUG=ghcr.io/awakari/int-mastodon
export VERSION=latest
docker tag awakari/int-mastodon "${SLUG}":"${VERSION}"
docker push "${SLUG}":"${VERSION}"
