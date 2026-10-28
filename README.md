# zsuggest

Context-aware command suggestions for zsh. As you type, zsuggest shows the command you most likely want, ranked by **how often you run it**, **which directory you are in**, and **how recently you used it**.

```
~/Projects/api $ sam b█uild --use-container        ← grey suggestion, press → to accept

[Ctrl-Space]
  ★  32× here   sam build --use-container
  ★  12× here   sam build
  ★   9× here   sam deploy --guided
     41× all    sam build && sam deploy
```

Most autosuggestion tools pick the most recent match from a global history. zsuggest ranks by directory instead, so in each project the suggestions are the commands you actually run there. For example, `npm run dev` in one repo and `sam build` in another.

## How it works

zsuggest has two parts:

1. **`zsuggest`**, a small Go binary that reads a directory-tagged history log and ranks matching commands.
2. **`zsuggest.zsh`**, a zsh plugin that adds two things:
   - grey "ghost text" suggestions, using [zsh-autosuggestions](https://github.com/zsh-users/zsh-autosuggestions) with zsuggest as the strategy
   - a ranked picker on **Ctrl-Space**, using [fzf](https://github.com/junegunn/fzf)

### Ranking

Every past use of a matching command adds to its score:

| Where it was run                          | Weight |
| ----------------------------------------- | ------ |
| Current directory                         | 1.0    |
| Parent or child of the current directory  | 0.35   |
| Anywhere else                             | 0.08   |

Each use is also scaled by recency. A use's weight halves every 2,500 commands but never drops below 25%. Old habits still count, and recent work counts more. Ties go to the most recently used command.

In the picker, `★` marks commands you have run in the current directory, with a count of uses there. Other entries show their total count.

Lookups take a few milliseconds and run asynchronously, so typing never blocks.

## Requirements

- zsh
- Go 1.24+ (to build)
- [zsh-autosuggestions](https://github.com/zsh-users/zsh-autosuggestions)
- [fzf](https://github.com/junegunn/fzf) (optional, for the Ctrl-Space picker)

On macOS:

```sh
brew install zsh-autosuggestions fzf
```

## Installation

1. Install the binary somewhere on your `PATH`:

   ```sh
   go install github.com/dsandor/zsuggest@latest
   ```

   Or build it from a clone:

   ```sh
   git clone https://github.com/dsandor/zsuggest.git ~/Projects/zsuggest
   cd ~/Projects/zsuggest
   go build -o ~/bin/zsuggest .
   ```

2. Add the history logger and the plugin to `~/.zshrc`. The logger records each command with the directory it ran in. That is the data zsuggest ranks from.

   ```zsh
   # Log every command with its working directory
   function zshaddhistory() {
     echo "${1%%$'\n'}⋮${PWD}   " >> ~/.zsh_history_ext
   }

   # Load zsuggest (after your prompt setup, e.g. starship/oh-my-zsh)
   source ~/Projects/zsuggest/zsuggest.zsh
   ```

   If you use `go install`, download `zsuggest.zsh` from this repo and source it from wherever you saved it.

3. Open a new terminal. Suggestions start appearing once the log has some history, and improve the more you use the shell.

## Usage

| Key                  | Action                                              |
| -------------------- | --------------------------------------------------- |
| type                 | Best match appears as grey text                     |
| `→` or `End`         | Accept the whole suggestion                         |
| `Alt-f`              | Accept the next word of the suggestion              |
| `Ctrl-Space`         | Open the ranked picker for the current input        |
| `Ctrl-Space` (empty) | Show the top commands for the current directory     |

Inside the picker you can type to filter. `Enter` puts the selected command on the command line without running it.

### Command line

The binary can also be used on its own:

```sh
zsuggest top  [--dir DIR] [--file FILE] -- PREFIX          # best match
zsuggest list [--dir DIR] [--file FILE] [-n N] -- PREFIX   # ranked list (NUL-separated)
```

## Configuration

| Setting                          | Default               | Purpose                                     |
| -------------------------------- | --------------------- | ------------------------------------------- |
| `ZSUGGEST_HISTFILE`              | `~/.zsh_history_ext`  | History log to read                         |
| `ZSUGGEST_AUTOSUGGEST_PATH`      | auto-detected         | Path to `zsh-autosuggestions.zsh`           |
| `ZSH_AUTOSUGGEST_HIGHLIGHT_STYLE`| `fg=244`              | Colour of the ghost text                    |

To use a different key for the picker, rebind it after sourcing the plugin:

```zsh
bindkey '^G' zsuggest-menu   # Ctrl-G
```

The ranking weights are constants at the top of `main.go`. Adjust them and rebuild if you prefer frequency or recency to count more.

## History log format

Each record is a single line, `<command>⋮<directory>` followed by three spaces. A multi-line command spans several lines, and only its last line carries the `⋮<directory>` marker. Blank commands are ignored.

## Development

```sh
go test ./...
go build -o ~/bin/zsuggest .
```

## License

[MIT](LICENSE)
