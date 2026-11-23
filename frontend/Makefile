# Frontend build. Run from the frontend directory. No Node required.
#
#   make setup   download the Tailwind CLI and daisyUI plugin (once)
#   make css     build static/css/app.css (minified)
#   make watch   rebuild CSS on every template change
#   make icons   rebuild static/icons.svg from Lucide
#   make vendor  re-download htmx, Alpine and the Geist font

TAILWIND_VERSION := v4.3.5
DAISYUI_VERSION  := v5.7.50
HTMX_VERSION     := 2.0.11
ALPINE_VERSION   := 3.17.5
GEIST_VERSION    := 1.9.2

UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)
ifeq ($(UNAME_S),Darwin)
  TW_OS := macos
else
  TW_OS := linux
endif
ifneq (,$(filter $(UNAME_M),arm64 aarch64))
  TW_ARCH := arm64
else
  TW_ARCH := x64
endif

TAILWIND := bin/tailwindcss
DAISYUI  := styles/vendor/daisyui.mjs styles/vendor/daisyui-theme.mjs

.PHONY: setup css watch icons vendor clean

setup: $(TAILWIND) $(DAISYUI)

$(TAILWIND):
	mkdir -p bin
	curl -sSLo $@ https://github.com/tailwindlabs/tailwindcss/releases/download/$(TAILWIND_VERSION)/tailwindcss-$(TW_OS)-$(TW_ARCH)
	chmod +x $@

styles/vendor/%.mjs:
	mkdir -p styles/vendor
	curl -sSLo $@ https://github.com/saadeghi/daisyui/releases/download/$(DAISYUI_VERSION)/$*.mjs

css: setup
	$(TAILWIND) -i styles/app.css -o static/css/app.css --minify

watch: setup
	$(TAILWIND) -i styles/app.css -o static/css/app.css --watch

icons:
	python3 scripts/build-icons.py

vendor:
	curl -sSLo static/js/htmx.min.js https://cdn.jsdelivr.net/npm/htmx.org@$(HTMX_VERSION)/dist/htmx.min.js
	curl -sSLo static/js/alpine.min.js https://cdn.jsdelivr.net/npm/alpinejs@$(ALPINE_VERSION)/dist/cdn.min.js
	curl -sSLo static/fonts/Geist-Variable.woff2 https://cdn.jsdelivr.net/npm/geist@$(GEIST_VERSION)/dist/fonts/geist-sans/Geist-Variable.woff2

clean:
	rm -rf bin styles/vendor
