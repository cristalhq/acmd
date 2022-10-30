#compdef %[1]s
_%[1]s() {
  completions="$(SHELL=$(which zsh) %[1]s __complete query "$words[@]" 2>/dev/null)"

  if [[ "$completions" == "()" ]]; then
    # Default to file-based completion
    _files
  else
    # Otherwise complete flags and subcommands.
    _describe command "$completions"
  fi
}

_%[1]s "$@"
