_%[1]s() {
	# Read autocompletion into COMPREPLY
	IFS=$'\n' COMPREPLY=( $(SHELL=$(which bash) %[1]s _autocomplete query "${COMP_WORDS[@]}" 2>/dev/null) )

	if [[ ${#COMPREPLY[@]} -eq 0 ]]; then
		local cur=${COMP_WORDS[COMP_CWORD]}
		COMPREPLY=( $( compgen -o plusdirs -f -- $cur ) )
	fi
}

complete -F _%[1]s %[1]s
