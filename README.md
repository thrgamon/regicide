# Regicide

Regicide is a simple commandline regex tester, similar to Regex 101[^1].

`$ regicide`

[![asciicast](https://asciinema.org/a/3gGtIoEzwRVDw4wjuxeM2pSpq.svg)](https://asciinema.org/a/3gGtIoEzwRVDw4wjuxeM2pSpq)

## Installation

At the moment, the only option is to download the repo and install the thing yourself. Sorry about that.

```bash
$ git clone https://github.com/thrgamon/regicide 
$ cd regicide
$ go install
```

## Usage

```bash
$ regicide # This will start the CLI tool with a blank regex and test cases
$ regicide --cases "$(cat somefile.txt)" # You can pass cases in with the cases flag
$ regicide --debug # The debug flag will create a log file in /tmp/regicide-debug.log - probably only useful for me right now
```

When the CLI starts there are some keybindings that are necessary to know.

```
tab - switches between the regex and the results input
ctrl+c - quits the programme

ctrl+l - enables the multiline regex flag
ctrl+n - enables case insensitive regex flag
ctrl+u - enables the ungreedy regex flag
ctrl+s - enables the "dot matches newline" regex flag
```

The regex flags follow the ones defined [here](https://pkg.go.dev/regexp/syntax).
```
i              case-insensitive (default false)
m              multi-line mode: ^ and $ match begin/end line in addition to begin/end text (default false)
s              let . match \n (default false)
U              ungreedy: swap meaning of x* and x*?, x+ and x+?, etc (default false)
```

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

[^1]: Other regex testers are available.

## License
[MIT](https://choosealicense.com/licenses/mit/)
