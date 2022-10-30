function __fish_%[1]s
  SHELL=(which fish) %[1]s _autocomplete query \
    (commandline --current-process --cut-at-cursor --tokenize) \
    (commandline --current-process --current-token) \
    2>/dev/null
end

complete --command %[1]s --arguments '(__fish_%[1]s)'
