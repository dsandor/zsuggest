# zsuggest: history suggestions ranked by frequency, directory and recency.
# Ghost text comes from zsh-autosuggestions using the zsuggest strategy;
# Ctrl-Space opens a ranked list (fzf) of matches for the current directory.

(( $+commands[zsuggest] )) || return

if (( ! $+functions[_zsh_autosuggest_start] )); then
  for f in \
    ${ZSUGGEST_AUTOSUGGEST_PATH:-} \
    /opt/homebrew/share/zsh-autosuggestions/zsh-autosuggestions.zsh \
    /usr/local/share/zsh-autosuggestions/zsh-autosuggestions.zsh \
    /usr/share/zsh-autosuggestions/zsh-autosuggestions.zsh; do
    [[ -r $f ]] && { source $f; break; }
  done
  unset f
fi
(( $+functions[_zsh_autosuggest_start] )) || return

_zsh_autosuggest_strategy_zsuggest() {
  typeset -g suggestion="$(command zsuggest top --dir "$PWD" -- "$1" 2>/dev/null)"
}

ZSH_AUTOSUGGEST_STRATEGY=(zsuggest completion)
ZSH_AUTOSUGGEST_USE_ASYNC=1
ZSH_AUTOSUGGEST_BUFFER_MAX_SIZE=200
ZSH_AUTOSUGGEST_HIGHLIGHT_STYLE='fg=244'

zsuggest-menu() {
  local selected
  if (( $+commands[fzf] )); then
    selected="$(command zsuggest list --dir "$PWD" -n 50 -- "$BUFFER" |
      fzf --read0 --print0 --height=40% --reverse --no-sort --tiebreak=index \
          --delimiter=$'\t' --nth=2.. --prompt="${BUFFER:-❯ }" \
          --header="★ = used in this directory" | tr -d '\0')"
  fi
  if [[ -n $selected ]]; then
    BUFFER="${selected#*$'\t'}"
    CURSOR=$#BUFFER
  fi
  zle reset-prompt
}
zle -N zsuggest-menu
bindkey '^@' zsuggest-menu      # Ctrl-Space
