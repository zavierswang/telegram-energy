.PHONY: all build clean run check cover lint docker help

dateTime=`date +%F_%T`
ARCH="linux-amd64"

all: build

build:
	mkdir -p build
	xgo -targets=linux/amd64 -ldflags="-w -s" -out=build/telegram-energy -pkg=cmd/telegram-energy/main.go .
	tar czvf build/telegram-energy_${dateTime}.tar.gz \
		build/telegram-energy-${ARCH} \
		template \
		assets \
		restart.sh
