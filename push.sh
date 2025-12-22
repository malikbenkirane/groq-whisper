#!/bin/sh

# workaround for ./groq-whisper.exe dev build --push
VERSION=$(./groq-whisper.exe version)
echo pushing $VERSION
./groq-whisper.exe dev build && \
				gcloud storage cp ./groq-$VERSION gs://groq-whisper && \
				gcloud storage cp ./setup/groq-setup-$VERSION gs://groq-whisper
