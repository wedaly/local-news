GO_SRC := $(shell find . -name "*.go")
GETTEXT_PO := $(shell find . -name "*.po")
GETTEXT_MO := $(patsubst %.po,%.mo,$(GETTEXT_PO))

.PHONY: all clean fmt test

all: $(GETTEXT_MO) localnews

localnews: $(GO_SRC)
	go build -o bin/localnews cmd/main.go

fmt:
	go fmt ./...

test:
	go test ./...

messages.pot: $(GO_SRC)
	xgettext $^ \
    --language=C \
    --output=messages.pot \
    --from-code=UTF-8 \
	--add-comments=translators \
    --keyword=Gettext \
    --keyword=NGettext:1,2

$(GETTEXT_PO): messages.pot
	msgmerge -U $@ $^ && touch $@

%.mo: %.po
	msgfmt $^ -o $@ && touch $@

coverage:
	go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out

clean:
	rm -rf bin/* messages.pot
