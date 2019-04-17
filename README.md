# local-news
Terminal-based RSS/Atom feed reader

# Getting Started

## Docker

1. [Install Docker](https://docs.docker.com/v17.12/install/)
2. Set your current working directory to the root of this git repository: `cd /path/to/local-news`
2. Build the docker image: `docker build . -t localnews`
3. Run the program in Docker: `docker run -it localnews`
4. Run the program using a pseudo-locale: `docker run -it -e LANG=eo localnews`

NOTES:
* Copy-and-paste won't work when using Docker because the container won't have access to the host system's clipboard.
* Opening a feed item in a browser also won't work in Docker, because no browser is installed in the image.
* The Docker image includes locales for `de_DE.UTF-8`, `es_ES.UTF-8`, and `fr_FR.UTF-8` as well.  Choosing these locales will affect datetime formatting, number formatting, and collation (sort) order.  However, the UI strings have not (yet) been translated for these locales, so the text will appear in English.

## Linux

If you are on Linux, you will need to install:

* [GNU gettext](https://www.gnu.org/software/gettext)
* [Go](http://golang.org/)
* [xdg-utils](https://freedesktop.org/wiki/Software/xdg-utils/)

To build the program: `make`

To run the program after it's built: `./bin/localnews`

To run tests: `make tests`

# Localization

* Translation files are in `configs/locale/{locale}/LC_MESSAGES`
* `make` will automatically update the ".po" and ".mo" files
* Locale-specific config (color schemes) is in `configs/etc/{locale}/config.xml`
* We're using Esperanto (`eo`) as a pseudo-language to test internationalization.  Set the environment variable `LANG=eo` to see the UI text and colors change.

# Known Issues

Pasting directly to the terminal (e.g. middle-click in X-Windows) will sometimes truncate the pasted text.  This is due to an [bug in the underlying TUI library](https://github.com/gdamore/tcell/issues/200).  As a workaround, you can use "Ctrl-v" to paste directly.  Note that this requires the terminal to send the ctrl-v command to the application, which some terminals don't support.
