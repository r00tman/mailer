# mailer
Fast and lean vi-style terminal email reader

## Getting Started

Installing/updating:
```shell
$ go get -v -u github.com/r00tman/mailer
```

Now you have `mailer` in `$GOPATH/bin` (which is usually `~/go/bin`).

If you haven't already, you can add `$GOPATH/bin` to your `$PATH` by appending this to `.bashrc`/`.zshrc`/etc.:
```bash
export PATH="${GOPATH:-$HOME/go}/bin:$PATH"
```

You may want to install `w3m`, since it is required to read html emails.

Now you can just run the app:
```shell
$ mailer
```

BTW, feel free to alias/symlink `mailer` to `ml` (or whatever you like), since it is much quicker to type.

## Usage
Most keybindings are vi/w3m style.

This means that:
 - `jk`/`Down Up` are for down, up (numbers work too, i.e., `10j` scrolls 10 lines down),
 - `hl`/`Esc Enter`/`Left Right` are for closing, opening emails,
 - `q`/`:q` are for quitting,
 - `PgUp PgDown`/`Ctr-U Ctr-D`/`{ }`/`b Space` are for page scrolling,
 - `gg G` are for getting to the start and to the end,
 - `Ctrl-L`/`;` aligns view to the selection (cycles between modes when pressed repeatedly),
 - `/?` are for forward, backward search,
 - `nN` are for next, previous search match.

Tip: You can scroll to the text/html part by typing `/html<Enter><Ctrl-L>`.
