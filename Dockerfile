From archlinux/base

ENV LANG en_US.UTF-8

WORKDIR /usr/src/local-news

# Assumes the build context directory is the git repository root
COPY . .

RUN pacman -Syu --noconfirm base-devel git gettext go
RUN echo "de_DE.UTF-8 UTF-8" >> /etc/locale.gen && \
    echo "es_ES.UTF-8 UTF-8" >> /etc/locale.gen && \
    echo "fr_FR.UTF-8 UTF-8" >> /etc/locale.gen && \
    echo "eo UTF-8" >> /etc/locale.gen && \
    locale-gen
RUN make

CMD ["./bin/localnews"]
