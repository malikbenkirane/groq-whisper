#!/usr/bin/env bash

export PATH=/mingw64/bin:$PATH
export PKG_CONFIG_PATH=$PKG_CONFIG_PATH:/mingw64/lib/pkgconfig
export PATH=/mingw64/lib/go/bin:$PATH
export PATH=$PATH:~/google-cloud-sdk/bin

# workaround for ./groq-whisper.exe dev build --push
git pull
go build
VERSION=$(./groq-whisper.exe version)
echo pushing $VERSION
./groq-whisper.exe dev build && \
	./groq-whisper.exe dev zip && \
	gcloud storage cp groq-setup-$VERSION.zip gs://groq-whisper && \
	gcloud storage cp setup/groq-setup-$VERSION.exe gs://groq-whisper && \
	gcloud storage cp groq-$VERSION.exe gs://groq-whisper
