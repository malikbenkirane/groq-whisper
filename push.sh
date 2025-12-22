#!/usr/bin/env bash

# workaround for ./groq-whisper.exe dev build --push
git pull
go build
VERSION=$(./groq-whisper.exe version)
echo pushing $VERSION
./groq-whisper.exe dev build && \
				gcloud storage cp groq-$VERSION.exe gs://groq-whisper && \
				gcloud storage cp setup/groq-setup-$VERSION.exe gs://groq-whisper
