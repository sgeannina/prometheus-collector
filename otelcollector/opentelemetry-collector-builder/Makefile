all: otelcollector fluentbitplugin promconfigvalidator

.PHONY: otelcollector
otelcollector:
	rm -rf otelcollector
	go get
	go build -buildmode=pie -ldflags '-linkmode external -extldflags=-Wl' -o otelcollector .

fluentbitplugin:
	make -C ../fluent-bit/src

promconfigvalidator:
	make -C ../prom-config-validator-builder