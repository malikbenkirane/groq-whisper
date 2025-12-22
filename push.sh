#!/usr/bin/env bash

# workaround for ./groq-whisper.exe dev build --push
git pull
go build
VERSION=$(./groq-whisper.exe version)
echo pushing $VERSION
./groq-whisper.exe dev build && \
	./groq-whisper.exe dev zip && \
	gcloud storage cp groq-setup-$VERSION.zip gs://groq-whisper
